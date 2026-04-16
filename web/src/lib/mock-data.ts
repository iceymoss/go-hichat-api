// Mock data for IM application demo

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
  type: 'text' | 'image' | 'voice' | 'system';
  imageUrl?: string;
  replyTo?: { senderName: string; content: string };
  /** 发送状态: sending=发送中, sent=已发送, failed=发送失败 */
  status?: 'sending' | 'sent' | 'failed';
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

export interface Moment {
  id: string;
  userId: string;
  userName: string;
  userAvatar: string;
  content: string;
  images?: string[];
  timestamp: Date;
  likes: { userId: string; userName: string }[];
  comments: { userId: string; userName: string; content: string }[];
  location?: string;
}

export interface ProfileMenuItem {
  icon: string;
  label: string;
  value?: string;
  href?: string;
  badge?: number;
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

// Conversations
export const conversations: Conversation[] = [
  {
    id: 'c1',
    type: 'private',
    name: '张三',
    avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=Zhang',
    lastMessage: '明天下午3点开会，记得准备一下PPT',
    lastMessageTime: new Date(Date.now() - 3 * 60000),
    unreadCount: 3,
    pinned: true,
    muted: false,
    online: true,
  },
  {
    id: 'c2',
    type: 'group',
    name: '产品研发群',
    avatar: 'https://api.dicebear.com/9.x/identicon/svg?seed=dev-team',
    lastMessage: '李四: 新版本已经部署到测试环境了',
    lastMessageTime: new Date(Date.now() - 15 * 60000),
    unreadCount: 12,
    pinned: true,
    muted: false,
    members: 48,
  },
  {
    id: 'c3',
    type: 'private',
    name: '李四',
    avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=Li',
    lastMessage: '这个设计稿你看了吗？感觉配色还可以优化一下',
    lastMessageTime: new Date(Date.now() - 1 * 3600000),
    unreadCount: 0,
    pinned: false,
    muted: false,
    online: true,
  },
  {
    id: 'c4',
    type: 'private',
    name: '王五',
    avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=Wang',
    lastMessage: '好的，收到',
    lastMessageTime: new Date(Date.now() - 2 * 3600000),
    unreadCount: 0,
    pinned: false,
    muted: false,
    online: false,
  },
  {
    id: 'c5',
    type: 'group',
    name: '周末约饭群',
    avatar: 'https://api.dicebear.com/9.x/identicon/svg?seed=food-group',
    lastMessage: '周六晚上7点，老地方见！',
    lastMessageTime: new Date(Date.now() - 5 * 3600000),
    unreadCount: 0,
    pinned: false,
    muted: true,
    members: 8,
  },
  {
    id: 'c6',
    type: 'private',
    name: '赵六',
    avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=Zhao',
    lastMessage: '那个项目进展怎么样了？',
    lastMessageTime: new Date(Date.now() - 8 * 3600000),
    unreadCount: 1,
    pinned: false,
    muted: false,
    online: false,
  },
  {
    id: 'c7',
    type: 'private',
    name: '文件传输助手',
    avatar: 'https://api.dicebear.com/9.x/shapes/svg?seed=file-transfer',
    lastMessage: '[文件] 项目需求文档v2.0.pdf',
    lastMessageTime: new Date(Date.now() - 24 * 3600000),
    unreadCount: 0,
    pinned: false,
    muted: false,
  },
  {
    id: 'c8',
    type: 'group',
    name: '技术交流群',
    avatar: 'https://api.dicebear.com/9.x/identicon/svg?seed=tech-group',
    lastMessage: '有人用过 Bun 吗？比 Node 快多了',
    lastMessageTime: new Date(Date.now() - 2 * 86400000),
    unreadCount: 0,
    pinned: false,
    muted: true,
    members: 156,
  },
  {
    id: 'c9',
    type: 'private',
    name: '小七',
    avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=Qi',
    lastMessage: '哈哈哈那个表情包太搞笑了',
    lastMessageTime: new Date(Date.now() - 3 * 86400000),
    unreadCount: 0,
    pinned: false,
    muted: false,
    online: true,
  },
  {
    id: 'c10',
    type: 'private',
    name: '陈经理',
    avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=Chen',
    lastMessage: '下周一之前把方案发给我',
    lastMessageTime: new Date(Date.now() - 4 * 86400000),
    unreadCount: 0,
    pinned: false,
    muted: false,
    online: false,
  },
];

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

// Keep a backward-compatible export for any remaining references
export const chatMessages: Message[] = conversationMessagesMap['c1'] || [];

// Group member display name map (for group messages)
export const groupMemberNames: Record<string, string> = {
  ct3: '陈经理',
  ct13: '李四',
  ct22: '王五',
  ct27: '张三',
  ct28: '赵六',
  ct8: '韩梅梅',
  ct18: '乔峰',
  ct20: '孙悟空',
  ct29: '周芷若',
  ct19: '小七',
  ct6: '方圆',
};

// Group conversation → member list (for call member selection, etc.)
export interface GroupMember {
  id: string;
  name: string;
  online: boolean;
}

export const groupMembersMap: Record<string, GroupMember[]> = {
  c2: [
    { id: 'ct3', name: '陈经理', online: false },
    { id: 'ct13', name: '李四', online: true },
    { id: 'ct22', name: '王五', online: false },
    { id: 'ct27', name: '张三', online: true },
    { id: 'ct28', name: '赵六', online: false },
    { id: 'ct8', name: '韩梅梅', online: true },
    { id: 'ct6', name: '方圆', online: true },
    { id: 'ct18', name: '乔峰', online: true },
    { id: 'ct20', name: '孙悟空', online: true },
    { id: 'ct29', name: '周芷若', online: true },
    { id: 'ct19', name: '小七', online: true },
  ],
  c5: [
    { id: 'ct8', name: '韩梅梅', online: true },
    { id: 'ct22', name: '王五', online: false },
    { id: 'ct13', name: '李四', online: true },
    { id: 'ct27', name: '张三', online: true },
    { id: 'ct6', name: '方圆', online: true },
    { id: 'ct19', name: '小七', online: true },
    { id: 'ct1', name: '阿杰', online: true },
    { id: 'ct20', name: '孙悟空', online: true },
  ],
  c8: [
    { id: 'ct18', name: '乔峰', online: true },
    { id: 'ct20', name: '孙悟空', online: true },
    { id: 'ct29', name: '周芷若', online: true },
    { id: 'ct13', name: '李四', online: true },
    { id: 'ct22', name: '王五', online: false },
    { id: 'ct3', name: '陈经理', online: false },
    { id: 'ct8', name: '韩梅梅', online: true },
    { id: 'ct6', name: '方圆', online: true },
    { id: 'ct27', name: '张三', online: true },
    { id: 'ct28', name: '赵六', online: false },
    { id: 'ct16', name: '聂小倩', online: true },
    { id: 'ct23', name: '魏无羡', online: true },
  ],
};

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

// Moments
export const moments: Moment[] = [
  {
    id: 'mo1',
    userId: 'ct27',
    userName: '张三',
    userAvatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=Zhang',
    content: '今天天气真好，和团队一起去爬山了 🏔️ 爬到山顶看到了绝美的日落，所有的疲惫都值得了！',
    images: [
      'https://picsum.photos/seed/mountain1/400/300',
      'https://picsum.photos/seed/sunset2/400/300',
      'https://picsum.photos/seed/nature3/400/300',
    ],
    timestamp: new Date(Date.now() - 2 * 3600000),
    likes: [
      { userId: 'ct13', userName: '李四' },
      { userId: 'ct6', userName: '方圆' },
      { userId: 'ct8', userName: '韩梅梅' },
    ],
    comments: [
      { userId: 'ct13', userName: '李四', content: '风景太美了！下次带我一起 😄' },
      { userId: 'ct8', userName: '韩梅梅', content: '好羡慕，我也想去！' },
    ],
    location: '深圳·梧桐山',
  },
  {
    id: 'mo2',
    userId: 'ct8',
    userName: '韩梅梅',
    userAvatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=HMM',
    content: '终于把这个项目交付了，感谢团队的每一个人！三年的心血，终于有了成果 🎉 特别感谢项目经理陈经理的指导和支持。',
    timestamp: new Date(Date.now() - 8 * 3600000),
    likes: [
      { userId: 'ct3', userName: '陈经理' },
      { userId: 'ct27', userName: '张三' },
      { userId: 'ct22', userName: '王五' },
      { userId: 'ct1', userName: '阿杰' },
      { userId: 'ct20', userName: '孙悟空' },
    ],
    comments: [
      { userId: 'ct3', userName: '陈经理', content: '辛苦了！这个项目做得非常出色 👏' },
      { userId: 'ct27', userName: '张三', content: '恭喜恭喜！庆祝一下 🍾' },
    ],
  },
  {
    id: 'mo3',
    userId: 'ct13',
    userName: '李四',
    userAvatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=Li',
    content: '分享一个设计灵感：扁平化风格配合微妙的动效，可以让界面既简洁又有层次感。推荐大家看看 Dribbble 上的最新趋势。#设计 #UI',
    images: [
      'https://picsum.photos/seed/design1/400/400',
      'https://picsum.photos/seed/design2/400/300',
    ],
    timestamp: new Date(Date.now() - 1 * 86400000),
    likes: [
      { userId: 'ct27', userName: '张三' },
      { userId: 'ct19', userName: '小七' },
    ],
    comments: [
      { userId: 'ct19', userName: '小七', content: '收藏了！最近正在研究设计趋势' },
    ],
  },
  {
    id: 'mo4',
    userId: 'ct20',
    userName: '孙悟空',
    userAvatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=SWK',
    content: '七十二变，也不如代码写得好 😤 今天的bug又修了一天...不过终于解决了，还是很有成就感的！',
    timestamp: new Date(Date.now() - 2 * 86400000),
    likes: [
      { userId: 'ct18', userName: '乔峰' },
      { userId: 'ct25', userName: '杨过' },
    ],
    comments: [
      { userId: 'ct18', userName: '乔峰', content: '哈哈，程序员的日常 😂' },
      { userId: 'ct25', userName: '杨过', content: '哪个bug？说出来让我开心一下 😏' },
    ],
  },
  {
    id: 'mo5',
    userId: 'ct16',
    userName: '聂小倩',
    userAvatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=NXQ',
    content: '周末的下午茶时光 ☕️',
    images: [
      'https://picsum.photos/seed/tea1/300/300',
      'https://picsum.photos/seed/cafe2/300/300',
      'https://picsum.photos/seed/dessert3/300/300',
      'https://picsum.photos/seed/afternoon4/300/300',
    ],
    timestamp: new Date(Date.now() - 3 * 86400000),
    likes: [
      { userId: 'ct8', userName: '韩梅梅' },
      { userId: 'ct29', userName: '周芷若' },
      { userId: 'ct23', userName: '魏无羡' },
      { userId: 'ct26', userName: '叶问' },
    ],
    comments: [],
    location: '深圳·海岸城',
  },
  {
    id: 'mo6',
    userId: 'ct23',
    userName: '魏无羡',
    userAvatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=WWX',
    content: '音乐是最好的治愈 ❤️ 推荐一首最近单曲循环的歌：坂本龙一 - Merry Christmas Mr. Lawrence',
    timestamp: new Date(Date.now() - 5 * 86400000),
    likes: [
      { userId: 'ct16', userName: '聂小倩' },
      { userId: 'ct29', userName: '周芷若' },
      { userId: 'ct15', userName: '马超' },
    ],
    comments: [
      { userId: 'ct16', userName: '聂小倩', content: '这首确实很好听！' },
      { userId: 'ct29', userName: '周芷若', content: '经典中的经典 🎵' },
    ],
  },
];

// Profile menu items
export const profileMenuItems: ProfileMenuItem[] = [
  { icon: 'Star', label: '收藏' },
  { icon: 'Image', label: '相册' },
  { icon: 'CreditCard', label: '卡包' },
  { icon: 'Emoji', label: '表情' },
];

export const settingsMenuItems: ProfileMenuItem[] = [
  { icon: 'Settings', label: '设置' },
];

// Helper function to format time
export function formatTime(date: Date): string {
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const minutes = Math.floor(diff / 60000);
  const hours = Math.floor(diff / 3600000);
  const days = Math.floor(diff / 86400000);

  if (minutes < 1) return '刚刚';
  if (minutes < 60) return `${minutes}分钟前`;
  if (hours < 24) {
    const today = new Date();
    if (date.getDate() === today.getDate()) {
      return `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`;
    }
    return `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`;
  }
  if (days === 1) return '昨天';
  if (days < 7) {
    const weekdays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];
    return weekdays[date.getDay()];
  }
  return `${(date.getMonth() + 1).toString().padStart(2, '0')}/${date.getDate().toString().padStart(2, '0')}`;
}

