const dict: Record<string, Record<string, string>> = {
  'zh-CN': {
    // Tabs
    'tab.chats': 'HiChat',
    'tab.contacts': '通讯录',
    'tab.moments': '发现',
    'tab.me': '我',

    // Profile page
    'profile.service': '服务',
    'profile.favorites': '收藏',
    'profile.album': '相册',
    'profile.cards': '卡包',
    'profile.emojis': '表情',
    'profile.settings': '设置',
    'profile.backup': '聊天记录备份与迁移',
    'profile.help': '帮助与反馈',
    'profile.about': '关于',
    'profile.plugins': '插件',
    'profile.switchAccount': '切换账号',
    'profile.logout': '退出登录',

    // Settings
    'settings.title': '设置',
    'settings.accountSecurity': '账号与安全',
    'settings.notification': '消息通知',
    'settings.privacy': '隐私',
    'settings.darkMode': '深色模式',
    'settings.general': '通用',
    'settings.phone': '手机号',
    'settings.email': '邮箱',
    'settings.changePassword': '修改密码',
    'settings.deleteAccount': '注销账号',
    'settings.notBound': '未绑定',

    // Dark mode
    'darkMode.system': '跟随系统',
    'darkMode.light': '始终浅色',
    'darkMode.dark': '始终深色',

    // Notifications
    'notify.enable': '接收新消息通知',
    'notify.sound': '声音',
    'notify.soundDesc': '收到消息时播放提示音',
    'notify.vibrate': '振动',
    'notify.vibrateDesc': '收到消息时振动',
    'notify.preview': '消息预览',
    'notify.previewDesc': '通知中显示消息内容',

    // Privacy
    'privacy.addMethods': '添加我的方式',
    'privacy.byPhone': '通过手机号找到我',
    'privacy.byId': '通过 HiChat ID 找到我',
    'privacy.moments': '朋友圈',
    'privacy.allVisible': '所有人可见',
    'privacy.friendsOnly': '仅好友可见',
    'privacy.privateOnly': '仅自己可见',
    'privacy.readReceipt': '消息已读回执',
    'privacy.readReceiptDesc': '对方能看到你的消息何时被你阅读；关闭后你也看不到对方的已读状态',

    // General
    'general.language': '语言',
    'general.fontSize': '字体大小',
    'general.fontSmall': '小',
    'general.fontMedium': '标准',
    'general.fontLarge': '大',
    'general.clearCache': '清除缓存',

    // Common
    'common.save': '保存',
    'common.cancel': '取消',
    'common.confirm': '确定',
    'common.back': '返回',
    'common.search': '搜索',
    'common.loading': '加载中...',
    'common.noData': '暂无数据',
    'common.loadMore': '加载更多',

    // Favorites
    'fav.title': '我的收藏',
    'fav.add': '添加收藏',
    'fav.empty': '暂无收藏',
    'fav.emptyHint': '点击右上角 + 添加你的第一个收藏',

    // Album
    'album.title': '我的相册',
    'album.empty': '暂无照片和视频',
    'album.emptyHint': '发布动态时添加的图片和视频会出现在这里',

    // Emojis
    'emoji.title': '我的表情',
    'emoji.empty': '暂无自定义表情',
    'emoji.emptyHint': '点击右上角 + 上传你的表情包',
    'emoji.upload': '上传表情',
    'emoji.manage': '管理',
    'emoji.done': '完成',

    // Trend / Moments message center
    'trend.notify.title': '通知',
    'trend.notify.markAllRead': '全部已读',
    'trend.notify.tab.all': '全部',
    'trend.notify.tab.comment': '评论',
    'trend.notify.tab.like': '点赞',
    'trend.notify.empty.all': '暂无通知',
    'trend.notify.empty.comment': '暂无评论通知',
    'trend.notify.empty.like': '暂无点赞通知',
    'trend.notify.emptyHint': '这里空空如也',
    'trend.notify.allReadDone': '已全部标记为已读',
    'trend.notify.markFailed': '标记失败',
    'trend.notify.act.like': '赞了你的动态',
    'trend.notify.act.comment': '评论了你的动态',
    'trend.notify.act.reply': '回复了你的评论',
    'trend.notify.act.atTrend': '在动态中提到了你',
    'trend.notify.act.atComment': '在评论中提到了你',
  },

  en: {
    'tab.chats': 'HiChat',
    'tab.contacts': 'Contacts',
    'tab.moments': 'Discover',
    'tab.me': 'Me',

    'profile.service': 'Services',
    'profile.favorites': 'Favorites',
    'profile.album': 'Album',
    'profile.cards': 'Cards',
    'profile.emojis': 'Stickers',
    'profile.settings': 'Settings',
    'profile.backup': 'Backup & Migration',
    'profile.help': 'Help & Feedback',
    'profile.about': 'About',
    'profile.plugins': 'Plugins',
    'profile.switchAccount': 'Switch Account',
    'profile.logout': 'Log Out',

    'settings.title': 'Settings',
    'settings.accountSecurity': 'Account & Security',
    'settings.notification': 'Notifications',
    'settings.privacy': 'Privacy',
    'settings.darkMode': 'Dark Mode',
    'settings.general': 'General',
    'settings.phone': 'Phone',
    'settings.email': 'Email',
    'settings.changePassword': 'Change Password',
    'settings.deleteAccount': 'Delete Account',
    'settings.notBound': 'Not bound',

    'darkMode.system': 'Follow System',
    'darkMode.light': 'Always Light',
    'darkMode.dark': 'Always Dark',

    'notify.enable': 'Enable Notifications',
    'notify.sound': 'Sound',
    'notify.soundDesc': 'Play sound for new messages',
    'notify.vibrate': 'Vibrate',
    'notify.vibrateDesc': 'Vibrate for new messages',
    'notify.preview': 'Preview',
    'notify.previewDesc': 'Show message content in notifications',

    'privacy.addMethods': 'Find Me By',
    'privacy.byPhone': 'Phone Number',
    'privacy.byId': 'HiChat ID',
    'privacy.moments': 'Moments',
    'privacy.allVisible': 'Everyone',
    'privacy.friendsOnly': 'Friends Only',
    'privacy.privateOnly': 'Only Me',
    'privacy.readReceipt': 'Read Receipts',
    'privacy.readReceiptDesc': 'Lets others see when you have read their messages; disabling also hides their read status from you',

    'general.language': 'Language',
    'general.fontSize': 'Font Size',
    'general.fontSmall': 'Small',
    'general.fontMedium': 'Standard',
    'general.fontLarge': 'Large',
    'general.clearCache': 'Clear Cache',

    'common.save': 'Save',
    'common.cancel': 'Cancel',
    'common.confirm': 'OK',
    'common.back': 'Back',
    'common.search': 'Search',
    'common.loading': 'Loading...',
    'common.noData': 'No data',
    'common.loadMore': 'Load More',

    'fav.title': 'Favorites',
    'fav.add': 'Add Favorite',
    'fav.empty': 'No favorites yet',
    'fav.emptyHint': 'Tap + to add your first favorite',

    'album.title': 'Album',
    'album.empty': 'No photos or videos',
    'album.emptyHint': 'Photos and videos from your moments will appear here',

    'emoji.title': 'Stickers',
    'emoji.empty': 'No custom stickers',
    'emoji.emptyHint': 'Tap + to upload your stickers',
    'emoji.upload': 'Upload',
    'emoji.manage': 'Manage',
    'emoji.done': 'Done',

    // Trend / Moments message center
    'trend.notify.title': 'Notifications',
    'trend.notify.markAllRead': 'Mark all read',
    'trend.notify.tab.all': 'All',
    'trend.notify.tab.comment': 'Comments',
    'trend.notify.tab.like': 'Likes',
    'trend.notify.empty.all': 'No notifications',
    'trend.notify.empty.comment': 'No comment notifications',
    'trend.notify.empty.like': 'No like notifications',
    'trend.notify.emptyHint': 'Nothing here yet',
    'trend.notify.allReadDone': 'All marked as read',
    'trend.notify.markFailed': 'Failed to mark as read',
    'trend.notify.act.like': 'liked your moment',
    'trend.notify.act.comment': 'commented on your moment',
    'trend.notify.act.reply': 'replied to your comment',
    'trend.notify.act.atTrend': 'mentioned you in a moment',
    'trend.notify.act.atComment': 'mentioned you in a comment',
  },
};

export function t(key: string, lang: string = 'zh-CN'): string {
  return dict[lang]?.[key] || dict['zh-CN']?.[key] || key;
}
