package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// ---------------- 配置 ----------------
// 地址优先读环境变量（容器内部署用服务名），缺省走本机端口（本地 go run）。

const (
	password = "hichat2024"
	smsCode  = "888888"

	chatSingle = 1
	chatGroup  = 2
	mTypeText  = 1
)

var (
	userAPI   = env("USER_API", "http://localhost:8887")
	socialAPI = env("SOCIAL_API", "http://localhost:8889")
	imAPI     = env("IM_API", "http://localhost:8890")
	trendAPI  = env("TREND_API", "http://localhost:8891")
	wsURL     = env("WS_URL", "ws://localhost:10090/ws")
	imgDir    = env("IMG_DIR", "docs/imgs")

	// 种注册验证码的两种方式：
	//   REDIS_ADDR 非空    -> 直连 redis（容器内，如 redis:6379），免依赖 RESP SET
	//   否则               -> docker exec <REDIS_CONTAINER> redis-cli（本机 compose）
	redisAddr      = os.Getenv("REDIS_ADDR")
	redisContainer = env("REDIS_CONTAINER", "go-hichat-api-redis-1")
)

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

var httpClient = &http.Client{Timeout: 20 * time.Second}

// 已上传到静态服务器的图片 URL 池（头像 + 封面），供图文动态复用。
var imagePool []string

func main() {
	trendsOnly := flag.Bool("trends-only", false, "只重灌动态/评论/点赞（复用已注册用户，不重建好友/群/聊天）")
	flag.Parse()

	fmt.Println("==> HiChat mock 数据生成开始")

	avatars := mustListAvatars()
	fmt.Printf("发现 %d 张图片，%d 个人设\n", len(avatars), len(personas))

	// 1. 注册 / 登录 + 头像 + 资料
	step("注册用户、上传头像、补全资料")
	for i, p := range personas {
		registerOrLogin(p)
		if *trendsOnly {
			// 复用已有头像 URL，避免重复上传
			p.AvatarURL = fetchAvatar(p)
			if p.AvatarURL != "" {
				imagePool = append(imagePool, p.AvatarURL)
			}
			fmt.Printf("  ✓ %s (uid=%s)\n", p.Nickname, p.Uid)
			continue
		}
		avatarFile := avatars[i%len(avatars)]
		url, err := uploadAvatar(p, filepath.Join(imgDir, avatarFile))
		if err != nil {
			warn("上传头像失败 %s: %v", p.Nickname, err)
		} else {
			p.AvatarURL = url
			imagePool = append(imagePool, url)
		}
		updateProfile(p)
		fmt.Printf("  ✓ %s (uid=%s) %s\n", p.Nickname, p.Uid, p.AvatarURL)
	}

	if !*trendsOnly {
		// 2. 好友关系
		step("建立好友关系 + 待处理申请")
		buildFriends()

		// 3. 群
		step("创建群、邀请成员、群公告、入群申请")
		buildGroups()

		// 4. 聊天消息（WebSocket）
		step("灌入单聊 / 群聊消息")
		buildChats()
	}

	// 5. 动态 / 评论 / 点赞
	step("发布动态、评论、点赞")
	buildTrends()

	fmt.Println("\n==> 全部完成！等待 MQ 落库 (3s)...")
	time.Sleep(3 * time.Second)
	fmt.Println("\n========== 主角登录账号 ==========")
	fmt.Printf("  手机号: %s\n  密码:   %s\n  昵称:   %s\n", personas[0].Phone, password, personas[0].Nickname)
	fmt.Println("==================================")
}

// fetchAvatar 读取用户当前头像 URL（trends-only 模式下复用）。
func fetchAvatar(p *persona) string {
	data, err := doJSON(http.MethodGet, userAPI+"/api/v1/user/detail", p.Token, nil)
	if err != nil {
		return ""
	}
	var r struct {
		Info struct {
			Avatar string `json:"avatar"`
		} `json:"info"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return ""
	}
	return r.Info.Avatar
}

// ---------------- 用户：注册 / 头像 / 资料 ----------------

func registerOrLogin(p *persona) {
	// 先种 Redis 验证码，再注册；已注册则回退登录。
	seedSMSCode(p.Phone)
	body := map[string]any{"phone": p.Phone, "password": password, "nickname": p.Nickname, "phoneCode": smsCode}
	data, err := doJSON(http.MethodPost, userAPI+"/api/v1/user/register", "", body)
	if err != nil {
		// 已注册：登录
		ld, lerr := doJSON(http.MethodPost, userAPI+"/api/v1/user/login", "", map[string]any{"phone": p.Phone, "password": password})
		if lerr != nil {
			fatal("注册和登录都失败 %s: register=%v login=%v", p.Phone, err, lerr)
		}
		data = ld
	}
	var tok struct {
		Token string `json:"token"`
	}
	must(json.Unmarshal(data, &tok))
	p.Token = tok.Token
	p.Uid = uidFromToken(tok.Token)
}

func uploadAvatar(p *persona, path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(fw, f); err != nil {
		return "", err
	}
	w.Close()

	req, _ := http.NewRequest(http.MethodPost, userAPI+"/api/v1/user/avatar/upload", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+p.Token)
	data, err := doRaw(req)
	if err != nil {
		return "", err
	}
	var r struct {
		Url string `json:"url"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return "", err
	}
	return r.Url, nil
}

