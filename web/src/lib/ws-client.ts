/**
 * WebSocket IM Client — RigorAck 三次通信协议
 *
 * 对接后端 apps/im/ws 的完整协议实现:
 *   1. 客户端发送消息 (ackSeq=0, frameType=0x0 FrameData)
 *   2. 服务端回 ACK   (ackSeq=1, frameType=0x3 FrameAck)
 *   3. 客户端确认 ACK (ackSeq=2, frameType=0x3 FrameAck) → 服务端处理业务
 */

// ========== 帧类型 (与后端 message.go 对应) ==========
export const FrameType = {
  Data: 0x0,
  Ping: 0x1,
  Err: 0x2,
  Ack: 0x3,
  NoAck: 0x4,
  CAck: 0x5,
} as const;

// ========== 消息类型 (与后端 constants/im.go 对应) ==========
export const MsgType = {
  Text: 1,
  File: 2,
  Voice: 3,
  Image: 4,
  Memes: 5,
  // ContentMakeRead = 6：已读回执，不是正常消息类型
  ContentMakeRead: 6,
  // Video = 8：视频（追加在 6/7 控制类型之后，与后端 constants/im.go 对应）
  Video: 8,
  // Call = 10：音视频通话记录（content 为 JSON {callType,status,duration}）
  Call: 10,
} as const;

// ContentType 附加类型（与 MsgType 独立），ws.Chat.contentType 字段
export const ContentType = {
  Normal: 0,
  MakeRead: 6,
  /** 发送方回响：服务端写库后把消息回推给发送方，携带真实 MongoDB MsgId */
  MsgAck: 7,
  /** 撤回事件：服务端把被撤回的 msgId + 操作者回推，前端原位置为撤回态 */
  Recall: 9,
} as const;

// ========== 聊天类型 ==========
export const ChatType = {
  Single: 1,
  Group: 2,
} as const;

// ========== TypeScript 接口 ==========

/** 后端 Message 结构 (message.go) */
export interface WsMessage {
  id: string;
  ackSeq?: number;
  frameType: number;
  method?: string;
  userId?: string;
  formId?: string;
  data?: unknown;
}

/** 后端 ws.Chat 数据结构 (ws.go) */
export interface WsChatData {
  conversationId: string;
  chatType: number;
  sendId: string;
  recvId: string;
  sendTime: number;
  contentType?: number;   // 7 表示这是给发送方的回响（ContentType.MsgAck），9 表示撤回事件
  recalledBy?: string;    // 撤回操作者 uid，仅 contentType=Recall 时有值
  msg: {
    mType: number;
    content: string;
    quote?: string;
    readRecords?: Record<string, string>;
    atUsers?: string[];   // 被 @ 的成员 uid 列表（群聊）
    atAll?: boolean;      // 是否 @所有人（群聊）
  };
}

/** 后端 ws.MarkRead 数据结构 */
export interface WsMarkReadData {
  chatType: number;
  recvId: string;
  conversationId: string;
  sendId: string;
  msgIds: string[];
  readRecords: Record<string, string>;
}

type ConnState = 'disconnected' | 'connecting' | 'connected';
type MessageHandler = (data: unknown, raw: WsMessage) => void;

interface PendingAck {
  resolve: () => void;
  reject: (err: Error) => void;
  timer: ReturnType<typeof setTimeout>;
}

export interface IMWebSocketOptions {
  url?: string;
  token?: string;
  ackTimeout?: number;
  heartbeatInterval?: number;
  reconnect?: boolean;
  reconnectInterval?: number;
  reconnectMaxRetries?: number;
  onStateChange?: (state: ConnState, prev: ConnState) => void;
  onError?: (err: unknown, msgId?: string) => void;
}

// ========== 唯一 ID 生成 ==========
let _seq = 0;
function genMsgId(): string {
  return `msg_${Date.now()}_${++_seq}`;
}

// ========== 主类 ==========

export class IMWebSocket {
  private url: string;
  private token: string;
  private ws: WebSocket | null = null;
  private state: ConnState = 'disconnected';

  // RigorAck
  private pendingAck = new Map<string, PendingAck>();
  private ackTimeout: number;

  // 心跳
  private heartbeatInterval: number;
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null;

  // 重连
  private reconnectEnabled: boolean;
  private reconnectInterval: number;
  private reconnectMaxRetries: number;
  private reconnectRetries = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private manualClose = false;

