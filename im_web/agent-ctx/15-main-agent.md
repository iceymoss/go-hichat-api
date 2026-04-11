Task ID: 15
Agent: Main Agent
Task: Implement user profile card "more actions" menu with sub-dialogs

Work Log:
- Created `src/components/im/ProfileActionMenu.tsx` — dropdown menu component:
  - 200px wide, 44px per item, white bg, rounded corners, subtle shadow
  - Data-driven menu items via config array with 3 groupings (friends, stranger, blocked)
  - Friends: 设置备注和标签, 把他推荐给朋友, 朋友权限, 加入/移出黑名单, divider, 投诉, 删除好友(red)
  - Strangers: 添加好友(highlighted), divider, 加入黑名单, 投诉
  - Hover: #F5F7FA normal, #FFF0F0 danger items
  - Danger items: #FF5252 text color
  - Position computed from click event coords, clamped to viewport
  - Invisible overlay for click-outside, Escape key to close
  - z-index: 9999, animate-fade-in class
  - All Lucide icons (Pencil, Share2, ShieldCheck, Ban, Flag, Trash2, UserPlus)

- Created `src/components/im/SetRemarkDialog.tsx` — remark/tag setting dialog:
  - Uses shadcn Dialog with custom styled content (border-radius 14, no border)
  - Remark name input (pre-filled with current remark/name)
  - Tag section with 4 preset pills (同事, 朋友, 家人, 重要) — toggle on/off
  - Selected tags get #3390EC blue background, unselected #F5F7FA
  - Cancel + Save buttons, onSave(remark, tags) callback
  - State resets when dialog opens

- Created `src/components/im/ConfirmDialog.tsx` — reusable confirmation dialog:
  - Uses shadcn AlertDialog with custom styled content
  - Configurable title, description, confirm/cancel text
  - confirmVariant: 'danger' (#E53935 red) or 'default' (#3390EC blue)
  - onConfirm callback with auto-close

- Updated `src/components/im/UserProfileCard.tsx` — integrated ActionMenu:
  - Added state: showMenu, menuPos, localBlocked, showSetRemark, showDeleteConfirm, showBlockConfirm
  - "..." button now opens ProfileActionMenu (position computed from click event)
  - All menu actions wired: SetRemark → SetRemarkDialog, Delete → ConfirmDialog(danger), Block → ConfirmDialog(default)
  - Others (recommend, permissions, report) fall through to parent props or show toast
  - New optional props: isBlocked, onRecommend, onSetPermissions, onToggleBlock, onReport, onDeleteFriend
  - Kept backward-compatible onSettings prop

- Updated `src/components/im/FloatingProfileCard.tsx` — passes new props
- Updated `src/components/im/ContactDetailPanel.tsx` — passes new props

Stage Summary:
- Complete "more actions" dropdown menu system for user profile card
- 3 new components: ProfileActionMenu, SetRemarkDialog, ConfirmDialog
- 3 updated components: UserProfileCard, FloatingProfileCard, ContactDetailPanel
- All menu items functional with appropriate dialogs and toasts
- Zero lint errors, clean compilation
