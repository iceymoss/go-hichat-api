# 品牌 Logo 落地与全站 PWA 化

## 状态
- 创建日期: 2026-06-18
- 状态: 草稿
- 关联: 设计师交付的 6 个 SVG 位于 `web/green/`

## 目标
把设计师交付的 HiChat 品牌 logo 套件正式接入前端 `web/` 与项目仓库，替换掉目前指向外部 CDN（`z-cdn.chatglm.cn`）的临时 favicon，并在登录/注册页、IM 主界面侧边栏、页面底部、浏览器标签页、PWA 安装入口、项目 README 统一呈现品牌标识，使产品具备完整的品牌一致性与"可安装到桌面"的 PWA 能力。同时把设计师临时交付的 `web/green/` 目录规整到规范的资产目录。

## 非目标
- 不重新设计 logo 本身（直接采用设计师交付件，仅做格式/尺寸适配）。
- 不重构 `AuthPage.tsx` / `IMLayout.tsx` 的整体布局，只在既有结构里插入品牌位。
- 不做 Service Worker / 离线缓存 / 推送（PWA 仅做到"可安装 + 图标 + manifest"，离线能力后续迭代）。
- 不改动后端任何服务。

## 设计资产盘点

设计师在 `web/green/` 交付 6 个 SVG，主色为品牌绿 `#1BB45B`，深色底为 `#15161A`。经分析其设计意图与建议用途：

| 文件 | 画布 | 形态 | 建议用途 |
|------|------|------|---------|
| `hichat-green-appicon.svg` | 1024×1024 | 绿色圆角方块 + 白色"hi"对话气泡标记（填充 tile） | PWA app icon / Apple touch icon / Android maskable icon（需栅格化成 PNG） |
| `hichat-green-bold.svg` | 64×64 | 透明底、加粗描边(3.2) 的绿色标记 | 小尺寸 favicon、侧边栏窄栏图标（小尺寸下加粗更清晰） |
| `hichat-green-full.svg` | 64×64 | 透明底、标准描边(2.5) 的绿色标记 | 标准独立标记（页面内中等尺寸） |
| `hichat-green-simple.svg` | 64×64 | 透明底、最简标记（体积最小） | 内联小标记、loading 占位 |
| `hichat-green-lockup.svg` | 376×96 | 透明底、图标 + "HiChat" 文字横版组合 | 浅色背景的横版品牌（注册卡片、浅色 header） |
| `hichat-green-lockup-dark.svg` | 376×96 | 深底圆角 tile 内嵌图标 + 文字横版 | 深色背景的横版品牌（深色登录卡片、深色页面） |

## 用户故事
- 作为**访客**，我在登录/注册页能立刻看到 HiChat 品牌标识，以便确认我访问的是正确、可信的产品。
- 作为**已登录用户**，我在 IM 主界面侧边栏顶部能看到品牌图标，强化产品归属感。
- 作为**浏览器用户**，我在标签页/书签里看到的是 HiChat 自己的图标，而不是第三方 CDN 的临时图标。
- 作为**移动端/桌面用户**，我可以把 HiChat "添加到主屏幕/安装"，安装后的图标与名称是规范的 HiChat 品牌图标。
- 作为**登录/注册页访客**，我在页面底部能看到带品牌标记的版权/署名栏，传达正规与可信感。
- 作为**GitHub 访客/潜在贡献者**，我打开仓库 README 顶部就能看到 HiChat 品牌横版 logo，快速建立品牌认知。
- 作为**维护者**，品牌源文件集中在一个规范目录里，便于复用与版本管理，而不是散落在临时交付文件夹。

## 核心流程（落地点）

### 1. 浏览器标签 / favicon
当前 `src/app/layout.tsx:22` 的 `icons.icon` 指向外部 CDN，需改为本地品牌图标。

### 2. 登录 / 注册页品牌位（卡片顶部标题上方）
在 `AuthPage.tsx` 每个表单卡片的 `<h1 className="auth-h">` 上方居中插入横版 lockup：
- 深色登录卡片（`LoginBox`，代码注释 `LOGIN box (dark)`）→ `lockup-dark`
- 浅色注册/重置卡片 → `lockup`
（最终深浅由各卡片实际背景决定，见"待定事项"。）

### 3. IM 主界面侧边栏顶部
`IMLayout.tsx` 桌面端为 56px 宽深色图标栏（`im-sidebar`，"TG dark"）。在 `<aside className="im-sidebar ...">` 的导航项上方插入小尺寸品牌标记（`bold` 或 `appicon` 小 tile，深底上绿色/白色均清晰）。移动端底部 tab 栏不放 logo（空间有限）。

