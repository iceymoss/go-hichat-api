# 登录 / 注册 / 重置密码页重构（顺滑旋转卡片）

## 状态
- 创建日期: 2026-06-17
- 状态: 已实现（前端 MVP，方案 A）
- 最终决策：**去掉左侧品牌面板**，纯还原 demo 的整屏渐变 + 居中半透明旋转卡片；**保留 HiChat 全部字段**（注册 5 项 / 重置手机或邮箱），卡片高度自适应。邮箱真正可登录待后端排期。

## 目标
将现有 Telegram 风格双栏认证页（`web/src/components/auth/AuthPage.tsx`）的右侧表单区，重构为参考 demo（`Demo_0x01` / sunyctf 顺滑简约登录注册页）的**顺滑旋转卡片**交互——蓝黄渐变、Mac 红绿灯、胶囊输入框、登录/注册/重置三态围绕卡片左下角旋转切换——同时**保留左侧 HiChat 品牌面板**，并**适配 HiChat 现有字段与接口**（手机号/邮箱、验证码、昵称、确认密码）。

## 非目标
- 不改后端 RPC / `.api` 契约（**唯一例外见「待定事项」中的邮箱登录**）。
- 不改 `web/src/app/api/auth/*` 代理路由的入参/出参字段（登录、注册、重置、send-code 的 body 保持现状）。
- 不改 `useIMStore` 的 `authView` 状态机语义（仍是 `login` / `register` / `forgot-password` 三态）。
- 不做第三方/社交登录（Google、Apple 等）。
- 不改密码强度规则、验证码 60s 倒计时等既有业务校验逻辑。

## 用户故事
- 作为**未登录用户**，我打开 HiChat 时看到左侧品牌面板 + 右侧一张渐变卡片，卡片以顺滑动画载入；输入手机号/邮箱与密码即可登录。
- 作为**新用户**，我点击卡片角落的「去注册」，卡片顺滑旋转翻到注册面，填写手机号、验证码、昵称、密码、确认密码完成注册。
- 作为**忘记密码的用户**，我在登录面点「忘记密码？」，卡片旋转翻到重置面，凭手机号/邮箱 + 验证码设置新密码。
- 三个面板之间的来回切换都用**同一套旋转滑动动画**，视觉与 demo「一模一样」，且卡片高度随字段数量自适应。

## 核心流程

### 载入
1. 页面挂载（`web/src/app/page.tsx` 未登录分支渲染 `AuthPage`）。
2. 左侧 `BrandPanel`（粒子动画 + slogan + 特性）保持不变。
3. 右侧区域为蓝黄渐变背景（`linear-gradient(120deg,#487eb0,#fbc531)`）+ 居中卡片；卡片以 `opacity 0→1 + translateY` 顺滑淡入（对应 demo 的 `.container-show`）。

### 登录（默认面）
1. 卡片正面为深色登录面（`rgba(17,39,59,0.8)`），Mac 红绿灯（hover 显现）。
2. 用户在「账号」框输入手机号或邮箱，在「密码」框输入密码。
3. 点击「登录」→ `POST /api/auth/login`，成功调 `login(d.data)`，失败显示错误。
4. 角落「去注册」→ 旋转到注册面；「忘记密码？」→ 旋转到重置面。

### 注册（旋转面）
1. 白色注册面，字段：手机号(+86) → 验证码（含「获取验证码」按钮 + 60s 倒计时）→ 昵称 → 密码 → 确认密码。
2. 点「获取验证码」→ `POST /api/auth/send-code { target: phone, type:'register' }`。
3. 点「注册」→ 校验（昵称非空、密码 8-20 含字母数字、两次一致）→ `POST /api/auth/register { phone, password, nickname, phoneCode }` → 成功 `login(d.data)`。
4. 角落「去登录」→ 旋转回登录面。