// Moments images per user (for profile card display)
export const userMomentsImages: Record<string, string[]> = {
  'ct27': ['https://picsum.photos/seed/mo1a/200/200', 'https://picsum.photos/seed/mo1b/200/200', 'https://picsum.photos/seed/mo1c/200/200'],
  'ct13': ['https://picsum.photos/seed/mo2a/200/200', 'https://picsum.photos/seed/mo2b/200/200', 'https://picsum.photos/seed/mo2c/200/200'],
  'ct8': ['https://picsum.photos/seed/mo3a/200/200', 'https://picsum.photos/seed/mo3b/200/200', 'https://picsum.photos/seed/mo3c/200/200'],
  'ct16': ['https://picsum.photos/seed/mo4a/200/200', 'https://picsum.photos/seed/mo4b/200/200', 'https://picsum.photos/seed/mo4c/200/200'],
  'ct23': ['https://picsum.photos/seed/mo5a/200/200', 'https://picsum.photos/seed/mo5b/200/200', 'https://picsum.photos/seed/mo5c/200/200'],
  'ct20': ['https://picsum.photos/seed/mo6a/200/200', 'https://picsum.photos/seed/mo6b/200/200', 'https://picsum.photos/seed/mo6c/200/200'],
  'ct1': ['https://picsum.photos/seed/mo7a/200/200', 'https://picsum.photos/seed/mo7b/200/200', 'https://picsum.photos/seed/mo7c/200/200'],
  'ct18': ['https://picsum.photos/seed/mo8a/200/200', 'https://picsum.photos/seed/mo8b/200/200', 'https://picsum.photos/seed/mo8c/200/200'],
  'ct25': ['https://picsum.photos/seed/mo9a/200/200', 'https://picsum.photos/seed/mo9b/200/200', 'https://picsum.photos/seed/mo9c/200/200'],
  'ct19': ['https://picsum.photos/seed/mo10a/200/200', 'https://picsum.photos/seed/mo10b/200/200', 'https://picsum.photos/seed/mo10c/200/200'],
};

