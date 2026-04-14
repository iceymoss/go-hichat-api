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
  msg: {
    mType: number;
    content: string;
    readRecords?: Record<string, string>;
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
  onError?: (err: unknown) => void;
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

  /** 发送业务消息 (RigorAck 三次通信)，Promise 在 ACK 完成后 resolve */
  send(method: string, data: unknown): Promise<void> {
    const id = genMsgId();
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
      case FrameType.Err:
        console.error('[WS] server error:', msg.data);
        this.onError?.(msg.data);
        break;
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
    if (this.reconnectRetries >= this.reconnectMaxRetries) return;
    this.reconnectRetries++;
    const delay = Math.min(this.reconnectInterval * 2 ** (this.reconnectRetries - 1), 30_000);
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
