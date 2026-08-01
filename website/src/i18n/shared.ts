/**
 * 语言中立的常量 —— URL、图片尺寸、演示账号。
 *
 * 这些值在中英文之间完全相同，放在这里避免 zh.ts / en.ts 各存一份导致漂移。
 */

export const links = {
  github: "https://github.com/iceymoss/go-hichat-api",
  githubIssues: "https://github.com/iceymoss/go-hichat-api/issues",
  contributing: "https://github.com/iceymoss/go-hichat-api/issues/207",
  license: "https://github.com/iceymoss/go-hichat-api/blob/main/LICENSE",
  releases: "https://github.com/iceymoss/go-hichat-api/releases",
  docsApi: "https://github.com/iceymoss/go-hichat-api/blob/main/docs/api.md",
  docsDevGuide:
    "https://github.com/iceymoss/go-hichat-api/blob/main/docs/development-guide.md",
  docsDevGuideZh:
    "https://github.com/iceymoss/go-hichat-api/blob/main/docs/development-guide.zh-CN.md",
  readmeZh:
    "https://github.com/iceymoss/go-hichat-api/blob/main/docs/README.zh-CN.md",
  dockerDeploy:
    "https://github.com/iceymoss/go-hichat-api/blob/main/deploy/docker/README.md",
  streamingFlows:
    "https://github.com/iceymoss/go-hichat-api/tree/main/apps/streaming/docs",
} as const;

/**
 * 截图尺寸。
 *
 * 原图为 docs/screenshots 下约 3024x1718 的 PNG；`bun run optimize:screenshots`
 * 将其压缩为 public/screenshots 下 1600px 宽的 WebP。这两个值与该输出一致，
 * 传给 next/image 以便浏览器在图片到达前预留正确的盒子尺寸。
 */
export const SCREENSHOT_W = 1600;
export const SCREENSHOT_H = 909;

export const demoAccount = {
  phone: "13800138000",
  password: "hichat2024",
  url: "http://localhost:2470",
} as const;
