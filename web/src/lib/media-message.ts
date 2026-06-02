/**
 * 富媒体消息内容约定
 *
 * 文本消息：msgContent 直接是纯文本（向后兼容，不变）。
 * 富媒体消息：msgContent 是一段 JSON 字符串，按媒体类型携带元数据。
 * 前后端共用同一份字段命名（见 docs/specs/rich-media-message.md）。
 */

import type { Message } from './mock-data';

/** 媒体消息类型（与 Message['type'] 中的富媒体子集对应） */
export type MediaKind = 'image' | 'video' | 'file' | 'voice' | 'memes';

export interface MediaContent {
  url: string;
  /** 图片缩略图 / 表情包小图 */
  thumbUrl?: string;
  /** 视频封面 */
  coverUrl?: string;
  /** 文件原始名 */
  name?: string;
  /** 字节数 */
  size?: number;
  /** 图片/视频宽高 */
  width?: number;
  height?: number;
  /** 语音/视频时长（秒） */
  duration?: number;
}

/** 把媒体元数据序列化为 msgContent 字符串 */
export function buildMediaContent(meta: MediaContent): string {
  return JSON.stringify(meta);
}

/** 解析富媒体 msgContent；非 JSON 或解析失败返回 null（降级为文本/不支持） */
export function parseMediaContent(content: string): MediaContent | null {
  if (!content || content[0] !== '{') return null;
  try {
    const obj = JSON.parse(content);
    if (obj && typeof obj.url === 'string') return obj as MediaContent;
    return null;
  } catch {
    return null;
  }
}

/** 会话列表 / 引用预览里富媒体消息的占位文案 */
export function mediaPreview(type: Message['type'], content: string): string {
  switch (type) {
    case 'image':
      return '[图片]';
    case 'video':
      return '[视频]';
    case 'voice':
      return '[语音]';
    case 'file': {
      const meta = parseMediaContent(content);
      return meta?.name ? `[文件] ${meta.name}` : '[文件]';
    }
    default:
      return content;
  }
}