  // 事件
  private handlers = new Map<string, MessageHandler>();
  private onStateChange: IMWebSocketOptions['onStateChange'];
  private onError: IMWebSocketOptions['onError'];

  constructor(opts: IMWebSocketOptions = {}) {
    this.url = opts.url || 'ws://localhost:10090/ws';
    this.token = opts.token || '';
    this.ackTimeout = opts.ackTimeout || 30_000;
    this.heartbeatInterval = opts.heartbeatInterval || 20_000;
    this.reconnectEnabled = opts.reconnect !== false;
    this.reconnectInterval = opts.reconnectInterval || 3_000;
    this.reconnectMaxRetries = opts.reconnectMaxRetries || 10;
    this.onStateChange = opts.onStateChange;
    this.onError = opts.onError;
  }

  // ===================== Public API =====================

  connect(token?: string) {
    if (token) this.token = token;
    if (!this.token) { console.error('[WS] missing token'); return; }
    if (this.state !== 'disconnected') return;

    this.manualClose = false;
    this.setState('connecting');

    try {
      const wsUrl = `${this.url}?token=${encodeURIComponent(this.token)}`;
      this.ws = new WebSocket(wsUrl);
      this.ws.onopen = () => this.handleOpen();
      this.ws.onmessage = (e) => this.handleMessage(e);
      this.ws.onclose = (e) => this.handleClose(e);
      this.ws.onerror = (e) => { this.onError?.(e); };
    } catch (err) {
      console.error('[WS] connect error:', err);
      this.setState('disconnected');
    }
  }

  disconnect() {
    this.manualClose = true;
    this.stopHeartbeat();
    this.clearReconnect();
    this.rejectAllPending('连接已关闭');
    this.ws?.close();
    this.ws = null;
    this.setState('disconnected');
  }

  /** 注册消息 handler (method → handler) */
  on(method: string, handler: MessageHandler) {
    this.handlers.set(method, handler);
  }

  off(method: string) {
    this.handlers.delete(method);
  }

  /** 等待 WebSocket 连接就绪（OPEN 状态） */
  private waitForOpen(timeout = 5000): Promise<void> {
    if (this.ws?.readyState === WebSocket.OPEN) return Promise.resolve();
    // 如果已断开，尝试重连
    if (this.state === 'disconnected') this.connect();
    return new Promise((resolve, reject) => {
      const timer = setTimeout(() => reject(new Error('连接超时')), timeout);
      const check = () => {
        if (this.ws?.readyState === WebSocket.OPEN) {
          clearTimeout(timer);
          resolve();
        } else if (this.state === 'disconnected') {
          clearTimeout(timer);
          reject(new Error('连接失败'));
        } else {
          setTimeout(check, 100);
        }
      };
      check();
    });
  }

  /** 发送业务消息 (RigorAck 三次通信)，Promise 在 ACK 完成后 resolve。
   *  可传 msgId 让帧 id = 调用方的本地消息 id，便于业务错误帧按消息 id 关联（红感叹号）。 */
  async send(method: string, data: unknown, msgId?: string): Promise<void> {
    // 等待连接就绪
    await this.waitForOpen();

    const id = msgId || genMsgId();
    const msg: WsMessage = {
      id,
      ackSeq: 0,
      frameType: FrameType.Data,
      method,
      data,
    };

    return new Promise((resolve, reject) => {
      const timer = setTimeout(() => {
        this.pendingAck.delete(id);
        reject(new Error(`ACK 超时: ${id}`));
      }, this.ackTimeout);

      this.pendingAck.set(id, { resolve, reject, timer });

      if (!this.sendRaw(msg)) {
        clearTimeout(timer);
        this.pendingAck.delete(id);
        reject(new Error('发送失败: 连接未建立'));
      }
    });
  }

  /** 发送不走 ACK 的消息 (心跳 / NoAck) */
  sendNoAck(method: string, data?: unknown) {
    this.sendRaw({
      id: genMsgId(),
      frameType: FrameType.NoAck,
      method,
      data: data ?? {},
    });
  }

  get connected() { return this.state === 'connected'; }
  get currentState() { return this.state; }

  // ===================== 内部 =====================

  private sendRaw(msg: WsMessage): boolean {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return false;
    try {
      this.ws.send(JSON.stringify(msg));
      return true;
    } catch { return false; }
  }

  private setState(s: ConnState) {
    const prev = this.state;
    this.state = s;
    if (prev !== s) this.onStateChange?.(s, prev);
  }

  // ---------- WebSocket 事件 ----------

