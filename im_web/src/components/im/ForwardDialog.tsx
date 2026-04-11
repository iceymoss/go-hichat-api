'use client';

import React from 'react';
import { X } from 'lucide-react';
import { conversations, type Message } from '@/lib/mock-data';
import { getAvatarColor } from '@/lib/utils';

interface ForwardDialogProps {
  message: Message;
  onClose: () => void;
  onForward: (targetConvId: string) => void;
}

export default function ForwardDialog({ message, onClose, onForward }: ForwardDialogProps) {
  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 9999,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'rgba(0,0,0,0.6)',
      }}
      onClick={onClose}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          background: '#2C3E50',
          borderRadius: 14,
          padding: '20px 20px 16px',
          width: 340,
          maxWidth: '90vw',
          maxHeight: '70vh',
          display: 'flex',
          flexDirection: 'column',
          boxShadow: '0 8px 32px rgba(0,0,0,0.4)',
          animation: 'fadeIn 0.15s ease',
        }}
      >
        {/* Header */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 4 }}>
          <h3 style={{ color: '#FFFFFF', fontSize: 16, fontWeight: 600 }}>转发消息</h3>
          <button
            onClick={onClose}
            style={{
              width: 32, height: 32, borderRadius: '50%', border: 'none',
              background: 'transparent', color: 'rgba(255,255,255,0.5)', cursor: 'pointer',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              transition: 'all 0.15s',
            }}
            onMouseEnter={(e) => {
              (e.currentTarget as HTMLButtonElement).style.background = 'rgba(255,255,255,0.1)';
              (e.currentTarget as HTMLButtonElement).style.color = '#FFFFFF';
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLButtonElement).style.background = 'transparent';
              (e.currentTarget as HTMLButtonElement).style.color = 'rgba(255,255,255,0.5)';
            }}
          >
            <X size={18} />
          </button>
        </div>

        {/* Message preview */}
        <div style={{
          color: 'rgba(255,255,255,0.5)', fontSize: 13, marginBottom: 12,
          overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
          padding: '6px 0',
        }}>
          {message.content}
        </div>

        {/* Conversation list */}
        <div style={{
          flex: 1, overflowY: 'auto', maxHeight: 320,
          borderRadius: 10, background: 'rgba(0,0,0,0.15)',
          padding: '4px 0',
        }}>
          {conversations.map((conv) => (
            <button
              key={conv.id}
              onClick={() => onForward(conv.id)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 12,
                width: '100%',
                height: 48,
                padding: '0 14px',
                border: 'none',
                borderRadius: 8,
                background: 'transparent',
                cursor: 'pointer',
                transition: 'background 0.15s',
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLButtonElement).style.background = 'rgba(255,255,255,0.08)';
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLButtonElement).style.background = 'transparent';
              }}
            >
              <div style={{
                width: 36, height: 36, borderRadius: '50%', flexShrink: 0,
                background: getAvatarColor(conv.name),
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                color: '#FFFFFF', fontSize: 14, fontWeight: 600,
              }}>
                {conv.name[0]}
              </div>
              <div style={{ flex: 1, textAlign: 'left', minWidth: 0 }}>
                <div style={{ color: '#FFFFFF', fontSize: 14, fontWeight: 500 }}>
                  {conv.name}
                </div>
              </div>
              {conv.type === 'group' && (
                <span style={{
                  fontSize: 11, color: 'rgba(255,255,255,0.4)',
                  background: 'rgba(255,255,255,0.08)',
                  padding: '2px 6px', borderRadius: 6,
                }}>
                  群聊
                </span>
              )}
            </button>
          ))}
        </div>

        {/* Cancel button */}
        <button
          onClick={onClose}
          style={{
            marginTop: 12,
            width: '100%',
            padding: '10px 0',
            borderRadius: 10,
            border: '1px solid rgba(255,255,255,0.15)',
            background: 'transparent',
            color: 'rgba(255,255,255,0.7)',
            fontSize: 14,
            fontWeight: 500,
            cursor: 'pointer',
            transition: 'background 0.15s',
          }}
          onMouseEnter={(e) => {
            (e.currentTarget as HTMLButtonElement).style.background = 'rgba(255,255,255,0.05)';
          }}
          onMouseLeave={(e) => {
            (e.currentTarget as HTMLButtonElement).style.background = 'transparent';
          }}
        >
          取消
        </button>
      </div>

      <style>{`
        @keyframes fadeIn {
          from { opacity: 0; transform: scale(0.95); }
          to { opacity: 1; transform: scale(1); }
        }
      `}</style>
    </div>
  );
}
