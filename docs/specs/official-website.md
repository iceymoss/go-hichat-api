# HiChat 官方营销网站

## 状态
- 创建日期: 2026-08-01
- 状态: 草稿

## 目标

为 HiChat 2.0 搭建一个面向开发者的营销落地页，展示项目核心功能与架构，降低开发者了解和部署项目的门槛，驱动 GitHub Star 与社区参与。

## 非目标

- 不是应用本体（应用前端在 `web/`）
- 不内置用户认证、登录、注册功能
- 不作为 API 文档站（`docs/api.md` 已有）
- 不做服务端渲染的动态数据（全静态导出即可）
- 不在 MVP 阶段做多语言（留待后续迭代）

---

## 用户故事

作为**开源开发者**，我想要访问 HiChat 官网，了解这个 IM 项目的能力与架构，以便决定是否使用或贡献。

作为**技术决策者**，我想要快速看到功能截图和架构图，以便判断是否满足我们的业务需求。

作为**新贡献者**，我想要一键找到 GitHub 仓库和快速开始文档，以便立刻上手。

---

## 核心流程（Happy Path）

1. 用户通过搜索引擎或 GitHub README 中的链接打开官网首页
2. Hero 区展示项目名称、一句话定位、Logo、两个 CTA 按钮（Star on GitHub / Quick Start）
3. 用户向下滚动，看到功能亮点卡片（6–8 个）
4. 用户继续滚动，看到真实截图展示（分类轮播/grid）
5. 用户点击「功能详情」进入独立的功能页，或点击「快速开始」进入文档页
6. 快速开始页展示 docker compose 三步部署，附 Demo 账号信息
7. 底部 Footer 有 GitHub 链接、文档链接、Apache 2.0 License

---

## 异常处理

| 场景 | 处理方式 |
|------|---------|
| 截图图片加载失败 | 使用 Next.js `<Image>` 自带 fallback，占位灰色区域 |
| 外部 GitHub 链接不可达 | 静态链接，无需处理（用户网络问题） |
| 移动端大截图溢出 | 响应式样式 + 横向滚动容器限宽 |
| 静态导出部署失败 | 配置 `output: 'export'`，检查动态路由是否都有 `generateStaticParams` |

---

## 技术设计

### 目录结构

```
go-hichat-api/
├── apps/                   ← 不变
├── web/                    ← 不变（应用前端）
└── website/                ← 官网（新建）
    ├── package.json
    ├── bun.lock
    ├── next.config.ts       ← output: 'export'
    ├── tailwind.config.ts
    ├── tsconfig.json
    ├── public/
    │   ├── brand/           ← 从 web/public/brand/ 复制
    │   └── screenshots/     ← 从 docs/screenshots/ 复制（按需精选）
    └── src/
        ├── app/
        │   ├── layout.tsx
        │   ├── page.tsx              ← 首页
        │   ├── features/page.tsx     ← 功能详情页
        │   └── quick-start/page.tsx  ← 快速开始
        ├── components/
        │   ├── Hero.tsx
        │   ├── FeatureGrid.tsx
        │   ├── ScreenshotGallery.tsx
        │   ├── ArchitectureDiagram.tsx
        │   ├── QuickStartSteps.tsx
        │   └── Footer.tsx
        └── lib/
            └── content.ts   ← 所有静态文案（便于后续 i18n）
```

### 技术栈

| 层 | 选择 | 理由 |
|----|------|------|
| 框架 | Next.js 16 + React 19 | 与 `web/` 保持一致，团队熟悉 |
| 包管理 | Bun | 与项目约定一致 |
| 样式 | Tailwind CSS v4 | 与 `web/` 一致，足以胜任营销页 |
| UI 组件 | **shadcn/ui** | 零 bundle 开销、Tailwind 原生、完全可定制 |
| 部署 | 静态导出 (`output: 'export'`) | 可直接 host 在 GitHub Pages / Vercel / CDN |
| 图片 | `next/image` + WebP 转换 | 截图文件较大（最大 3 MB），需优化 |

---

### UI 设计规范

