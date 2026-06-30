package main

// persona 是一个 mock 用户的人设。Sex: 1=男 2=女。
type persona struct {
	Phone      string
	Nickname   string
	Sex        int
	Intro      string
	Region     string
	Occupation string
	Tags       []string

	// 运行时填充
	Token     string
	Uid       string
	AvatarURL string
}

// 14 个中文真实人设，索引 0 为「主角」。
var personas = []*persona{
	{Phone: "13800138000", Nickname: "林晚晴", Sex: 2, Region: "浙江·杭州", Occupation: "产品经理",
		Intro: "把复杂留给自己，把简单留给用户。", Tags: []string{"产品", "效率控", "咖啡", "马拉松"}},
	{Phone: "13800138001", Nickname: "陈默", Sex: 1, Region: "广东·深圳", Occupation: "后端工程师",
		Intro: "代码写得好，bug 跑不了。", Tags: []string{"Golang", "分布式", "机械键盘"}},
	{Phone: "13800138002", Nickname: "苏小棠", Sex: 2, Region: "上海", Occupation: "UI/UX 设计师",
		Intro: "好的设计是看不见的。", Tags: []string{"设计", "手绘", "猫", "胶片"}},
	{Phone: "13800138003", Nickname: "王梓睿", Sex: 1, Region: "北京", Occupation: "连续创业者",
		Intro: "保持饥饿，保持热爱。", Tags: []string{"创业", "AI", "健身", "读书"}},
	{Phone: "13800138004", Nickname: "周岚", Sex: 2, Region: "四川·成都", Occupation: "市场运营",
		Intro: "生活要有烟火气，也要有诗和远方。", Tags: []string{"运营", "火锅", "露营", "美妆"}},
	{Phone: "13800138005", Nickname: "李航", Sex: 1, Region: "云南·大理", Occupation: "风光摄影师",
		Intro: "用快门收藏每一束光。", Tags: []string{"摄影", "旅行", "无人机", "徒步"}},
	{Phone: "13800138006", Nickname: "赵悦", Sex: 2, Region: "广东·广州", Occupation: "自由插画师",
		Intro: "画画是我和世界对话的方式。", Tags: []string{"插画", "iPad", "二次元", "奶茶"}},
	{Phone: "13800138007", Nickname: "吴桐", Sex: 1, Region: "重庆", Occupation: "私人健身教练",
		Intro: "自律给我自由。", Tags: []string{"健身", "增肌", "营养", "篮球"}},
	{Phone: "13800138008", Nickname: "郑一鸣", Sex: 1, Region: "北京", Occupation: "算法工程师",
		Intro: "相信数据，也相信直觉。", Tags: []string{"机器学习", "围棋", "咖啡", "科幻"}},
	{Phone: "13800138009", Nickname: "何雨曦", Sex: 2, Region: "湖南·长沙", Occupation: "美食博主",
		Intro: "唯爱与美食不可辜负。", Tags: []string{"美食", "探店", "烘焙", "vlog"}},
	{Phone: "13800138010", Nickname: "孙乐", Sex: 1, Region: "湖北·武汉", Occupation: "在读研究生",
		Intro: "在变好的路上，一步一步来。", Tags: []string{"学习", "游戏", "吉他", "考研上岸"}},
	{Phone: "13800138011", Nickname: "冯哲", Sex: 1, Region: "上海", Occupation: "早期投资人",
		Intro: "投人、投未来。", Tags: []string{"投资", "商业", "高尔夫", "红酒"}},
	{Phone: "13800138012", Nickname: "许念", Sex: 2, Region: "福建·厦门", Occupation: "瑜伽老师",
		Intro: "呼吸之间，回到当下。", Tags: []string{"瑜伽", "冥想", "海边", "素食"}},
	{Phone: "13800138013", Nickname: "杨帆", Sex: 1, Region: "在路上", Occupation: "旅行博主",
		Intro: "世界那么大，我想去看看。", Tags: []string{"旅行", "户外", "潜水", "纪录片"}},
}

// dialogue 描述一段单聊对话：A、B 为 personas 索引，Lines 为按顺序的发言（from 为说话人索引）。
type dialogue struct {
	A, B  int
	Lines []line
}

type line struct {
	From int
	Text string
}

