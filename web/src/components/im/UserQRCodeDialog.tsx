'use client';

import React, { useEffect, useRef, useState } from 'react';
import QRCode from 'qrcode';
import { X, Download, Copy } from 'lucide-react';
import { toast } from 'sonner';
import { getAvatarColor } from '@/lib/utils';

interface UserQRCodeDialogProps {
  open: boolean;
  onClose: () => void;
  userId: string;
  userName: string;
  userAvatar: string;
  userIntroduction?: string;
}

export default function UserQRCodeDialog({
  open, onClose, userId, userName, userAvatar, userIntroduction,
}: UserQRCodeDialogProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [dataUrl, setDataUrl] = useState('');

  useEffect(() => {
    if (!open || !canvasRef.current) return;

    const payload = JSON.stringify({ type: 'user', id: `hichat_${userId}` });

    QRCode.toCanvas(canvasRef.current, payload, {
      width: 200,
      margin: 2,
      color: { dark: '#1C2733', light: '#FFFFFF' },
      errorCorrectionLevel: 'M',
    }, (err) => {
      if (err) return;
      // Generate data URL for download
      setDataUrl(canvasRef.current?.toDataURL('image/png') || '');
    });
  }, [open, userId]);

  useEffect(() => {
    if (!open) return;
    const h = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose(); };
    window.addEventListener('keydown', h);
    return () => window.removeEventListener('keydown', h);
  }, [open, onClose]);

  if (!open) return null;

  const avatarColor = getAvatarColor(userName);

  const handleDownload = () => {
    if (!dataUrl) return;
    const a = document.createElement('a');
    a.href = dataUrl;
    a.download = `HiChat_${userName}_QR.png`;
    a.click();
    toast.success('二维码已保存');
  };

  const handleCopyId = () => {
    navigator.clipboard.writeText(`hichat_${userId}`);
    toast.success('HiChat ID 已复制');
  };

  return (
    <div
      onClick={onClose}
      style={{
        position: 'fixed', inset: 0, zIndex: 1000,
        background: 'rgba(0,0,0,0.5)', backdropFilter: 'blur(4px)',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        animation: 'fade-in 0.15s ease-out',
      }}
    >
      <div
        onClick={e => e.stopPropagation()}
        style={{
          width: 340, maxWidth: 'calc(100vw - 32px)',
          background: '#fff', borderRadius: 20,
          boxShadow: '0 12px 48px rgba(0,0,0,0.2)',
          overflow: 'hidden',
        }}
      >
        {/* Header gradient */}
        <div style={{
          height: 80,
          background: 'linear-gradient(135deg, #1BB45B 0%, #1BB45B 50%, #149a4c 100%)',
          position: 'relative',
        }}>
          {/* Close button */}
          <button onClick={onClose} style={{
            position: 'absolute', top: 12, right: 12,
            width: 28, height: 28, borderRadius: '50%', border: 'none',
            background: 'rgba(255,255,255,0.2)', color: '#fff',
            cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center',
            transition: 'background 0.15s',
          }}
            onMouseEnter={e => { e.currentTarget.style.background = 'rgba(255,255,255,0.35)'; }}
            onMouseLeave={e => { e.currentTarget.style.background = 'rgba(255,255,255,0.2)'; }}
          ><X size={14} /></button>

          {/* Decorative circles */}
          <div style={{ position: 'absolute', width: 120, height: 120, top: -40, right: -20, borderRadius: '50%', background: 'rgba(255,255,255,0.06)', pointerEvents: 'none' }} />
          <div style={{ position: 'absolute', width: 80, height: 80, bottom: -30, left: 20, borderRadius: '50%', background: 'rgba(255,255,255,0.04)', pointerEvents: 'none' }} />
        </div>

        {/* Avatar (overlapping header) */}
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', marginTop: -36, position: 'relative', zIndex: 1 }}>
          <div style={{
            width: 72, height: 72, borderRadius: 18,
            background: userAvatar ? '#fff' : avatarColor,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            color: '#fff', fontSize: 28, fontWeight: 600,
            border: '3px solid #fff', boxShadow: '0 2px 12px rgba(0,0,0,0.12)',
            overflow: 'hidden',
          }}>
            {userAvatar
              ? <img src={userAvatar} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
              : userName[0]
            }
          </div>

          {/* Name */}
          <div style={{ fontSize: 18, fontWeight: 600, color: '#1C2733', marginTop: 10 }}>{userName}</div>

          {/* ID */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 4, marginTop: 4 }}>
            <span style={{ fontSize: 13, color: '#646A73' }}>HiChat ID: hichat_{userId}</span>
            <button onClick={handleCopyId} style={{
              background: 'none', border: 'none', cursor: 'pointer', color: '#A2ACB5', padding: 2, display: 'flex',
              transition: 'color 0.15s',
            }}
              onMouseEnter={e => { (e.currentTarget as HTMLButtonElement).style.color = '#1BB45B'; }}
              onMouseLeave={e => { (e.currentTarget as HTMLButtonElement).style.color = '#A2ACB5'; }}
            ><Copy size={12} /></button>
          </div>

          {/* Introduction */}
          {userIntroduction && (
            <div style={{ fontSize: 12, color: '#8F959E', marginTop: 4, maxWidth: 240, textAlign: 'center', lineHeight: 1.4 }}>
              {userIntroduction}
            </div>
          )}
        </div>

        {/* QR Code */}
        <div style={{
          display: 'flex', flexDirection: 'column', alignItems: 'center',
          padding: '20px 0 12px',
        }}>
          <div style={{
            padding: 12, background: '#fff', borderRadius: 12,
            border: '1px solid rgba(0,0,0,0.06)',
            boxShadow: '0 1px 4px rgba(0,0,0,0.04)',
          }}>
            <canvas ref={canvasRef} style={{ display: 'block' }} />
          </div>

          <div style={{ fontSize: 12, color: '#A2ACB5', marginTop: 10 }}>
            扫一扫，加我为好友
          </div>
        </div>

        {/* Download button */}
        <div style={{ padding: '0 24px 20px' }}>
          <button onClick={handleDownload} style={{
            width: '100%', height: 40, borderRadius: 20, border: '1.5px solid #1BB45B',
            background: '#fff', color: '#1BB45B', fontSize: 14, fontWeight: 500,
            cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 6,
            transition: 'all 0.2s',
          }}
            onMouseEnter={e => { e.currentTarget.style.background = 'rgba(27,180,91,0.06)'; }}
            onMouseLeave={e => { e.currentTarget.style.background = '#fff'; }}
          >
            <Download size={15} />
            保存二维码
          </button>
        </div>
      </div>
    </div>
  );
}