### 4. PWA 安装
新增 `manifest`，提供多尺寸图标，使浏览器识别为可安装应用。

### 5. 页面底部品牌 footer（登录/注册页）
`AuthPage.tsx` 当前无任何页脚。在认证页面底部新增品牌版权栏：小尺寸 `mark` + `HiChat` 文字 + `© 2026 HiChat`（文案走 i18n）。深色页面用浅色文字版。IM 主界面（登录后）不强加 footer（聊天界面寸土寸金）。

### 6. 项目 README 品牌头图
`README.md`（英文）与 `docs/README.zh-CN.md`（中文）顶部 `# go-hichat-api` 标题处插入居中横版 lockup 图：
```markdown
<p align="center">
  <img src="assets/brand/hichat-green-lockup.svg" alt="HiChat" width="320" />
</p>
```
（README 引用走仓库相对路径，指向规整后的 `assets/brand/`。）

## 异常处理
| 场景 | 处理方式 |
|------|---------|
| 资产文件缺失 / 路径错误 | 构建期由 Next 静态资源校验暴露；`<img>` 加 `alt="HiChat"` 文本兜底 |
| 深色 logo 放到浅色背景（或反之）导致对比度差 | 按卡片实际背景选 lockup/lockup-dark；统一封装到 `<Logo>` 组件，集中控制 |
| SVG 栅格化工具不可用 | manifest 的 PNG 图标生成依赖栅格化工具（见实现步骤），若缺工具则降级为 SVG-only 图标并记录 |
| 旧外部 CDN favicon 被缓存 | 文件名变更天然破缓存；必要时附带版本查询串 |

## 技术设计

### 资产组织
品牌源文件需要从临时交付目录 `web/green/` 规整到规范位置。统一方案（单一真相 → 各端引用副本）：

- **仓库级品牌源文件（单一真相）**：`assets/brand/`（仓库根目录新建）。把 `web/green/` 的 6 个原始 SVG 移动到此处，作为设计母版，README 也从这里引用。规整后删除临时的 `web/green/` 目录。
  - `assets/brand/hichat-green-{appicon,bold,full,simple,lockup,lockup-dark}.svg`
- **前端运行时副本**：`web/public/brand/`（从 `assets/brand/` 复制需要被前端加载的几个）：
  - `web/public/brand/lockup.svg`（← `hichat-green-lockup.svg`）
  - `web/public/brand/lockup-dark.svg`（← `hichat-green-lockup-dark.svg`）
  - `web/public/brand/mark.svg`（← `hichat-green-bold.svg`，侧边栏/底栏小尺寸用）
- **Next.js App Router 文件约定**（自动注入 `<link>`）：
  - `web/src/app/icon.svg`（← `hichat-green-full.svg` 或 `bold`）→ 自动成为 favicon
  - `web/src/app/apple-icon.png`（← `appicon` 栅格化 180×180）→ Apple 触摸图标
- **PWA manifest 图标**（栅格化自 `hichat-green-appicon.svg`）放 `web/public/brand/`：
  - `icon-192.png`、`icon-512.png`、`icon-maskable-512.png`

> 备注：`web/public/` 下的副本是构建产物式引用，因 README/Next 约定/前端三处的引用基准路径不同，复制副本比软链更稳妥；母版改动时以 `assets/brand/` 为准重新同步。

### 复用组件
新增 `src/components/brand/Logo.tsx`，统一品牌呈现，避免散落硬编码路径：

```tsx
type LogoVariant = 'lockup' | 'lockup-dark' | 'mark';
// 渲染 <img src="/brand/..."> 或内联 SVG，带 alt="HiChat"、可控尺寸
```

登录/注册页与侧边栏都通过 `<Logo variant=... />` 引用。

### manifest（`src/app/manifest.ts`，Next App Router 约定）
```ts
import type { MetadataRoute } from 'next';
export default function manifest(): MetadataRoute.Manifest {
  return {
    name: 'HiChat - 即时通讯',
    short_name: 'HiChat',
    description: '安全、快速、现代化的即时通讯平台',
    start_url: '/',
    display: 'standalone',
    background_color: '#15161A',
    theme_color: '#1BB45B',
    icons: [
      { src: '/brand/icon-192.png', sizes: '192x192', type: 'image/png' },
      { src: '/brand/icon-512.png', sizes: '512x512', type: 'image/png' },
      { src: '/brand/icon-maskable-512.png', sizes: '512x512', type: 'image/png', purpose: 'maskable' },
    ],
  };
}
```

