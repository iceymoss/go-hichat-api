'use client';

import React, { useState, useMemo } from 'react';
import { Dialog, DialogContent } from '@/components/ui/dialog';
import { Phone, Video, Check, Users } from 'lucide-react';
import { getAvatarColor } from '@/lib/utils';
import type { GroupMember } from '@/lib/mock-data';

interface CallDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  type: 'voice' | 'video';
  contactName: string;
  isGroup?: boolean;
  members?: GroupMember[];
}

export function CallDialog({ open, onOpenChange, type, contactName, isGroup = false, members = [] }: CallDialogProps) {
  const isVoice = type === 'voice';
  const [selectAll, setSelectAll] = useState(true);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set(members.map(m => m.id)));

  // When dialog opens (or members change), reset selection
  const effectiveMembers = useMemo(() => members, [members]);
  const allSelected = selectedIds.size === effectiveMembers.length && effectiveMembers.length > 0;

  const handleToggleAll = () => {
    if (selectAll || allSelected) {
      setSelectAll(false);
      setSelectedIds(new Set());
    } else {
      setSelectAll(true);
      setSelectedIds(new Set(effectiveMembers.map(m => m.id)));
    }
  };

  const handleToggleMember = (id: string) => {
    setSelectedIds(prev => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
    setSelectAll(false);
  };

  const selectedCount = selectedIds.size;
  const avatarColor = getAvatarColor(contactName);

  const handleCall = () => {
    // TODO: implement actual call logic
    onOpenChange(false);
  };

  // Reset state when dialog opens
  const handleOpenChange = (v: boolean) => {
    if (v) {
      setSelectAll(true);
      setSelectedIds(new Set(effectiveMembers.map(m => m.id)));
    }
    onOpenChange(v);
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent
        showCloseButton={false}
        className="!border-none !bg-transparent !p-0 !shadow-none !gap-0 data-[state=open]:!animate-none data-[state=closed]:!animate-none"
      >
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            maxWidth: isGroup ? 380 : 340,
            width: '100%',
            backgroundColor: '#FFFFFF',
            borderRadius: 16,
            boxShadow: '0 8px 40px rgba(0,0,0,0.15)',
            overflow: 'hidden',
          }}
        >
          {/* ── Header: Avatar + Title ── */}
          <div
            style={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              padding: '28px 24px 20px',
            }}
          >
            <div
              style={{
                width: 72,
                height: 72,
                borderRadius: '50%',
                backgroundColor: avatarColor,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                marginBottom: 16,
              }}
            >
              {isGroup ? (
                <Users size={30} color="#FFFFFF" />
              ) : isVoice ? (
                <Phone size={30} color="#FFFFFF" />
              ) : (
                <Video size={30} color="#FFFFFF" />
              )}
            </div>
            <div style={{ fontSize: 18, fontWeight: 600, color: '#1C2733', marginBottom: 4 }}>
              {isGroup
                ? `${isVoice ? '语音通话' : '视频通话'} - ${contactName}`
                : `${isVoice ? '语音通话' : '视频通话'}`}
            </div>
            {!isGroup && (
              <div style={{ fontSize: 13, color: '#708499' }}>
                向 {contactName} 发起{isVoice ? '语音' : '视频'}通话？
              </div>
            )}
            {isGroup && (
              <div style={{ fontSize: 13, color: '#708499' }}>
                选择要邀请的成员
              </div>
            )}
          </div>

          {/* ── Member Selection (Group only) ── */}
          {isGroup && effectiveMembers.length > 0 && (
            <div style={{ padding: '0 16px', maxHeight: 260, overflowY: 'auto' }}>
              {/* Select All option */}
              <button
                onClick={handleToggleAll}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 12,
                  width: '100%',
                  height: 44,
                  padding: '0 12px',
                  border: 'none',
                  borderRadius: 10,
                  background: (selectAll || allSelected) ? 'rgba(51,144,236,0.06)' : 'transparent',
                  cursor: 'pointer',
                  transition: 'background 0.2s',
                  marginBottom: 4,
                }}
              >
                {/* Avatar circle with check */}
                <div style={{ position: 'relative', width: 36, height: 36, flexShrink: 0 }}>
                  <div style={{
                    width: 36, height: 36, borderRadius: '50%',
                    background: '#3390EC',
                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                  }}>
                    <Users size={16} color="#FFFFFF" />
                  </div>
                  {(selectAll || allSelected) && (
                    <div style={{
                      position: 'absolute', bottom: -2, right: -2,
                      width: 18, height: 18, borderRadius: '50%',
                      background: '#3390EC', border: '2px solid #FFFFFF',
                      display: 'flex', alignItems: 'center', justifyContent: 'center',
                    }}>
                      <Check size={10} color="#FFFFFF" strokeWidth={3} />
                    </div>
                  )}
                </div>
                <div style={{ flex: 1, textAlign: 'left' }}>
                  <div style={{ fontSize: 14, fontWeight: 600, color: '#1C2733' }}>
                    所有成员
                  </div>
                  <div style={{ fontSize: 11, color: '#A2ACB5', marginTop: 1 }}>
                    共 {effectiveMembers.length} 人
                  </div>
                </div>
                {(selectAll || allSelected) && (
                  <span style={{
                    fontSize: 11, color: '#3390EC', fontWeight: 600,
                    background: 'rgba(51,144,236,0.08)',
                    padding: '2px 8px', borderRadius: 10,
                  }}>
                    已选 {effectiveMembers.length}
                  </span>
                )}
              </button>

              {/* Individual member list */}
              {effectiveMembers.map((member) => {
                const isSelected = selectedIds.has(member.id);
                return (
                  <button
                    key={member.id}
                    onClick={() => handleToggleMember(member.id)}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 12,
                      width: '100%',
                      height: 44,
                      padding: '0 12px',
                      border: 'none',
                      borderRadius: 10,
                      background: isSelected ? 'rgba(51,144,236,0.04)' : 'transparent',
                      cursor: 'pointer',
                      transition: 'background 0.15s',
                    }}
                  >
                    <div style={{ position: 'relative', width: 36, height: 36, flexShrink: 0 }}>
                      <div style={{
                        width: 36, height: 36, borderRadius: '50%',
                        background: getAvatarColor(member.name),
                        display: 'flex', alignItems: 'center', justifyContent: 'center',
                        color: '#FFFFFF', fontSize: 14, fontWeight: 600,
                        opacity: isSelected ? 1 : 0.5,
                        transition: 'opacity 0.15s',
                      }}>
                        {member.name[0]}
                      </div>
                      {isSelected && (
                        <div style={{
                          position: 'absolute', bottom: -2, right: -2,
                          width: 18, height: 18, borderRadius: '50%',
                          background: '#3390EC', border: '2px solid #FFFFFF',
                          display: 'flex', alignItems: 'center', justifyContent: 'center',
                        }}>
                          <Check size={10} color="#FFFFFF" strokeWidth={3} />
                        </div>
                      )}
                    </div>
                    <div style={{ flex: 1, textAlign: 'left', minWidth: 0 }}>
                      <div style={{
                        fontSize: 14, fontWeight: 500, color: isSelected ? '#1C2733' : '#A2ACB5',
                        transition: 'color 0.15s',
                      }}>
                        {member.name}
                      </div>
                    </div>
                    {/* Online indicator */}
                    <div style={{
                      width: 8, height: 8, borderRadius: '50%',
                      background: member.online ? '#4DCD5E' : '#C8D1DA',
                      flexShrink: 0,
                    }} />
                  </button>
                );
              })}
            </div>
          )}

          {/* ── Bottom: Action Buttons ── */}
          <div
            style={{
              display: 'flex',
              gap: 12,
              padding: '16px 24px 24px',
              borderTop: isGroup ? '1px solid rgba(0,0,0,0.06)' : 'none',
            }}
          >
            <button
              onClick={() => onOpenChange(false)}
              style={{
                flex: 1,
                height: 44,
                borderRadius: 12,
                border: '1px solid rgba(0,0,0,0.08)',
                backgroundColor: 'transparent',
                color: '#708499',
                fontSize: 14,
                fontWeight: 500,
                cursor: 'pointer',
                transition: 'all 0.2s',
              }}
            >
              取消
            </button>
            <button
              onClick={handleCall}
              disabled={isGroup && selectedCount === 0}
              style={{
                flex: 1,
                height: 44,
                borderRadius: 12,
                border: 'none',
                backgroundColor: isGroup && selectedCount === 0 ? '#C8D1DA' : '#3390EC',
                color: '#FFFFFF',
                fontSize: 14,
                fontWeight: 600,
                cursor: isGroup && selectedCount === 0 ? 'not-allowed' : 'pointer',
                transition: 'all 0.2s',
                opacity: isGroup && selectedCount === 0 ? 0.7 : 1,
              }}
              onMouseEnter={(e) => {
                if (!(isGroup && selectedCount === 0)) {
                  (e.currentTarget as HTMLButtonElement).style.backgroundColor = '#2B7FD4';
                }
              }}
              onMouseLeave={(e) => {
                if (!(isGroup && selectedCount === 0)) {
                  (e.currentTarget as HTMLButtonElement).style.backgroundColor = '#3390EC';
                }
              }}
            >
              {isGroup
                ? selectedCount > 0 ? `呼叫 (${selectedCount}人)` : '请选择成员'
                : '呼叫'}
            </button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