// 单聊对话脚本（主角参与的为主，另有两段非主角对话）。
var dialogues = []dialogue{
	{A: 0, B: 1, Lines: []line{
		{0, "陈默在吗？2.0 的消息已读回执这块，后端接口能赶在周五前给到吗？"},
		{1, "在的～接口基本写完了，今天联调，明天给你 mock 数据。"},
		{0, "太好了，那我先把交互稿同步给设计。"},
		{1, "OK，有字段问题随时拉我。"},
		{0, "👍 辛苦啦"},
	}},
	{A: 0, B: 2, Lines: []line{
		{2, "晚晴，会话列表的红点样式我出了两版，发你看看？"},
		{0, "好呀，我比较倾向数字角标那版，更直观。"},
		{2, "我也觉得！那我就按这版细化了。"},
		{0, "记得深色模式也对一下颜色～"},
		{2, "收到，今晚给你切图。"},
	}},
	{A: 0, B: 5, Lines: []line{
		{5, "大理这几天的云绝了，给你拍了张洱海的日落 🌅"},
		{0, "哇也太治愈了吧，求原图当壁纸！"},
		{5, "哈哈稍等，回酒店传给你。"},
		{0, "下次团建就去大理，我提议的！"},
		{5, "随时欢迎，我当向导。"},
	}},
	{A: 0, B: 9, Lines: []line{
		{9, "长沙新开的一家文和友隔壁的小店，小龙虾绝了！周末来不来？"},
		{0, "馋哭了……可惜这周末要赶版本 😭"},
		{9, "那我先替你吃，给你拍视频馋你。"},
		{0, "你最坏了哈哈哈"},
	}},
	{A: 0, B: 3, Lines: []line{
		{3, "晚晴，下周三方便约个 30 分钟聊聊增长这块吗？"},
		{0, "可以的，周三下午三点？"},
		{3, "完美，我发会议链接给你。"},
		{0, "好嘞👌"},
	}},
	{A: 0, B: 4, Lines: []line{
		{4, "周末成都的活动物料我整理好啦，已经发群里咯～"},
		{0, "辛苦周岚！转化数据记得活动后同步一下。"},
		{4, "没问题，安排上！"},
	}},
	// 非主角对话
	{A: 2, B: 6, Lines: []line{
		{6, "小棠，那套插画风格的图标我画好了，要不要联名出个表情包？😆"},
		{2, "好哇好哇！我来排版，咱俩整一套。"},
		{6, "成交！晚上语音对一下风格。"},
	}},
	{A: 5, B: 13, Lines: []line{
		{13, "航哥，川藏线我下个月出发，要不要一起？"},
		{5, "心动了，我把摄影器材清单发你。"},
		{13, "哈哈那必须的，路上有大片拍。"},
	}},
}

// groupChat 描述一段群聊。Owner/Members 为 personas 索引。
type groupChat struct {
	Name         string
	Icon         string // 留空则用群主头像
	Owner        int
	Invite       []int
	Announcement string
	Lines        []line // From 为 personas 索引

	gid string // 运行时填充
}

var groups = []groupChat{
	{
		Name:         "HiChat 产品研发群",
		Owner:        0,
		Invite:       []int{1, 2, 3, 4, 8},
		Announcement: "📌 2.0 版本冲刺中：本周五完成消息回执联调，下周一提测。有问题随时 @ 我。",
		Lines: []line{
			{0, "各位，2.0 版本进入冲刺阶段啦，目标月底上线 🚀"},
			{1, "后端这边消息回执和已读未读基本就绪，明天联调。"},
			{2, "设计稿我今晚全部切完，深色模式也一起。"},
			{8, "推荐流的算法我先灰度 5%，观察下点击率。"},
			{3, "增长侧我准备了一波内测邀请，配合上线节奏。"},
			{4, "运营物料和公众号推文我来安排，上线当天同步发。"},
			{0, "完美，大家辛苦！周五同步进度 💪"},
		},
	},
	{
		Name:         "周末爬山摄影 🏔",
		Owner:        5,
		Invite:       []int{0, 6, 13},
		Announcement: "本周日苍山徒步，早八集合，记得带好装备和水～",
		Lines: []line{
			{5, "周日苍山，天气晴，适合出片！谁去？"},
			{13, "我去我去，带上无人机。"},
			{6, "我也想去，主要想拍点素材画画 🎨"},
			{0, "我争取！周末看版本进度，尽量去透透气。"},
			{5, "好嘞，早八玉带路入口集合。"},
		},
	},
	{
		Name:         "深漂干饭俱乐部 🍜",
		Owner:        9,
		Invite:       []int{0, 2, 4},
		Announcement: "本群唯一宗旨：好好吃饭。探店报名接龙～",
		Lines: []line{
			{9, "新探到一家潮汕牛肉火锅，人均 80，评分 4.9！"},
			{4, "听起来就很顶，这周五约吗？"},
			{2, "约！我下班直接过去。"},
			{0, "我也来，太久没好好吃顿饭了 🥹"},
			{9, "那就周五晚七点，我去定位子。"},
		},
	},
}

