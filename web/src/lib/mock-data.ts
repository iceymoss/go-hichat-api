// Shared domain types for the IM/social/moments UI, plus a few seed constants
// still used as fallbacks (currentUser, conversationMessagesMap, contacts,
// contactGroups) and the formatTime helper. The bulk of the old demo mock data
// has been removed for production.

export interface User {
  id: string;
  name: string;
  avatar: string;
  status?: string;
  signature?: string;
  region?: string;
  gender?: 'male' | 'female';
  phone?: string;
  tag?: string;
}

export interface Message {
  id: string;
  content: string;
  senderId: string;
  timestamp: Date;
  type: 'text' | 'image' | 'video' | 'file' | 'voice' | 'memes' | 'system';
  imageUrl?: string;
  replyTo?: { senderName: string; content: string; msgId?: string; senderId?: string; mType?: Message['type']; thumbUrl?: string };
  /** 发送状态: sending=发送中, sent=已发送, failed=发送失败 */
  status?: 'sending' | 'sent' | 'failed';
  /** 已读状态（用于发送方显示 ✓✓）；私聊为布尔，群聊用 readCount/readTotal 细化 */
  isRead?: boolean;
  /** 群聊已读人数（不含发送者自己） */
  readCount?: number;
  /** 群聊总人数（不含发送者自己） */
  readTotal?: number;
  /** 是否已撤回；为 true 时原位渲染"XX撤回了一条消息" */
  recalled?: boolean;
  /** 撤回操作者 uid，用于区分"你/对方/管理员撤回了一条消息" */
  recalledBy?: string;
  /** 被 @ 的成员 uid 列表（群聊） */
  atUsers?: string[];
  /** 是否 @所有人（群聊） */
  atAll?: boolean;
}

export interface Conversation {
  id: string;
  type: 'private' | 'group';
  name: string;
  avatar: string;
  lastMessage: string;
  lastMessageTime: Date;
  unreadCount: number;
  pinned: boolean;
  muted: boolean;
  members?: number;
  online?: boolean;
  /** 是否有未读的 @我（群聊），进会话后清除 */
  hasAtMe?: boolean;
}

export interface ContactGroup {
  title: string;
  icon: string;
  contacts?: Contact[];
}

export interface Contact {
  id: string;
  name: string;
  avatar: string;
  status?: string;
  pinyin: string;
  letter: string;
  online?: boolean;
  // New fields for profile card
  gender?: 'male' | 'female';
  age?: number;
  phone?: string;
  email?: string;
  region?: string;
  signature?: string;
  account?: string;  // HiChat ID
  remark?: string;   // 备注 (friend remark name)
  nickname?: string; // 用户真实昵称（网名）；name 可能是 remark||nickname 的展示名
  introduction?: string;
  occupation?: string;
  tags?: string;
  // Friend settings from API
  blacklisted?: boolean;
  moments_permission?: number;
  notify_enabled?: boolean;
  pinned?: boolean;
  muted?: boolean;
  friend_tags?: string[];
  friend_uid?: string;
}

// Current user
export const currentUser: User = {
  id: 'me',
  name: '我',
  avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=Felix',
  status: 'online',
  signature: '保持热爱，奔赴山海 ✨',
  region: '广东 深圳',
  gender: 'male',
  phone: '138****8888',
};

