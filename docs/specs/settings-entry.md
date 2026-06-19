# 侧边栏「设置」入口（齿轮 Popover）

## 状态
- 创建日期: 2026/06/19
- 状态: 草稿
- 类型: 前端（web/）信息架构调整 + 接通死按钮，**无后端 schema 变更**

## 目标
把左侧侧边栏底部那个**目前点了没反应的齿轮按钮**接通，作为"系统/应用设置"的统一入口；同时把散落在「我的」Tab 底部的设置/工具/账号项收拢过来，理清"系统设置（齿轮）"与"个人内容（我的）"的边界。

## 背景（现状）
- 齿轮按钮 `web/src/components/im/IMLayout.tsx:337` 是死按钮，无 `onClick`。
- 完整的 `SettingsPage`（`web/src/components/im/SettingsPage.tsx`）**已存在并可用**，含 5 个分组：账号与安全 / 通知 / 隐私 / 深色模式 / 通用。数据走 `settings-store.ts` ＋ 后端 `/api/v1/user/settings`（存一个 JSON 字符串 blob，GET/PUT 已就绪，见 `src/app/api/user/settings/route.ts`）。
- 该 SettingsPage 目前**只能从「我的」Tab → 设置入口**（`meSubPage==='settings'`）进入。
- 「我的」Tab（`ProfilePage.tsx`）底部还散落："设置"入口、工具区（备份/帮助/关于/插件）、切换账号、退出登录——这些**不在** SettingsPage 里。
- i18n 文案在 `web/src/lib/i18n.ts`（TS 字典，非 locales JSON）。

## 非目标
- 不改后端任何接口 / 数据库表（设置仍是 `/api/v1/user/settings` 的自由 JSON blob）。
- 不实现真正的"检查更新"OTA（版本号前端写死）。
- 不实现真正的多账号"切换账号"逻辑（沿用现有占位行为）。
- 不重做 SettingsPage 已有的 5 个分组内容。
- 不动「我的」Tab 的个人内容（资料卡、服务/钱包、我的朋友圈、收藏/相册/卡包/表情）。

## 用户故事
作为已登录用户，我想点击侧边栏底部的齿轮，弹出一个小菜单快速进入设置、查看关于版本、获取帮助、或切换/退出账号，以便不必再去「我的」Tab 里翻找这些系统级操作。

## 核心流程

### A. 齿轮 Popover（主流程）
1. 用户点击侧边栏底部齿轮 → 在齿轮**上方**弹出 Popover 小菜单。
2. 菜单项（最终确认）：
   - **设置** → 打开完整 `SettingsPage`（右侧大面板，含已有 5 分组）
   - **关于 HiChat** → 打开"关于"页（产品介绍 + 版本号 + 检查更新）
   - **帮助与反馈** → 帮助入口（MVP 占位/静态）
   - ── 分隔线 ──
   - **切换账号**（沿用现有占位）
   - **退出登录**（danger，调用 `logout()`）
3. 点击菜单外区域 / Esc → 关闭 Popover。

### B. 进入设置大面板
1. Popover 点"设置" → 关闭 Popover，渲染 `SettingsPage`。
2. 由于齿轮独立于「我的」Tab，需要一个**不依赖 `activeTab==='me'`** 的渲染开关（见技术设计）。
3. SettingsPage 内部子页（账号安全/通知/…）行为不变。返回 → 关闭设置面板，回到原来的 Tab。

### C. 关于页（纯前端）
1. 展示：Logo、产品名、版本号（写死常量）、`检查更新`按钮。
2. 点"检查更新" → toast 提示"已是最新版本"。
3. 底部：用户协议 / 隐私政策链接（MVP 可为占位链接）。

### D. 「我的」Tab 信息架构搬家（彻底搬家）
从 `ProfilePage.tsx` **移除**以下，统一由齿轮体系承载：
- 设置入口（行 177-192）
- 工具区：备份 / 帮助 / 关于 / 插件（行 197-235）
- 切换账号 / 退出登录按钮（行 240-264）

「我的」保留：资料卡、服务/钱包、我的朋友圈、收藏/相册/卡包/表情。

> 备份 / 插件 的归属见「待定事项」。

## 异常处理
| 场景 | 处理方式 |
|------|---------|
| Popover 打开时切换 Tab / 点击别处 | 关闭 Popover |
| 设置面板打开时点齿轮再点设置 | 幂等，仍打开设置面板（不叠加） |
| 退出登录失败 | 沿用现有 `logout()` 行为，不新增逻辑 |
| 关于页"检查更新"（无后端） | 固定 toast "已是最新版本" |
| 未登录态 | 该入口仅在已登录布局（IMLayout）出现，无需额外判空 |

## 技术设计

### 数据模型
无新增。设置仍走 `/api/v1/user/settings`（JSON blob）。版本号为前端常量（如 `APP_VERSION`，写在常量文件或读 `package.json` version）。