// Get unique letters from contacts
export function getAlphabetIndex(): string[] {
  const letters = new Set(contacts.map(c => c.letter));
  return Array.from(letters).sort();
}

/* ═══════════════════════════════════════
   Friend Requests
   ═══════════════════════════════════════ */

export type FriendRequestStatus = 'pending' | 'accepted' | 'rejected' | 'ignored';
export type FriendRequestClass = 'received' | 'sent';

export interface FriendRequest {
  id: string;
  class: FriendRequestClass;       // 'received' = 我收到的, 'sent' = 我发起的
  nickname: string;
  avatar: string;
  sex: 'male' | 'female' | 'unknown';
  region: string;
  occupation: string;
  introduction: string;
  tags: string[];
  reqMsg: string;                   // 申请附言
  handleMsg: string;                // 处理附言
  status: FriendRequestStatus;
  readState: boolean;               // false = unread
  reqTime: Date;
  hiChatId?: string;
  email?: string;
  phone?: string;
}

export const friendRequests: FriendRequest[] = [
  {
    id: 'fr1',
    class: 'received',
    nickname: '林小雨',
    avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=LXY',
    sex: 'female',
    region: '浙江杭州',
    occupation: 'UI设计师',
    introduction: '热爱设计，追求极致的视觉体验',
    tags: ['设计师', '杭州'],
    reqMsg: '你好！我在产品研发群里看到你的消息，对IM项目很感兴趣，可以加个好友交流一下吗？',
    handleMsg: '',
    status: 'pending',
    readState: false,
    reqTime: new Date(Date.now() - 15 * 60000),
    hiChatId: 'hichat_linyu',
    email: 'linyu@example.com',
    phone: '159****4478',
  },
  {
    id: 'fr2',
    class: 'received',
    nickname: '赵文博',
    avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=ZWB',
    sex: 'male',
    region: '北京',
    occupation: '全栈工程师',
    introduction: '代码即艺术',
    tags: ['开发者', '北京'],
    reqMsg: '你好，我是从技术交流群加你的，可以认识一下吗？',
    handleMsg: '',
    status: 'pending',
    readState: false,
    reqTime: new Date(Date.now() - 2 * 3600000),
    hiChatId: 'hichat_zhaowb',
    email: 'zhaowb@example.com',
  },
  {
    id: 'fr3',
    class: 'received',
    nickname: '孙悦',
    avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=SY',
    sex: 'female',
    region: '上海',
    occupation: '产品经理',
    introduction: '用户体验至上',
    tags: ['产品', '上海'],
    reqMsg: '嗨，上次技术沙龙上认识的，我是做产品策划的，想和你多聊聊IM领域的产品方向',
    handleMsg: '很高兴认识你，欢迎加入！',
    status: 'accepted',
    readState: true,
    reqTime: new Date(Date.now() - 1 * 86400000),
    hiChatId: 'hichat_sunyue',
    email: 'sunyue@example.com',
    phone: '138****7712',
  },
  {
    id: 'fr4',
    class: 'sent',
    nickname: '周鸿飞',
    avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=ZHF',
    sex: 'male',
    region: '广东深圳',
    occupation: '后端工程师',
    introduction: '架构设计爱好者',
    tags: ['后端', '深圳'],
    reqMsg: '你好，我在社区看到你分享的IM项目，做得非常棒！想向你请教一些技术问题',
    handleMsg: '',
    status: 'pending',
    readState: true,
    reqTime: new Date(Date.now() - 5 * 3600000),
    hiChatId: 'hichat_zhouhf',
  },
  {
    id: 'fr5',
    class: 'received',
    nickname: '王丽华',
    avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=WLH',
    sex: 'female',
    region: '四川成都',
    occupation: '前端工程师',
    introduction: 'Vue/React 双修，专注前端开发5年',
    tags: ['前端', '成都', 'Vue'],
    reqMsg: '你好，想加你好友一起学习前端技术',
    handleMsg: '',
    status: 'rejected',
    readState: true,
    reqTime: new Date(Date.now() - 3 * 86400000),
    hiChatId: 'hichat_wanglh',
  },
  {
    id: 'fr6',
    class: 'sent',
    nickname: '李明哲',
    avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=LMZ',
    sex: 'male',
    region: '江苏南京',
    occupation: '测试工程师',
    introduction: '质量是产品的生命线',
    tags: ['测试', '南京'],
    reqMsg: '同为互联网从业者，希望能多交流',
    handleMsg: '好的，欢迎！',
    status: 'accepted',
    readState: true,
    reqTime: new Date(Date.now() - 5 * 86400000),
    hiChatId: 'hichat_limz',
    phone: '135****8834',
  },
  {
    id: 'fr7',
    class: 'received',
    nickname: '陈雅文',
    avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=CYW',
    sex: 'female',
    region: '湖北武汉',
    occupation: '数据分析师',
    introduction: '数据驱动决策',
    tags: ['数据', '武汉'],
    reqMsg: '嗨，我们公司正在做一个IM相关的数据分析项目，想跟你聊聊',
    handleMsg: '',
    status: 'ignored',
    readState: true,
    reqTime: new Date(Date.now() - 7 * 86400000),
    hiChatId: 'hichat_chenyw',
  },
  {
    id: 'fr8',
    class: 'sent',
    nickname: '张鹏飞',
    avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=ZPF',
    sex: 'male',
    region: '福建厦门',
    occupation: '移动端开发',
    introduction: 'iOS / Android 双平台开发者',
    tags: ['移动端', 'iOS', '厦门'],
    reqMsg: '你的IM项目有没有考虑做移动端？我们可以合作',
    handleMsg: '',
    status: 'pending',
    readState: true,
    reqTime: new Date(Date.now() - 12 * 3600000),
    hiChatId: 'hichat_zhangpf',
  },
  {
    id: 'fr9',
    class: 'received',
    nickname: '刘思琪',
    avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=LSQ',
    sex: 'female',
    region: '广东广州',
    occupation: '运营专员',
    introduction: '让产品被更多人看到',
    tags: ['运营', '广州'],
    reqMsg: '你好呀，我是做互联网运营的，感觉你的IM项目很有潜力，想多了解了解',
    handleMsg: '',
    status: 'pending',
    readState: true,
    reqTime: new Date(Date.now() - 8 * 3600000),
    hiChatId: 'hichat_liusq',
    email: 'liusq@example.com',
    phone: '137****2256',
  },
  {
    id: 'fr10',
    class: 'sent',
    nickname: '黄志强',
    avatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=HZQ',
    sex: 'male',
    region: '陕西西安',
    occupation: '架构师',
    introduction: '15年互联网老兵',
    tags: ['架构', '西安'],
    reqMsg: '看了你的技术方案，想和你深入讨论一下系统架构',
    handleMsg: '期待交流',
    status: 'accepted',
    readState: true,
    reqTime: new Date(Date.now() - 10 * 86400000),
    hiChatId: 'hichat_huangzq',
    email: 'huangzq@example.com',
  },
];

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