  private handleOpen() {
    this.setState('connected');
    this.reconnectRetries = 0;
    this.startHeartbeat();
  }

  private handleMessage(event: MessageEvent) {
    let msg: WsMessage;
    try { msg = JSON.parse(event.data); } catch { return; }

    switch (msg.frameType) {
      case FrameType.Ack:
        this.handleAck(msg);
        break;
      case FrameType.Err: {
        // 业务错误帧（如发送鉴权拦截）带原始消息 id。注意：传输层 ACK 已先于业务校验完成并 resolve 了
        // send promise（pendingAck 已删除），故不走 pendingAck，而是把 id 上交，由上层按消息 id 标记失败（红感叹号）。
        if (msg.id) {
          const pending = this.pendingAck.get(msg.id);
          if (pending) {
            clearTimeout(pending.timer);
            this.pendingAck.delete(msg.id);
            pending.reject(new Error(typeof msg.data === 'string' ? msg.data : '发送失败'));
          }
        }
        this.onError?.(msg.data, msg.id);
        break;
      }
      case FrameType.Data:
      case FrameType.NoAck:
        this.routeMessage(msg);
        break;
      // FramePing: 忽略（服务端不会主动发 FramePing）
    }
  }

  private handleClose(_event: CloseEvent) {
    this.stopHeartbeat();
    this.rejectAllPending('连接断开');
    this.setState('disconnected');
    if (!this.manualClose && this.reconnectEnabled) {
      this.scheduleReconnect();
    }
  }

  // ---------- RigorAck ----------

  /**
   * 收到服务端 ACK (ackSeq=1) → 回复 ackSeq=2 完成三次握手
   *
   * 后端 server.go readAck():
   *   message.AckSeq == 0 → AckSeq++ → 发回 {ackSeq:1, FrameAck}
   *   等待客户端 msgSeq.AckSeq > message.AckSeq → 处理业务
   */
  private handleAck(msg: WsMessage) {
    const pending = this.pendingAck.get(msg.id);
    if (!pending) return; // 超时后到达的 ACK 或重发

    if ((msg.ackSeq ?? 0) > 0) {
      // 第三步：客户端确认
      this.sendRaw({
        id: msg.id,
        ackSeq: (msg.ackSeq ?? 0) + 1,
        frameType: FrameType.Ack,
      });
      clearTimeout(pending.timer);
      this.pendingAck.delete(msg.id);
      pending.resolve();
    }
  }

  // ---------- 消息路由 ----------

  private routeMessage(msg: WsMessage) {
    const method = msg.method || '';
    const handler = this.handlers.get(method);
    if (handler) {
      try { handler(msg.data, msg); } catch (e) { console.error(`[WS] handler "${method}" error:`, e); }
    }
  }

  // ---------- 心跳 ----------

  private startHeartbeat() {
    this.stopHeartbeat();
    this.heartbeatTimer = setInterval(() => {
      if (this.state === 'connected') this.sendNoAck('chat.ping');
    }, this.heartbeatInterval);
  }

  private stopHeartbeat() {
    if (this.heartbeatTimer) { clearInterval(this.heartbeatTimer); this.heartbeatTimer = null; }
  }

  // ---------- 重连 ----------

  private scheduleReconnect() {
    const exponent = Math.min(this.reconnectRetries, Math.max(0, this.reconnectMaxRetries - 1));
    this.reconnectRetries = Math.min(this.reconnectRetries + 1, this.reconnectMaxRetries);
    const delay = Math.min(this.reconnectInterval * 2 ** exponent, 30_000);
    this.reconnectTimer = setTimeout(() => this.connect(), delay);
  }

  private clearReconnect() {
    if (this.reconnectTimer) { clearTimeout(this.reconnectTimer); this.reconnectTimer = null; }
    this.reconnectRetries = 0;
  }

  // ---------- 清理 ----------

  private rejectAllPending(reason: string) {
    for (const [, p] of this.pendingAck) {
      clearTimeout(p.timer);
      p.reject(new Error(reason));
    }
    this.pendingAck.clear();
  }
}

// ========== 单例 ==========

let _inst: IMWebSocket | null = null;

export function getIMWs(): IMWebSocket {
  if (!_inst) _inst = new IMWebSocket();
  return _inst;
}

export function createIMWs(opts: IMWebSocketOptions): IMWebSocket {
  _inst?.disconnect();
  _inst = new IMWebSocket(opts);
  return _inst;
}