### 重置密码（第三旋转面）
1. 字段：手机号/邮箱 → 验证码（获取 + 倒计时）→ 新密码 → 确认新密码。
2. 点「获取验证码」→ `POST /api/auth/send-code { target, type }`（target 为手机号或邮箱）。
3. 点「重置密码」→ 校验 → `POST /api/auth/reset-pwd { phone|email, code, password }` → 成功显示「重置成功」并提供「去登录」按钮（旋转回登录面或独立成功态，沿用现有成功视图）。

### 切换动画（对齐 demo）
- 三个面板都 `position:absolute` 铺满卡片，`transform-origin: 0 100%`（左下角）。
- 当前面 `rotate(0deg)`，非当前面预置 `rotate(90deg)`（藏在卡片左下角外）。
- 切换时：离开的面 `rotate(-90deg)`，进入的面 `rotate(0deg)`，`transition: .4s`。由 `authView` 驱动 class 切换（用 React 状态/CSS class，**不用 demo 的 jQuery**）。
- hover 时卡片内 title 缩小、输入框上移、角落 change 按钮显现（对齐 demo `.container:hover` 系列效果，按 HiChat 字段量调参）。

## 异常处理
| 场景 | 处理方式 |
|------|---------|
| 账号/密码为空 | 前端拦截，`auth.err.incomplete` |
| 手机号格式错误 | `^1[3-9]\d{9}$` 校验，`auth.err.phone` |
| 邮箱格式错误（登录/重置选邮箱时） | `^[^\s@]+@[^\s@]+\.[^\s@]+$` 校验，新增 `auth.err.email` key |
| 验证码未发送/格式错 | 按钮 disabled 直到账号合法；倒计时内禁止重发 |
| 密码长度/字符不符 | `auth.err.pwdLen / pwdLetter / pwdDigit` |
| 两次密码不一致 | `auth.err.pwdMismatch` |
| 登录/注册/重置接口失败 | 显示后端 `message`，错误行 `shake` 动画 |
| 网络异常 | `auth.err.network` |
| 切换面板时正在加载 | 切换不打断进行中的请求；各面板维护独立 state（沿用现状） |

## 技术设计

### 涉及文件（前端）
| 文件 | 改动 |
|------|------|
| `web/src/components/auth/AuthPage.tsx` | 主要重构：用旋转卡片替换右侧三视图布局；保留 `BrandPanel` / `PInput` 业务、`sendCode`、`useCountdown`、校验逻辑 |
| `web/src/app/globals.css` | 新增 demo 风格 class（卡片容器、`.login-box`/`.sign-box`/第三面、Mac 红绿灯、胶囊输入、旋转切换、渐变背景、hover 效果）；可复用现有 `hc-btn-primary`/`hc-btn-code`/`auth-input-reset`/`shake` |
| `web/src/lib/i18n.ts` | 新增 `auth.err.email`、`auth.account`（账号=手机号/邮箱）等少量 key（zh-CN + en） |
| `web/src/app/page.tsx` | 无需改动（仍渲染 `AuthPage`） |

### 数据模型
无数据库变更。复用现有接口字段：
- 登录 `POST /api/auth/login`：`{ phone, password }`（MVP 仍只传 phone；邮箱登录见待定）
- 注册 `POST /api/auth/register`：`{ phone, password, nickname, phoneCode }`
- 重置 `POST /api/auth/reset-pwd`：`{ phone?, email?, code, password }`
- 验证码 `POST /api/auth/send-code`：`{ target, type }`

### API 接口
| 方法 | 路径 | 说明 | 是否改动 |
|------|------|------|---------|
| POST | /api/auth/login | 登录 | 不改（除非做邮箱登录） |
| POST | /api/auth/register | 注册 | 不改 |
| POST | /api/auth/reset-pwd | 重置密码 | 不改 |
| POST | /api/auth/send-code | 发送验证码 | 不改 |