// ── Groups ──
export const groups: GroupInfo[] = [
  { id: 'c2', name: '产品研发群', icon: 'https://api.dicebear.com/9.x/identicon/svg?seed=dev-team', isVerify: true, notification: '每周一上午10点站会，禁止发送与工作无关的内容。', createUid: 'ct3' },
  { id: 'c5', name: '周末约饭群', icon: 'https://api.dicebear.com/9.x/identicon/svg?seed=food-group', isVerify: false, notification: '每周六晚7点聚餐，地点群内投票决定。', createUid: 'ct8' },
  { id: 'c8', name: '技术交流群', icon: 'https://api.dicebear.com/9.x/identicon/svg?seed=tech-group', isVerify: true, notification: '技术讨论为主，禁止广告和灌水。', createUid: 'me' },
  { id: 'g_design', name: '设计灵感圈', icon: 'https://api.dicebear.com/9.x/identicon/svg?seed=design-circle', isVerify: false, notification: '', createUid: 'ct13' },
  { id: 'g_run', name: '周末跑步群', icon: 'https://api.dicebear.com/9.x/identicon/svg?seed=run-group', isVerify: false, notification: '每周日早上8点南山公园集合', createUid: 'ct6' },
];

// ── Group Members ──
export const groupMembers: Record<string, GroupMemberInfo[]> = {
  'c2': [
    { id: 1, groupId: 'c2', userId: 'ct3', nickname: '陈总', roleLevel: 2, online: false },
    { id: 2, groupId: 'c2', userId: 'me', nickname: '', roleLevel: 2, online: true },
    { id: 3, groupId: 'c2', userId: 'ct13', nickname: '', roleLevel: 2, online: true },
    { id: 4, groupId: 'c2', userId: 'ct22', nickname: '', roleLevel: 1, online: false },
    { id: 5, groupId: 'c2', userId: 'ct27', nickname: '', roleLevel: 1, online: true },
    { id: 6, groupId: 'c2', userId: 'ct28', nickname: '', roleLevel: 1, online: false },
    { id: 7, groupId: 'c2', userId: 'ct8', nickname: '', roleLevel: 1, online: true },
    { id: 8, groupId: 'c2', userId: 'ct6', nickname: '', roleLevel: 1, online: true },
    { id: 9, groupId: 'c2', userId: 'ct18', nickname: '', roleLevel: 1, online: true },
    { id: 10, groupId: 'c2', userId: 'ct20', nickname: '猴哥', roleLevel: 1, online: true },
    { id: 11, groupId: 'c2', userId: 'ct29', nickname: '', roleLevel: 1, online: true },
  ],
  'c5': [
    { id: 20, groupId: 'c5', userId: 'ct8', nickname: '', roleLevel: 2, online: true },
    { id: 21, groupId: 'c5', userId: 'me', nickname: '', roleLevel: 1, online: true },
    { id: 22, groupId: 'c5', userId: 'ct22', nickname: '', roleLevel: 1, online: false },
    { id: 23, groupId: 'c5', userId: 'ct13', nickname: '', roleLevel: 1, online: true },
    { id: 24, groupId: 'c5', userId: 'ct27', nickname: '', roleLevel: 1, online: true },
    { id: 25, groupId: 'c5', userId: 'ct6', nickname: '', roleLevel: 2, online: true },
    { id: 26, groupId: 'c5', userId: 'ct19', nickname: '', roleLevel: 1, online: true },
    { id: 27, groupId: 'c5', userId: 'ct20', nickname: '', roleLevel: 1, online: true },
  ],
  'c8': [
    { id: 30, groupId: 'c8', userId: 'me', nickname: '', roleLevel: 2, online: true },
    { id: 31, groupId: 'c8', userId: 'ct18', nickname: '', roleLevel: 2, online: true },
    { id: 32, groupId: 'c8', userId: 'ct20', nickname: '', roleLevel: 2, online: true },
    { id: 33, groupId: 'c8', userId: 'ct29', nickname: '', roleLevel: 1, online: true },
    { id: 34, groupId: 'c8', userId: 'ct13', nickname: '', roleLevel: 1, online: true },
    { id: 35, groupId: 'c8', userId: 'ct22', nickname: '', roleLevel: 1, online: false },
    { id: 36, groupId: 'c8', userId: 'ct3', nickname: '', roleLevel: 1, online: false },
    { id: 37, groupId: 'c8', userId: 'ct8', nickname: '', roleLevel: 1, online: true },
    { id: 38, groupId: 'c8', userId: 'ct6', nickname: '', roleLevel: 1, online: true },
    { id: 39, groupId: 'c8', userId: 'ct27', nickname: '', roleLevel: 1, online: true },
    { id: 40, groupId: 'c8', userId: 'ct16', nickname: '', roleLevel: 1, online: true },
    { id: 41, groupId: 'c8', userId: 'ct23', nickname: '', roleLevel: 1, online: true },
  ],
  'g_design': [
    { id: 50, groupId: 'g_design', userId: 'ct13', nickname: '', roleLevel: 2, online: true },
    { id: 51, groupId: 'g_design', userId: 'me', nickname: '', roleLevel: 2, online: true },
    { id: 52, groupId: 'g_design', userId: 'ct4', nickname: '', roleLevel: 1, online: true },
    { id: 53, groupId: 'g_design', userId: 'ct16', nickname: '', roleLevel: 1, online: true },
    { id: 54, groupId: 'g_design', userId: 'ct29', nickname: '', roleLevel: 1, online: true },
    { id: 55, groupId: 'g_design', userId: 'ct19', nickname: '', roleLevel: 1, online: true },
  ],
  'g_run': [
    { id: 60, groupId: 'g_run', userId: 'ct6', nickname: '', roleLevel: 2, online: true },
    { id: 61, groupId: 'g_run', userId: 'me', nickname: '', roleLevel: 1, online: true },
    { id: 62, groupId: 'g_run', userId: 'ct25', nickname: '', roleLevel: 1, online: true },
    { id: 63, groupId: 'g_run', userId: 'ct15', nickname: '', roleLevel: 2, online: false },
    { id: 64, groupId: 'g_run', userId: 'ct1', nickname: '', roleLevel: 1, online: true },
  ],
};

