import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useAuthStore } from '../stores/auth'

export const useFeedStore = defineStore('feed', () => {
  const authStore = useAuthStore()
  // 生成模拟数据
  function randomFrom(arr) { return arr[Math.floor(Math.random() * arr.length)] }
  function randomTime(daysAgo = 10) {
    const now = Date.now() - Math.floor(Math.random() * daysAgo * 24 * 3600 * 1000)
    return new Date(now).toISOString()
  }
  const userList = [
    { id: 'u1', name: '好友1', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Friend1' },
    { id: 'u2', name: '好友2', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Friend2' },
    { id: 'u3', name: '好友3', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Friend3' },
    { id: 'u4', name: '好友4', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Friend4' },
    { id: 'u5', name: '好友5', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Friend5' },
    { id: 'u6', name: '好友6', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Friend6' },
    { id: 'u7', name: '好友7', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Friend7' },
    { id: 'u8', name: '好友8', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Friend8' },
    { id: 'u9', name: '好友9', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Friend9' },
    { id: 'u10', name: '好友10', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=Friend10' }
  ]
  const sampleContents = [
    '今天的心情超级棒！',
    '分享一张美食照 #美食',
    '和朋友们一起去爬山，风景超美',
    '最近在追剧，剧情太精彩了',
    '工作顺利，继续加油！',
    '生活需要仪式感',
    '阳光正好，适合发呆',
    '新技能get√',
    '周末宅家，快乐加倍',
    '记录美好瞬间',
    '运动让人快乐',
    '今天又学到新知识',
    '和家人一起的时光最温馨',
    '偶遇老友，开心！',
    '小确幸的一天'
  ]
  const sampleReplies = [
    '说得好！',
    '同感同感',
    '哈哈哈',
    '点赞！',
    '太棒了',
    '羡慕ing',
    '有机会一起',
    '支持你',
    '加油！',
    '真不错'
  ]
  const posts = ref(Array.from({ length: 50 }).map((_, i) => {
    const user = randomFrom(userList)
    const likeCount = Math.floor(Math.random() * 20) + 3
    const likeUsers = Array.from({ length: likeCount }).map(() => randomFrom(userList).name)
    const commentCount = Math.floor(Math.random() * 8) + 1
    const comments = Array.from({ length: commentCount }).map((_, ci) => {
      const commentUser = randomFrom(userList)
      const replyCount = Math.floor(Math.random() * 4)
      const replies = Array.from({ length: replyCount }).map((_, ri) => {
        const replyUser = randomFrom(userList)
        return {
          id: Date.now() + i*1000 + ci*100 + ri,
          user: replyUser.name,
          content: `@${commentUser.name} ${randomFrom(sampleReplies)}`,
          time: `${Math.floor(Math.random()*23)+1}:${Math.floor(Math.random()*6)}${Math.floor(Math.random()*10)}`,
          likes: []
        }
      })
      return {
        id: Date.now() + i*100 + ci,
        user: commentUser.name,
        content: randomFrom(sampleContents),
        time: `${Math.floor(Math.random()*23)+1}:${Math.floor(Math.random()*6)}${Math.floor(Math.random()*10)}`,
        replies
      }
    })
    return {
      id: i + 1,
      type: 'friend',
      user: { 
        name: user.name, 
        avatar: user.avatar 
      },
      time: `${Math.floor(Math.random()*59)+1}分钟前`,
      content: randomFrom(sampleContents),
      image: i % 4 === 0 ? `https://picsum.photos/seed/friend${i+1}/480/220` : '',
      likes: likeUsers,
      liked: false, // 当前用户是否点赞
      comments
    }
  }))
  const communityPosts = ref(Array.from({ length: 50 }).map((_, i) => {
    const user = randomFrom(userList)
    const likeCount = Math.floor(Math.random() * 25) + 5
    const likeUsers = Array.from({ length: likeCount }).map(() => `社区成员${Math.floor(Math.random()*20)+1}`)
    const commentCount = Math.floor(Math.random() * 10) + 1
    const comments = Array.from({ length: commentCount }).map((_, ci) => {
      const commentUser = `社区成员${Math.floor(Math.random()*20)+1}`
      const replyCount = Math.floor(Math.random() * 4)
      const replies = Array.from({ length: replyCount }).map((_, ri) => {
        const replyUser = `社区回复${Math.floor(Math.random()*20)+1}`
        return {
          id: Date.now() + i*1000 + ci*100 + ri,
          user: replyUser,
          content: `@${commentUser} 回复内容${ri+1}`,
          time: `${Math.floor(Math.random()*23)+1}:${Math.floor(Math.random()*6)}${Math.floor(Math.random()*10)}`,
          likes: []
        }
      })
      return {
        id: Date.now() + i*100 + ci,
        user: commentUser,
        content: `社区评论${ci+1} @${user.name}`,
        time: `${Math.floor(Math.random()*23)+1}:${Math.floor(Math.random()*6)}${Math.floor(Math.random()*10)}`,
        replies
      }
    })
    return {
      id: i + 101,
      type: 'community',
      user: { 
        name: `社区用户${i+1}`, 
        avatar: `https://api.dicebear.com/7.x/personas/svg?seed=Community${i+1}` 
      },
      time: `${Math.floor(Math.random()*59)+1}分钟前`,
      content: `这里是社区动态${i+1}，欢迎大家参与讨论！${i%2===0?'#讨论':'#分享'}${i%7===0?'#热点':''}`,
      image: i % 5 === 0 ? `https://picsum.photos/seed/community${i+1}/480/220` : '',
      likes: likeUsers,
      liked: false,
      comments
    }
  }))

  // getter
  const getFriendFeeds = computed(() => posts.value)
  const getCommunityFeeds = computed(() => communityPosts.value)
  
  // 当前正在查看的图片索引
  const currentImageIndex = ref(0)
  // 当前查看的图片列表
  const viewingImages = ref([])
  
  // 发布新动态
  const createPost = (content, images, type = 'friend') => {
    const currentUser = authStore.user || { name: '我', avatar: 'https://api.dicebear.com/7.x/personas/svg?seed=You' }
    
    const newPost = {
      id: Date.now(),
      type,
      user: { 
        name: currentUser.name, 
        avatar: currentUser.avatar 
      },
      time: '刚刚',
      content,
      image: images && images.length > 0 ? images[0] : '', // 保持向后兼容
      images: images || [], // 支持多张图片
      likes: [],
      liked: false,
      comments: []
    }
    
    if (type === 'friend') {
      posts.value.unshift(newPost)
    } else {
      communityPosts.value.unshift(newPost)
    }
    
    return newPost
  }
  
  // 点赞/取消点赞
  const toggleLike = (postId, type = 'friend') => {
    const targetPosts = type === 'friend' ? posts.value : communityPosts.value
    const post = targetPosts.find(p => p.id === postId)
    if (!post) return
    
    const currentUser = authStore.user?.name || '我'
    
    // 判断当前用户是否已点赞
    const hasLiked = post.likes.includes(currentUser)
    if (hasLiked) {
      // 取消点赞
      post.likes = post.likes.filter(u => u !== currentUser)
      post.liked = false
    } else {
      // 点赞，确保无重复
      post.likes = [currentUser, ...post.likes.filter(u => u !== currentUser)]
      post.liked = true
    }
  }
  
  // 添加评论
  const addComment = (postId, content, type = 'friend') => {
    const targetPosts = type === 'friend' ? posts.value : communityPosts.value
    const post = targetPosts.find(p => p.id === postId)
    if (!post) return
    
    const currentUser = authStore.user?.name || '我'
    
    const newComment = {
      id: Date.now(),
      user: currentUser,
      content,
      time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
      replies: []
    }
    
    post.comments.unshift(newComment)
  }
  
  // 删除评论
  const deleteComment = (postId, commentId, type = 'friend') => {
    const targetPosts = type === 'friend' ? posts.value : communityPosts.value
    const post = targetPosts.find(p => p.id === postId)
    if (!post) return
    
    const index = post.comments.findIndex(c => c.id === commentId)
    if (index > -1) {
      post.comments.splice(index, 1)
    }
  }
  
  // 添加回复
  const addReply = (postId, commentId, content, type = 'friend') => {
    const targetPosts = type === 'friend' ? posts.value : communityPosts.value
    const post = targetPosts.find(p => p.id === postId)
    if (!post) return
    
    const comment = post.comments.find(c => c.id === commentId)
    if (!comment) return
    
    const currentUser = authStore.user?.name || '我'
    
    const newReply = {
      id: Date.now(),
      user: currentUser,
      content,
      time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
      likes: []
    }
    
    comment.replies.push(newReply)
  }
  
  // 删除回复
  const deleteReply = (postId, commentId, replyId, type = 'friend') => {
    const targetPosts = type === 'friend' ? posts.value : communityPosts.value
    const post = targetPosts.find(p => p.id === postId)
    if (!post) return
    
    const comment = post.comments.find(c => c.id === commentId)
    if (!comment) return
    
    const index = comment.replies.findIndex(r => r.id === replyId)
    if (index > -1) {
      comment.replies.splice(index, 1)
    }
  }
  
  // 新增：评论点赞/取消点赞
  const toggleCommentLike = (postId, commentId, type = 'friend') => {
    const targetPosts = type === 'friend' ? posts.value : communityPosts.value
    const post = targetPosts.find(p => p.id === postId)
    if (!post) return
    
    const comment = post.comments.find(c => c.id === commentId)
    if (!comment) return
    
    const currentUser = authStore.user?.name || '我'
    
    if (!comment.likes) {
      comment.likes = []
    }
    
    const index = comment.likes.indexOf(currentUser)
    if (index > -1) {
      comment.likes.splice(index, 1)
    } else {
      comment.likes.push(currentUser)
    }
  }
  
  // 新增：回复点赞/取消点赞
  const toggleReplyLike = (postId, commentId, replyId, type = 'friend') => {
    const targetPosts = type === 'friend' ? posts.value : communityPosts.value
    const post = targetPosts.find(p => p.id === postId)
    if (!post) return
    
    const comment = post.comments.find(c => c.id === commentId)
    if (!comment) return
    
    const reply = comment.replies.find(r => r.id === replyId)
    if (!reply) return
    
    const currentUser = authStore.user?.name || '我'
    
    if (!reply.likes) {
      reply.likes = []
    }
    
    const index = reply.likes.indexOf(currentUser)
    if (index > -1) {
      reply.likes.splice(index, 1)
    } else {
      reply.likes.push(currentUser)
    }
  }
  
  // 打开图片查看器
  const openImageViewer = (postId, index = 0) => {
    const post = posts.value.find(p => p.id === postId) || communityPosts.value.find(p => p.id === postId)
    if (post && post.image) {
      viewingImages.value = [post.image]
      currentImageIndex.value = index
    }
  }
  
  // 关闭图片查看器
  const closeImageViewer = () => {
    viewingImages.value = []
  }
  
  return {
    posts,
    communityPosts,
    getFriendFeeds,
    getCommunityFeeds,
    viewingImages,
    currentImageIndex,
    createPost,
    toggleLike,
    addComment,
    deleteComment,
    addReply,
    deleteReply,
    toggleCommentLike,
    toggleReplyLike,
    openImageViewer,
    closeImageViewer
  }
})