// Per-conversation messages map (keyed by conversation id)
export const conversationMessagesMap: Record<string, Message[]> = {
  // c1 — 张三 (private)
  c1: [
    { id: 'c1-m1', content: '嗨，最近项目进展怎么样了？', senderId: 'zhang', timestamp: new Date(Date.now() - 2 * 3600000), type: 'text' },
    { id: 'c1-m2', content: '还不错，基本功能已经开发完了，正在做UI优化', senderId: 'me', timestamp: new Date(Date.now() - 1.9 * 3600000), type: 'text' },
    { id: 'c1-m3', content: '那太好了！我记得你之前说要做类似微信的IM应用？', senderId: 'zhang', timestamp: new Date(Date.now() - 1.8 * 3600000), type: 'text' },
    { id: 'c1-m4', content: '对的，后端API已经写好了，现在主要是前端界面的设计和开发', senderId: 'me', timestamp: new Date(Date.now() - 1.7 * 3600000), type: 'text' },
    { id: 'c1-m5', content: '前端用的是什么技术栈？', senderId: 'zhang', timestamp: new Date(Date.now() - 1.6 * 3600000), type: 'text' },
    { id: 'c1-m6', content: 'Next.js 16 + TypeScript + Tailwind CSS + shadcn/ui，非常现代化的方案', senderId: 'me', timestamp: new Date(Date.now() - 1.5 * 3600000), type: 'text' },
    { id: 'c1-m7', content: '听起来很棒！shadcn/ui 确实很好用，组件质量很高 👍', senderId: 'zhang', timestamp: new Date(Date.now() - 1.4 * 3600000), type: 'text' },
    { id: 'c1-m8', content: '是的，而且可以自定义主题，非常适合做这种需要高度定制化的应用', senderId: 'me', timestamp: new Date(Date.now() - 1.3 * 3600000), type: 'text' },
    { id: 'c1-m9', content: '对了，你UI设计有参考什么应用吗？', senderId: 'zhang', timestamp: new Date(Date.now() - 1 * 3600000), type: 'text' },
    { id: 'c1-m10', content: '主要是参考微信的设计理念，但加入了一些现代化的元素', senderId: 'me', timestamp: new Date(Date.now() - 0.9 * 3600000), type: 'text' },
    { id: 'c1-m11', content: '明天下午3点开会，记得准备一下PPT', senderId: 'zhang', timestamp: new Date(Date.now() - 3 * 60000), type: 'text' },
  ],

  // c2 — 产品研发群 (group)
  c2: [
    { id: 'c2-m1', content: '大家早上好！今天的站会改到10点', senderId: 'ct3', timestamp: new Date(Date.now() - 4 * 3600000), type: 'text' },
    { id: 'c2-m2', content: '收到，我调整一下日程', senderId: 'me', timestamp: new Date(Date.now() - 3.9 * 3600000), type: 'text' },
    { id: 'c2-m3', content: '好的，我也知道了', senderId: 'ct13', timestamp: new Date(Date.now() - 3.8 * 3600000), type: 'text' },
    { id: 'c2-m4', content: '顺便说一下，新版本已经部署到测试环境了', senderId: 'ct13', timestamp: new Date(Date.now() - 1 * 3600000), type: 'text' },
    { id: 'c2-m5', content: '大家帮忙测试一下，有问题及时反馈', senderId: 'ct13', timestamp: new Date(Date.now() - 0.9 * 3600000), type: 'text' },
    { id: 'c2-m6', content: '好的，我下午开始测', senderId: 'me', timestamp: new Date(Date.now() - 0.8 * 3600000), type: 'text' },
    { id: 'c2-m7', content: '我已经测了几个模块，登录和注册都正常', senderId: 'ct27', timestamp: new Date(Date.now() - 15 * 60000), type: 'text' },
  ],

  // c3 — 李四 (private)
  c3: [
    { id: 'c3-m1', content: '这个设计稿你看了吗？', senderId: 'ct13', timestamp: new Date(Date.now() - 5 * 3600000), type: 'text' },
    { id: 'c3-m2', content: '看了，整体感觉不错', senderId: 'me', timestamp: new Date(Date.now() - 4.9 * 3600000), type: 'text' },
    { id: 'c3-m3', content: '感觉配色还可以优化一下', senderId: 'ct13', timestamp: new Date(Date.now() - 4.8 * 3600000), type: 'text' },
    { id: 'c3-m4', content: '嗯，蓝色系确实更符合产品调性', senderId: 'me', timestamp: new Date(Date.now() - 4.7 * 3600000), type: 'text' },
    { id: 'c3-m5', content: '对，我再调一版给你看', senderId: 'ct13', timestamp: new Date(Date.now() - 4.5 * 3600000), type: 'text' },
  ],

  // c4 — 王五 (private)
  c4: [
    { id: 'c4-m1', content: '帮我看看这个bug，我搞了一下午了', senderId: 'ct22', timestamp: new Date(Date.now() - 6 * 3600000), type: 'text' },
    { id: 'c4-m2', content: '什么问题？截图给我看看', senderId: 'me', timestamp: new Date(Date.now() - 5.9 * 3600000), type: 'text' },
    { id: 'c4-m3', content: '就是接口返回数据不对，排序有问题', senderId: 'ct22', timestamp: new Date(Date.now() - 5.8 * 3600000), type: 'text' },
    { id: 'c4-m4', content: '我看到了，是少了一个排序参数。你加上 `orderBy: desc` 试试', senderId: 'me', timestamp: new Date(Date.now() - 5.7 * 3600000), type: 'text' },
    { id: 'c4-m5', content: '好的，收到', senderId: 'ct22', timestamp: new Date(Date.now() - 2 * 3600000), type: 'text' },
  ],

  // c5 — 周末约饭群 (group)
  c5: [
    { id: 'c5-m1', content: '这周六聚餐去哪吃？', senderId: 'ct8', timestamp: new Date(Date.now() - 8 * 3600000), type: 'text' },
    { id: 'c5-m2', content: '我想吃火锅！', senderId: 'ct22', timestamp: new Date(Date.now() - 7.9 * 3600000), type: 'text' },
    { id: 'c5-m3', content: '可以可以，海底捞走起', senderId: 'me', timestamp: new Date(Date.now() - 7.8 * 3600000), type: 'text' },
    { id: 'c5-m4', content: '周六晚上7点，老地方见！', senderId: 'ct8', timestamp: new Date(Date.now() - 5 * 3600000), type: 'text' },
  ],

  // c6 — 赵六 (private)
  c6: [
    { id: 'c6-m1', content: '那个项目进展怎么样了？', senderId: 'ct28', timestamp: new Date(Date.now() - 10 * 3600000), type: 'text' },
    { id: 'c6-m2', content: '已经完成80%了，还剩一些细节需要调整', senderId: 'me', timestamp: new Date(Date.now() - 9.9 * 3600000), type: 'text' },
    { id: 'c6-m3', content: '预计什么时候能上线？', senderId: 'ct28', timestamp: new Date(Date.now() - 9.8 * 3600000), type: 'text' },
    { id: 'c6-m4', content: '下周应该可以，到时候先发个内测版本', senderId: 'me', timestamp: new Date(Date.now() - 9.7 * 3600000), type: 'text' },
  ],

  // c7 — 文件传输助手
  c7: [
    { id: 'c7-m1', content: '[文件] 项目需求文档v2.0.pdf', senderId: 'me', timestamp: new Date(Date.now() - 24 * 3600000), type: 'text' },
    { id: 'c7-m2', content: '[文件] 技术方案评审记录.docx', senderId: 'me', timestamp: new Date(Date.now() - 23 * 3600000), type: 'text' },
    { id: 'c7-m3', content: '[文件] 接口设计文档.xlsx', senderId: 'me', timestamp: new Date(Date.now() - 22 * 3600000), type: 'text' },
  ],

  // c8 — 技术交流群 (group)
  c8: [
    { id: 'c8-m1', content: '有人用过 Bun 吗？', senderId: 'ct18', timestamp: new Date(Date.now() - 3 * 86400000), type: 'text' },
    { id: 'c8-m2', content: '比 Node 快多了，强烈推荐', senderId: 'ct20', timestamp: new Date(Date.now() - 2.9 * 86400000), type: 'text' },
    { id: 'c8-m3', content: '我一直在用，开发体验很好', senderId: 'me', timestamp: new Date(Date.now() - 2.8 * 86400000), type: 'text' },
    { id: 'c8-m4', content: '但生态还是不如 Node 成熟', senderId: 'ct29', timestamp: new Date(Date.now() - 2.7 * 86400000), type: 'text' },
  ],

  // c9 — 小七 (private)
  c9: [
    { id: 'c9-m1', content: '哈哈哈那个表情包太搞笑了', senderId: 'ct19', timestamp: new Date(Date.now() - 3 * 86400000), type: 'text' },
    { id: 'c9-m2', content: '是吧！我也笑死', senderId: 'me', timestamp: new Date(Date.now() - 2.9 * 86400000), type: 'text' },
    { id: 'c9-m3', content: '你有没有别的？发我看看', senderId: 'ct19', timestamp: new Date(Date.now() - 2.8 * 86400000), type: 'text' },
  ],

  // c10 — 陈经理 (private)
  c10: [
    { id: 'c10-m1', content: '下周一之前把方案发给我', senderId: 'ct3', timestamp: new Date(Date.now() - 5 * 86400000), type: 'text' },
    { id: 'c10-m2', content: '好的陈经理，我尽快整理', senderId: 'me', timestamp: new Date(Date.now() - 4.9 * 86400000), type: 'text' },
    { id: 'c10-m3', content: '方案的重点放在技术选型和架构设计上', senderId: 'ct3', timestamp: new Date(Date.now() - 4.8 * 86400000), type: 'text' },
    { id: 'c10-m4', content: '收到，我这两天花时间好好写一下', senderId: 'me', timestamp: new Date(Date.now() - 4.7 * 86400000), type: 'text' },
  ],
};