// ── Group Applications ──
export const groupApplications: GroupApplication[] = [
  {
    id: 301, userId: 'ct14', userName: '林黛玉', userAvatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=LDY',
    groupId: 'c8', groupName: '技术交流群', groupIcon: 'https://api.dicebear.com/9.x/identicon/svg?seed=tech-group',
    reqMsg: '对技术交流很感兴趣，希望能加入学习', reqTime: new Date(Date.now() - 30 * 60000),
    joinSource: 1, handleResult: 0, readState: false,
  },
  {
    id: 302, userId: 'ct5', userName: '大头', userAvatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=DT',
    groupId: 'c2', groupName: '产品研发群', groupIcon: 'https://api.dicebear.com/9.x/identicon/svg?seed=dev-team',
    reqMsg: '测试同学想了解产品开发流程', reqTime: new Date(Date.now() - 90 * 60000),
    joinSource: 2, inviterName: '李四', handleResult: 0, readState: false,
  },
  {
    id: 303, userId: 'ct21', userName: '唐僧', userAvatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=TS',
    groupId: 'c8', groupName: '技术交流群', groupIcon: 'https://api.dicebear.com/9.x/identicon/svg?seed=tech-group',
    reqMsg: '通过邀请链接申请加入', reqTime: new Date(Date.now() - 2 * 3600000),
    joinSource: 3, handleResult: 0, readState: false,
  },
  {
    id: 304, userId: 'ct9', userName: '黄小明', userAvatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=HXM',
    groupId: 'c5', groupName: '周末约饭群', groupIcon: 'https://api.dicebear.com/9.x/identicon/svg?seed=food-group',
    reqMsg: '喜欢美食，想加入约饭', reqTime: new Date(Date.now() - 1 * 86400000),
    joinSource: 1, handleResult: 1, readState: true,
  },
  {
    id: 305, userId: 'me', userName: '我', userAvatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=Felix',
    groupId: 'g_design', groupName: '设计灵感圈', groupIcon: 'https://api.dicebear.com/9.x/identicon/svg?seed=design-circle',
    reqMsg: '李四邀请我加入设计群', reqTime: new Date(Date.now() - 3 * 86400000),
    joinSource: 2, inviterName: '李四', handleResult: 1, readState: true,
  },
  {
    id: 306, userId: 'ct24', userName: '谢逊', userAvatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=XX',
    groupId: 'c2', groupName: '产品研发群', groupIcon: 'https://api.dicebear.com/9.x/identicon/svg?seed=dev-team',
    reqMsg: '想加入产品研发群学习', reqTime: new Date(Date.now() - 5 * 86400000),
    joinSource: 1, handleResult: 2, readState: true,
  },
  {
    id: 307, userId: 'me', userName: '我', userAvatar: 'https://api.dicebear.com/9.x/adventurer/svg?seed=Felix',
    groupId: 'g_run', groupName: '周末跑步群', groupIcon: 'https://api.dicebear.com/9.x/identicon/svg?seed=run-group',
    reqMsg: '想加入跑步群一起运动', reqTime: new Date(Date.now() - 4 * 86400000),
    joinSource: 1, handleResult: 1, readState: true,
  },
];