> 参考风格：IM 工具类官网（OpenIM、Element、Zulip 等）——深色主调 + 品牌绿强调色 + 截图为主视觉。

#### 色彩体系

品牌绿从 SVG 资源中提取（`assets/brand/hichat-green-appicon.svg`）：

```
品牌主色   --brand:        #1BB45B   (hichat-green)
品牌悬停   --brand-hover:  #17A050   (加深 10%)
深色背景   --bg:           #15161A   (来自 brand SVG)
卡片背景   --surface:      #1E1F24
边框       --border:       #2A2B30
主文字     --fg:           #F4F4F5
次文字     --muted:        #A1A1AA
代码块背景 --code-bg:      #0E0F12
```

映射到 Tailwind CSS v4 的 `@theme` 扩展：

```css
/* website/src/app/globals.css */
@import "tailwindcss";

@theme {
  --color-brand:       #1BB45B;
  --color-brand-hover: #17A050;
  --color-bg:          #15161A;
  --color-surface:     #1E1F24;
  --color-border:      #2A2B30;
  --color-fg:          #F4F4F5;
  --color-muted:       #A1A1AA;
  --color-code-bg:     #0E0F12;
}
```

shadcn/ui 的 `components.json` 设置 `style: "default"` + `baseColor: "zinc"`，然后覆盖 CSS 变量中的 `--primary` 指向品牌绿。

#### 字体

```
无衬线（正文/UI）: system-ui / Inter（通过 next/font/google 引入）
等宽（代码块）:    JetBrains Mono 或 Geist Mono（next/font/local）
```

#### 整体美学原则

| 原则 | 做法 |
|------|------|
| **深色为主** | `<html>` 背景 `#15161A`，不做浅色切换（MVP） |
| **截图是主角** | 截图区域用轻微圆角 + `ring-1 ring-border` 边框，不加花哨阴影 |
| **品牌绿克制使用** | 仅用于 CTA 按钮、强调文字、图标、分隔线点缀 |
| **卡片层次** | bg → surface → 微弱 border，用透明度而非重阴影营造层次 |
| **代码块** | 背景 `#0E0F12`，token 颜色用 `shiki` 或 `highlight.js`（`github-dark` 主题） |
| **动效克制** | 仅 `transition-colors`、scroll-fade-in（Intersection Observer），不用重动画库 |

#### 关键组件样式说明

```
Navbar:
  - 背景: bg-bg/80 backdrop-blur-md (毛玻璃)
  - 粘性: sticky top-0 z-50
  - CTA 按钮: bg-brand text-white hover:bg-brand-hover

Hero:
  - 标题渐变: from-fg to-muted (text-transparent bg-clip-text)
  - 副标题: text-muted
  - 主 CTA: bg-brand, 次 CTA: border border-border hover:bg-surface
  - 首屏截图: 带 ring-1 ring-border, 圆角 rounded-xl, 轻微 drop-shadow

FeatureCard:
  - bg-surface border border-border rounded-xl p-6
  - 图标: 品牌绿 24px (lucide-react)
  - 悬停: hover:border-brand/50 transition-colors

ScreenshotGallery:
  - Tab: shadcn/ui Tabs 组件，激活 tab 用品牌绿下划线
  - 截图: next/image + rounded-lg + ring-1 ring-border

CodeBlock (Quick Start):
  - bg-code-bg rounded-xl p-4 font-mono text-sm
  - 语法高亮: shiki + github-dark 主题
  - 复制按钮: shadcn/ui Button variant="ghost"

Footer:
  - bg-bg border-t border-border
  - Logo 用 dark 变体: hichat-green-lockup-dark.svg
```

### 页面与版块规划

#### 首页 `/`

```
[Navbar]       Logo + GitHub Star 按钮 + Quick Start 链接
[Hero]         大标题 / 副标题 / 两个 CTA / 首屏截图（对话界面）
[Highlights]   6张卡片：IM 消息、社交关系、动态空间、实时音视频、异步消息管道、微服务架构
[Screenshots]  分类 Tab（消息 / 好友 / 动态 / 音视频）+ 截图网格
[Architecture] ASCII 架构图（从 README 复用）或 SVG 版本
[Tech Stack]   图标行：Go / go-zero / Next.js / Kafka / MongoDB / Redis / WebRTC …
[CTA Strip]    "一条命令启动" + docker compose 代码块
[Footer]       GitHub / Docs / License / 中文 README
```