### `layout.tsx` metadata 调整
- 删除 `icons.icon` 的外部 CDN URL（改用 `app/icon.svg` 文件约定，或显式指向本地）。
- 新增 `themeColor: '#1BB45B'`（或经 `viewport` 导出）。
- 顺手修正 `authors`（当前为 `Z.ai Team` 模板残留 → `HiChat`，待用户确认）。

### 实现步骤（每步可独立 commit）
1. [ ] 资产规整：`web/green/` 6 个 SVG 移动到 `assets/brand/`，删除 `web/green/`；复制前端需引用者到 `web/public/brand/`；放置 `app/icon.svg`
2. [ ] 栅格化 `appicon.svg` → `apple-icon.png` / `icon-192/512/maskable` PNG（用 `sharp`/`resvg`/`rsvg-convert`，见待定）
3. [ ] 新增 `src/components/brand/Logo.tsx` 复用组件
4. [ ] `AuthPage.tsx`：登录(dark)/注册(light)/重置卡片标题上方插入 `<Logo>`
5. [ ] `AuthPage.tsx`：页面底部新增品牌版权 footer（mark + 文案，走 i18n）
6. [ ] `IMLayout.tsx`：侧边栏顶部插入小尺寸 `<Logo variant="mark">`
7. [ ] `src/app/manifest.ts`：新增 PWA manifest
8. [ ] `layout.tsx`：移除外部 CDN favicon，接入本地图标 + themeColor，修正 authors
9. [ ] `README.md` + `docs/README.zh-CN.md`：顶部插入居中 lockup 头图（引用 `assets/brand/`）
10. [ ] i18n：新增 footer/版权文案 key 到 `web/src/lib/i18n.ts`（zh-CN + en）
11. [ ] 自测：标签页图标、登录/注册/侧边栏/底栏视觉、README 渲染、Lighthouse PWA 可安装性

### 参考的现有模式
- `src/app/layout.tsx:16-24` — 现有 `metadata` 配置（含待替换的外部 CDN favicon）
- `src/components/auth/AuthPage.tsx:138-160` — `LoginBox` 卡片结构（`<h1 className="auth-h">` 即插入锚点）
- `src/components/im/IMLayout.tsx:291-292` — `im-sidebar` 56px 深色图标栏（侧边栏 logo 锚点）
- `src/app/globals.css` — `.auth-*` / `.im-sidebar` 样式来源

## 测试计划
- [ ] 浏览器标签页显示 HiChat 本地图标（非外部 CDN），无 404
- [ ] 登录页（深底）lockup-dark 对比度正常；注册页（浅底）lockup 对比度正常
- [ ] 侧边栏 logo 在 56px 深色栏内显示清晰、不挤压导航项
- [ ] 登录/注册页底部品牌 footer 显示正常，中/英文案正确
- [ ] 移动端布局未被 logo / footer 破坏
- [ ] Chrome DevTools → Application → Manifest 可识别，提示"可安装"；安装后图标/名称正确
- [ ] README.md / docs/README.zh-CN.md 在 GitHub 上头图正常渲染（SVG 相对路径生效）
- [ ] 中/英切换下 alt 文本与布局正常
- [ ] `cd web && bun run build` 通过，无静态资源告警
- [ ] `assets/brand/` 已就位，`web/green/` 已删除，无残留引用

## 待定事项
- 注册/重置卡片的实际背景是深是浅？决定用 `lockup` 还是 `lockup-dark`（需看 `globals.css` 的 `.auth-*` 实际渲染，或运行时确认）。
- SVG → PNG 栅格化用哪个工具链？项目用 bun，可 `bunx sharp-cli` / `@resvg/resvg-js`，或本机 `rsvg-convert`/`inkscape`。需确认环境可用工具。
- `layout.tsx` 的 `authors: Z.ai Team` 是否改为 HiChat？
- 侧边栏 logo 用绿色透明标记（`bold`）还是绿色圆角 tile（`appicon` 缩小）？取决于深色栏视觉效果。

## MVP 范围
MVP（一次性交付，用户已确认"完整 PWA + 全站品牌"）：
- 资产规整：`web/green/` → `assets/brand/` + `web/public/brand/` ✅
- 替换外部 CDN favicon 为本地图标 ✅
- 登录/注册页卡片顶部品牌 lockup ✅
- 登录/注册页底部品牌 footer ✅
- IM 侧边栏顶部小 logo ✅
- PWA manifest + 多尺寸图标（可安装） ✅
- README（中/英）头图 ✅

后续迭代（非本次）：
- Service Worker / 离线缓存 / 安装提示横幅
- OpenGraph / 社交分享卡片图（og:image）
- 深浅主题切换时 logo 自动换色