### 实现步骤（每步可独立 commit）
1. [ ] `globals.css` 增加 demo 风格样式：渐变背景容器、卡片、`.login-box`/`.sign-box`/`.reset-box` 旋转、Mac 红绿灯、胶囊输入、hover 与切换过渡、卡片自适应高度（`min-height` + 内容撑高）。
2. [ ] 重构 `AuthPage.tsx`：保留 `BrandPanel`，右侧改为「渐变背景 + 居中旋转卡片」容器；登录/注册/重置三面板挂在同一卡片内，由 `authView` 驱动旋转 class。
3. [ ] 接入字段与校验：登录面（账号+密码）、注册面（手机号/验证码/昵称/密码/确认）、重置面（手机号或邮箱/验证码/新密码/确认）——复用既有 state、`sendCode`、`useCountdown`、`pwdOk`。
4. [ ] i18n：补 `auth.err.email`、`auth.account` 等 key（zh-CN + en）。
5. [ ] 移动端适配：窄屏隐藏品牌面板（沿用 `useIsMobile`），卡片占满。
6. [ ] 自测三态切换、三接口联调、错误态与倒计时。

### 参考的现有模式
- `web/src/components/auth/AuthPage.tsx` — 现有 `BrandPanel`、`PInput`、`useCountdown`、`sendCode`、`pwdOk` 校验、三视图字段与接口调用，全部复用。
- `web/src/app/globals.css:480+` — `auth-input-reset` / `hc-btn-primary` / `hc-btn-code` / `shake` 既有 class，新样式与之协调。
- demo `Demo_0x01`：`transform-origin:0 100%` + `rotate(±90deg)` + `.4s` 过渡的旋转切换；`.container-show` 载入淡入；clip-path 角落 change 按钮；Mac 红绿灯 hover 显现。**仅借鉴 CSS/交互，逻辑用 React 状态，不引入 jQuery。**

## 测试计划
- [ ] 登录面：手机号+密码成功登录；空字段、错误手机号、错误密码均有正确提示。
- [ ] 注册面：获取验证码倒计时；昵称空、密码弱、两次不一致拦截；正常注册成功登录。
- [ ] 重置面：手机号与邮箱两种 target 都能发码；重置成功后可回登录。
- [ ] 三态来回旋转切换动画顺滑、无残影、无错位；卡片高度随字段自适应。
- [ ] 移动端窄屏：品牌面板隐藏，卡片可用。
- [ ] i18n：zh-CN / en 两种语言文案完整，无硬编码中文。
- [ ] `cd web && bun run build`（或 lint）通过。

## 待定事项
- **邮箱登录（重要）**：用户希望登录「手机号或邮箱+密码」，但后端 `LoginReq` 仅有 `Phone`+`Password`（`apps/user/api/domain.api:39`），`/api/auth/login` 代理也只透传 phone。要支持邮箱登录需后端改 `.api`/`.proto` + 登录逻辑（按 phone 或 email 查用户）。
  - 方案 A（推荐，MVP 内）：登录账号框 UI 支持手机号或邮箱输入，但 **MVP 后端仅 phone 生效**；邮箱登录列为后续后端迭代（另开 `/new-api` 或 `/new-rpc`）。
  - 方案 B：本次一并改后端登录支持邮箱（超出「纯前端重构」范围，需用户确认）。
  - **请用户拍板 A / B。**
- demo 的 hover 放大效果在字段多（注册 5 项）时可能拥挤，是否保留 hover 缩放 title/上移输入，或仅保留旋转切换 + 渐变 + 红绿灯视觉？（实现时按观感微调，倾向弱化 hover 形变以容纳多字段。）
- 重置成功态：复用现有「成功页 + 去登录」独立视图，还是旋转回登录面并 toast 提示？（默认复用现有成功视图。）

## MVP 范围
- ✅ 左品牌面板保留 + 右侧 demo 风格渐变旋转卡片。
- ✅ 登录（手机号/邮箱输入框，后端仅 phone 生效）、注册、重置三态旋转切换，视觉对齐 demo，卡片自适应高度。
- ✅ 复用现有四个 `/api/auth/*` 代理与全部既有校验、验证码、i18n。
- ⏳ 邮箱真正可登录（依赖后端，待定 A/B 后排期）。
- ⏳ hover 形变细节按观感微调。