#### 功能详情页 `/features`

每个功能域一个分区（IM、社交、动态、音视频、异步任务），包含截图 + 功能列表。

#### 快速开始页 `/quick-start`

```
Step 1: git clone …
Step 2: docker compose up -d --build
Step 3: 打开 http://localhost:2470，登录 demo 账号
```
附：Demo 账号（`13800138000 / hichat2024`）和功能说明。

---

### 可复用的现有资产

| 资产 | 来源路径 | 用途 |
|------|---------|------|
| 24 张 PNG 截图 | `docs/screenshots/*.png` | Screenshots 分区、功能详情页 |
| Logo SVG（含深色版） | `assets/brand/hichat-green-lockup*.svg` | Navbar、Hero、Footer |
| Web 优化 SVG/PNG | `web/public/brand/` | 可直接复制到 `website/public/brand/` |
| README 功能表 | `README.md` | Highlights 卡片文案 |
| README 架构图 | `README.md` | Architecture 分区 |
| docker compose 命令 | `docker-compose.yaml` | Quick Start 代码块 |

---

### 实现步骤（每步可独立 commit）

1. [ ] 初始化 `website/` Next.js + Bun 项目，配置 `output: 'export'`、Tailwind v4
2. [ ] 复制品牌资源到 `website/public/brand/`，选取精简截图集（~12 张）到 `website/public/screenshots/`
3. [ ] 实现 Navbar 和 Footer 组件
4. [ ] 实现 Hero 分区（标题、副标题、CTA、首屏截图）
5. [ ] 实现 Highlights 功能卡片网格（6–8 张，带图标）
6. [ ] 实现 Screenshots Gallery（分类 Tab + 网格布局）
7. [ ] 实现 Architecture 分区（代码块或 SVG 渲染）
8. [ ] 实现 Tech Stack 徽章行 + "一条命令启动" CTA 条
9. [ ] 实现 `/features` 功能详情页（六大模块分区 + 截图）
10. [ ] 实现 `/quick-start` 快速开始页（三步 + 代码块 + demo 账号）
11. [ ] 响应式调整（移动端截图处理、Navbar 汉堡菜单）
12. [ ] 在仓库根 `README.md` 加官网链接，在 `hichat2.sh` 或 `docker-compose.yaml` 添加注释指向官网

---

## 测试计划

- [ ] `bun run build` 静态导出无报错
- [ ] 首页所有截图正常渲染（无 404）
- [ ] 所有内部链接可访问（`/features`、`/quick-start`）
- [ ] 所有外部链接有效（GitHub、`docs/api.md`）
- [ ] 移动端（375px）首页不横向溢出
- [ ] 桌面端（1440px）截图画廊排版正常
- [ ] Lighthouse 性能分数 ≥ 90（图片优化是关键）

---

## 待定事项

- **官网域名**：暂无；MVP 可先部署到 GitHub Pages（`iceymoss.github.io/go-hichat-api`）或 Vercel
- **在线 Demo 链接**：README 中暂无公开 Demo；官网中可链接到 GitHub 用 Codespaces 一键体验，待确认
- **多语言（中/英）**：不在 MVP 内；文案提取到 `lib/content.ts`，便于后续接入 next-intl
- **暗色模式**：品牌 Logo 有深色变体，Tailwind dark mode 开关；MVP 先只做浅色，深色后续迭代
- **博客 / 更新日志**：不在 MVP 内

---

## MVP 范围

✅ 首页（Hero + 功能亮点 + 截图画廊 + 架构图 + 一键启动 CTA）
✅ `/features` 功能详情页
✅ `/quick-start` 快速开始页（含 Demo 账号）
✅ 所有页面响应式（mobile-first）
✅ 静态导出可直接部署到 GitHub Pages / Vercel

⏳ 后续：多语言、暗色模式、博客、在线 Demo、SEO sitemap