func updateProfile(p *persona) {
	tags, _ := json.Marshal(p.Tags)
	cover := ""
	if len(imagePool) > 0 {
		cover = imagePool[len(imagePool)-1]
	}
	body := map[string]any{
		"name":          p.Nickname,
		"avatar":        p.AvatarURL,
		"sex":           p.Sex,
		"introduction":  p.Intro,
		"region":        p.Region,
		"occupation":    p.Occupation,
		"tags":          string(tags),
		"moments_cover": cover,
	}
	if _, err := doJSON(http.MethodPut, userAPI+"/api/v1/user/update", p.Token, body); err != nil {
		warn("更新资料失败 %s: %v", p.Nickname, err)
	}
}

// ---------------- 好友 ----------------

func buildFriends() {
	// 主角与 1..10 互为好友
	for i := 1; i <= 10; i++ {
		addFriend(0, i, friendReqMsg(0, i))
	}
	// 交叉好友
	crosses := [][2]int{{1, 2}, {1, 8}, {2, 6}, {5, 6}, {5, 13}, {3, 11}, {4, 9}, {7, 10}, {9, 2}}
	for _, c := range crosses {
		addFriend(c[0], c[1], friendReqMsg(c[0], c[1]))
	}
	// 收到的待处理申请（不通过）：11、12 -> 主角
	applyFriend(11, 0, "你好，我是冯哲，看了你的产品分享，想认识一下～")
	applyFriend(12, 0, "晚晴你好，我是瑜伽老师许念，朋友推荐认识的 🧘")
	// 发出的待处理申请（不通过）：主角 -> 13
	applyFriend(0, 13, "杨帆你好，很喜欢你的旅行动态，交个朋友！")
	fmt.Println("  ✓ 好友关系与待处理申请就绪")
}

// addFriend 让 a 申请加 b，并由 b 通过。
func addFriend(a, b int, msg string) {
	applyFriend(a, b, msg)
	// b 拉取收到的待处理申请，找到来自 a 的那条并通过
	reqID := findPendingReq(b, personas[a].Uid)
	if reqID == 0 {
		warn("未找到 %s->%s 的好友申请", personas[a].Nickname, personas[b].Nickname)
		return
	}
	body := map[string]any{
		"friend_req_id": reqID,
		"handle_result": 1, // 同意
		"remark":        "",
	}
	if _, err := doJSON(http.MethodPut, socialAPI+"/v1/social/friend/putIn", personas[b].Token, body); err != nil {
		warn("通过好友申请失败 %s<-%s: %v", personas[b].Nickname, personas[a].Nickname, err)
	}
}

func applyFriend(a, b int, msg string) {
	body := map[string]any{"user_uid": personas[b].Uid, "req_msg": msg, "req_time": time.Now().Unix()}
	if _, err := doJSON(http.MethodPost, socialAPI+"/v1/social/friend/putIn", personas[a].Token, body); err != nil {
		warn("好友申请失败 %s->%s: %v", personas[a].Nickname, personas[b].Nickname, err)
	}
}

