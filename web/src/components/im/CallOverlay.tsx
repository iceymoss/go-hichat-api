'use client';

/**
 * 全局通话浮层：来电振铃 + 呼出 + 通话中。
 * 挂在 IMLayout 顶层，任意页面都能弹。状态来自 useCallStore（由 im ws call.signal + CallEngine 驱动）。
 */

import React, { useEffect, useRef, useState } from 'react';
import { Phone, PhoneOff, Mic, MicOff, Video, VideoOff, User } from 'lucide-react';
import { toast } from 'sonner';
import { useCallStore } from '@/lib/call-store';
import { getAvatarColor } from '@/lib/utils';

function fmtDuration(sec: number): string {
  const m = Math.floor(sec / 60);
  const s = sec % 60;
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
}

export function CallOverlay() {
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

  const accept = useCallStore(s => s.accept);
  const reject = useCallStore(s => s.reject);
  const hangup = useCallStore(s => s.hangup);
  const toggleMute = useCallStore(s => s.toggleMute);
  const toggleCamera = useCallStore(s => s.toggleCamera);
  const clearError = useCallStore(s => s.clearError);

  const localVideoRef = useRef<HTMLVideoElement>(null);
  const remoteVideoRef = useRef<HTMLVideoElement>(null);
  const remoteAudioRef = useRef<HTMLAudioElement>(null);
  const [elapsed, setElapsed] = useState(0);

  // 绑定媒体流到 video 元素
  useEffect(() => {
    if (localVideoRef.current && localVideoRef.current.srcObject !== localStream) {
      localVideoRef.current.srcObject = localStream;
    }
  }, [localStream, phase]);
  useEffect(() => {
    if (remoteVideoRef.current && remoteVideoRef.current.srcObject !== remoteStream) {
      remoteVideoRef.current.srcObject = remoteStream;
    }
  }, [remoteStream, phase]);
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
      toast(errorMsg);
      clearError();
    }
  }, [errorMsg, clearError]);

  if (phase === 'idle' || phase === 'ended') return null;

  const isVideo = mediaType === 'video';
  const name = peer?.name || '对方';
  const avatarColor = getAvatarColor(name);
  const showRemoteVideo = isVideo && phase === 'connected' && remoteMedia.video;

  const statusText =
    phase === 'incoming' ? `邀请你${isVideo ? '视频' : '语音'}通话`
    : phase === 'outgoing' ? '正在呼叫…'
    : phase === 'connecting' ? '连接中…'
    : fmtDuration(elapsed);

  return (
    <div
      style={{
        position: 'fixed', inset: 0, zIndex: 3000,
        background: '#1C2733',
        display: 'flex', flexDirection: 'column', alignItems: 'center',
        color: '#FFFFFF',
      }}
    >
      {/* 远端音频：始终隐藏播放（语音通话靠它出声；视频通话的 <video> 设为 muted 避免双声） */}
      <audio ref={remoteAudioRef} autoPlay playsInline style={{ display: 'none' }} />

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
            <CircleBtn color="#F5483B" onClick={reject} label="拒绝"><PhoneOff size={26} color="#FFF" /></CircleBtn>
            <CircleBtn color="#1BB45B" onClick={accept} label="接听"><Phone size={26} color="#FFF" /></CircleBtn>
          </>
        ) : (
          <>
            <CircleBtn color="rgba(255,255,255,0.15)" onClick={toggleMute} label={muted ? '取消静音' : '静音'}>
              {muted ? <MicOff size={24} color="#FFF" /> : <Mic size={24} color="#FFF" />}
            </CircleBtn>
            {isVideo && (
              <CircleBtn color="rgba(255,255,255,0.15)" onClick={toggleCamera} label={cameraOff ? '开摄像头' : '关摄像头'}>
                {cameraOff ? <VideoOff size={24} color="#FFF" /> : <Video size={24} color="#FFF" />}
              </CircleBtn>
            )}
            <CircleBtn color="#F5483B" onClick={hangup} label="挂断"><PhoneOff size={26} color="#FFF" /></CircleBtn>
          </>
        )}
      </div>
    </div>
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