### 状态管理（im-store）
新增一个与 Tab 解耦的设置面板开关，例如：
- `settingsOpen: boolean`
- `openSettings()` / `closeSettings()`

齿轮"设置"→ `openSettings()`；`SettingsPage` 作为右侧面板/覆盖层在 `settingsOpen` 为真时渲染（不再依赖 `meSubPage==='settings'`）。

> 复用 vs 双入口：本期 `meSubPage==='settings'` 路径随「我的」搬家一并移除，避免两套入口（决策：彻底搬家）。

### 组件
| 组件 | 文件 | 说明 |
|------|------|------|
| 齿轮 Popover | 新增 `web/src/components/im/SettingsMenuPopover.tsx` | 锚定齿轮上方的小菜单 |
| 关于页 | 新增 `web/src/components/im/AboutPage.tsx` | Logo/版本/检查更新/协议链接 |
| 设置面板 | 复用 `SettingsPage.tsx` | 改为由 `settingsOpen` 驱动渲染 |
| 侧边栏 | 改 `IMLayout.tsx` | 齿轮按钮加 `onClick` 触发 Popover；新增 `settingsOpen`/About 的渲染 |
| 我的 | 改 `ProfilePage.tsx` | 移除设置/工具/账号项 |

### i18n（`src/lib/i18n.ts`，中英都加）
新增键（示意）：`menu.settings`、`menu.about`、`menu.help`、`menu.switchAccount`、`menu.logout`、`about.version`、`about.checkUpdate`、`about.upToDate`、`about.terms`、`about.privacy`。复用已有 `settings.*` / `profile.logout` 等。

### 实现步骤（每步可独立 commit）
1. [ ] i18n：在 `i18n.ts` 加 Popover / 关于页文案（zh + en）
2. [ ] im-store：加 `settingsOpen` + `openSettings/closeSettings`
3. [ ] `SettingsMenuPopover.tsx`：齿轮上方 Popover 菜单
4. [ ] `IMLayout.tsx`：齿轮 `onClick` 接 Popover；`settingsOpen` 渲染 `SettingsPage`
5. [ ] `AboutPage.tsx`：版本写死 + 检查更新 toast + 协议占位链接
6. [ ] `ProfilePage.tsx`：移除设置入口 / 工具区 / 切换账号 / 退出登录
7. [ ] 清理 `meSubPage==='settings'` 旧渲染分支（IMLayout）

### 参考的现有模式
- `SettingsPage.tsx` — 子页导航 / `PageHeader` / `MenuItem` / `RadioRow` 样式，关于页与新菜单沿用同款视觉
- `ProfilePage.tsx` — 列表项 `im-profile-menu-item`、退出登录 `logout()` 用法
- `IMLayout.tsx:468-492` — 右侧面板 `meSubPage` 子页渲染模式（设置面板照此挂载）
- `FloatingProfileCard` / `showUserCard` — 浮层/锚定弹层的现有实现可参考 Popover 定位

## 测试计划（前端手动，账号 17585710998 / dsafasf）
- [ ] 点齿轮弹出 Popover；点外部 / Esc 关闭
- [ ] Popover → 设置：打开 SettingsPage，5 分组正常，返回回到原 Tab
- [ ] Popover → 关于：版本号显示；检查更新 toast "已是最新"
- [ ] Popover → 帮助与反馈：占位/静态正常
- [ ] Popover → 退出登录：正常登出
- [ ] 「我的」Tab 不再出现设置/工具/切换账号/退出登录，个人内容完好
- [ ] 中英文切换下 Popover 与关于页文案均已翻译（无硬编码中文）

## 待定事项
- **备份（聊天记录备份）/ 插件** 的归属：搬进设置大面板（新增"更多/通用"项）还是本期下线？（建议：本期下线或挂占位，留待后续）
- **帮助与反馈** 落地形态：mailto 外链 / 静态 FAQ 页 / 占位 toast？（建议 MVP 静态页或占位）
- **切换账号** 是否本期实现真多账号？（建议：保持现有占位，不在本 spec 范围）
- **移动端布局**：IMLayout 为桌面布局，是否存在移动端布局需同步接通齿轮？（需确认）
- Popover 锚定方向与溢出处理（齿轮在最底部，菜单向上展开）。

## MVP 范围
1. 接通齿轮死按钮 → 上方 Popover 小菜单（设置 / 关于 / 帮助 / 切换账号 / 退出登录）。
2. "设置"复用已有 `SettingsPage`，改为 `settingsOpen` 驱动、与「我的」解耦。
3. 新增纯前端"关于"页（版本写死 + 检查更新 toast + 协议占位）。
4. 「我的」Tab 彻底搬家：移除设置/工具/切换账号/退出登录。
5. 全量 i18n（中英）。

不进 MVP：后端版本接口、真·检查更新、真·多账号切换、备份/插件落地。