// trendSpec 描述一条动态。Author 为 personas 索引。WithImages>0 表示带几张图（从已上传图片池里取）。
type trendSpec struct {
	Author     int
	Content    string
	WithImages int
	Scope      int // 1 仅自己 2 仅好友 3 所有人
	// 评论者索引 + 内容；点赞者索引
	Comments  []comment
	Likers    []int
}

type comment struct {
	From int
	Text string
}

var trends = []trendSpec{
	{Author: 0, Scope: 3, WithImages: 0,
		Content: "做了三年产品，越来越相信一句话：用户不会为功能买单，只会为「更好的自己」买单。共勉。",
		Comments: []comment{{3, "深以为然，受教了 🙏"}, {1, "所以需求文档能不能写简单点（狗头）"}, {8, "这条我转工作群了。"}},
		Likers:   []int{1, 2, 3, 4, 5, 8, 9}},
	{Author: 5, Scope: 3, WithImages: 3,
		Content: "大理的风花雪月，是要亲眼看过才懂的。洱海边住了三天，每天都舍不得回。📷",
		Comments: []comment{{0, "美哭了，求壁纸！"}, {13, "下次带上我！"}, {6, "这光影绝了，我要画下来。"}},
		Likers:   []int{0, 2, 4, 6, 9, 13}},
	{Author: 9, Scope: 3, WithImages: 2,
		Content: "今日份探店：藏在巷子里的小馆子，一份糖油粑粑治愈整个加班的周一。🍡",
		Comments: []comment{{4, "看饿了！求定位。"}, {0, "长沙人均美食家实锤。"}},
		Likers:   []int{0, 2, 4, 7, 10}},
	{Author: 2, Scope: 3, WithImages: 1,
		Content: "新的图标规范出炉，统一了圆角和描边。细节是设计的灵魂，强迫症狂喜。✨",
		Comments: []comment{{6, "舒服了，看着就干净。"}, {0, "辛苦小棠，体验提升一大截。"}},
		Likers:   []int{0, 1, 3, 6, 8}},
	{Author: 3, Scope: 3, WithImages: 0,
		Content: "第三次创业，最大的体会是：方向比努力重要，但努力让你有资格谈方向。",
		Comments: []comment{{11, "好选手，约下午茶聊聊。"}, {0, "梓睿牛！"}},
		Likers:   []int{0, 1, 4, 8, 11}},
	{Author: 7, Scope: 3, WithImages: 1,
		Content: "坚持晨练第 100 天打卡。身体是革命的本钱，各位久坐的码农们动起来！💪",
		Comments: []comment{{1, "扎心了，这就去站起来。"}, {10, "教练带带我！"}},
		Likers:   []int{0, 1, 3, 8, 10}},
	{Author: 6, Scope: 3, WithImages: 2,
		Content: "摸鱼画了套小动物表情包，已经免费上架啦，欢迎大家斗图时翻我牌子～🐱",
		Comments: []comment{{2, "已下载，太可爱了！"}, {9, "求联名美食版！"}},
		Likers:   []int{0, 2, 4, 9, 13}},
	{Author: 13, Scope: 3, WithImages: 3,
		Content: "第 28 个国家打卡完成。旅行教会我的不是看了多少风景，而是接纳了多少种活法。🌍",
		Comments: []comment{{5, "下次同行！"}, {0, "羡慕这份自由。"}},
		Likers:   []int{0, 2, 5, 6, 9}},
	{Author: 4, Scope: 2, WithImages: 1,
		Content: "成都的周末，火锅配盖碗茶，巴适得很。生活嘛，就该有点烟火气。🌶",
		Comments: []comment{{9, "下次去成都你做向导！"}, {0, "馋了馋了。"}},
		Likers:   []int{0, 5, 9}},
	{Author: 8, Scope: 3, WithImages: 0,
		Content: "推荐系统上线灰度，CTR 提升 12%。数据是会说话的，但要听懂它说的是不是真话。",
		Comments: []comment{{0, "漂亮！"}, {3, "这波增长稳了。"}},
		Likers:   []int{0, 1, 3}},
	{Author: 12, Scope: 3, WithImages: 1,
		Content: "清晨海边的一次冥想，把焦虑都交还给海风。愿你也能在忙碌里给自己留几分钟。🧘",
		Comments: []comment{{0, "需要这种松弛感。"}},
		Likers:   []int{0, 2, 4}},
	{Author: 10, Scope: 3, WithImages: 0,
		Content: "考研上岸啦！谢谢那个没有放弃的自己，也谢谢一路鼓励我的朋友们。🎓",
		Comments: []comment{{7, "牛啊兄弟！"}, {0, "恭喜恭喜🎉"}},
		Likers:   []int{0, 1, 3, 7, 8}},
}
