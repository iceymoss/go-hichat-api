'use client';

import React, { useEffect } from 'react';

interface ConfirmDialogProps {
  open: boolean;
  onClose: () => void;
  title: string;
  description: string;
  confirmText?: string;
  cancelText?: string;
  confirmVariant?: 'danger' | 'default';
  /** 隐藏取消按钮，退化为只有一个确认按钮的纯提示弹框（如「知道了」） */
  hideCancel?: boolean;
  onConfirm: () => void;
}

export default function ConfirmDialog({
  open,
  onClose,
  title,
  description,
  confirmText = '确认',
  cancelText = '取消',
  confirmVariant = 'danger',
  hideCancel = false,
  onConfirm,
}: ConfirmDialogProps) {
  // Close on Escape
  useEffect(() => {
    if (!open) return;
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', handleKey);
    return () => window.removeEventListener('keydown', handleKey);
  }, [open, onClose]);

  if (!open) return null;

  const handleConfirm = () => {
    onConfirm();
    onClose();
  };

  const confirmBg = confirmVariant === 'danger' ? '#E53935' : '#3390EC';
  const confirmHoverBg = confirmVariant === 'danger' ? '#C62828' : '#2A7BD6';

  return (
    <div
      className="animate-fade-in"
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 10001,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      {/* Overlay */}
      <div
        style={{
          position: 'absolute',
          inset: 0,
          backgroundColor: 'rgba(0,0,0,0.5)',
        }}
        onClick={onClose}
      />

      {/* Dialog content — centered */}
      <div
        style={{
          position: 'relative',
          zIndex: 10002,
          maxWidth: 380,
          width: 'calc(100% - 2rem)',
          backgroundColor: '#FFFFFF',
          borderRadius: 14,
          padding: '24px',
          boxShadow: '0 8px 32px rgba(0,0,0,0.16)',
          border: 'none',
        }}
      >
        {/* Title */}
        <div
          style={{
            fontSize: 17,
            fontWeight: 600,
            color: '#1F2329',
          }}
        >
          {title}
        </div>

        {/* Description */}
        <div
          style={{
            fontSize: 14,
            color: '#646A73',
            lineHeight: 1.6,
            marginTop: 8,
          }}
        >
          {description}
        </div>

        {/* Footer buttons */}
        <div
          style={{
            display: 'flex',
            gap: 10,
            marginTop: 20,
            justifyContent: 'flex-end',
          }}
        >
          {!hideCancel && (
            <button
              onClick={onClose}
              style={{
                padding: '8px 20px',
                borderRadius: 8,
                border: '1px solid rgba(0,0,0,0.1)',
                background: '#FFFFFF',
                color: '#646A73',
                fontSize: 14,
                fontWeight: 500,
                cursor: 'pointer',
                transition: 'all 0.15s',
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLButtonElement).style.background = '#F5F7FA';
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLButtonElement).style.background = '#FFFFFF';
              }}
            >
              {cancelText}
            </button>
          )}
          <button
            onClick={handleConfirm}
            style={{
              padding: '8px 20px',
              borderRadius: 8,
              border: 'none',
              background: confirmBg,
              color: '#FFFFFF',
              fontSize: 14,
              fontWeight: 600,
              cursor: 'pointer',
              transition: 'background 0.15s',
            }}
            onMouseEnter={(e) => {
              (e.currentTarget as HTMLButtonElement).style.background = confirmHoverBg;
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLButtonElement).style.background = confirmBg;
            }}
          >
            {confirmText}
          </button>
        </div>
      </div>
    </div>
  );
}
