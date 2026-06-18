---
name: i18n
description: 给 web/ 前端组件接入多语言（中英）。当用户说加多语言、国际化、i18n、翻译、某页面没翻译、硬编码中文时触发。
---

给 `web/` 前端的用户可见文本接入多语言。本项目**不用** `react-i18next`，也**没有** `web/src/i18n/locales/*.json`（CLAUDE.md / frontend.md 里那句已过时，以本文为准）。

## 机制（真相）

- 字典：`web/src/lib/i18n.ts` —— 一个扁平对象 `dict['zh-CN'] / dict['en']`，key 用点号命名（如 `trend.publish`）。
- 取值函数：`t(key, lang)`，缺失时回退 `zh-CN`，再回退 key 本身。**不支持占位符插值**。
- Hook：`web/src/hooks/use-i18n.ts` 的 `useT()`，返回 `(key) => translate(key, lang)`，lang 来自 `useSettingsStore(s => s.language)`。
- 占位符：自己用 `.replace()`，例如 `t('trend.imageLimit').replace('{count}', String(n))`。

## 步骤

### 1. 找出硬编码中文

```bash
cd web
# 列某组件里所有非注释的中文行
grep -nE "[一-鿿]" src/components/im/Xxx.tsx | grep -vE "^\s*[0-9]+:\s*(//|/\*|\*)"
```

逐条判断是否用户可见（按钮/标题/placeholder/title/toast/空态/确认弹窗都算）。**不该翻译的**：语言名本身（`简体中文` / `English`）、代码注释、日志、接口字段名。

### 2. 在 `lib/i18n.ts` 加 key

在 `zh-CN` 块末尾和 `en` 块末尾**各加一组**，分组写注释，两边 key 必须一一对应：

```ts
// zh-CN 块
'xxx.title': '我的相册',
'xxx.itemCount': '{count} 项',
// en 块
'xxx.title': 'Album',
'xxx.itemCount': '{count} items',
```

- 命名空间按模块取前缀（`trend.* / fav.* / pe.* / sec.* / upc.*` 等）。
- **先复用已有 key**，文案一致就别新建：通用 `common.save/cancel/confirm/back/loadMore`、时间 `group.time.justNow/minutesAgo/hoursAgo/monthDay`、在线 `chat.online/offline`。

### 3. 组件里接 `useT`

```tsx
import { useT } from '@/hooks/use-i18n';

export default function Xxx() {
  const t = useT();           // 必须在所有 early return 之前（Hooks 规则）
  ...
  <span>{t('xxx.title')}</span>
  <input placeholder={t('xxx.search')} />
  toast.success(t('xxx.saved'));
  <div>{t('xxx.itemCount').replace('{count}', String(n))}</div>
}
```

**每个子组件都要各自 `const t = useT()`** —— 子组件拿不到父组件的 `t`。

### 4. 校验

```bash
cd web && npx tsc --noEmit -p tsconfig.json 2>&1 | grep <你改的文件名>
# 再确认无残留（排除注释/语言名/兜底）
grep -nE "[一-鿿]" src/components/im/Xxx.tsx | grep -vE "//|/\*|\*|简体中文"
```

预存的既有报错（如 `aspectSquare`/`ringColor`）与本次无关，只看自己文件。

## 常见坑（务必照做）

- **模块级常量表**（`const TYPE_MAP = {1:{label:'文本'}}`）：函数体外拿不到 `t`。把 `label:'文本'` 改成 `labelKey:'trend.type.text'`，渲染时 `t(cfg.labelKey)`。
- **模块级辅助函数**（`fmtTime` / `getUserName`）：加一个 `t` 参数 `fmtTime(date, t)`，所有调用点传进去。
- **变量名遮蔽**：`.map(t => ...)`、`const t = list.find(...)` 会把翻译函数 `t` 遮蔽。重命名迭代/局部变量（`tr` / `ty`），别动翻译函数名。
- **early return 之前调 Hook**：`useT()` 不能放在 `if (!open) return null` 之后。
- **富文本拆分**：`你的数据将<strong>永久删除</strong>` 这种，拆成 `t('a')<strong>{t('b')}</strong>t('c')` 三段 key。
- **批量替换**：用 `perl -0pi -e "s/.../.../g"`（**不要加 `-CSD`**，否则中文字节与解码不匹配会全部 miss）。

## 严格约束

- 新增用户可见文本**必须** `t()` 包裹，不许硬编码中文。
- `zh-CN` 与 `en` 两边 key 必须对齐，缺一边会回退中文。
- 包管理用 bun，UI 用 Semi（见 [`frontend.md`](../../rules/frontend.md)）。
- commit 用约定式前缀（`feat:` / `fix:`），英文，无署名。
