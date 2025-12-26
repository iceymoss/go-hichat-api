import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useContactsStore = defineStore('contacts', () => {
  // 好友列表
  const friends = ref([
    {
      id: 1001,
      name: '小李',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Lily',
      remark: '前端开发',
      status: 'online',
      tags: ['同事', '前端'],
      lastActive: '刚刚',
      gender: 'female',
      email: 'xiaoli@example.com',
      phone: '13800000001',
      signature: '热爱前端，喜欢React和Vue'
    },
    {
      id: 1002,
      name: '王经理',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Wang',
      remark: '产品经理',
      status: 'away',
      tags: ['领导'],
      lastActive: '1小时前',
      gender: 'male',
      email: 'wang@example.com',
      phone: '13800000002',
      signature: '产品为王，体验至上'
    },
    {
      id: 1003,
      name: '张工程师',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Zhang',
      remark: '后端开发',
      status: 'online',
      tags: ['同事', 'Go'],
      lastActive: '刚刚',
      gender: 'male',
      email: 'zhang@example.com',
      phone: '13800000003',
      signature: '专注Go开发，热爱技术分享'
    },
    {
      id: 1004,
      name: '小陈',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Chen',
      remark: 'UI设计师',
      status: 'offline',
      tags: ['设计'],
      lastActive: '2天前',
      gender: 'female',
      email: 'chen@example.com',
      phone: '13800000004',
      signature: '设计让生活更美好'
    },
    {
      id: 101,
      name: '小明',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Ming',
      remark: '后端开发',
      tags: ['Go', '后端'],
      status: 'online',
      lastActive: '1分钟前',
      gender: 'male',
      email: 'xiaoming@example.com',
      phone: '13800000005',
      signature: '代码改变世界'
    },
    {
      id: 102,
      name: '小红',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Hong',
      remark: '产品经理',
      tags: ['产品'],
      status: 'away',
      lastActive: '5分钟前',
      gender: 'female',
      email: 'xiaohong@example.com',
      phone: '13800000006',
      signature: '用心做产品'
    },
    {
      id: 103,
      name: '小刚',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Gang',
      remark: '测试工程师',
      tags: ['测试'],
      status: 'offline',
      lastActive: '10分钟前',
      gender: 'male',
      email: 'gang@example.com',
      phone: '13800000007',
      signature: '测试让产品更可靠'
    },
    {
      id: 104,
      name: '小美',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Mei',
      remark: 'UI设计',
      tags: ['设计'],
      status: 'online',
      lastActive: '2分钟前',
      gender: 'female',
      email: 'mei@example.com',
      phone: '13800000008',
      signature: '美是设计的灵魂'
    },
    {
      id: 105,
      name: '小强',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Qiang',
      remark: '前端开发',
      tags: ['Vue', '前端'],
      status: 'away',
      lastActive: '8分钟前',
      gender: 'male',
      email: 'qiang@example.com',
      phone: '13800000009',
      signature: '前端路上，一起成长'
    },
    {
      id: 106,
      name: '小芳',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Fang',
      remark: '运营',
      tags: ['运营'],
      status: 'offline',
      lastActive: '20分钟前',
      gender: 'female',
      email: 'fang@example.com',
      phone: '13800000010',
      signature: '用心服务每一位用户'
    },
    {
      id: 107,
      name: '小军',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Jun',
      remark: '后端开发',
      tags: ['Go'],
      status: 'online',
      lastActive: '3分钟前',
      gender: 'male',
      email: 'jun@example.com',
      phone: '13800000011',
      signature: '后端技术爱好者'
    },
    {
      id: 108,
      name: '小雪',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Xue',
      remark: '产品助理',
      tags: ['产品'],
      status: 'away',
      lastActive: '12分钟前',
      gender: 'female',
      email: 'xue@example.com',
      phone: '13800000012',
      signature: '助理也有大梦想'
    },
    {
      id: 109,
      name: '小林',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Lin',
      remark: '测试',
      tags: ['测试'],
      status: 'offline',
      lastActive: '30分钟前',
      gender: 'male',
      email: 'lin@example.com',
      phone: '13800000013',
      signature: '细致入微，追求极致'
    },
    {
      id: 110,
      name: '小天',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Tian',
      remark: '前端开发',
      tags: ['Vue'],
      status: 'online',
      lastActive: '刚刚',
      gender: 'male',
      email: 'tian@example.com',
      phone: '13800000014',
      signature: '热爱前端，乐于分享'
    },
    {
      id: 111,
      name: '阿丽',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Ali',
      remark: '',
      tags: ['朋友'],
      status: 'online',
      lastActive: '刚刚',
      gender: 'female',
      email: 'ali@example.com',
      phone: '13800000015',
      signature: '生活美好'
    },
    {
      id: 112,
      name: '安安',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Anan',
      remark: '',
      tags: ['朋友'],
      status: 'away',
      lastActive: '5分钟前',
      gender: 'female',
      email: 'anan@example.com',
      phone: '13800000016',
      signature: '平安喜乐'
    },
    {
      id: 113,
      name: '安南',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Annan',
      remark: '',
      tags: ['朋友'],
      status: 'offline',
      lastActive: '1小时前',
      gender: 'male',
      email: 'annan@example.com',
      phone: '13800000017',
      signature: '南方人'
    },
    {
      id: 114,
      name: '兵哥',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Bingge',
      remark: '',
      tags: ['朋友'],
      status: 'online',
      lastActive: '刚刚',
      gender: 'male',
      email: 'bingge@example.com',
      phone: '13800000018',
      signature: '军人作风'
    },
    {
      id: 115,
      name: '鲍姐',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Baojie',
      remark: '',
      tags: ['朋友'],
      status: 'away',
      lastActive: '10分钟前',
      gender: 'female',
      email: 'baojie@example.com',
      phone: '13800000019',
      signature: '大姐大'
    },
    {
      id: 116,
      name: '八爷',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Bayi',
      remark: '',
      tags: ['朋友'],
      status: 'offline',
      lastActive: '2小时前',
      gender: 'male',
      email: 'bayi@example.com',
      phone: '13800000020',
      signature: '老八'
    },
    {
      id: 117,
      name: '张三',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=ZhangSan',
      remark: '',
      tags: ['朋友'],
      status: 'online',
      lastActive: '刚刚',
      gender: 'male',
      email: 'zhangsan@example.com',
      phone: '13800000021',
      signature: '张三李四'
    },
    {
      id: 118,
      name: '赵高',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=ZhaoGao',
      remark: '',
      tags: ['朋友'],
      status: 'away',
      lastActive: '15分钟前',
      gender: 'male',
      email: 'zhaogao@example.com',
      phone: '13800000022',
      signature: '高个子'
    },
    {
      id: 119,
      name: 'Alice',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Alice',
      remark: '',
      tags: ['朋友'],
      status: 'online',
      lastActive: '刚刚',
      gender: 'female',
      email: 'alice@example.com',
      phone: '13800000023',
      signature: 'Hello World'
    },
    {
      id: 120,
      name: 'Bob',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Bob',
      remark: '',
      tags: ['朋友'],
      status: 'away',
      lastActive: '8分钟前',
      gender: 'male',
      email: 'bob@example.com',
      phone: '13800000024',
      signature: 'Bob the Builder'
    },
    {
      id: 121,
      name: 'Charlie',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Charlie',
      remark: '',
      tags: ['朋友'],
      status: 'offline',
      lastActive: '1小时前',
      gender: 'male',
      email: 'charlie@example.com',
      phone: '13800000025',
      signature: 'Charlie Brown'
    }
  ])
  
  // 好友请求列表
  const friendRequests = ref([
    {
      id: 2001,
      name: '刘同学',
      avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Liu',
      remark: '求职者',
      message: '您好，我对Go开发很感兴趣',
      status: 'pending',
      time: '10分钟前'
    }
  ])
  
  // 搜索关键词
  const searchKeyword = ref('')
  
  // 分组好友列表
  const groupedFriends = computed(() => {
    const groups = {
      online: [],
      away: [],
      offline: []
    }
    
    friends.value.forEach(friend => {
      if (friend.status === 'online') groups.online.push(friend)
      else if (friend.status === 'away') groups.away.push(friend)
      else groups.offline.push(friend)
    })
    
    return groups
  })
  
  // 过滤搜索结果
  const searchResults = computed(() => {
    if (!searchKeyword.value.trim()) return []
    
    const keyword = searchKeyword.value.toLowerCase()
    return friends.value.filter(friend => 
      friend.name.toLowerCase().includes(keyword) || 
      friend.remark.toLowerCase().includes(keyword) ||
      friend.tags.some(tag => tag.toLowerCase().includes(keyword))
    )
  })
  
  // 添加好友请求
  const addFriendRequest = (request) => {
    friendRequests.value.unshift(request)
  }
  
  // 处理好友请求
  const handleFriendRequest = (requestId, accept) => {
    const requestIndex = friendRequests.value.findIndex(r => r.id === requestId)
    
    if (requestIndex !== -1) {
      const request = friendRequests.value[requestIndex]
      
      if (accept) {
        // 添加到好友列表
        friends.value.unshift({
          id: Date.now(),
          name: request.name,
          avatar: request.avatar,
          remark: request.remark || request.name,
          status: 'online',
          tags: ['新好友'],
          lastActive: '刚刚'
        })
      }
      
      // 移除请求
      friendRequests.value.splice(requestIndex, 1)
    }
  }
  
  // 发送好友请求
  const sendFriendRequest = (userId) => {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          success: true,
          message: '好友请求已发送'
        })
      }, 1000)
    })
  }
  
  return {
    friends,
    friendRequests,
    searchKeyword,
    groupedFriends,
    searchResults,
    addFriendRequest,
    handleFriendRequest,
    sendFriendRequest
  }
})