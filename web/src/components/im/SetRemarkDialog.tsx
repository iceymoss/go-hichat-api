'use client';

import React, { useState, useEffect, useRef } from 'react';
import { Input } from '@/components/ui/input';
import { X } from 'lucide-react';
import { toast } from 'sonner';

interface SetRemarkDialogProps {
  open: boolean;
  onClose: () => void;
  currentRemark?: string;
  onSave: (remark: string, tags: string[]) => void;
}

const PRESET_TAGS = ['同事', '朋友', '家人', '重要'];

export default function SetRemarkDialog({
  open,
  onClose,
  currentRemark = '',
  onSave,
}: SetRemarkDialogProps) {
  const [remark, setRemark] = useState(currentRemark);
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleToggleTag = (tag: string) => {
    setSelectedTags((prev) =>
      prev.includes(tag) ? prev.filter((t) => t !== tag) : [...prev, tag]
    );
  };

  const handleSave = () => {
    if (!remark.trim() && selectedTags.length === 0) {
      toast.info('请输入备注名或选择标签');
      return;
    }
    onSave(remark.trim(), selectedTags);
    onClose();
  };

  // Auto-focus input when dialog opens, reset form
  useEffect(() => {
    if (!open) return;
    // Use rAF to avoid synchronous setState in effect (React 19 lint rule)
    const raf = requestAnimationFrame(() => {
      setRemark(currentRemark);
      setSelectedTags([]);
    });
    const timer = setTimeout(() => inputRef.current?.focus(), 150);
    return () => {
      cancelAnimationFrame(raf);
      clearTimeout(timer);
    };
  }, [open, currentRemark]);

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
          maxWidth: 400,
          width: 'calc(100% - 2rem)',
          backgroundColor: '#FFFFFF',
          borderRadius: 14,
          padding: '24px',
          boxShadow: '0 8px 32px rgba(0,0,0,0.16)',
          border: 'none',
        }}
      >
        {/* Header */}
        <div style={{ marginBottom: 4 }}>
          <div
            style={{
              fontSize: 17,
              fontWeight: 600,
              color: '#1F2329',
            }}
          >
            设置备注
          </div>
        </div>

        {/* Close button */}
        <button
          onClick={onClose}
          style={{
            position: 'absolute',
            top: 16,
            right: 16,
            width: 28,
            height: 28,
            borderRadius: '50%',
            border: 'none',
            background: 'transparent',
            color: '#8F959E',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            transition: 'all 0.15s',
          }}
          onMouseEnter={(e) => {
            (e.currentTarget as HTMLButtonElement).style.background = 'rgba(0,0,0,0.06)';
            (e.currentTarget as HTMLButtonElement).style.color = '#646A73';
          }}
          onMouseLeave={(e) => {
            (e.currentTarget as HTMLButtonElement).style.background = 'transparent';
            (e.currentTarget as HTMLButtonElement).style.color = '#8F959E';
          }}
        >
          <X size={18} />
        </button>

        {/* Remark input */}
        <div style={{ marginTop: 16, marginBottom: 16 }}>
          <label
            style={{
              display: 'block',
              fontSize: 13,
              fontWeight: 500,
              color: '#646A73',
              marginBottom: 8,
            }}
          >
            备注名
          </label>
          <Input
            ref={inputRef}
            value={remark}
            onChange={(e) => setRemark(e.target.value)}
            placeholder="输入备注名"
            maxLength={20}
            style={{
              height: 40,
              borderRadius: 8,
              border: '1px solid rgba(0,0,0,0.1)',
              fontSize: 14,
              color: '#1F2329',
              outline: 'none',
            }}
          />
        </div>

        {/* Tags section */}
        <div style={{ marginBottom: 24 }}>
          <label
            style={{
              display: 'block',
              fontSize: 13,
              fontWeight: 500,
              color: '#646A73',
              marginBottom: 10,
            }}
          >
            标签
          </label>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
            {PRESET_TAGS.map((tag) => {
              const isSelected = selectedTags.includes(tag);
              return (
                <button
                  key={tag}
                  onClick={() => handleToggleTag(tag)}
                  style={{
                    padding: '6px 16px',
                    borderRadius: 20,
                    border: 'none',
                    background: isSelected
                      ? '#1BB45B'
                      : '#F5F7FA',
                    color: isSelected
                      ? '#FFFFFF'
                      : '#646A73',
                    fontSize: 13,
                    fontWeight: 500,
                    cursor: 'pointer',
                    transition: 'all 0.15s',
                  }}
                  onMouseEnter={(e) => {
                    if (!isSelected) {
                      (e.currentTarget as HTMLButtonElement).style.background = '#EEF0F4';
                    }
                  }}
                  onMouseLeave={(e) => {
                    if (!isSelected) {
                      (e.currentTarget as HTMLButtonElement).style.background = '#F5F7FA';
                    }
                  }}
                >
                  {tag}
                </button>
              );
            })}
          </div>
        </div>

        {/* Footer buttons */}
        <div
          style={{
            display: 'flex',
            gap: 10,
            justifyContent: 'flex-end',
          }}
        >
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
            取消
          </button>
          <button
            onClick={handleSave}
            style={{
              padding: '8px 20px',
              borderRadius: 8,
              border: 'none',
              background: '#1BB45B',
              color: '#FFFFFF',
              fontSize: 14,
              fontWeight: 600,
              cursor: 'pointer',
              transition: 'background 0.15s',
            }}
            onMouseEnter={(e) => {
              (e.currentTarget as HTMLButtonElement).style.background = '#149A4C';
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLButtonElement).style.background = '#1BB45B';
            }}
          >
            保存
          </button>
        </div>
      </div>
    </div>
  );
}