// ── Group Invite Links ──
export const groupInviteLinks: Record<string, GroupInviteLink[]> = {
  'c2': [
    { token: 'tk_a1b2c3d4', groupId: 'c2', createdBy: 'ct3', createdAt: new Date(Date.now() - 86400000), expireAt: new Date(Date.now() + 86400000), maxUses: 50, usedCount: 3, revoked: false },
    { token: 'tk_e5f6g7h8', groupId: 'c2', createdBy: 'ct13', createdAt: new Date(Date.now() - 172800000), expireAt: null, maxUses: 0, usedCount: 8, revoked: false },
  ],
  'c8': [
    { token: 'tk_i9j0k1l2', groupId: 'c8', createdBy: 'me', createdAt: new Date(Date.now() - 3600000), expireAt: new Date(Date.now() + 604800000), maxUses: 20, usedCount: 2, revoked: false },
  ],
  'g_design': [
    { token: 'tk_m3n4o5p6', groupId: 'g_design', createdBy: 'ct13', createdAt: new Date(Date.now() - 259200000), expireAt: null, maxUses: 0, usedCount: 5, revoked: true },
  ],
};

// ── Group Announcements ──
export const groupAnnouncements: Record<string, GroupAnnouncement[]> = {
  'c2': [
    { id: 401, groupId: 'c2', content: '新版本已部署到测试环境，请大家帮忙测试，有问题及时反馈。', createdBy: 'ct13', createdAt: new Date(Date.now() - 86400000), pinned: true },
    { id: 402, groupId: 'c2', content: '群规：1.禁止广告 2.禁止人身攻击 3.工作讨论为主', createdBy: 'ct3', createdAt: new Date(Date.now() - 604800000), pinned: true },
    { id: 403, groupId: 'c2', content: '恭喜产品v2.0顺利发布！感谢每一位成员的付出！', createdBy: 'ct3', createdAt: new Date(Date.now() - 172800000), pinned: false },
  ],
  'c8': [
    { id: 410, groupId: 'c8', content: '本月技术分享主题：微服务架构实践。欢迎大家准备分享内容。', createdBy: 'me', createdAt: new Date(Date.now() - 259200000), pinned: true },
  ],
  'g_run': [
    { id: 420, groupId: 'g_run', content: '本周日跑步路线：南山公园→深圳湾，全程约10公里。', createdBy: 'ct6', createdAt: new Date(Date.now() - 86400000), pinned: true },
  ],
};

// ── Group Member Settings (for current user) ──
export const groupMemberSettings: Record<string, GroupMemberSetting> = {
  'c2': { groupId: 'c2', groupNickname: '', groupRemark: '工作群' },
  'c5': { groupId: 'c5', groupNickname: '吃货', groupRemark: '约饭群' },
  'c8': { groupId: 'c8', groupNickname: '技术小能手', groupRemark: '' },
  'g_design': { groupId: 'g_design', groupNickname: '', groupRemark: '设计群' },
  'g_run': { groupId: 'g_run', groupNickname: '', groupRemark: '' },
};

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
  type: 'reply' | 'like';
  trendId: number;
  trendContent: string;       // preview text
  actor: { id: string; name: string; avatar: string };
  content?: string;           // comment text (for reply type)
  read: boolean;
  createTime: Date;
}

// Helper: get user info for trends
function trendUser(id: string): { id: string; name: string; avatar: string } {
  if (id === 'me') return { id: 'me', name: currentUser.name, avatar: currentUser.avatar };
  const c = contacts.find(ct => ct.id === id);
  return c ? { id: c.id, name: c.name, avatar: c.avatar } : { id, name: id, avatar: '' };
}

const NOW = Date.now();

