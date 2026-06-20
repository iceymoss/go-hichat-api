'use client';

/**
 * 全局通话浮层：来电振铃 + 呼出 + 通话中。
 * 挂在 IMLayout 顶层，任意页面都能弹。状态来自 useCallStore（由 im ws call.signal + CallEngine 驱动）。
 */

import React, { useEffect, useRef, useState } from 'react';
import { Phone, PhoneOff, Mic, MicOff, Video, VideoOff, User, Minimize2 } from 'lucide-react';
import { toast } from 'sonner';
import { useCallStore } from '@/lib/call-store';
import { getAvatarColor } from '@/lib/utils';
import { useT } from '@/hooks/use-i18n';
import { playRingback, playRingtone, stopRing } from '@/lib/ringtone';

function fmtDuration(sec: number): string {
  const m = Math.floor(sec / 60);
  const s = sec % 60;
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
}

export function CallOverlay() {
  const t = useT();
  const phase = useCallStore(s => s.phase);
  const peer = useCallStore(s => s.peer);
  const mediaType = useCallStore(s => s.mediaType);
  const localStream = useCallStore(s => s.localStream);
  const remoteStream = useCallStore(s => s.remoteStream);
  const muted = useCallStore(s => s.muted);
  const cameraOff = useCallStore(s => s.cameraOff);
  const startedAt = useCallStore(s => s.startedAt);
  const errorMsg = useCallStore(s => s.errorMsg);
  const remoteMedia = useCallStore(s => s.remoteMedia);
  const minimized = useCallStore(s => s.minimized);

  const accept = useCallStore(s => s.accept);
  const reject = useCallStore(s => s.reject);
  const hangup = useCallStore(s => s.hangup);
  const toggleMute = useCallStore(s => s.toggleMute);
  const toggleCamera = useCallStore(s => s.toggleCamera);
  const setMinimized = useCallStore(s => s.setMinimized);
  const clearError = useCallStore(s => s.clearError);

  const localVideoRef = useRef<HTMLVideoElement>(null);
  const remoteVideoRef = useRef<HTMLVideoElement>(null);
  const remoteAudioRef = useRef<HTMLAudioElement>(null);
  const [elapsed, setElapsed] = useState(0);

  // 悬浮球拖拽
  const widgetRef = useRef<HTMLDivElement>(null);
  const dragRef = useRef<{ sx: number; sy: number; ox: number; oy: number; moved: boolean } | null>(null);
  const [pos, setPos] = useState<{ x: number; y: number } | null>(null);

  // 恢复全屏后复位悬浮球位置（下次最小化回到右上角默认位）
  useEffect(() => {
    if (!minimized) setPos(null);
  }, [minimized]);

  const beginDrag = (clientX: number, clientY: number) => {
    const rect = widgetRef.current?.getBoundingClientRect();
    if (!rect) return;
    dragRef.current = { sx: clientX, sy: clientY, ox: rect.left, oy: rect.top, moved: false };
  };
  const moveDrag = (clientX: number, clientY: number) => {
    const d = dragRef.current;
    if (!d) return;
    const dx = clientX - d.sx, dy = clientY - d.sy;
    if (Math.abs(dx) > 4 || Math.abs(dy) > 4) d.moved = true;
    const w = widgetRef.current?.offsetWidth ?? 0, h = widgetRef.current?.offsetHeight ?? 0;
    const x = Math.min(Math.max(0, d.ox + dx), window.innerWidth - w);
    const y = Math.min(Math.max(0, d.oy + dy), window.innerHeight - h);
    setPos({ x, y });
  };
  const onWidgetMouseDown = (e: React.MouseEvent) => {
    beginDrag(e.clientX, e.clientY);
    const mm = (ev: MouseEvent) => moveDrag(ev.clientX, ev.clientY);
    const mu = () => { window.removeEventListener('mousemove', mm); window.removeEventListener('mouseup', mu); };
    window.addEventListener('mousemove', mm);
    window.addEventListener('mouseup', mu);
  };
  const onWidgetTouchStart = (e: React.TouchEvent) => {
    const t = e.touches[0];
    beginDrag(t.clientX, t.clientY);
    const tm = (ev: TouchEvent) => { const tt = ev.touches[0]; if (tt) moveDrag(tt.clientX, tt.clientY); };
    const te = () => { window.removeEventListener('touchmove', tm); window.removeEventListener('touchend', te); };
    window.addEventListener('touchmove', tm, { passive: true });
    window.addEventListener('touchend', te);
  };
  // 拖动过则不触发「点击放大」
  const onWidgetClickRestore = () => { if (!dragRef.current?.moved) setMinimized(false); };

  // 绑定媒体流到 video 元素（minimized 切换会换 video 元素，需重绑）
  useEffect(() => {
    if (localVideoRef.current && localVideoRef.current.srcObject !== localStream) {
      localVideoRef.current.srcObject = localStream;
    }
  }, [localStream, phase, minimized]);
  useEffect(() => {
    if (remoteVideoRef.current && remoteVideoRef.current.srcObject !== remoteStream) {
      remoteVideoRef.current.srcObject = remoteStream;
    }
  }, [remoteStream, phase, minimized]);
  // 远端音频始终由隐藏的 <audio> 播放（语音通话没有 <video> 元素，否则收到音轨也没声音）
  useEffect(() => {
    if (remoteAudioRef.current && remoteAudioRef.current.srcObject !== remoteStream) {
      remoteAudioRef.current.srcObject = remoteStream;
    }
  }, [remoteStream, phase]);

  // 计时
  useEffect(() => {
    if (phase !== 'connected' || !startedAt) { setElapsed(0); return; }
    const tick = () => setElapsed(Math.floor((Date.now() - startedAt) / 1000));
    tick();
    const id = setInterval(tick, 1000);
    return () => clearInterval(id);
  }, [phase, startedAt]);

  // 错误提示
  useEffect(() => {
    if (errorMsg) {
      toast(t(errorMsg));
      clearError();
    }
  }, [errorMsg, clearError, t]);

  // 铃声：呼出=回铃音，来电=来电铃声+振动，接通/结束即停
  useEffect(() => {
    if (phase === 'outgoing') playRingback();
    else if (phase === 'incoming') playRingtone();
    else stopRing();
    return () => stopRing();
  }, [phase]);

  if (phase === 'idle' || phase === 'ended') return null;

  const isVideo = mediaType === 'video';
  const name = peer?.name || t('call.peerFallback');
  const avatarColor = getAvatarColor(name);
  const showRemoteVideo = isVideo && phase === 'connected' && remoteMedia.video;

  const statusText =
    phase === 'incoming' ? (isVideo ? t('call.status.incomingVideo') : t('call.status.incomingVoice'))
    : phase === 'outgoing' ? t('call.status.calling')
    : phase === 'connecting' ? t('call.status.connecting')
    : fmtDuration(elapsed);

  return (
    <>
      {/* 远端音频：始终隐藏播放，最小化也不中断 */}
      <audio ref={remoteAudioRef} autoPlay playsInline style={{ display: 'none' }} />

      {minimized ? (
        /* ── 悬浮球（可拖拽），点击放大；来电时带接听/拒绝 ── */
        <div
          ref={widgetRef}
          onMouseDown={onWidgetMouseDown}
          onTouchStart={onWidgetTouchStart}
          title={name}
          style={{
            position: 'fixed', zIndex: 3000,
            ...(pos ? { left: pos.x, top: pos.y } : { top: 12, right: 12 }),
            borderRadius: 14, overflow: 'hidden', background: '#1C2733', color: '#FFF',
            boxShadow: '0 6px 24px rgba(0,0,0,0.35)', border: '1px solid rgba(255,255,255,0.12)',
            cursor: 'move', userSelect: 'none', touchAction: 'none',
          }}
        >
          <div onClick={onWidgetClickRestore} style={{ cursor: 'pointer' }}>
            {showRemoteVideo ? (
              <div style={{ position: 'relative' }}>
                <video ref={remoteVideoRef} autoPlay playsInline muted style={{ width: 120, height: 170, objectFit: 'cover', background: '#000', display: 'block' }} />
                <div style={{ position: 'absolute', bottom: 0, left: 0, right: 0, padding: '4px 6px', fontSize: 12, textAlign: 'center', background: 'linear-gradient(transparent,rgba(0,0,0,0.6))' }}>{statusText}</div>
              </div>
            ) : (
              <div style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '10px 12px', maxWidth: 200 }}>
                <div style={{ width: 36, height: 36, borderRadius: '50%', flexShrink: 0, overflow: 'hidden', background: peer?.avatar ? 'transparent' : avatarColor, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                  {peer?.avatar ? <img src={peer.avatar} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }} /> : (isVideo ? <Video size={18} color="#FFF" /> : <Phone size={18} color="#FFF" />)}
                </div>
                <div style={{ minWidth: 0 }}>
                  <div style={{ fontSize: 13, fontWeight: 600, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{name}</div>
                  <div style={{ fontSize: 12, color: 'rgba(255,255,255,0.7)' }}>{statusText}</div>
                </div>
              </div>
            )}
          </div>
          {phase === 'incoming' && (
            <div style={{ display: 'flex', borderTop: '1px solid rgba(255,255,255,0.1)' }}>
              <button onClick={reject} aria-label={t('call.action.reject')} style={{ flex: 1, height: 40, border: 'none', background: '#F5483B', color: '#FFF', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <PhoneOff size={18} />
              </button>
              <button onClick={accept} aria-label={t('call.action.accept')} style={{ flex: 1, height: 40, border: 'none', background: '#1BB45B', color: '#FFF', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Phone size={18} />
              </button>
            </div>
          )}
        </div>
      ) : (
        /* ── 全屏通话界面 ── */
        <div
          style={{
            position: 'fixed', inset: 0, zIndex: 3000,
            background: '#1C2733',
            display: 'flex', flexDirection: 'column', alignItems: 'center',
            color: '#FFFFFF',
          }}
        >
          {/* 最小化按钮（含来电响铃页；缩小后悬浮球带接听/拒绝） */}
          <button
            onClick={() => setMinimized(true)}
            aria-label="minimize"
            style={{
              position: 'absolute', top: 16, left: 16, zIndex: 4,
              width: 40, height: 40, borderRadius: '50%', border: 'none',
              background: 'rgba(255,255,255,0.15)', color: '#FFF', cursor: 'pointer',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
            }}
          >
            <Minimize2 size={20} />
          </button>

          {/* 远端视频铺满 */}
          {showRemoteVideo && (
            <video
              ref={remoteVideoRef}
              autoPlay
              playsInline
              muted
              style={{ position: 'absolute', inset: 0, width: '100%', height: '100%', objectFit: 'cover', background: '#000' }}
            />
          )}

          {/* 本地视频小窗（视频通话且已开摄像头） */}
          {isVideo && !cameraOff && (phase === 'connected' || phase === 'connecting' || phase === 'outgoing') && (
            <video
              ref={localVideoRef}
              autoPlay
              playsInline
              muted
              style={{
                position: 'absolute', top: 20, right: 16, width: 110, height: 160,
                objectFit: 'cover', borderRadius: 12, background: '#000',
                border: '1px solid rgba(255,255,255,0.2)', zIndex: 2,
              }}
            />
          )}

          {/* 头像 + 名称 + 状态（语音通话 或 视频未出图时） */}
          {!showRemoteVideo && (
            <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', marginTop: '22vh', zIndex: 1 }}>
              <div
                style={{
                  width: 96, height: 96, borderRadius: '50%', overflow: 'hidden',
                  background: peer?.avatar ? 'transparent' : avatarColor,
                  display: 'flex', alignItems: 'center', justifyContent: 'center', marginBottom: 18,
                }}
              >
                {peer?.avatar
                  ? <img src={peer.avatar} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                  : <User size={40} color="#FFFFFF" />}
              </div>
              <div style={{ fontSize: 22, fontWeight: 600, marginBottom: 8 }}>{name}</div>
              <div style={{ fontSize: 14, color: 'rgba(255,255,255,0.7)' }}>{statusText}</div>
            </div>
          )}

          {/* 视频出图时，状态/计时浮在顶部 */}
          {showRemoteVideo && (
            <div style={{ position: 'absolute', top: 24, left: 0, right: 0, textAlign: 'center', zIndex: 2, textShadow: '0 1px 4px rgba(0,0,0,0.6)' }}>
              <div style={{ fontSize: 17, fontWeight: 600 }}>{name}</div>
              <div style={{ fontSize: 13, color: 'rgba(255,255,255,0.85)' }}>{statusText}</div>
            </div>
          )}

          {/* 底部控制区 */}
          <div style={{ position: 'absolute', bottom: 48, left: 0, right: 0, display: 'flex', justifyContent: 'center', alignItems: 'center', gap: 28, zIndex: 3 }}>
            {phase === 'incoming' ? (
              <>
                <CircleBtn color="#F5483B" onClick={reject} label={t('call.action.reject')}><PhoneOff size={26} color="#FFF" /></CircleBtn>
                <CircleBtn color="#1BB45B" onClick={accept} label={t('call.action.accept')}><Phone size={26} color="#FFF" /></CircleBtn>
              </>
            ) : (
              <>
                <CircleBtn color="rgba(255,255,255,0.15)" onClick={toggleMute} label={muted ? t('call.action.unmute') : t('call.action.mute')}>
                  {muted ? <MicOff size={24} color="#FFF" /> : <Mic size={24} color="#FFF" />}
                </CircleBtn>
                {isVideo && (
                  <CircleBtn color="rgba(255,255,255,0.15)" onClick={toggleCamera} label={cameraOff ? t('call.action.cameraOn') : t('call.action.cameraOff')}>
                    {cameraOff ? <VideoOff size={24} color="#FFF" /> : <Video size={24} color="#FFF" />}
                  </CircleBtn>
                )}
                <CircleBtn color="#F5483B" onClick={hangup} label={t('call.action.hangup')}><PhoneOff size={26} color="#FFF" /></CircleBtn>
              </>
            )}
          </div>
        </div>
      )}
    </>
  );
}

function CircleBtn({ color, onClick, label, children }: { color: string; onClick: () => void; label: string; children: React.ReactNode }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 8 }}>
      <button
        onClick={onClick}
        aria-label={label}
        style={{
          width: 60, height: 60, borderRadius: '50%', border: 'none', cursor: 'pointer',
          background: color, display: 'flex', alignItems: 'center', justifyContent: 'center',
          transition: 'transform 0.1s',
        }}
        onMouseDown={(e) => { (e.currentTarget as HTMLButtonElement).style.transform = 'scale(0.92)'; }}
        onMouseUp={(e) => { (e.currentTarget as HTMLButtonElement).style.transform = 'scale(1)'; }}
        onMouseLeave={(e) => { (e.currentTarget as HTMLButtonElement).style.transform = 'scale(1)'; }}
      >
        {children}
      </button>
      <span style={{ fontSize: 12, color: 'rgba(255,255,255,0.75)' }}>{label}</span>
    </div>
  );
}