// Group conversation → member list (for call member selection, etc.)
export interface GroupMember {
  id: string;
  name: string;
  online: boolean;
}

// Contacts
export const contactGroups: ContactGroup[] = [
  { title: '新的朋友', icon: 'UserPlus' },
  { title: '群聊', icon: 'Users' },
  { title: '标签', icon: 'Tag' },
  { title: '公众号', icon: 'Megaphone' },
];

export const contacts: Contact[] = [
  { id: 'ct1', name: '阿杰', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=AJ', pinyin: 'ajie', letter: 'A', online: true, gender: 'male', age: 28, phone: '138****5612', region: '广东深圳', signature: '代码改变世界', account: 'hichat_ajie' },
  { id: 'ct2', name: '艾伦', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=AL', pinyin: 'ailun', letter: 'A', gender: 'male', age: 32, phone: '139****8823', region: '上海', account: 'hichat_ailun', signature: '保持好奇心' },
  { id: 'ct3', name: '陈经理', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=Chen', pinyin: 'chenjingli', letter: 'C', online: false, gender: 'male', age: 40, phone: '136****2091', region: '北京', account: 'hichat_chenjl', remark: '陈总' },
  { id: 'ct4', name: '陈思远', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=CSY', pinyin: 'chensiyuan', letter: 'C', online: true, gender: 'female', age: 26, phone: '158****7743', region: '浙江杭州', signature: '生活不止眼前的苟且', account: 'hichat_chensy', remark: '思远' },
  { id: 'ct5', name: '大头', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=DT', pinyin: 'datou', letter: 'D', gender: 'male', age: 24, phone: '137****3320', region: '四川成都', account: 'hichat_datou', signature: '吃好喝好' },
  { id: 'ct6', name: '方圆', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=FY', pinyin: 'fangyuan', letter: 'F', online: true, gender: 'male', age: 30, phone: '135****9981', region: '江苏南京', account: 'hichat_fangyuan', signature: '脚踏实地，仰望星空', remark: '方方' },
  { id: 'ct7', name: '顾小白', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=GXB', pinyin: 'guxiaobai', letter: 'G', gender: 'female', age: 25, phone: '152****4456', region: '湖北武汉', account: 'hichat_guxb' },
  { id: 'ct8', name: '韩梅梅', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=HMM', pinyin: 'hanmeimei', letter: 'H', online: true, gender: 'female', age: 27, phone: '139****1157', region: '广东广州', signature: '项目终于交付了！感谢团队 ❤️', account: 'hichat_hmm', remark: '梅梅' },
  { id: 'ct9', name: '黄小明', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=HXM', pinyin: 'huangxiaoming', letter: 'H', gender: 'male', age: 35, phone: '186****6629', region: '福建厦门', account: 'hichat_hxm' },
  { id: 'ct10', name: '贾玲', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=JL', pinyin: 'jialing', letter: 'J', gender: 'female', age: 38, phone: '158****2234', region: '北京', account: 'hichat_jl', signature: '开心最重要' },
  { id: 'ct11', name: '老王', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=LW', pinyin: 'laowang', letter: 'L', online: true, gender: 'male', age: 45, phone: '133****8870', region: '山东济南', account: 'hichat_laowang', signature: '老骥伏枥', remark: '王哥' },
  { id: 'ct12', name: '李雷', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=LL', pinyin: 'leilei', letter: 'L', gender: 'male', age: 23, phone: '177****5548', region: '广东深圳', account: 'hichat_leilei', signature: 'Hello World!' },
  { id: 'ct13', name: '李四', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=Li', pinyin: 'lisi', letter: 'L', online: true, gender: 'male', age: 29, phone: '138****3376', region: '浙江杭州', signature: '设计是一种态度 🎨', account: 'hichat_lisi' },
  { id: 'ct14', name: '林黛玉', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=LDY', pinyin: 'lindaiyu', letter: 'L', gender: 'female', age: 22, phone: '156****9912', region: '江苏苏州', signature: '花谢花飞花满天', account: 'hichat_ldy' },
  { id: 'ct15', name: '马超', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=MC', pinyin: 'machao', letter: 'M', gender: 'male', age: 31, phone: '185****2207', region: '陕西西安', account: 'hichat_machao', signature: '音乐是最好的治愈 ❤️' },
  { id: 'ct16', name: '聂小倩', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=NXQ', pinyin: 'niexiaoqian', letter: 'N', online: true, gender: 'female', age: 24, phone: '159****4435', region: '广东深圳', signature: '周末的下午茶时光 ☕️', account: 'hichat_nxq', remark: '小倩' },
  { id: 'ct17', name: '欧阳锋', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=OYF', pinyin: 'ouyangfeng', letter: 'O', gender: 'male', age: 42, phone: '136****7761', region: '甘肃兰州', account: 'hichat_oyf' },
  { id: 'ct18', name: '乔峰', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=QF', pinyin: 'qiaofeng', letter: 'Q', online: true, gender: 'male', age: 36, phone: '188****1134', region: '辽宁大连', account: 'hichat_qf', signature: '大碗喝酒，大口吃肉', remark: '峰哥' },
  { id: 'ct19', name: '小七', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=Qi', pinyin: 'qi', letter: 'Q', gender: 'female', age: 23, phone: '157****6689', region: '四川成都', signature: '每天都要元气满满！', account: 'hichat_qiqi', remark: '七七' },
  { id: 'ct20', name: '孙悟空', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=SWK', pinyin: 'sunwukong', letter: 'S', online: true, gender: 'male', age: 33, phone: '139****5523', region: '北京', signature: '七十二变，也不如代码写得好', account: 'hichat_swk', remark: '猴哥' },
  { id: 'ct21', name: '唐僧', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=TS', pinyin: 'tangseng', letter: 'T', gender: 'male', age: 39, phone: '186****8847', region: '河南洛阳', account: 'hichat_ts', signature: '心诚则灵' },
  { id: 'ct22', name: '王五', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=Wang', pinyin: 'wangwu', letter: 'W', gender: 'male', age: 27, phone: '138****2290', region: '广东深圳', account: 'hichat_wangwu' },
  { id: 'ct23', name: '魏无羡', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=WWX', pinyin: 'weiwuxian', letter: 'W', online: true, gender: 'male', age: 26, phone: '155****7716', region: '江苏南京', signature: '音乐是最好的治愈 ❤️', account: 'hichat_wwx', remark: '羡羡' },
  { id: 'ct24', name: '谢逊', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=XX', pinyin: 'xiexun', letter: 'X', gender: 'male', age: 44, phone: '133****4428', region: '新疆乌鲁木齐', account: 'hichat_xx' },
  { id: 'ct25', name: '杨过', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=YG', pinyin: 'yangguo', letter: 'Y', online: true, gender: 'male', age: 30, phone: '159****5567', region: '湖北武汉', signature: '问世间情为何物', account: 'hichat_yg' },
  { id: 'ct26', name: '叶问', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=YW', pinyin: 'yewen', letter: 'Y', gender: 'male', age: 41, phone: '137****8834', region: '广东佛山', account: 'hichat_yw', signature: '咏春拳，一代宗师' },
  { id: 'ct27', name: '张三', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=Zhang', pinyin: 'zhangsan', letter: 'Z', online: true, gender: 'male', age: 28, phone: '136****7745', region: '广东深圳', signature: '保持热爱，奔赴山海 ✨', account: 'hichat_zhangsan' },
  { id: 'ct28', name: '赵六', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=Zhao', pinyin: 'zhaoliu', letter: 'Z', gender: 'male', age: 34, phone: '185****1123', region: '浙江杭州', account: 'hichat_zhaoliu' },
  { id: 'ct29', name: '周芷若', avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=ZZR', pinyin: 'zhouzhiruo', letter: 'Z', online: true, gender: 'female', age: 25, phone: '158****6678', region: '湖南长沙', signature: '经典中的经典 🎵', account: 'hichat_zzr', remark: '若若' },
];

// Helper function to format time
export function formatTime(date: Date): string {
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const minutes = Math.floor(diff / 60000);
  const hm = `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`;

  // 判断是否同一天（基于日期而非小时差）
  const isToday = date.getFullYear() === now.getFullYear() && date.getMonth() === now.getMonth() && date.getDate() === now.getDate();
  const yesterday = new Date(now);
  yesterday.setDate(yesterday.getDate() - 1);
  const isYesterday = date.getFullYear() === yesterday.getFullYear() && date.getMonth() === yesterday.getMonth() && date.getDate() === yesterday.getDate();

  if (isToday) {
    if (minutes < 1) return '刚刚';
    if (minutes < 60) return `${minutes}分钟前`;
    return hm;
  }
  if (isYesterday) return `昨天 ${hm}`;

  const days = Math.floor(diff / 86400000);
  if (days < 7) {
    const weekdays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];
    return `${weekdays[date.getDay()]} ${hm}`;
  }
  const md = `${(date.getMonth() + 1).toString().padStart(2, '0')}/${date.getDate().toString().padStart(2, '0')}`;
  if (date.getFullYear() !== now.getFullYear()) return `${date.getFullYear()}/${md} ${hm}`;
  return `${md} ${hm}`;
}

/* ═══════════════════════════════════════
   Group Management
   ═══════════════════════════════════════ */

export type GroupRoleLevel = 0 | 1 | 2; // 0=member, 1=admin, 2=owner
export type GroupAppResult = 0 | 1 | 2 | 3; // 0=pending, 1=accepted, 2=rejected, 3=ignored
export type GroupAppClass = 'received' | 'sent';
export type GroupJoinSource = 1 | 2 | 3; // 1=apply, 2=invite, 3=link

export interface GroupInfo {
  id: string;
  name: string;
  icon: string;
  description?: string;   // 群描述
  isVerify: boolean;
  notification: string;
  createUid: string;
  groupNickname?: string;
  groupRemark?: string;
}

export interface GroupMemberInfo {
  id: number;
  groupId: string;
  userId: string;
  nickname: string;
  avatar?: string;
  roleLevel: GroupRoleLevel;
  online: boolean;
  groupNickname?: string;
  groupRemark?: string;
}

export interface GroupApplication {
  id: number;
  userId: string;
  userName: string;
  userAvatar: string;
  groupId: string;
  groupName: string;
  groupIcon: string;
  reqMsg: string;
  reqTime: Date;
  joinSource: GroupJoinSource;
  inviterName?: string;
  handleResult: GroupAppResult;
  readState: boolean;
}

export interface GroupInviteLink {
  token: string;
  groupId: string;
  createdBy: string;
  createdAt: Date;
  expireAt: Date | null;
  maxUses: number;
  usedCount: number;
  revoked: boolean;
}

export interface GroupAnnouncement {
  id: number;
  groupId: string;
  content: string;
  createdBy: string;
  createdAt: Date;
  pinned: boolean;
}

export interface GroupMemberSetting {
  groupId: string;
  groupNickname: string;
  groupRemark: string;
}

/* ═══════════════════════════════════════
   Moments / Trends Feed
   ═══════════════════════════════════════ */

export type TrendType = 1 | 2 | 3 | 4 | 5; // 1=text, 2=image, 3=article, 4=share, 5=video
export type TrendScope = 1 | 2 | 3; // 1=self-only, 2=friends, 3=all

export interface Trend {
  id: number;
  userId: string;
  type: TrendType;
  content: string;
  scope: TrendScope;
  createTime: Date;
  replyCount: number;
  agreeCount: number;
  positionName: string;
  title: string;
  atUsers: { id: string; name: string; avatar: string }[];
  resources: string[];       // images (type 2/5) or video cover
  coverUrl: string;          // article cover (type 3)
  shareUrl: string;          // link share (type 4)
  openReply: boolean;
  isTop: boolean;
  /** Poster display name — filled by the trend-api mapper when loaded from backend. */
  userName?: string;
  /** Poster avatar URL — filled by the trend-api mapper when loaded from backend. */
  userAvatar?: string;
}

export interface TrendComment {
  id: number;
  trendId: number;
  rootId: number;            // 0 = root comment
  father: number;            // 0 = no parent
  replyer: { id: string; name: string; avatar: string };
  user: { id: string; name: string; avatar: string }; // the person being replied to
  level: number;
  content: string;
  agreeCount: number;
  children: TrendComment[];
  createTime: Date;
}

export interface MomentsNotification {
  id: number;
  type: 'reply' | 'like' | 'comment' | 'at_trend' | 'at_comment';
  trendId: number;
  commentId?: number;         // 关联评论ID（评论/回复/@评论）
  trendContent: string;       // preview text
  actor: { id: string; name: string; avatar: string };
  content?: string;           // comment text (for reply type)
  read: boolean;
  createTime: Date;
}