// ── Trends ──
export const trends: Trend[] = [
  {
    id: 50010, userId: 'ct8', type: 2, content: '周末去了海边，日落太美了，分享几张照片给大家 🌅', scope: 3,
    createTime: new Date(NOW - 20 * 60000), replyCount: 3, agreeCount: 18, positionName: '厦门环岛路',
    title: '', atUsers: [], resources: ['https://picsum.photos/seed/t10a/600/400', 'https://picsum.photos/seed/t10b/600/400', 'https://picsum.photos/seed/t10c/600/400'],
    coverUrl: '', shareUrl: '', openReply: true, isTop: false,
  },
  {
    id: 50009, userId: 'ct22', type: 1, content: '终于把那个算法题解出来了，花了整整两天时间。关键在于发现单调栈的性质，把 O(n²) 优化到 O(n)。如果你也在刷题，可以一起交流。',
    scope: 3, createTime: new Date(NOW - 90 * 60000), replyCount: 5, agreeCount: 24, positionName: '',
    title: '', atUsers: [], resources: [], coverUrl: '', shareUrl: '', openReply: true, isTop: false,
  },
  {
    id: 50008, userId: 'me', type: 3, content: '最近在研究微前端架构，总结了一些实践经验，包括模块联邦的配置、样式隔离方案、以及性能优化的一些思考。',
    scope: 3, createTime: new Date(NOW - 5 * 3600000), replyCount: 8, agreeCount: 82, positionName: '',
    title: '微前端架构实践总结', atUsers: [trendUser('ct22')], resources: [],
    coverUrl: 'https://picsum.photos/seed/t08cover/800/300', shareUrl: '', openReply: true, isTop: true,
  },
  {
    id: 50007, userId: 'ct4', type: 4, content: '这篇关于产品思维的文章写得很好，推荐给做产品的朋友们看看',
    scope: 2, createTime: new Date(NOW - 12 * 3600000), replyCount: 2, agreeCount: 9, positionName: '',
    title: '', atUsers: [], resources: [], coverUrl: '', shareUrl: 'https://example.com/article/product-thinking', openReply: true, isTop: false,
  },
  {
    id: 50006, userId: 'ct20', type: 2, content: '用 Three.js 做了一个数据可视化大屏的效果，整体效果还不错',
    scope: 3, createTime: new Date(NOW - 24 * 3600000), replyCount: 4, agreeCount: 31, positionName: '',
    title: '', atUsers: [trendUser('me')], resources: ['https://picsum.photos/seed/t06a/600/400', 'https://picsum.photos/seed/t06b/600/400'],
    coverUrl: '', shareUrl: '', openReply: true, isTop: false,
  },
  {
    id: 50005, userId: 'ct6', type: 5, content: '今天跑步配速突破了5分钟，记录一下 🏃',
    scope: 3, createTime: new Date(NOW - 48 * 3600000), replyCount: 6, agreeCount: 15, positionName: '北京奥林匹克森林公园',
    title: '', atUsers: [], resources: ['https://picsum.photos/seed/t05v/800/450'],
    coverUrl: '', shareUrl: '', openReply: true, isTop: false,
  },
  {
    id: 50004, userId: 'ct8', type: 2, content: '新的 UI 设计稿完成了，这次走的是玻璃拟态风格，大家觉得怎么样？',
    scope: 2, createTime: new Date(NOW - 72 * 3600000), replyCount: 7, agreeCount: 22, positionName: '',
    title: '', atUsers: [], resources: ['https://picsum.photos/seed/t04a/600/400', 'https://picsum.photos/seed/t04b/600/400', 'https://picsum.photos/seed/t04c/600/400', 'https://picsum.photos/seed/t04d/600/400'],
    coverUrl: '', shareUrl: '', openReply: true, isTop: false,
  },
  {
    id: 50003, userId: 'me', type: 1, content: '读完了《重构：改善既有代码的设计》，收获很大。最核心的观点是：小步重构，每次改动后都要确保测试通过。',
    scope: 3, createTime: new Date(NOW - 96 * 3600000), replyCount: 3, agreeCount: 16, positionName: '',
    title: '', atUsers: [], resources: [], coverUrl: '', shareUrl: '', openReply: false, isTop: false,
  },
];

// ── Comments (keyed by trend id) ──
export const trendComments: Record<number, TrendComment[]> = {
  50010: [
    { id: 60010, trendId: 50010, rootId: 0, father: 0, replyer: trendUser('me'), user: trendUser('ct8'), level: 1, content: '太美了！厦门是个好地方', agreeCount: 3, createTime: new Date(NOW - 15 * 60000), children: [
      { id: 60011, trendId: 50010, rootId: 60010, father: 60010, replyer: trendUser('ct8'), user: trendUser('me'), level: 2, content: '下次一起去呀', agreeCount: 1, createTime: new Date(NOW - 10 * 60000), children: [] },
    ]},
    { id: 60012, trendId: 50010, rootId: 0, father: 0, replyer: trendUser('ct4'), user: trendUser('ct8'), level: 1, content: '第三张构图绝了', agreeCount: 2, createTime: new Date(NOW - 5 * 60000), children: [] },
  ],
  50009: [
    { id: 60020, trendId: 50009, rootId: 0, father: 0, replyer: trendUser('me'), user: trendUser('ct22'), level: 1, content: '哪道题？我也在刷', agreeCount: 5, createTime: new Date(NOW - 67 * 60000), children: [
      { id: 60021, trendId: 50009, rootId: 60020, father: 60020, replyer: trendUser('ct22'), user: trendUser('me'), level: 2, content: 'LeetCode 84 柱状图中最大矩形', agreeCount: 2, createTime: new Date(NOW - 58 * 60000), children: [
        { id: 60022, trendId: 50009, rootId: 60020, father: 60021, replyer: trendUser('me'), user: trendUser('ct22'), level: 3, content: '这道经典，我用栈做的', agreeCount: 0, createTime: new Date(NOW - 50 * 60000), children: [] },
      ]},
    ]},
    { id: 60023, trendId: 50009, rootId: 0, father: 0, replyer: trendUser('ct20'), user: trendUser('ct22'), level: 1, content: '单调栈确实是利器', agreeCount: 1, createTime: new Date(NOW - 30 * 60000), children: [] },
  ],
  50008: [
    { id: 60030, trendId: 50008, rootId: 0, father: 0, replyer: trendUser('ct22'), user: trendUser('me'), level: 1, content: '写得很好，正好在调研微前端方案', agreeCount: 8, createTime: new Date(NOW - 4 * 3600000), children: [
      { id: 60031, trendId: 50008, rootId: 60030, father: 60030, replyer: trendUser('me'), user: trendUser('ct22'), level: 2, content: 'qiankun 和 module-federation 对比一下？', agreeCount: 3, createTime: new Date(NOW - 3.5 * 3600000), children: [] },
    ]},
  ],
};

