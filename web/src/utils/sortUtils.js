import pinyin from 'pinyin'

/**
 * 获取字符串的拼音
 * @param {string} str - 要转换的字符串
 * @returns {string} - 拼音字符串
 */
export function getPinyin(str) {
  if (!str) return ''
  return pinyin(str, { style: pinyin.STYLE_NORMAL })
    .flat()
    .join('')
    .toLowerCase()
}

/**
 * 获取字符串的拼音首字母
 * @param {string} str - 要转换的字符串
 * @returns {string} - 拼音首字母
 */
export function getPinyinFirstLetter(str) {
  if (!str) return '#'
  const pinyinResult = pinyin(str.charAt(0), { style: pinyin.STYLE_FIRST_LETTER })
  return pinyinResult[0]?.[0]?.toUpperCase() || str.charAt(0).toUpperCase()
}

/**
 * 微信风格的好友排序函数
 * @param {Array} friends - 好友数组
 * @returns {Array} - 排序后的好友数组
 */
export function sortFriendsByPinyin(friends) {
  return friends.slice().sort((a, b) => {
    const nameA = getPinyin(a.remark || a.name)
    const nameB = getPinyin(b.remark || b.name)
    return nameA.localeCompare(nameB)
  })
}

/**
 * 按拼音首字母分组好友
 * @param {Array} friends - 排序后的好友数组
 * @returns {Array} - 分组后的好友数组
 */
export function groupFriendsByPinyin(friends) {
  const groups = {}
  
  friends.forEach(friend => {
    const displayName = friend.remark || friend.name
    const firstLetter = getPinyinFirstLetter(displayName)
    
    if (!groups[firstLetter]) {
      groups[firstLetter] = []
    }
    groups[firstLetter].push(friend)
  })
  
  // 转换为数组并按字母排序
  return Object.keys(groups)
    .sort()
    .map(letter => ({
      letter,
      friends: groups[letter]
    }))
} 