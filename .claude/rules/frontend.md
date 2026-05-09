---
description: React 前端代码规则
globs: ["web/**/*.{js,jsx,ts,tsx}"]
---

- 包管理用 bun，不要用 npm/yarn/pnpm
- UI 组件用 Semi Design（@douyinfe/semi-ui），不要引入其他 UI 库
- 国际化用 useTranslation() hook，调用 t('中文key')
- 翻译文件在 web/src/i18n/locales/{lang}.json，扁平 JSON 格式
- 新增用户可见文本必须用 t() 包裹，不要硬编码中文
- 样式优先用 Tailwind CSS class，其次用 Semi 内置样式