// ── Like users per trend ──
export const trendLikeUsers: Record<number, { id: string; name: string; avatar: string }[]> = {
  50010: [trendUser('me'), trendUser('ct4'), trendUser('ct29')],
  50009: [trendUser('me'), trendUser('ct8'), trendUser('ct22'), trendUser('ct20'), trendUser('ct29')],
  50008: [trendUser('ct8'), trendUser('ct22'), trendUser('ct4'), trendUser('ct20'), trendUser('ct29'),
    trendUser('ct1'), trendUser('ct2'), trendUser('ct3'), trendUser('ct5'), trendUser('ct6'), trendUser('ct7'), trendUser('ct9'),
    { id: 'like_extra_1', name: '赵文博', avatar: '' }, { id: 'like_extra_2', name: '孙悦', avatar: '' },
    { id: 'like_extra_3', name: '周鸿飞', avatar: '' }, { id: 'like_extra_4', name: '王丽华', avatar: '' },
    { id: 'like_extra_5', name: '李明哲', avatar: '' }, { id: 'like_extra_6', name: '陈雅文', avatar: '' },
    { id: 'like_extra_7', name: '张鹏飞', avatar: '' }, { id: 'like_extra_8', name: '刘思琪', avatar: '' },
    { id: 'like_extra_9', name: '黄志强', avatar: '' }, { id: 'like_extra_10', name: '林小雨', avatar: '' },
    { id: 'like_extra_11', name: '郑天宇', avatar: '' }, { id: 'like_extra_12', name: '吴佳琪', avatar: '' },
    { id: 'like_extra_13', name: '刘浩然', avatar: '' }, { id: 'like_extra_14', name: '杨雨欣', avatar: '' },
    { id: 'like_extra_15', name: '陈志远', avatar: '' }, { id: 'like_extra_16', name: '张雪', avatar: '' },
    { id: 'like_extra_17', name: '王鹏', avatar: '' }, { id: 'like_extra_18', name: '李晓峰', avatar: '' },
    { id: 'like_extra_19', name: '赵琳', avatar: '' }, { id: 'like_extra_20', name: '黄伟', avatar: '' },
    { id: 'like_extra_21', name: '周敏', avatar: '' }, { id: 'like_extra_22', name: '吴强', avatar: '' },
    { id: 'like_extra_23', name: '郑洋', avatar: '' }, { id: 'like_extra_24', name: '孙磊', avatar: '' },
    { id: 'like_extra_25', name: '马丽', avatar: '' }, { id: 'like_extra_26', name: '胡建国', avatar: '' },
    { id: 'like_extra_27', name: '朱明', avatar: '' }, { id: 'like_extra_28', name: '林涛', avatar: '' },
    { id: 'like_extra_29', name: '何静', avatar: '' }, { id: 'like_extra_30', name: '高翔', avatar: '' },
    { id: 'like_extra_31', name: '罗军', avatar: '' }, { id: 'like_extra_32', name: '谢敏', avatar: '' },
    { id: 'like_extra_33', name: '唐磊', avatar: '' }, { id: 'like_extra_34', name: '韩冰', avatar: '' },
    { id: 'like_extra_35', name: '曹颖', avatar: '' }, { id: 'like_extra_36', name: '许峰', avatar: '' },
    { id: 'like_extra_37', name: '邓辉', avatar: '' }, { id: 'like_extra_38', name: '冯雅', avatar: '' },
    { id: 'like_extra_39', name: '程刚', avatar: '' }, { id: 'like_extra_40', name: '蔡婷', avatar: '' },
    { id: 'like_extra_41', name: '彭勇', avatar: '' }, { id: 'like_extra_42', name: '潘亮', avatar: '' },
    { id: 'like_extra_43', name: '田静', avatar: '' }, { id: 'like_extra_44', name: '董洁', avatar: '' },
    { id: 'like_extra_45', name: '袁浩', avatar: '' }, { id: 'like_extra_46', name: '蒋丽', avatar: '' },
    { id: 'like_extra_47', name: '余强', avatar: '' }, { id: 'like_extra_48', name: '杜明', avatar: '' },
    { id: 'like_extra_49', name: '叶翔', avatar: '' }, { id: 'like_extra_50', name: '苏敏', avatar: '' },
    { id: 'like_extra_51', name: '任杰', avatar: '' }, { id: 'like_extra_52', name: '卢涛', avatar: '' },
    { id: 'like_extra_53', name: '沈婷', avatar: '' }, { id: 'like_extra_54', name: '丁勇', avatar: '' },
    { id: 'like_extra_55', name: '傅洁', avatar: '' }, { id: 'like_extra_56', name: '钟浩', avatar: '' },
    { id: 'like_extra_57', name: '姜丽', avatar: '' }, { id: 'like_extra_58', name: '崔强', avatar: '' },
    { id: 'like_extra_59', name: '谭明', avatar: '' }, { id: 'like_extra_60', name: '汪翔', avatar: '' },
    { id: 'like_extra_61', name: '范敏', avatar: '' }, { id: 'like_extra_62', name: '廖涛', avatar: '' },
    { id: 'like_extra_63', name: '石婷', avatar: '' }, { id: 'like_extra_64', name: '金浩', avatar: '' },
    { id: 'like_extra_65', name: '熊杰', avatar: '' }, { id: 'like_extra_66', name: '秦丽', avatar: '' },
    { id: 'like_extra_67', name: '陆明', avatar: '' }, { id: 'like_extra_68', name: '段翔', avatar: '' },
    { id: 'like_extra_69', name: '雷婷', avatar: '' }, { id: 'like_extra_70', name: '侯洁', avatar: '' },
  ],
  50007: [trendUser('me'), trendUser('ct8')],
  50006: [trendUser('me'), trendUser('ct8'), trendUser('ct22')],
  50005: [trendUser('me'), trendUser('ct8')],
  50004: [trendUser('me'), trendUser('ct22'), trendUser('ct20')],
  50003: [trendUser('ct8'), trendUser('ct22'), trendUser('ct4')],
};

// Initially liked by current user
export const initialLikedTrends: Set<number> = new Set([50008, 50006]);

// ── Notifications ──
export const momentsNotifications: MomentsNotification[] = [
  { id: 70001, type: 'reply', trendId: 50009, trendContent: '终于把那个算法题解出来了', actor: trendUser('me'), content: '哪道题？我也在刷', read: false, createTime: new Date(NOW - 67 * 60000) },
  { id: 70002, type: 'reply', trendId: 50008, trendContent: '最近在研究微前端架构', actor: trendUser('ct22'), content: '写得很好，正好在调研微前端方案', read: false, createTime: new Date(NOW - 4 * 3600000) },
  { id: 70003, type: 'reply', trendId: 50008, trendContent: '最近在研究微前端架构', actor: trendUser('me'), content: 'qiankun 和 module-federation 对比一下？', read: false, createTime: new Date(NOW - 3.5 * 3600000) },
  { id: 80001, type: 'like', trendId: 50008, trendContent: '最近在研究微前端架构', actor: trendUser('ct8'), read: true, createTime: new Date(NOW - 3.6 * 3600000) },
  { id: 80002, type: 'like', trendId: 50008, trendContent: '最近在研究微前端架构', actor: trendUser('ct4'), read: true, createTime: new Date(NOW - 3 * 3600000) },
  { id: 80003, type: 'like', trendId: 50006, trendContent: '用 Three.js 做了一个数据可视化', actor: trendUser('ct22'), read: true, createTime: new Date(NOW - 22 * 3600000) },
];
