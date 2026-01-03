import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useConversationStore = defineStore('conversation', () => {
  // 状态数据
  const conversations = ref([])
  const activeConversation = ref(null)
  const messages = ref({}) // 按会话ID存储消息
  const loadingMessages = ref({}) // 按会话ID存储加载状态
  
  // 计算属性
  const sortedConversations = computed(() => {
    return [...conversations.value].sort((a, b) => {
      // 置顶的会话排在前面
      if (a.isPinned && !b.isPinned) return -1
      if (!a.isPinned && b.isPinned) return 1
      
      // 按最后消息时间排序
      return new Date(b.lastMessageTime) - new Date(a.lastMessageTime)
    })
  })
  
  // 初始化会话数据
  function initializeConversations() {
    conversations.value = [
      {
        id: 1,
        name: '技术讨论群',
        avatar: 'https://api.dicebear.com/7.x/initials/svg?seed=Tech&backgroundType=gradientLinear&backgroundColor=4f46e5,7c3aed',
        lastMessage: 'Go的goroutine真是强大...',
        lastMessageType: 'text',
        lastMessageTime: new Date().toISOString(),
        unreadCount: 3,
        isOnline: true,
        isPinned: true,
        isMuted: false,
        type: 'group'
      },
      {
        id: 2,
        name: '小李',
        avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Lily&backgroundColor=f59e0b',
        lastMessage: '今晚一起吃饭吗？',
        lastMessageType: 'text',
        lastMessageTime: new Date(Date.now() - 1000 * 60 * 30).toISOString(), // 30分钟前
        unreadCount: 1,
        isOnline: true,
        isPinned: false,
        isMuted: false,
        type: 'private'
      },
      {
        id: 3,
        name: '公司通知群',
        avatar: 'https://api.dicebear.com/7.x/initials/svg?seed=Notice&backgroundType=gradientLinear&backgroundColor=ef4444,dc2626',
        lastMessage: '下周五下午3点有全体会议',
        lastMessageType: 'text',
        lastMessageTime: new Date(Date.now() - 1000 * 60 * 60 * 2).toISOString(), // 2小时前
        unreadCount: 0,
        isOnline: false,
        isPinned: false,
        isMuted: true,
        type: 'group'
      },
      {
        id: 4,
        name: '项目小组',
        avatar: 'https://api.dicebear.com/7.x/initials/svg?seed=Project&backgroundType=gradientLinear&backgroundColor=10b981,059669',
        lastMessage: '[图片]',
        lastMessageType: 'image',
        lastMessageTime: new Date(Date.now() - 1000 * 60 * 60 * 5).toISOString(), // 5小时前
        unreadCount: 7,
        isOnline: true,
        isPinned: false,
        isMuted: false,
        type: 'group'
      },
      {
        id: 5,
        name: '小王',
        avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Wang&backgroundColor=8b5cf6',
        lastMessage: '[语音]',
        lastMessageType: 'voice',
        lastMessageTime: new Date(Date.now() - 1000 * 60 * 60 * 8).toISOString(), // 8小时前
        unreadCount: 0,
        isOnline: false,
        isPinned: false,
        isMuted: false,
        type: 'private'
      },
      {
        id: 6,
        name: '前端交流群',
        avatar: 'https://api.dicebear.com/7.x/initials/svg?seed=Frontend&backgroundType=gradientLinear&backgroundColor=3b82f6,1d4ed8',
        lastMessage: 'Vue3的Composition API真的太好用了',
        lastMessageType: 'text',
        lastMessageTime: new Date(Date.now() - 1000 * 60 * 60 * 12).toISOString(), // 12小时前
        unreadCount: 5,
        isOnline: true,
        isPinned: false,
        isMuted: false,
        type: 'group'
      },
      {
        id: 7,
        name: '小陈',
        avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Chen&backgroundColor=06b6d4',
        lastMessage: '好的，我知道了',
        lastMessageType: 'text',
        lastMessageTime: new Date(Date.now() - 1000 * 60 * 60 * 18).toISOString(), // 18小时前
        unreadCount: 0,
        isOnline: true,
        isPinned: false,
        isMuted: false,
        type: 'private'
      },
      {
        id: 8,
        name: '产品讨论群',
        avatar: 'https://api.dicebear.com/7.x/initials/svg?seed=Product&backgroundType=gradientLinear&backgroundColor=f97316,ea580c',
        lastMessage: '新功能的原型设计已完成',
        lastMessageType: 'text',
        lastMessageTime: new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString(), // 1天前
        unreadCount: 12,
        isOnline: false,
        isPinned: true,
        isMuted: false,
        type: 'group'
      },
      {
        id: 9,
        name: '小赵',
        avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Zhao&backgroundColor=84cc16',
        lastMessage: '[文件] 需求文档.pdf',
        lastMessageType: 'file',
        lastMessageTime: new Date(Date.now() - 1000 * 60 * 60 * 30).toISOString(), // 1.25天前
        unreadCount: 2,
        isOnline: false,
        isPinned: false,
        isMuted: false,
        type: 'private'
      },
      {
        id: 10,
        name: '设计师群组',
        avatar: 'https://api.dicebear.com/7.x/initials/svg?seed=Design&backgroundType=gradientLinear&backgroundColor=ec4899,db2777',
        lastMessage: '新版UI设计稿请大家review',
        lastMessageType: 'text',
        lastMessageTime: new Date(Date.now() - 1000 * 60 * 60 * 36).toISOString(), // 1.5天前
        unreadCount: 0,
        isOnline: true,
        isPinned: false,
        isMuted: true,
        type: 'group'
      },
      {
        id: 11,
        name: '小刘',
        avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Liu&backgroundColor=f59e0b',
        lastMessage: '周末有空一起打球吗？',
        lastMessageType: 'text',
        lastMessageTime: new Date(Date.now() - 1000 * 60 * 60 * 48).toISOString(), // 2天前
        unreadCount: 1,
        isOnline: true,
        isPinned: false,
        isMuted: false,
        type: 'private'
      },
      {
        id: 12,
        name: '测试团队',
        avatar: 'https://api.dicebear.com/7.x/initials/svg?seed=QA&backgroundType=gradientLinear&backgroundColor=6366f1,4f46e5',
        lastMessage: '发现一个严重bug，需要紧急修复',
        lastMessageType: 'text',
        lastMessageTime: new Date(Date.now() - 1000 * 60 * 60 * 60).toISOString(), // 2.5天前
        unreadCount: 8,
        isOnline: false,
        isPinned: false,
        isMuted: false,
        type: 'group'
      },
      {
        id: 13,
        name: '小孙',
        avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Sun&backgroundColor=14b8a6',
        lastMessage: '[表情] 😄😄😄',
        lastMessageType: 'emoji',
        lastMessageTime: new Date(Date.now() - 1000 * 60 * 60 * 72).toISOString(), // 3天前
        unreadCount: 0,
        isOnline: false,
        isPinned: false,
        isMuted: false,
        type: 'private'
      },
      {
        id: 14,
        name: '运营推广群',
        avatar: 'https://api.dicebear.com/7.x/initials/svg?seed=Marketing&backgroundType=gradientLinear&backgroundColor=f43f5e,e11d48',
        lastMessage: '这个月的数据报告出来了',
        lastMessageType: 'text',
        lastMessageTime: new Date(Date.now() - 1000 * 60 * 60 * 96).toISOString(), // 4天前
        unreadCount: 15,
        isOnline: true,
        isPinned: false,
        isMuted: true,
        type: 'group'
      },
      {
        id: 15,
        name: '小周',
        avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Zhou&backgroundColor=a855f7',
        lastMessage: '明天的会议推迟到下周',
        lastMessageType: 'text',
        lastMessageTime: new Date(Date.now() - 1000 * 60 * 60 * 120).toISOString(), // 5天前
        unreadCount: 0,
        isOnline: true,
        isPinned: false,
        isMuted: false,
        type: 'private'
      }
    ]
    
    // 初始化消息数据
    initializeMessages()
  }
  
  // 初始化消息数据
  function initializeMessages() {
    // 技术讨论群的消息
    messages.value[1] = [
      {
        id: 1,
        content: '大家好，今天想和大家讨论一下Go语言的并发编程',
        type: 'text',
        sender: 'other',
        senderName: '张三',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=zhangsan',
        timestamp: new Date(Date.now() - 1000 * 60 * 10).toISOString(),
        status: 'sent'
      },
      {
        id: 2,
        content: 'Go的goroutine确实很强大，比传统的线程模型要轻量很多',
        type: 'text',
        sender: 'other',
        senderName: '李四',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=lisi',
        timestamp: new Date(Date.now() - 1000 * 60 * 8).toISOString(),
        status: 'sent'
      },
      {
        id: 3,
        content: '是的，而且channel的设计也很优雅，实现了CSP模型',
        type: 'text',
        sender: 'me',
        senderName: '我',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=me',
        timestamp: new Date(Date.now() - 1000 * 60 * 5).toISOString(),
        status: 'sent'
      },
      {
        id: 4,
        content: 'https://example.com/golang-concurrency.png',
        type: 'image',
        sender: 'other',
        senderName: '王五',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=wangwu',
        timestamp: new Date(Date.now() - 1000 * 60 * 3).toISOString(),
        status: 'sent'
      },
      {
        id: 5,
        content: 'Go的goroutine真是强大...',
        type: 'text',
        sender: 'other',
        senderName: '赵六',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=zhaoliu',
        timestamp: new Date().toISOString(),
        status: 'sent'
      }
    ]
    
    // 小李的私聊消息
    messages.value[2] = [
      {
        id: 1,
        content: '你好！最近项目进展怎么样？',
        type: 'text',
        sender: 'other',
        senderName: '小李',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Lily&backgroundColor=f59e0b',
        timestamp: new Date(Date.now() - 1000 * 60 * 60).toISOString(),
        status: 'sent'
      },
      {
        id: 2,
        content: '还不错，后端部分基本完成了，现在在开发前端部分',
        type: 'text',
        sender: 'me',
        senderName: '我',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=me',
        timestamp: new Date(Date.now() - 1000 * 60 * 50).toISOString(),
        status: 'sent'
      },
      {
        id: 3,
        content: '前端用的什么技术栈？Vue还是React？',
        type: 'text',
        sender: 'other',
        senderName: '小李',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Lily&backgroundColor=f59e0b',
        timestamp: new Date(Date.now() - 1000 * 60 * 40).toISOString(),
        status: 'sent'
      },
      {
        id: 4,
        content: '用的是Vue3，配合Pinia做状态管理',
        type: 'text',
        sender: 'me',
        senderName: '我',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=me',
        timestamp: new Date(Date.now() - 1000 * 60 * 35).toISOString(),
        status: 'sent'
      },
      {
        id: 5,
        content: 'blob:demo-file-url',
        type: 'file',
        fileName: '项目技术文档.pdf',
        fileSize: '2.5MB',
        fileExtension: 'pdf',
        mimeType: 'application/pdf',
        sender: 'other',
        senderName: '小李',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Lily&backgroundColor=f59e0b',
        timestamp: new Date(Date.now() - 1000 * 60 * 32).toISOString(),
        status: 'sent'
      },
      {
        id: 6,
        content: '今晚一起吃饭吗？',
        type: 'text',
        sender: 'other',
        senderName: '小李',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Lily&backgroundColor=f59e0b',
        timestamp: new Date(Date.now() - 1000 * 60 * 30).toISOString(),
        status: 'sent'
      }
    ]
    
    // 前端交流群的消息
    messages.value[6] = [
      {
        id: 1,
        content: 'Vue3的Composition API真的太好用了',
        type: 'text',
        sender: 'other',
        senderName: '前端开发者A',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=dev1',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 12).toISOString(),
        status: 'sent'
      },
      {
        id: 2,
        content: '是的，写起来比Vue2清晰多了',
        type: 'text',
        sender: 'other',
        senderName: '前端开发者B',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=dev2',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 12 + 300).toISOString(),
        status: 'sent'
      },
      {
        id: 3,
        content: '我也觉得，特别是在处理复杂状态时',
        type: 'text',
        sender: 'me',
        senderName: '我',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=me',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 12 + 600).toISOString(),
        status: 'sent'
      },
      {
        id: 4,
        content: 'TypeScript支持也更好了',
        type: 'text',
        sender: 'other',
        senderName: '前端开发者C',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=dev3',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 12 + 900).toISOString(),
        status: 'sent'
      },
      {
        id: 5,
        content: '有没有人试过新的<script setup>语法？',
        type: 'text',
        sender: 'other',
        senderName: '前端开发者D',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=dev4',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 12 + 1200).toISOString(),
        status: 'sent'
      }
    ]
    
    // 产品讨论群的消息
    messages.value[8] = [
      {
        id: 1,
        content: '新功能的原型设计已完成',
        type: 'text',
        sender: 'other',
        senderName: '产品经理',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=pm1',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString(),
        status: 'sent'
      },
      {
        id: 2,
        content: '大家有时间review一下吗？',
        type: 'text',
        sender: 'other',
        senderName: '产品经理',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=pm1',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24 + 300).toISOString(),
        status: 'sent'
      },
      {
        id: 3,
        content: '链接发一下',
        type: 'text',
        sender: 'other',
        senderName: '设计师A',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=designer1',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24 + 600).toISOString(),
        status: 'sent'
      },
      {
        id: 4,
        content: 'https://www.figma.com/prototype/abc123',
        type: 'text',
        sender: 'other',
        senderName: '产品经理',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=pm1',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24 + 900).toISOString(),
        status: 'sent'
      },
      {
        id: 5,
        content: '看起来不错，用户流程很清晰',
        type: 'text',
        sender: 'me',
        senderName: '我',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=me',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24 + 1200).toISOString(),
        status: 'sent'
      },
      {
        id: 6,
        content: '技术实现上有什么难点吗？',
        type: 'text',
        sender: 'other',
        senderName: '产品经理',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=pm1',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24 + 1500).toISOString(),
        status: 'sent'
      },
      {
        id: 7,
        content: '主要是数据同步的部分，需要仔细设计',
        type: 'text',
        sender: 'other',
        senderName: '后端开发者',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=backend1',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24 + 1800).toISOString(),
        status: 'sent'
      },
      {
        id: 8,
        content: '前端这边需要重构几个组件',
        type: 'text',
        sender: 'other',
        senderName: '前端负责人',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=frontend1',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24 + 2100).toISOString(),
        status: 'sent'
      },
      {
        id: 9,
        content: '预计开发周期多久？',
        type: 'text',
        sender: 'other',
        senderName: '项目经理',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=pm2',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24 + 2400).toISOString(),
        status: 'sent'
      },
      {
        id: 10,
        content: '大概2-3个sprint',
        type: 'text',
        sender: 'me',
        senderName: '我',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=me',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24 + 2700).toISOString(),
        status: 'sent'
      },
      {
        id: 11,
        content: '好的，那我们下周开始立项',
        type: 'text',
        sender: 'other',
        senderName: '项目经理',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=pm2',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24 + 3000).toISOString(),
        status: 'sent'
      },
      {
        id: 12,
        content: '收到，我会准备详细的技术方案',
        type: 'text',
        sender: 'other',
        senderName: '技术负责人',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=tech1',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24 + 3300).toISOString(),
        status: 'sent'
      },
      {
        id: 13,
        content: 'blob:demo-file-url',
        type: 'file',
        fileName: '需求分析报告.xlsx',
        fileSize: '890KB',
        fileExtension: 'xlsx',
        mimeType: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
        sender: 'other',
        senderName: '产品经理',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=pm1',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24 + 3600).toISOString(),
        status: 'sent'
      },
      {
        id: 14,
        content: '这个表格很详细，数据分析得不错',
        type: 'text',
        sender: 'me',
        senderName: '我',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=me',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24 + 3900).toISOString(),
        status: 'sent'
      },
      {
        id: 15,
        content: 'blob:demo-file-url',
        type: 'file',
        fileName: '项目文档.zip',
        fileSize: '5.2MB',
        fileExtension: 'zip',
        mimeType: 'application/zip',
        sender: 'other',
        senderName: '技术负责人',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=tech1',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24 + 4200).toISOString(),
        status: 'sent'
      },
      {
        id: 16,
        content: 'blob:demo-file-url',
        type: 'file',
        fileName: '演示稿.pptx',
        fileSize: '3.8MB',
        fileExtension: 'pptx',
        mimeType: 'application/vnd.openxmlformats-officedocument.presentationml.presentation',
        sender: 'other',
        senderName: '产品经理',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=pm1',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24 + 4500).toISOString(),
        status: 'sent'
      }
    ]
    
    // 测试团队的消息
    messages.value[12] = [
      {
        id: 1,
        content: '发现一个严重bug，需要紧急修复',
        type: 'text',
        sender: 'other',
        senderName: '测试工程师A',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=qa1',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 60).toISOString(),
        status: 'sent'
      },
      {
        id: 2,
        content: 'blob:demo-file-url',
        type: 'file',
        fileName: '测试报告.docx',
        fileSize: '1.2MB',
        fileExtension: 'docx',
        mimeType: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
        sender: 'other',
        senderName: '测试工程师A',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=qa1',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 60 + 200).toISOString(),
        status: 'sent'
      },
      {
        id: 3,
        content: '什么问题？',
        type: 'text',
        sender: 'me',
        senderName: '我',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=me',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 60 + 300).toISOString(),
        status: 'sent'
      },
      {
        id: 4,
        content: '用户登录后，个人信息页面无法加载',
        type: 'text',
        sender: 'other',
        senderName: '测试工程师A',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=qa1',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 60 + 600).toISOString(),
        status: 'sent'
      },
      {
        id: 5,
        content: '重现概率有多高？',
        type: 'text',
        sender: 'other',
        senderName: '开发工程师B',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=dev5',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 60 + 900).toISOString(),
        status: 'sent'
      },
      {
        id: 6,
        content: '100%重现，已经在JIRA上创建了bug单',
        type: 'text',
        sender: 'other',
        senderName: '测试工程师A',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=qa1',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 60 + 1200).toISOString(),
        status: 'sent'
      },
      {
        id: 7,
        content: 'blob:demo-file-url',
        type: 'file',
        fileName: 'bugfix.js',
        fileSize: '8.5KB',
        fileExtension: 'js',
        mimeType: 'application/javascript',
        sender: 'me',
        senderName: '我',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=me',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 60 + 1500).toISOString(),
        status: 'sent'
      },
      {
        id: 8,
        content: 'BUG-12345',
        type: 'text',
        sender: 'other',
        senderName: '测试工程师A',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=qa1',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 60 + 1500).toISOString(),
        status: 'sent'
      },
      {
        id: 9,
        content: '我看看，可能是权限验证的问题',
        type: 'text',
        sender: 'me',
        senderName: '我',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=me',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 60 + 1800).toISOString(),
        status: 'sent'
      },
      {
        id: 10,
        content: 'blob:demo-file-url',
        type: 'file',
        fileName: 'screen_recording.mp4',
        fileSize: '15.3MB',
        fileExtension: 'mp4',
        mimeType: 'video/mp4',
        sender: 'other',
        senderName: '测试工程师A',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=qa1',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 60 + 2000).toISOString(),
        status: 'sent'
      },
      {
        id: 11,
        content: '需要回滚到上个版本吗？',
        type: 'text',
        sender: 'other',
        senderName: '运维工程师',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=ops1',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 60 + 2100).toISOString(),
        status: 'sent'
      }
    ]
    
    // 小赵的消息
    messages.value[9] = [
      {
        id: 1,
        content: '你好，我这里有一份需求文档',
        type: 'text',
        sender: 'other',
        senderName: '小赵',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Zhao&backgroundColor=84cc16',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 30 - 600).toISOString(),
        status: 'sent'
      },
      {
        id: 2,
        content: 'blob:demo-file-url',
        type: 'file',
        fileName: '需求文档.pdf',
        fileSize: '2.1MB',
        fileExtension: 'pdf',
        mimeType: 'application/pdf',
        sender: 'other',
        senderName: '小赵',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Zhao&backgroundColor=84cc16',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 30).toISOString(),
        status: 'sent'
      },
      {
        id: 3,
        content: '我来看看',
        type: 'text',
        sender: 'me',
        senderName: '我',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=me',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 30 + 300).toISOString(),
        status: 'sent'
      },
      {
        id: 4,
        content: 'blob:demo-file-url',
        type: 'file',
        fileName: '项目计划.md',
        fileSize: '15KB',
        fileExtension: 'md',
        mimeType: 'text/markdown',
        sender: 'me',
        senderName: '我',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=me',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 30 + 600).toISOString(),
        status: 'sent'
      },
      {
        id: 5,
        content: '这个计划很详细，我们按这个来',
        type: 'text',
        sender: 'other',
        senderName: '小赵',
        senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Zhao&backgroundColor=84cc16',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 30 + 900).toISOString(),
        status: 'sent'
      }
    ]
  }
  
  // 设置活跃会话
  function setActiveConversation(conversationId) {
    const conversation = conversations.value.find(c => c.id === conversationId)
    activeConversation.value = conversation
    
    // 清除未读消息
    if (conversation) {
      conversation.unreadCount = 0
    }
  }
  
  // 获取会话消息
  function getConversationMessages(conversationId) {
    return messages.value[conversationId] || []
  }
  
  // 发送消息
  function sendMessage(conversationId, messageData) {
    if (!messages.value[conversationId]) {
      messages.value[conversationId] = []
    }
    
    const newMessage = {
      id: Date.now(),
      content: messageData.content,
      type: messageData.type || 'text',
      sender: 'me',
      senderName: '我',
      senderAvatar: 'https://api.dicebear.com/7.x/personas/svg?seed=me',
      timestamp: new Date().toISOString(),
      status: 'sending'
    }
    
    // 如果是文件消息，保存文件的完整信息
    if (messageData.type === 'file' && messageData.fileName) {
      newMessage.fileName = messageData.fileName
      newMessage.fileSize = messageData.fileSize
      newMessage.fileExtension = messageData.fileExtension
      newMessage.mimeType = messageData.mimeType
      newMessage.originalFile = messageData.originalFile
    }
    
    messages.value[conversationId].push(newMessage)
    
    // 更新会话的最后消息
    const conversation = conversations.value.find(c => c.id === conversationId)
    if (conversation) {
      if (messageData.type === 'text') {
        conversation.lastMessage = messageData.content
      } else if (messageData.type === 'file' && messageData.fileName) {
        conversation.lastMessage = `[文件] ${messageData.fileName}`
      } else {
        conversation.lastMessage = `[${getMessageTypeDisplay(messageData.type)}]`
      }
      conversation.lastMessageType = messageData.type
      conversation.lastMessageTime = new Date().toISOString()
    }
    
    // 模拟发送状态更新
    setTimeout(() => {
      newMessage.status = 'sent'
      
      // 模拟对方回复（仅私聊）
      if (conversation && conversation.type === 'private') {
        setTimeout(() => {
          const replyMessage = {
            id: Date.now() + 1,
            content: generateAutoReply(messageData.content, messageData.type),
            type: 'text',
            sender: 'other',
            senderName: conversation.name,
            senderAvatar: conversation.avatar,
            timestamp: new Date().toISOString(),
            status: 'sent'
          }
          
          messages.value[conversationId].push(replyMessage)
          
          // 更新会话信息
          conversation.lastMessage = replyMessage.content
          conversation.lastMessageType = replyMessage.type
          conversation.lastMessageTime = replyMessage.timestamp
          
          // 如果不是当前活跃会话，增加未读数
          if (!activeConversation.value || activeConversation.value.id !== conversationId) {
            conversation.unreadCount = (conversation.unreadCount || 0) + 1
          }
        }, 1500)
      }
    }, 500)
    
    return newMessage
  }
  
  // 加载历史消息
  function loadMoreMessages(conversationId) {
    if (loadingMessages.value[conversationId]) return
    
    loadingMessages.value[conversationId] = true
    
    return new Promise((resolve) => {
      setTimeout(() => {
        // 模拟加载历史消息
        const existingMessages = messages.value[conversationId] || []
        const newMessages = generateHistoryMessages(conversationId, existingMessages.length)
        
        messages.value[conversationId] = [...newMessages, ...existingMessages]
        loadingMessages.value[conversationId] = false
        
        resolve(newMessages)
      }, 1000)
    })
  }
  
  // 删除会话
  function deleteConversation(conversationId) {
    const index = conversations.value.findIndex(c => c.id === conversationId)
    if (index !== -1) {
      conversations.value.splice(index, 1)
      delete messages.value[conversationId]
      
      if (activeConversation.value && activeConversation.value.id === conversationId) {
        activeConversation.value = null
      }
    }
  }
  
  // 置顶/取消置顶会话
  function togglePinConversation(conversationId) {
    const conversation = conversations.value.find(c => c.id === conversationId)
    if (conversation) {
      conversation.isPinned = !conversation.isPinned
    }
  }
  
  // 静音/取消静音会话
  function toggleMuteConversation(conversationId) {
    const conversation = conversations.value.find(c => c.id === conversationId)
    if (conversation) {
      conversation.isMuted = !conversation.isMuted
    }
  }
  
  // 创建或获取会话
  function getOrCreateConversation(userId) {
    let conversation = conversations.value.find(c => c.id === userId)
    
    if (!conversation) {
      conversation = {
        id: userId,
        name: `用户${userId}`,
        avatar: `https://api.dicebear.com/7.x/personas/svg?seed=user${userId}`,
        lastMessage: '',
        lastMessageType: 'text',
        lastMessageTime: new Date().toISOString(),
        unreadCount: 0,
        isOnline: Math.random() > 0.5,
        isPinned: false,
        isMuted: false,
        type: 'private'
      }
      
      conversations.value.unshift(conversation)
      messages.value[userId] = []
    }
    
    return conversation
  }
  
  // 辅助函数
  function getMessageTypeDisplay(type) {
    const typeMap = {
      text: '文字',
      image: '图片',
      voice: '语音',
      video: '视频',
      file: '文件',
      emoji: '表情'
    }
    return typeMap[type] || '消息'
  }
  
  function generateAutoReply(content, type) {
    const replies = {
      text: [
        '收到！',
        '好的，明白了',
        '谢谢分享',
        '有道理',
        '我也这么觉得',
        '确实如此',
        '说得对'
      ],
      image: [
        '图片很不错！',
        '看起来很棒',
        '收到图片了',
        '很有意思的图片'
      ],
      voice: [
        '收到语音了',
        '声音很清楚',
        '听到了'
      ],
      file: [
        '文件已收到',
        '谢谢发送文件',
        '我来看看'
      ]
    }
    
    const typeReplies = replies[type] || replies.text
    return typeReplies[Math.floor(Math.random() * typeReplies.length)]
  }
  
  function generateHistoryMessages(conversationId, existingCount) {
    const messageTemplates = [
      '这个想法很不错',
      '我也有同样的想法',
      '确实需要考虑这个问题',
      '技术选型很重要',
      '性能优化是个持续的过程',
      '用户体验要放在第一位',
      '代码质量很关键',
      '测试覆盖率需要提高'
    ]
    
    const newMessages = []
    const count = Math.min(10, existingCount === 0 ? 5 : 10) // 第一次加载5条，后续10条
    
    for (let i = 0; i < count; i++) {
      const isMe = Math.random() > 0.5
      const messageTime = new Date(Date.now() - (existingCount + count - i) * 60 * 1000).toISOString()
      
      newMessages.push({
        id: Date.now() + i,
        content: messageTemplates[Math.floor(Math.random() * messageTemplates.length)],
        type: 'text',
        sender: isMe ? 'me' : 'other',
        senderName: isMe ? '我' : '历史用户',
        senderAvatar: isMe 
          ? 'https://api.dicebear.com/7.x/personas/svg?seed=me'
          : `https://api.dicebear.com/7.x/personas/svg?seed=history${i}`,
        timestamp: messageTime,
        status: 'sent'
      })
    }
    
    return newMessages
  }
  
  return {
    // 状态
    conversations,
    activeConversation,
    messages,
    loadingMessages,
    
    // 计算属性
    sortedConversations,
    
    // 方法
    initializeConversations,
    setActiveConversation,
    getConversationMessages,
    sendMessage,
    loadMoreMessages,
    deleteConversation,
    togglePinConversation,
    toggleMuteConversation,
    getOrCreateConversation
  }
})