// findPendingReq 在 b 收到的待处理申请里找 reqUid 发来的申请，返回 friend_req_id。
func findPendingReq(b int, reqUid string) int {
	url := fmt.Sprintf("%s/v1/social/friend/putIns?class=1&type=0", socialAPI)
	data, err := doJSON(http.MethodGet, url, personas[b].Token, nil)
	if err != nil {
		warn("拉取好友申请列表失败 %s: %v", personas[b].Nickname, err)
		return 0
	}
	var r struct {
		List []struct {
			Id     int64  `json:"id"`
			UserId string `json:"user_id"` // 申请人（在收到方列表里即对方）
		} `json:"list"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return 0
	}
	for _, it := range r.List {
		if it.UserId == reqUid {
			return int(it.Id)
		}
	}
	return 0
}

func friendReqMsg(a, b int) string {
	return fmt.Sprintf("我是%s，很高兴认识你！", personas[a].Nickname)
}

// ---------------- 群 ----------------

func buildGroups() {
	for gi := range groups {
		g := &groups[gi]
		owner := personas[g.Owner]
		icon := g.Icon
		if icon == "" {
			icon = owner.AvatarURL
		}
		// 创建群
		data, err := doJSON(http.MethodPost, socialAPI+"/v1/social/group", owner.Token,
			map[string]any{"name": g.Name, "icon": icon})
		if err != nil {
			warn("创建群失败 %s: %v", g.Name, err)
			continue
		}
		var r struct {
			GroupId string `json:"group_id"`
		}
		must(json.Unmarshal(data, &r))
		gid := r.GroupId
		g.gid = gid

		// 邀请成员（群主邀请 => 直接入群）
		var ids []string
		for _, m := range g.Invite {
			ids = append(ids, personas[m].Uid)
		}
		if _, err := doJSON(http.MethodPost, socialAPI+"/v1/social/group/invite", owner.Token,
			map[string]any{"group_id": gid, "friend_ids": ids}); err != nil {
			warn("邀请入群失败 %s: %v", g.Name, err)
		}

		// 群公告
		if g.Announcement != "" {
			if _, err := doJSON(http.MethodPost, socialAPI+"/v1/social/group/announcement", owner.Token,
				map[string]any{"group_id": gid, "content": g.Announcement}); err != nil {
				warn("发布群公告失败 %s: %v", g.Name, err)
			}
		}

		// 为每个成员建立群会话（让群出现在其会话列表）
		members := append([]int{g.Owner}, g.Invite...)
		for _, m := range members {
			setupConversation(personas[m].Token, personas[m].Uid, gid, chatGroup)
		}

		fmt.Printf("  ✓ 群「%s」gid=%s 成员%d人\n", g.Name, gid, len(members))
	}

	// 一条待审入群申请：群1开启验证后，由非成员 11 申请加入 -> 主角(群主)可见
	g0 := &groups[0]
	if g0.gid != "" {
		owner := personas[g0.Owner]
		if _, err := doJSON(http.MethodPost, socialAPI+"/v1/social/group/update", owner.Token,
			map[string]any{"group_id": g0.gid, "is_verify": 1}); err != nil {
			warn("开启群验证失败: %v", err)
		}
		applicant := personas[12] // 许念，非该群成员
		if _, err := doJSON(http.MethodPost, socialAPI+"/v1/social/group/putIn", applicant.Token,
			map[string]any{"group_id": g0.gid, "join_source": 1, "req_msg": "你好，我对产品很感兴趣，想加入学习～", "req_time": time.Now().Unix()}); err != nil {
			warn("提交入群申请失败: %v", err)
		} else {
			fmt.Printf("  ✓ 一条待审入群申请 -> 群「%s」\n", g0.Name)
		}
	}
}

func setupConversation(token, sendId, recvId string, chatType int) {
	body := map[string]any{"sendId": sendId, "recvId": recvId, "chatType": chatType}
	if _, err := doJSON(http.MethodPost, imAPI+"/v1/im/setup/conversation", token, body); err != nil {
		warn("建立会话失败 send=%s recv=%s: %v", sendId, recvId, err)
	}
}

// ---------------- 聊天（WebSocket，RigorAck 三次握手）----------------

func buildChats() {
	// 单聊
	for _, d := range dialogues {
		a, b := personas[d.A], personas[d.B]
		// 双向建立会话，确保两边都展示
		setupConversation(a.Token, a.Uid, b.Uid, chatSingle)
		setupConversation(b.Token, b.Uid, a.Uid, chatSingle)

		conns := map[int]*wsClient{}
		for _, idx := range []int{d.A, d.B} {
			if _, ok := conns[idx]; !ok {
				c, err := dialWS(personas[idx].Token)
				if err != nil {
					warn("连接 WS 失败 %s: %v", personas[idx].Nickname, err)
					continue
				}
				conns[idx] = c
			}
		}
		for _, ln := range d.Lines {
			c := conns[ln.From]
			if c == nil {
				continue
			}
			recv := personas[d.B].Uid
			if ln.From == d.B {
				recv = personas[d.A].Uid
			}
			c.sendChat(chatSingle, recv, ln.Text)
			time.Sleep(180 * time.Millisecond)
		}
		for _, c := range conns {
			c.close()
		}
		fmt.Printf("  ✓ 单聊 %s ⇄ %s (%d 条)\n", a.Nickname, b.Nickname, len(d.Lines))
	}

	// 群聊
	for gi := range groups {
		g := &groups[gi]
		if g.gid == "" {
			continue
		}
		conns := map[int]*wsClient{}
		for _, ln := range g.Lines {
			if _, ok := conns[ln.From]; !ok {
				c, err := dialWS(personas[ln.From].Token)
				if err != nil {
					warn("连接 WS 失败 %s: %v", personas[ln.From].Nickname, err)
					continue
				}
				conns[ln.From] = c
			}
		}
		for _, ln := range g.Lines {
			c := conns[ln.From]
			if c == nil {
				continue
			}
			c.sendChat(chatGroup, g.gid, ln.Text)
			time.Sleep(220 * time.Millisecond)
		}
		for _, c := range conns {
			c.close()
		}
		fmt.Printf("  ✓ 群聊「%s」(%d 条)\n", g.Name, len(g.Lines))
	}
}

var msgCounter int64

type wsClient struct {
	conn *websocket.Conn
}

func dialWS(token string) (*wsClient, error) {
	conn, _, err := websocket.DefaultDialer.Dial(wsURL+"?token="+token, nil)
	if err != nil {
		return nil, err
	}
	return &wsClient{conn: conn}, nil
}

func (w *wsClient) close() { _ = w.conn.Close() }

// sendChat 发送一条聊天消息并完成 RigorAck 三次握手。
func (w *wsClient) sendChat(chatType int, recvId, content string) {
	id := fmt.Sprintf("mock-%d-%d", time.Now().UnixNano(), atomic.AddInt64(&msgCounter, 1))
	msg := map[string]any{
		"id":        id,
		"frameType": 0, // FrameData
		"method":    "chat.user",
		"data": map[string]any{
			"conversationId": "",
			"chatType":       chatType,
			"recvId":         recvId,
			"sendTime":       time.Now().UnixNano(),
			"msg": map[string]any{
				"mType":       mTypeText,
				"content":     content,
				"readRecords": map[string]string{},
			},
		},
	}
	if err := w.conn.WriteJSON(msg); err != nil {
		warn("发送消息失败: %v", err)
		return
	}
	// 等待服务端 ack（ackSeq>=1），然后回 ackSeq=2 完成握手
	_ = w.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	for {
		var in map[string]any
		if err := w.conn.ReadJSON(&in); err != nil {
			break // 超时或读完，直接回 ack
		}
		if ft, _ := in["frameType"].(float64); int(ft) == 3 { // FrameAck
			if mid, _ := in["id"].(string); mid == id {
				break
			}
		}
	}
	_ = w.conn.WriteJSON(map[string]any{"id": id, "frameType": 3, "ackSeq": 2})
}

// ---------------- 动态 / 评论 / 点赞 ----------------

func buildTrends() {
	for _, t := range trends {
		author := personas[t.Author]
		ttype := 1 // TEXT
		var resources []string
		if t.WithImages > 0 {
			ttype = 2 // MIXED
			resources = pickImages(t.WithImages, t.Author)
		}
		body := map[string]any{
			"type":       ttype,
			"content":    t.Content,
			"scope":      t.Scope,
			"open_reply": true,
		}
		if len(resources) > 0 {
			body["resources"] = resources
		}
		data, err := doJSON(http.MethodPost, trendAPI+"/v1/trend", author.Token, body)
		if err != nil {
			warn("发布动态失败 %s: %v", author.Nickname, err)
			continue
		}
		var r struct {
			TrendID int `json:"trend_id"`
		}
		must(json.Unmarshal(data, &r))
		tid := r.TrendID

		// 点赞
		authorUID, _ := strconv.Atoi(author.Uid)
		for _, l := range t.Likers {
			if _, err := doJSON(http.MethodPost, trendAPI+"/v1/trend/like", personas[l].Token,
				map[string]any{"trend_id": tid, "author_id": authorUID, "like_type": 1}); err != nil {
				warn("点赞失败 t=%d by %s: %v", tid, personas[l].Nickname, err)
			}
		}
		// 评论
		for _, c := range t.Comments {
			cbody := map[string]any{"trend_id": tid, "user_id": authorUID, "content": c.Text}
			if _, err := doJSON(http.MethodPost, trendAPI+"/v1/trend/comment", personas[c.From].Token, cbody); err != nil {
				// trend-api 偶发 503（突发限流），稍等重试一次
				time.Sleep(400 * time.Millisecond)
				if _, err2 := doJSON(http.MethodPost, trendAPI+"/v1/trend/comment", personas[c.From].Token, cbody); err2 != nil {
					warn("评论失败 t=%d by %s: %v", tid, personas[c.From].Nickname, err2)
				}
			}
			time.Sleep(150 * time.Millisecond)
		}
		fmt.Printf("  ✓ 动态 #%d by %s (👍%d 💬%d)\n", tid, author.Nickname, len(t.Likers), len(t.Comments))
	}
}

// pickImages 从图片池里取 n 张，尽量从作者头像之后错开，避免每条都一样。
func pickImages(n, offset int) []string {
	if len(imagePool) == 0 {
		return nil
	}
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, imagePool[(offset+i)%len(imagePool)])
	}
	return out
}

// ---------------- 基础设施 ----------------

func mustListAvatars() []string {
	entries, err := os.ReadDir(imgDir)
	must(err)
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" && ext != ".gif" {
			continue
		}
		// go-zero 默认请求体上限 1MB，跳过过大的图片。
		if info, err := e.Info(); err == nil && info.Size() > 900*1024 {
			fmt.Printf("  (跳过过大图片 %s, %.1fMB)\n", e.Name(), float64(info.Size())/1024/1024)
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)
	if len(files) == 0 {
		fatal("%s 下没有图片", imgDir)
	}
	return files
}

// seedSMSCode 往 Redis 写入注册验证码（绕过真实短信）。
// 容器内（REDIS_ADDR 非空）直连 redis；本机走 docker exec。
func seedSMSCode(phone string) {
	key := "verify:sms:" + phone
	if redisAddr != "" {
		if err := redisSetEX(redisAddr, key, smsCode, 600); err != nil {
			warn("种验证码失败(direct %s) %s: %v", redisAddr, phone, err)
		}
		return
	}
	cmd := exec.Command("docker", "exec", redisContainer, "redis-cli", "set", key, smsCode, "EX", "600")
	if out, err := cmd.CombinedOutput(); err != nil {
		warn("种验证码失败 %s: %v (%s)", phone, err, string(out))
	}
}

// redisSetEX 用最简 RESP 协议执行 `SET key val EX ttl`，无需引入 redis 客户端依赖。
func redisSetEX(addr, key, val string, ttlSec int) error {
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	args := []string{"SET", key, val, "EX", strconv.Itoa(ttlSec)}
	var b strings.Builder
	fmt.Fprintf(&b, "*%d\r\n", len(args))
	for _, a := range args {
		fmt.Fprintf(&b, "$%d\r\n%s\r\n", len(a), a)
	}
	if _, err := conn.Write([]byte(b.String())); err != nil {
		return err
	}
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	buf := make([]byte, 64)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}
	if n == 0 || buf[0] != '+' { // 期望 +OK\r\n
		return fmt.Errorf("unexpected redis reply: %q", string(buf[:n]))
	}
	return nil
}

// doJSON 发送 JSON 请求，解析 {code,msg,data} 信封，返回 data 原文。body 为 nil 时不带请求体。
func doJSON(method, url, token string, body any) (json.RawMessage, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return doRaw(req)
}

func doRaw(req *http.Request) (json.RawMessage, error) {
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}

	// user-api 用 {code,msg,data} 信封；social/im/trend 直接返回原始对象。
	// 通过是否同时存在 code+msg+data 三个键来区分。
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err == nil {
		_, hasCode := m["code"]
		_, hasMsg := m["msg"]
		data, hasData := m["data"]
		if hasCode && hasMsg && hasData {
			var code int
			_ = json.Unmarshal(m["code"], &code)
			if code != 200 && code != 0 {
				var msg string
				_ = json.Unmarshal(m["msg"], &msg)
				return nil, fmt.Errorf("业务错误 code=%d msg=%s", code, msg)
			}
			return data, nil
		}
	}
	return raw, nil
}

// uidFromToken 解出 JWT 里的 hichat2.com 声明（即用户 id）。
func uidFromToken(token string) string {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ""
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return ""
	}
	if v, ok := claims["hichat2.com"].(string); ok {
		return v
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}

func step(s string)  { fmt.Printf("\n--- %s ---\n", s) }
func warn(f string, a ...any) { fmt.Printf("  [warn] "+f+"\n", a...) }
func fatal(f string, a ...any) {
	fmt.Printf("[FATAL] "+f+"\n", a...)
	os.Exit(1)
}
func must(err error) {
	if err != nil {
		fatal("%v", err)
	}
}
