# Task ID: 8 - frontend-styling-expert

## Task: Update ProfilePage.tsx to Telegram Web style

### Changes Made:

#### ProfilePage.tsx (complete rewrite)
- Removed all WeChat green (#07C160) → TG blue (#3390EC)
- Profile card: clean white, no gradient, enlarged avatar (64px) with blue ring, bold name, blue HiChat ID
- Menu sections: white cards with `shadow-sm`, blue icon backgrounds `rgba(51,144,236,0.1)`, inline dividers
- Settings section: blue accent icons, 深色模式 with "跟随系统" value
- Tools section: same clean style
- Bottom buttons: outlined switch account with icon, red destructive logout with icon
- All critical styles use inline `style={{}}`

#### globals.css
- Updated `.im-profile-menu-item`: padding 11px 14px, blue hover tint, no border-bottom

### Result:
- 0 errors, 1 false-positive warning (lucide `Image` icon name)
- All existing logic preserved unchanged
- Consistent TG blue (#3390EC) color scheme
