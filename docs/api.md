# HiChat API 接口文档

> 版本: v1.0  
> 更新时间: 2026-06-04  
> 框架: go-zero (REST + gRPC + WebSocket)  
> 认证方式: JWT Token (请求头 `Authorization: Bearer <token>`)

---

## 目录

- [1. 用户服务 (User Service)](#1-用户服务-user-service)
  - [1.1 认证模块](#11-认证模块)
  - [1.2 用户信息模块](#12-用户信息模块)
  - [1.3 验证码模块](#13-验证码模块)
- [2. 社交服务 (Social Service)](#2-社交服务-social-service)
  - [2.1 好友管理模块](#21-好友管理模块)
  - [2.2 好友申请模块](#22-好友申请模块)
  - [2.3 好友设置模块](#23-好友设置模块)
  - [2.4 群组管理模块](#24-群组管理模块)
  - [2.5 群成员管理模块](#25-群成员管理模块)
  - [2.6 群申请模块](#26-群申请模块)
  - [2.7 群邀请链接模块](#27-群邀请链接模块)
  - [2.8 群公告模块](#28-群公告模块)
  - [2.9 群设置模块](#29-群设置模块)
- [3. 动态服务 (Trend Service)](#3-动态服务-trend-service)
  - [3.1 动态管理模块](#31-动态管理模块)
  - [3.2 评论模块](#32-评论模块)
  - [3.3 点赞模块](#33-点赞模块)
- [4. 即时通讯服务 (IM Service)](#4-即时通讯服务-im-service)
  - [4.1 聊天记录模块 (REST)](#41-聊天记录模块-rest)
  - [4.2 会话管理模块 (REST)](#42-会话管理模块-rest)
  - [4.3 WebSocket 通讯模块](#43-websocket-通讯模块)
  - [4.4 富媒体上传模块 (REST)](#44-富媒体上传模块-rest)
- [附录: 公共数据结构](#附录-公共数据结构)

---

## 通用说明

### 响应格式

所有 REST API 返回统一 JSON 格式:

```json
{
  "code": 200,
  "msg": "success",
  "data": { ... }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| code | int | 状态码，200 表示成功 |
| msg | string | 状态描述信息 |
| data | object | 业务数据，具体结构见各接口定义 |

### 认证说明

- **公开接口**: 无需认证，可直接调用
- **受保护接口**: 需在请求头中携带 JWT Token

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

---

# 1. 用户服务 (User Service)

> 基础路径: `/api/v1/user`

## 1.1 认证模块

### 1.1.1 用户注册

- **接口**: `POST /api/v1/user/register`
- **认证**: 不需要
- **描述**: 新用户注册账号，需先通过手机验证码验证

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| phone | string | 是 | 手机号码 |
| password | string | 是 | 登录密码 |
| nickname | string | 是 | 用户昵称 |
| sex | int | 否 | 性别 (0-未知, 1-男, 2-女)，注册时可不设置，后续在个人主页编辑 |
| avatar | string | 否 | 头像URL，注册时可不设置，后续在个人主页编辑 |
| phoneCode | string | 是 | 手机验证码（6位数） |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| token | string | JWT 登录凭证 |
| expire | int64 | Token 过期时间（Unix 时间戳，秒） |

**请求示例**

```bash
curl -X POST http://localhost:8080/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13800138000",
    "password": "Abc123456",
    "nickname": "小明",
    "phoneCode": "123456"
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTAwMDEiLCJleHAiOjE3MTI4MjQwMDB9.abc123",
    "expire": 1712824000
  }
}
```

---

### 1.1.2 用户登录

- **接口**: `POST /api/v1/user/login`
- **认证**: 不需要
- **描述**: 已注册用户通过手机号和密码登录

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| phone | string | 是 | 手机号码 |
| password | string | 是 | 登录密码 |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| token | string | JWT 登录凭证 |
| expire | int64 | Token 过期时间（Unix 时间戳，秒） |

**请求示例**

```bash
curl -X POST http://localhost:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13800138000",
    "password": "Abc123456"
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTAwMDEiLCJleHAiOjE3MTI4MjQwMDB9.abc123",
    "expire": 1712824000
  }
}
```

---

### 1.1.3 重置密码

- **接口**: `PUT /api/v1/user/reset_pwd`
- **认证**: 不需要
- **描述**: 通过手机号或邮箱重置密码，需先获取验证码

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| password | string | 是 | 新密码 |
| phone | string | 否 | 手机号（与 email 二选一） |
| email | string | 否 | 邮箱（与 phone 二选一） |
| code | string | 是 | 验证码 |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X PUT http://localhost:8080/api/v1/user/reset_pwd \
  -H "Content-Type: application/json" \
  -d '{
    "password": "NewPass789",
    "phone": "13800138000",
    "code": "654321"
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 1.1.4 注销账号

- **接口**: `DELETE /api/v1/user/logout`
- **认证**: 需要
- **描述**: 注销当前登录用户的账号

**请求参数**

无

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X DELETE http://localhost:8080/api/v1/user/logout \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

## 1.2 用户信息模块

### 1.2.1 获取用户详情

- **接口**: `GET /api/v1/user/detail`
- **认证**: 需要
- **描述**: 获取当前登录用户的详细信息

**请求参数**

无

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| info | object | 用户信息对象，结构见下方 |

**info 字段明细 (User)**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 用户唯一ID |
| mobile | string | 手机号 |
| nickname | string | 昵称 |
| sex | int | 性别 (0-未知, 1-男, 2-女) |
| avatar | string | 头像URL |
| lastLogin | string | 最后登录时间 |
| introduction | string | 个性签名/简介 |
| email | string | 邮箱地址 |
| region | string | 所在地区 |
| occupation | string | 职业 |
| tags | string | 个人标签（JSON 数组字符串） |

**请求示例**

```bash
curl -X GET http://localhost:8080/api/v1/user/detail \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "info": {
      "id": "10001",
      "mobile": "13800138000",
      "nickname": "小明",
      "sex": 1,
      "avatar": "https://cdn.hichat.com/avatar/10001.jpg",
      "lastLogin": "2026-04-09 10:30:00",
      "introduction": "热爱生活",
      "email": "xiaoming@example.com",
      "region": "广东省深圳市",
      "occupation": "工程师",
      "tags": "[\"Go\",\"React\"]"
    }
  }
}
```

---

### 1.2.2 更新用户信息

- **接口**: `PUT /api/v1/user/update`
- **认证**: 需要
- **描述**: 更新当前用户的个人资料，所有字段均为可选

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 否 | 昵称 |
| phone | string | 否 | 手机号 |
| avatar | string | 否 | 头像URL |
| type | string | 否 | 用户类型 |
| sex | int | 否 | 性别 (0-未知, 1-男, 2-女) |
| introduction | string | 否 | 个性签名/简介 |
| password | string | 否 | 新密码 |
| region | string | 否 | 地区 |
| occupation | string | 否 | 职业 |
| tags | string | 否 | 个人标签（JSON 数组字符串） |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X PUT http://localhost:8080/api/v1/user/update \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "name": "小明Pro",
    "introduction": "全栈工程师",
    "region": "广东省深圳市"
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 1.2.3 上传头像

- **接口**: `POST /api/v1/user/avatar/upload`
- **认证**: 需要
- **描述**: 上传用户头像文件，支持 multipart/form-data，文件大小限制 10MB

**请求参数 (Form-Data)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | file | 是 | 头像图片文件 |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| url | string | 上传成功后的头像访问URL |

**请求示例**

```bash
curl -X POST http://localhost:8080/api/v1/user/avatar/upload \
  -H "Authorization: Bearer eyJhbGci..." \
  -F "file=@/path/to/avatar.jpg"
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "url": "https://cdn.hichat.com/avatar/10001_20260409.jpg"
  }
}
```

---

### 1.2.4 搜索用户

- **接口**: `GET /api/v1/user/search`
- **认证**: 需要
- **描述**: 搜索用户，支持昵称模糊匹配，手机号、邮箱和用户ID精准匹配

**请求参数 (Query)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 否 | 用户昵称（模糊匹配） |
| phone | string | 否 | 手机号（精准匹配） |
| email | string | 否 | 邮箱（精准匹配） |
| ids | []string | 否 | 用户ID列表（精准匹配） |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| users | []User | 用户列表，User 结构见 [1.2.1 获取用户详情](#121-获取用户详情) |

**请求示例**

```bash
curl -X GET "http://localhost:8080/api/v1/user/search?name=小明&phone=13800138000" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "users": [
      {
        "id": "10001",
        "mobile": "13800138000",
        "nickname": "小明",
        "sex": 1,
        "avatar": "https://cdn.hichat.com/avatar/10001.jpg",
        "lastLogin": "2026-04-09 10:30:00",
        "introduction": "热爱生活",
        "email": "xiaoming@example.com",
        "region": "广东省深圳市",
        "occupation": "工程师",
        "tags": "[\"Go\",\"React\"]"
      }
    ]
  }
}
```

---

### 1.2.5 绑定/更换邮箱

- **接口**: `POST /api/v1/user/email/bind`
- **认证**: 需要
- **描述**: 为当前用户绑定或更换邮箱地址，需先发送邮箱验证码

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 邮箱地址 |
| code | string | 是 | 邮箱验证码 |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/api/v1/user/email/bind \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newemail@example.com",
    "code": "123456"
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

## 1.3 验证码模块

### 1.3.1 发送邮箱验证码

- **接口**: `POST /api/v1/user/email/code`
- **认证**: 不需要
- **描述**: 向指定邮箱发送验证码

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 目标邮箱地址 |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/api/v1/user/email/code \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com"
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 1.3.2 验证邮箱验证码

- **接口**: `POST /api/v1/user/email/verify`
- **认证**: 不需要
- **描述**: 验证邮箱验证码是否正确

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 邮箱地址 |
| code | string | 是 | 验证码 |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/api/v1/user/email/verify \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "code": "123456"
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 1.3.3 发送手机验证码

- **接口**: `POST /api/v1/user/phone/code`
- **认证**: 不需要
- **描述**: 向指定手机号发送验证码

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| phone | string | 是 | 手机号码 |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/api/v1/user/phone/code \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13800138000"
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 1.3.4 验证手机验证码

- **接口**: `POST /api/v1/user/phone/verify`
- **认证**: 不需要
- **描述**: 验证手机验证码是否正确

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| phone | string | 是 | 手机号码 |
| code | string | 是 | 验证码 |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/api/v1/user/phone/verify \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13800138000",
    "code": "123456"
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

# 2. 社交服务 (Social Service)

> 基础路径: `/v1/social`  
> 所有接口均需要 JWT 认证

## 2.1 好友管理模块

### 2.1.1 获取好友列表

- **接口**: `GET /v1/social/friends`
- **认证**: 需要
- **描述**: 获取当前用户的全部好友列表

**请求参数**

无

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| list | []Friends | 好友列表 |

**Friends 字段明细**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 好友的用户ID |
| friend_uid | string | 好友用户ID（与id相同，兼容字段） |
| nickname | string | 昵称 |
| avatar | string | 头像URL |
| remark | string | 备注名 |
| sex | int32 | 性别 (0-未知, 1-男, 2-女) |
| email | string | 邮箱 |
| phone | string | 手机号 |
| introduction | string | 个性签名 |
| region | string | 地区 |
| occupation | string | 职业 |
| tags | string | 个人标签（JSON 数组字符串） |
| status | int32 | 用户状态 (0-禁用, 1-正常) |
| type | int32 | 用户类型 (0-普通用户, 1-管理员) |
| last_login | int64 | 最后登录时间（Unix 时间戳） |
| blacklisted | bool | 是否已拉黑 |
| moments_permission | int32 | 朋友圈权限 (0-允许查看, 1-仅聊天, 2-屏蔽朋友圈) |
| notify_enabled | bool | 消息通知是否开启 |
| pinned | bool | 是否置顶 |
| muted | bool | 是否静音 |
| friend_tags | []string | 好友标签数组 |

**请求示例**

```bash
curl -X GET http://localhost:8080/v1/social/friends \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "list": [
      {
        "id": "10002",
        "friend_uid": "10002",
        "nickname": "小红",
        "avatar": "https://cdn.hichat.com/avatar/10002.jpg",
        "remark": "同事小红",
        "sex": 2,
        "email": "xiaohong@example.com",
        "phone": "13900139000",
        "introduction": "设计师",
        "region": "北京市",
        "occupation": "UI设计师",
        "tags": "[\"设计\",\"摄影\"]",
        "status": 1,
        "type": 0,
        "last_login": 1712620800,
        "blacklisted": false,
        "moments_permission": 0,
        "notify_enabled": true,
        "pinned": true,
        "muted": false,
        "friend_tags": ["同事", "设计部"]
      }
    ]
  }
}
```

---

### 2.1.2 获取好友在线状态

- **接口**: `GET /v1/social/friends/online`
- **认证**: 需要
- **描述**: 获取当前用户所有好友的在线/离线状态

**请求参数**

无

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| onLineList | map[string]bool | 好友在线状态，key 为用户ID，value 为是否在线 |

**请求示例**

```bash
curl -X GET http://localhost:8080/v1/social/friends/online \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "onLineList": {
      "10002": true,
      "10003": false,
      "10004": true
    }
  }
}
```

---

### 2.1.3 删除好友

- **接口**: `POST /v1/social/friend/delete`
- **认证**: 需要
- **描述**: 删除指定好友关系

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| friend_uid | string | 是 | 好友用户ID |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/friend/delete \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "friend_uid": "10002"
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 2.1.4 修改好友备注

- **接口**: `POST /v1/social/friend/remark`
- **认证**: 需要
- **描述**: 修改好友的备注名

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| friend_uid | string | 是 | 好友用户ID |
| remark | string | 是 | 新的备注名 |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/friend/remark \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "friend_uid": "10002",
    "remark": "同事小红"
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 2.1.5 设置好友标签

- **接口**: `POST /v1/social/friend/tags`
- **认证**: 需要
- **描述**: 为好友设置标签

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| friend_uid | string | 是 | 好友用户ID |
| tags | []string | 是 | 标签数组 |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/friend/tags \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "friend_uid": "10002",
    "tags": ["同事", "设计部"]
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 2.1.6 举报好友

- **接口**: `POST /v1/social/friend/report`
- **认证**: 需要
- **描述**: 举报指定好友

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| friend_uid | string | 是 | 好友用户ID |
| reason | string | 否 | 举报原因 |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/friend/report \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "friend_uid": "10002",
    "reason": "发送垃圾广告信息"
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 2.1.7 分享好友名片

- **接口**: `POST /v1/social/friend/share`
- **认证**: 需要
- **描述**: 将好友名片分享给其他人（预留接口）

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| friend_uid | string | 是 | 被分享的好友用户ID |
| target_uid | string | 否 | 分享目标用户ID |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/friend/share \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "friend_uid": "10002",
    "target_uid": "10003"
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

## 2.2 好友申请模块

### 2.2.1 发送好友申请

- **接口**: `POST /v1/social/friend/putIn`
- **认证**: 需要
- **描述**: 向目标用户发送好友申请

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_uid | string | 是 | 目标用户ID |
| req_msg | string | 否 | 申请留言 |
| req_time | int64 | 否 | 申请时间（Unix 时间戳） |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/friend/putIn \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "user_uid": "10002",
    "req_msg": "你好，我是小明，希望加你为好友"
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 2.2.2 处理好友申请

- **接口**: `PUT /v1/social/friend/putIn`
- **认证**: 需要
- **描述**: 同意、拒绝或忽略好友申请

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| friend_req_id | int32 | 否 | 好友申请记录ID |
| handle_result | int32 | 否 | 处理结果 (1-同意, 2-拒绝, 3-忽略) |
| remark | string | 否 | 备注名（同意时可设置） |
| tags | []string | 否 | 好友标签（同意时可设置） |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X PUT http://localhost:8080/v1/social/friend/putIn \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "friend_req_id": 101,
    "handle_result": 1,
    "remark": "小明",
    "tags": ["同事"]
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 2.2.3 获取好友申请列表

- **接口**: `GET /v1/social/friend/putIns`
- **认证**: 需要
- **描述**: 获取好友申请列表，支持按处理状态和发起/接收分类筛选

**请求参数 (Query)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | int32 | 否 | 处理类型 (0-待处理, 1-已通过, 2-已拒绝, 3-已忽略) |
| class | string | 否 | 申请列表类型 ("0"-我发起的申请, "1"-我收到的申请) |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| list | []FriendRequests | 好友申请列表 |

**FriendRequests 字段明细**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 申请记录ID |
| user_id | string | 被申请者用户ID |
| req_uid | string | 申请发起者用户ID |
| req_msg | string | 申请留言 |
| message_status | int | 消息状态 (0-已删除, 1-正常显示, 2-忽略不显示) |
| req_time | int64 | 申请时间（Unix 时间戳） |
| handle_result | int | 处理结果 (0-待处理, 1-已同意, 2-已拒绝, 3-已忽略) |
| status | string | 状态英文文本 (pending/accepted/rejected/ignored) |
| status_text | string | 状态中文文本 (待处理/已同意/已拒绝/已忽略) |
| handle_msg | string | 处理附言 |
| read_state | int | 已读状态 (0-未读, 1-已读) |
| nickname | string | 申请者昵称 |
| avatar | string | 申请者头像 |
| sex | int32 | 申请者性别 (0-未知, 1-男, 2-女) |
| email | string | 申请者邮箱 |
| phone | string | 申请者手机号 |
| introduction | string | 申请者个性签名 |
| region | string | 申请者地区 |
| occupation | string | 申请者职业 |
| tags | string | 申请者个人标签（JSON 数组字符串） |
| user_status | int32 | 申请者用户状态 (0-禁用, 1-正常) |
| user_type | int32 | 申请者用户类型 (0-普通用户, 1-管理员) |
| last_login | int64 | 申请者最后登录时间（Unix 时间戳） |

**请求示例**

```bash
curl -X GET "http://localhost:8080/v1/social/friend/putIns?type=0&class=1" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "list": [
      {
        "id": 101,
        "user_id": "10001",
        "req_uid": "10002",
        "req_msg": "你好，我想加你为好友",
        "message_status": 1,
        "req_time": 1712620800,
        "handle_result": 0,
        "status": "pending",
        "status_text": "待处理",
        "handle_msg": "",
        "read_state": 0,
        "nickname": "小红",
        "avatar": "https://cdn.hichat.com/avatar/10002.jpg",
        "sex": 2,
        "email": "xiaohong@example.com",
        "phone": "13900139000",
        "introduction": "UI设计师",
        "region": "北京市",
        "occupation": "设计师",
        "tags": "[\"设计\"]",
        "user_status": 1,
        "user_type": 0,
        "last_login": 1712620800
      }
    ]
  }
}
```

---

### 2.2.4 标记好友申请已读

- **接口**: `PUT /v1/social/friend/putIn/read`
- **认证**: 需要
- **描述**: 将好友申请标记为已读，传 0 或不传 friend_req_id 表示全部标记已读

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| friend_req_id | int32 | 否 | 好友申请记录ID，0 或不传表示全部已读 |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X PUT http://localhost:8080/v1/social/friend/putIn/read \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "friend_req_id": 101
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 2.2.5 删除好友申请

- **接口**: `POST /v1/social/friend/putIn/delete`
- **认证**: 需要
- **描述**: 删除指定好友申请记录

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| friend_req_id | int32 | 否 | 好友申请记录ID |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/friend/putIn/delete \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "friend_req_id": 101
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 2.2.6 获取好友申请未读数量

- **接口**: `GET /v1/social/friend/putIn/messageCount`
- **认证**: 需要
- **描述**: 获取当前用户未处理的好友申请消息数量

**请求参数**

无

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| count | int32 | 未读消息数量 |

**请求示例**

```bash
curl -X GET http://localhost:8080/v1/social/friend/putIn/messageCount \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "count": 3
  }
}
```

---

## 2.3 好友设置模块

### 2.3.1 拉黑/取消拉黑好友

- **接口**: `POST /v1/social/friend/block`
- **认证**: 需要
- **描述**: 将好友加入或移出黑名单

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| friend_uid | string | 是 | 好友用户ID |
| block | bool | 是 | true-拉黑, false-取消拉黑 |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/friend/block \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "friend_uid": "10002",
    "block": true
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 2.3.2 设置朋友圈权限

- **接口**: `POST /v1/social/friend/momentsPermission`
- **认证**: 需要
- **描述**: 设置好友的朋友圈可见权限

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| friend_uid | string | 是 | 好友用户ID |
| permission | int32 | 是 | 权限值 (0-允许查看, 1-仅聊天, 2-屏蔽朋友圈) |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/friend/momentsPermission \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "friend_uid": "10002",
    "permission": 1
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 2.3.3 开关消息通知

- **接口**: `POST /v1/social/friend/notification`
- **认证**: 需要
- **描述**: 开启或关闭好友的消息推送通知

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| friend_uid | string | 是 | 好友用户ID |
| enabled | bool | 是 | true-开启通知, false-关闭通知 |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/friend/notification \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "friend_uid": "10002",
    "enabled": false
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 2.3.4 置顶/取消置顶好友

- **接口**: `POST /v1/social/friend/pin`
- **认证**: 需要
- **描述**: 将好友置顶或取消置顶

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| friend_uid | string | 是 | 好友用户ID |
| pinned | bool | 是 | true-置顶, false-取消置顶 |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/friend/pin \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "friend_uid": "10002",
    "pinned": true
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 2.3.5 静音/取消静音好友

- **接口**: `POST /v1/social/friend/mute`
- **认证**: 需要
- **描述**: 对好友消息进行静音或取消静音

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| friend_uid | string | 是 | 好友用户ID |
| muted | bool | 是 | true-静音, false-取消静音 |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/friend/mute \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "friend_uid": "10002",
    "muted": true
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

## 2.4 群组管理模块

### 2.4.1 创建群组

- **接口**: `POST /v1/social/group`
- **认证**: 需要
- **描述**: 创建一个新的群组

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 否 | 群名称 |
| icon | string | 否 | 群头像URL |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| group_id | string | 新创建的群ID |

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/group \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "name": "技术交流群",
    "icon": "https://cdn.hichat.com/group/tech.jpg"
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "group_id": "g_20001"
  }
}
```

---

### 2.4.2 获取群组列表

- **接口**: `GET /v1/social/groups`
- **认证**: 需要
- **描述**: 获取当前用户加入的所有群组

**请求参数**

无

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| list | []Groups | 群组列表 |

**Groups 字段明细**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 群ID |
| name | string | 群名称 |
| icon | string | 群头像URL |
| status | int64 | 群状态 |
| group_type | int64 | 群类型 |
| is_verify | bool | 是否需要验证才能加入 |
| notification | string | 群公告内容 |
| notification_uid | string | 设置群公告的用户ID |
| create_uid | string | 创建群的用户ID |

**请求示例**

```bash
curl -X GET http://localhost:8080/v1/social/groups \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "list": [
      {
        "id": "g_20001",
        "name": "技术交流群",
        "icon": "https://cdn.hichat.com/group/tech.jpg",
        "status": 1,
        "group_type": 0,
        "is_verify": true,
        "notification": "本群禁止发广告",
        "notification_uid": "10001",
        "create_uid": "10001"
      }
    ]
  }
}
```

---

### 2.4.3 获取群组详情

- **接口**: `GET /v1/social/group/detail`
- **认证**: 需要
- **描述**: 获取指定群组的详细信息及成员列表

**请求参数 (Query)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 是 | 群ID |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| group | Groups | 群组信息，结构见 [2.4.2](#242-获取群组列表) |
| members | []GroupMembers | 群成员列表 |

**GroupMembers 字段明细**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 成员记录ID |
| group_id | string | 群ID |
| user_id | string | 成员用户ID |
| nickname | string | 成员昵称 |
| user_avatar_url | string | 成员头像URL |
| role_level | int | 角色等级 (1-普通成员, 2-管理员, 3-群主) |
| inviter_uid | string | 邀请人用户ID |
| operator_uid | string | 操作人用户ID |
| user | User | 用户基本信息 |

**User 字段明细 (Social 服务中的用户)**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 用户ID |
| nickname | string | 昵称 |
| sex | int | 性别 (0-未知, 1-男, 2-女) |
| avatar | string | 头像URL |
| introduction | string | 个性签名 |
| is_current_user | int | 是否为当前登录用户 (0-否, 1-是) |

**请求示例**

```bash
curl -X GET "http://localhost:8080/v1/social/group/detail?group_id=g_20001" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "group": {
      "id": "g_20001",
      "name": "技术交流群",
      "icon": "https://cdn.hichat.com/group/tech.jpg",
      "status": 1,
      "group_type": 0,
      "is_verify": true,
      "notification": "本群禁止发广告",
      "notification_uid": "10001",
      "create_uid": "10001"
    },
    "members": [
      {
        "id": 1,
        "group_id": "g_20001",
        "user_id": "10001",
        "nickname": "小明",
        "user_avatar_url": "https://cdn.hichat.com/avatar/10001.jpg",
        "role_level": 3,
        "inviter_uid": "",
        "operator_uid": "",
        "user": {
          "id": "10001",
          "nickname": "小明",
          "sex": 1,
          "avatar": "https://cdn.hichat.com/avatar/10001.jpg",
          "introduction": "全栈工程师",
          "is_current_user": 1
        }
      }
    ]
  }
}
```

---

### 2.4.4 更新群组信息

- **接口**: `POST /v1/social/group/update`
- **认证**: 需要
- **描述**: 更新群组基本信息（管理员/群主）

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 是 | 群ID |
| name | string | 否 | 新群名称 |
| icon | string | 否 | 新群头像URL |
| notification | string | 否 | 群公告内容 |
| is_verify | int32 | 否 | 是否需要验证加群 (-1不修改, 0不需要验证, 1需要验证) |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/group/update \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "g_20001",
    "name": "技术交流2群",
    "notification": "欢迎新同学！",
    "is_verify": 1
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 2.4.5 解散群组

- **接口**: `POST /v1/social/group/disband`
- **认证**: 需要
- **描述**: 解散群组（仅群主可操作）

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 是 | 群ID |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/group/disband \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "g_20001"
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

## 2.5 群成员管理模块

### 2.5.1 获取群成员列表

- **接口**: `GET /v1/social/group/users`
- **认证**: 需要
- **描述**: 获取指定群的全部成员列表

**请求参数 (Query/JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 否 | 群ID |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| List | []GroupMembers | 群成员列表，结构见 [2.4.3](#243-获取群组详情) |

**请求示例**

```bash
curl -X GET "http://localhost:8080/v1/social/group/users?group_id=g_20001" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "List": [
      {
        "id": 1,
        "group_id": "g_20001",
        "user_id": "10001",
        "nickname": "小明",
        "user_avatar_url": "https://cdn.hichat.com/avatar/10001.jpg",
        "role_level": 3,
        "inviter_uid": "",
        "operator_uid": "",
        "user": {
          "id": "10001",
          "nickname": "小明",
          "sex": 1,
          "avatar": "https://cdn.hichat.com/avatar/10001.jpg",
          "introduction": "全栈工程师",
          "is_current_user": 0
        }
      }
    ]
  }
}
```

---

### 2.5.2 获取群在线成员

- **接口**: `GET /v1/social/group/users/online`
- **认证**: 需要
- **描述**: 获取群内所有成员的在线状态

**请求参数 (Query/JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 否 | 群ID |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| onLineList | map[string]bool | 成员在线状态，key 为用户ID，value 为是否在线 |

**请求示例**

```bash
curl -X GET "http://localhost:8080/v1/social/group/users/online?group_id=g_20001" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "onLineList": {
      "10001": true,
      "10002": false
    }
  }
}
```

---

### 2.5.3 邀请好友入群

- **接口**: `POST /v1/social/group/invite`
- **认证**: 需要
- **描述**: 邀请好友加入群组

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 是 | 群ID |
| friend_ids | []string | 是 | 要邀请的好友用户ID列表 |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/group/invite \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "g_20001",
    "friend_ids": ["10003", "10004"]
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 2.5.4 踢出群成员

- **接口**: `POST /v1/social/group/kick`
- **认证**: 需要
- **描述**: 将成员踢出群组（管理员/群主）

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 是 | 群ID |
| member_ids | []string | 是 | 被踢成员的用户ID列表 |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/group/kick \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "g_20001",
    "member_ids": ["10005"]
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 2.5.5 退出群组

- **接口**: `POST /v1/social/group/quit`
- **认证**: 需要
- **描述**: 当前用户退出指定群组

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 是 | 群ID |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/group/quit \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "g_20001"
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 2.5.6 转让群主

- **接口**: `POST /v1/social/group/transferOwner`
- **认证**: 需要
- **描述**: 将群主权限转让给其他成员（仅群主可操作）

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 是 | 群ID |
| new_owner_id | string | 是 | 新群主用户ID |
| keep_old_owner_as_admin | bool | 否 | 原群主是否保留为管理员，默认 true |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/group/transferOwner \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "g_20001",
    "new_owner_id": "10002",
    "keep_old_owner_as_admin": true
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 2.5.7 设置/取消管理员

- **接口**: `POST /v1/social/group/setAdmin`
- **认证**: 需要
- **描述**: 将群成员设为管理员或取消管理员（仅群主可操作）

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 是 | 群ID |
| member_ids | []string | 是 | 目标成员用户ID列表 |
| is_admin | bool | 是 | true-设为管理员, false-取消管理员 |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/group/setAdmin \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "g_20001",
    "member_ids": ["10002", "10003"],
    "is_admin": true
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 2.5.8 获取群 @ 列表

- **接口**: `GET /v1/social/group/atList`
- **认证**: 需要
- **描述**: 获取群内可被 @ 的成员列表，支持关键词搜索

**请求参数 (Query)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 是 | 群ID |
| keyword | string | 否 | 搜索关键词（匹配昵称/群昵称） |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| list | []AtMember | 可 @ 的成员列表 |

**AtMember 字段明细**

| 字段 | 类型 | 说明 |
|------|------|------|
| user_id | string | 用户ID |
| nickname | string | 用户昵称 |
| group_nickname | string | 群内昵称 |
| avatar | string | 头像URL |
| role_level | int32 | 角色等级 (1-普通成员, 2-管理员, 3-群主) |

**请求示例**

```bash
curl -X GET "http://localhost:8080/v1/social/group/atList?group_id=g_20001&keyword=小" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "list": [
      {
        "user_id": "10001",
        "nickname": "小明",
        "group_nickname": "明哥",
        "avatar": "https://cdn.hichat.com/avatar/10001.jpg",
        "role_level": 3
      },
      {
        "user_id": "10002",
        "nickname": "小红",
        "group_nickname": "",
        "avatar": "https://cdn.hichat.com/avatar/10002.jpg",
        "role_level": 1
      }
    ]
  }
}
```

---

## 2.6 群申请模块

### 2.6.1 申请加入群组

- **接口**: `POST /v1/social/group/putIn`
- **认证**: 需要
- **描述**: 申请加入群组，支持普通申请、邀请入群、Token 入群三种方式

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 否 | 群ID（普通申请/邀请入群时必传，token入群时可为空） |
| req_id | string | 否 | 申请者/被邀请者ID（邀请入群时必传，普通申请时为空则使用当前用户ID） |
| req_msg | string | 否 | 申请消息 |
| req_time | int64 | 否 | 请求时间（Unix 时间戳） |
| join_source | int64 | 否 | 加入来源 (1-申请入群, 2-邀请入群, 3-邀请链接/二维码入群) |
| inviter_uid | string | 否 | 邀请人用户ID（邀请入群/token入群时必传） |
| token | string | 否 | 邀请链接 token（token 入群时必传，此时 group_id 可为空） |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| group_id | int | 群ID |
| is_pass | int | 是否直接通过 (1-直接进群, 0-进入申请流程) |

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/group/putIn \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "g_20001",
    "req_msg": "我想加入技术交流群",
    "join_source": 1
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "group_id": 20001,
    "is_pass": 0
  }
}
```

---

### 2.6.2 处理入群申请

- **接口**: `PUT /v1/social/group/putIn`
- **认证**: 需要
- **描述**: 同意或拒绝入群申请（管理员/群主）

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_req_id | int32 | 否 | 入群申请记录ID |
| group_id | string | 否 | 群ID |
| handle_result | int32 | 否 | 处理结果 (1-同意, 2-拒绝, 3-忽略) |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X PUT http://localhost:8080/v1/social/group/putIn \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "group_req_id": 201,
    "group_id": "g_20001",
    "handle_result": 1
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 2.6.3 获取入群申请列表

- **接口**: `GET /v1/social/group/putIns`
- **认证**: 需要
- **描述**: 获取指定群的入群申请列表

**请求参数 (Query/JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 否 | 群ID |
| type | []int32 | 否 | 处理状态过滤 (0-未处理, 1-已通过, 2-已拒绝, 3-已忽略) |
| class | int32 | 否 | 申请分类 |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| list | []GroupRequests | 入群申请列表 |

**GroupRequests 字段明细**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 申请记录ID |
| user_id | string | 申请者用户ID |
| group_id | string | 群ID |
| req_msg | string | 申请消息 |
| req_time | int64 | 申请时间（Unix 时间戳） |
| join_source | int64 | 加入来源 (1-申请, 2-邀请, 3-链接/二维码) |
| inviter_user_id | string | 邀请人用户ID |
| handle_user_id | string | 处理人用户ID |
| handle_time | int64 | 处理时间（Unix 时间戳） |
| handle_result | int64 | 处理结果 (0-未处理, 1-同意, 2-拒绝, 3-忽略) |
| user | User | 申请者用户信息 |
| group | Groups | 群组信息 |

**请求示例**

```bash
curl -X GET "http://localhost:8080/v1/social/group/putIns?group_id=g_20001" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "list": [
      {
        "id": 201,
        "user_id": "10005",
        "group_id": "g_20001",
        "req_msg": "我想加入技术交流群",
        "req_time": 1712620800,
        "join_source": 1,
        "inviter_user_id": "",
        "handle_user_id": "",
        "handle_time": 0,
        "handle_result": 0,
        "user": {
          "id": "10005",
          "nickname": "小王",
          "sex": 1,
          "avatar": "https://cdn.hichat.com/avatar/10005.jpg",
          "introduction": "后端工程师",
          "is_current_user": 0
        },
        "group": {
          "id": "g_20001",
          "name": "技术交流群",
          "icon": "https://cdn.hichat.com/group/tech.jpg",
          "status": 1,
          "group_type": 0,
          "is_verify": true,
          "notification": "",
          "notification_uid": "",
          "create_uid": "10001"
        }
      }
    ]
  }
}
```

---

### 2.6.4 获取用户的入群申请列表

- **接口**: `GET /v1/social/group/putInsByUid`
- **认证**: 需要
- **描述**: 按用户ID查询入群申请记录

**请求参数 (Query/JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| ids | []string | 否 | 用户ID列表 |
| class | string | 否 | 类别 ("1"-我发起的申请, "2"-我接受到的申请) |
| type | string | 否 | 状态 ("0"-未处理, "1"-已通过, "2"-已拒绝, "3"-已忽略) |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| list | []GroupRequests | 入群申请列表，结构同 [2.6.3](#263-获取入群申请列表) |

**请求示例**

```bash
curl -X GET "http://localhost:8080/v1/social/group/putInsByUid?class=1&type=0" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "list": [
      {
        "id": 201,
        "user_id": "10001",
        "group_id": "g_20001",
        "req_msg": "希望加入",
        "req_time": 1712620800,
        "join_source": 1,
        "handle_result": 0,
        "user": { "id": "10001", "nickname": "小明", "sex": 1, "avatar": "", "introduction": "", "is_current_user": 1 },
        "group": { "id": "g_20001", "name": "技术交流群", "icon": "", "status": 1, "group_type": 0, "is_verify": true, "notification": "", "notification_uid": "", "create_uid": "10001" }
      }
    ]
  }
}
```

---

### 2.6.5 通过邀请链接加入群组

- **接口**: `POST /v1/social/group/joinByToken`
- **认证**: 需要
- **描述**: 通过邀请链接 token 加入群组

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| token | string | 是 | 邀请链接 token |
| req_msg | string | 否 | 入群申请消息 |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| group_id | int32 | 群ID |
| is_pass | int32 | 是否直接通过 (1-直接进群, 0-进入申请流程) |

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/group/joinByToken \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "token": "abc123def456",
    "req_msg": "通过链接加入"
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "group_id": 20001,
    "is_pass": 1
  }
}
```

---

## 2.7 群邀请链接模块

### 2.7.1 创建邀请链接

- **接口**: `POST /v1/social/group/inviteLink/create`
- **认证**: 需要
- **描述**: 创建群邀请链接/二维码（管理员/群主）

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 是 | 群ID |
| expire_seconds | int64 | 否 | 过期时间（秒），0 表示永不过期 |
| max_uses | int32 | 否 | 最大使用次数，0 表示无限制 |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| link | GroupInviteLink | 邀请链接信息 |

**GroupInviteLink 字段明细**

| 字段 | 类型 | 说明 |
|------|------|------|
| token | string | 邀请 token |
| group_id | string | 群ID |
| created_by | string | 创建者用户ID |
| created_at | int64 | 创建时间（Unix 时间戳） |
| expire_at | int64 | 过期时间（Unix 时间戳），0 表示永不过期 |
| max_uses | int32 | 最大使用次数，0 表示无限制 |
| used_count | int32 | 已使用次数 |
| revoked | bool | 是否已撤销 |
| revoked_at | int64 | 撤销时间（Unix 时间戳） |

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/group/inviteLink/create \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "g_20001",
    "expire_seconds": 86400,
    "max_uses": 50
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "link": {
      "token": "abc123def456ghi789",
      "group_id": "g_20001",
      "created_by": "10001",
      "created_at": 1712620800,
      "expire_at": 1712707200,
      "max_uses": 50,
      "used_count": 0,
      "revoked": false,
      "revoked_at": 0
    }
  }
}
```

---

### 2.7.2 获取邀请链接列表

- **接口**: `GET /v1/social/group/inviteLinks`
- **认证**: 需要
- **描述**: 获取群的所有邀请链接列表（管理员/群主）

**请求参数 (Query)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 是 | 群ID |
| include_revoked | bool | 否 | 是否包含已撤销的链接 |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| list | []GroupInviteLink | 邀请链接列表，结构见 [2.7.1](#271-创建邀请链接) |

**请求示例**

```bash
curl -X GET "http://localhost:8080/v1/social/group/inviteLinks?group_id=g_20001&include_revoked=false" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "list": [
      {
        "token": "abc123def456ghi789",
        "group_id": "g_20001",
        "created_by": "10001",
        "created_at": 1712620800,
        "expire_at": 1712707200,
        "max_uses": 50,
        "used_count": 3,
        "revoked": false,
        "revoked_at": 0
      }
    ]
  }
}
```

---

### 2.7.3 撤销邀请链接

- **接口**: `POST /v1/social/group/inviteLink/revoke`
- **认证**: 需要
- **描述**: 撤销指定邀请链接（管理员/群主/链接创建者）

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 是 | 群ID |
| token | string | 是 | 邀请 token |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/group/inviteLink/revoke \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "g_20001",
    "token": "abc123def456ghi789"
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

## 2.8 群公告模块

### 2.8.1 创建群公告

- **接口**: `POST /v1/social/group/announcement`
- **认证**: 需要
- **描述**: 发布群公告（管理员/群主）

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 是 | 群ID |
| content | string | 是 | 公告内容 |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 公告ID |

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/group/announcement \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "g_20001",
    "content": "本周五下午3点线上会议，请大家准时参加"
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "id": 301
  }
}
```

---

### 2.8.2 获取群公告列表

- **接口**: `GET /v1/social/group/announcements`
- **认证**: 需要
- **描述**: 获取群的公告列表

**请求参数 (Query)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 是 | 群ID |
| include_deleted | bool | 否 | 是否包含已删除的公告 |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| list | []GroupAnnouncement | 公告列表 |

**GroupAnnouncement 字段明细**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 公告ID |
| group_id | string | 群ID |
| content | string | 公告内容 |
| created_by | string | 发布者用户ID |
| created_at | int64 | 发布时间（Unix 时间戳） |
| pinned | bool | 是否置顶 |
| pinned_at | int64 | 置顶时间（Unix 时间戳） |

**请求示例**

```bash
curl -X GET "http://localhost:8080/v1/social/group/announcements?group_id=g_20001" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "list": [
      {
        "id": 301,
        "group_id": "g_20001",
        "content": "本周五下午3点线上会议，请大家准时参加",
        "created_by": "10001",
        "created_at": 1712620800,
        "pinned": true,
        "pinned_at": 1712620900
      }
    ]
  }
}
```

---

### 2.8.3 置顶/取消置顶公告

- **接口**: `POST /v1/social/group/announcement/pin`
- **认证**: 需要
- **描述**: 置顶或取消置顶群公告（管理员/群主）

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 是 | 群ID |
| announcement_id | int64 | 是 | 公告ID |
| pinned | bool | 是 | true-置顶, false-取消置顶 |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/group/announcement/pin \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "g_20001",
    "announcement_id": 301,
    "pinned": true
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

## 2.9 群设置模块

### 2.9.1 获取我的群成员设置

- **接口**: `GET /v1/social/group/memberSetting`
- **认证**: 需要
- **描述**: 获取当前用户在指定群内的个人设置（群昵称、群备注）

**请求参数 (Query)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 是 | 群ID |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| setting | GroupMemberSetting | 群成员设置信息 |

**GroupMemberSetting 字段明细**

| 字段 | 类型 | 说明 |
|------|------|------|
| group_id | string | 群ID |
| user_id | string | 用户ID |
| group_nickname | string | 群内昵称 |
| group_remark | string | 群备注名 |
| updated_at | int64 | 最后更新时间（Unix 时间戳） |

**请求示例**

```bash
curl -X GET "http://localhost:8080/v1/social/group/memberSetting?group_id=g_20001" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "setting": {
      "group_id": "g_20001",
      "user_id": "10001",
      "group_nickname": "明哥",
      "group_remark": "技术群",
      "updated_at": 1712620800
    }
  }
}
```

---

### 2.9.2 更新我的群成员设置

- **接口**: `POST /v1/social/group/memberSetting`
- **认证**: 需要
- **描述**: 修改当前用户在指定群内的群昵称或群备注

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 是 | 群ID |
| group_nickname | string | 否 | 新的群内昵称 |
| group_remark | string | 否 | 新的群备注名 |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/social/group/memberSetting \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "g_20001",
    "group_nickname": "明哥",
    "group_remark": "技术交流群"
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

# 3. 动态服务 (Trend Service)

> 基础路径: `/v1`  
> 所有接口均需要 JWT 认证

## 3.1 动态管理模块

### 3.1.1 发布动态

- **接口**: `POST /v1/trend`
- **认证**: 需要
- **描述**: 发布新动态，支持图文、视频、文章、分享等多种类型

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | int | 是 | 动态类型 (1-纯文本, 2-图文混合, 3-文章, 4-分享, 5-视频, 6-广告) |
| content | string | 否 | 文本内容 |
| scope | int | 是 | 可见范围 (1-仅自己, 2-仅好友, 3-所有人) |
| resources | []string | 否 | 资源URL列表（图片/视频） |
| position_name | string | 否 | 位置名称 |
| longitude | float64 | 否 | 经度 |
| latitude | float64 | 否 | 纬度 |
| title | string | 否 | 文章标题（type=3 时使用） |
| at_user_ids | []int32 | 否 | 被 @ 的用户ID列表 |
| open_reply | bool | 否 | 是否开放评论区 |
| cover_url | string | 否 | 封面图URL |
| share_url | string | 否 | 第三方分享链接（type=4 时使用） |
| device | string | 否 | 发布设备信息 |
| ip | string | 否 | 发布者IP |

> **类型与必填校验**（后端按 `type` 校验内容是否为空）：
> - `type=1` 纯文本：`content` 必填
> - `type=2` 图文混合：`content` 与 `resources` 至少其一
> - `type=3` 文章：`title` 必填，且 `content` 与 `resources` 至少其一
> - `type=4` 分享：`share_url` 必填
> - `type=5` 视频：`resources` 必填
> - `type=0` 与 `type=6`（广告）**不对普通用户开放**，调用会返回「不支持该类型的动态」。广告动态由运营侧投放，不经此接口。

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| trend_id | int | 新动态ID |
| code | int | 操作状态码 |

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/trend \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "type": 2,
    "content": "今天天气真好，出去爬山了！",
    "scope": 3,
    "resources": [
      "https://cdn.hichat.com/img/001.jpg",
      "https://cdn.hichat.com/img/002.jpg"
    ],
    "position_name": "深圳梧桐山",
    "longitude": 114.2,
    "latitude": 22.6,
    "open_reply": true
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "trend_id": 50001,
    "code": 0
  }
}
```

---

### 3.1.2 删除动态

- **接口**: `POST /v1/trend/delete`
- **认证**: 需要
- **描述**: 删除指定动态（仅动态作者可操作）

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| trend_id | int | 是 | 动态ID |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| success | bool | 是否删除成功 |

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/trend/delete \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "trend_id": 50001
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "success": true
  }
}
```

---

### 3.1.3 更新动态

- **接口**: `PUT /v1/trend/update`
- **认证**: 需要
- **描述**: 更新动态设置（置顶、评论开关、可见范围）

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| trend_id | uint32 | 是 | 动态ID |
| is_top | int32 | 否 | 是否置顶 (0-否, 1-是) |
| open_reply | int32 | 否 | 是否开启评论区 (0-关闭, 1-开启) |
| scope | int32 | 否 | 可见范围 (1-可见, 2-不可见) |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X PUT http://localhost:8080/v1/trend/update \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "trend_id": 50001,
    "is_top": 1,
    "open_reply": 1
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 3.1.4 获取动态详情

- **接口**: `GET /v1/trend/detail`
- **认证**: 需要
- **描述**: 获取单条动态的详细信息

**请求参数 (Query/JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| trend_id | int | 是 | 动态ID |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| trend | Trend | 动态详情 |

**Trend 字段明细**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 动态唯一ID |
| user_id | User | 发布者用户信息 |
| type | int | 动态类型 (1-纯文本, 2-图文混合, 3-文章, 4-分享, 5-视频, 6-广告) |
| content | string | 文本内容 |
| scope | int | 可见范围 (1-仅自己, 2-仅好友, 3-所有人) |
| create_time | int64 | 创建时间（Unix 时间戳） |
| reply_count | int32 | 评论数量 |
| agree_count | int32 | 点赞数量 |
| position_name | string | 位置名称 |
| position_point | []float32 | 经纬度数组 [经度, 纬度] |
| title | string | 文章标题 |
| at_user | []User | 被 @ 的用户列表 |
| resources | []string | 资源URL列表 |
| cover_url | string | 封面图URL |
| share_url | string | 第三方分享链接 |
| open_reply | int32 | 是否允许评论 (0-不允许, 1-允许) |
| device_id | string | 发布设备ID |
| ip | string | 发布者IP |
| is_top | int | 是否置顶 (0-否, 1-是) |

**Trend 中 User 字段明细**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 用户ID |
| nickname | string | 昵称 |
| sex | int | 性别 (0-未知, 1-男, 2-女) |
| avatar | string | 头像URL |

**请求示例**

```bash
curl -X GET "http://localhost:8080/v1/trend/detail?trend_id=50001" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "trend": {
      "id": 50001,
      "user_id": {
        "id": "10001",
        "nickname": "小明",
        "sex": 1,
        "avatar": "https://cdn.hichat.com/avatar/10001.jpg"
      },
      "type": 2,
      "content": "今天天气真好，出去爬山了！",
      "scope": 3,
      "create_time": 1712620800,
      "reply_count": 5,
      "agree_count": 12,
      "position_name": "深圳梧桐山",
      "position_point": [114.2, 22.6],
      "title": "",
      "at_user": [],
      "resources": [
        "https://cdn.hichat.com/img/001.jpg",
        "https://cdn.hichat.com/img/002.jpg"
      ],
      "cover_url": "",
      "share_url": "",
      "open_reply": 1,
      "device_id": "",
      "ip": "120.230.10.1",
      "is_top": 0
    }
  }
}
```

---

### 3.1.5 获取最新动态列表

- **接口**: `GET /v1/trends/latest`
- **认证**: 需要
- **描述**: 获取最新动态流，支持游标分页。当前实现的数据范围为**当前用户本人 + 其好友**发布的动态（并非全站推荐流）。

**请求参数 (Query/JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| last_trend_id | int | 否 | 上一页最后一条动态ID（用于分页） |
| count | int | 否 | 每页数量 |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| list | []Trend | 动态列表，Trend 结构见 [3.1.4](#314-获取动态详情) |
| last_trend_id | int | 本页最后一条动态ID（用于下次请求分页） |
| has_more | bool | 是否还有更多数据 |

**请求示例**

```bash
curl -X GET "http://localhost:8080/v1/trends/latest?count=10" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "list": [
      {
        "id": 50002,
        "user_id": { "id": "10002", "nickname": "小红", "sex": 2, "avatar": "https://cdn.hichat.com/avatar/10002.jpg" },
        "type": 1,
        "content": "今天学了新的设计技巧",
        "scope": 3,
        "create_time": 1712624400,
        "reply_count": 2,
        "agree_count": 8,
        "position_name": "",
        "position_point": [],
        "title": "",
        "at_user": [],
        "resources": [],
        "cover_url": "",
        "share_url": "",
        "open_reply": 1,
        "device_id": "",
        "ip": "",
        "is_top": 0
      }
    ],
    "last_trend_id": 50002,
    "has_more": true
  }
}
```

---

### 3.1.6 获取用户动态列表

- **接口**: `GET /v1/user/trends`
- **认证**: 需要
- **描述**: 获取指定用户的动态列表，包含置顶动态

**请求参数 (Query/JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| target_user_id | int | 是 | 目标用户ID |
| last_id | int | 否 | 上一页最后一条动态ID（用于分页） |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| list | []Trend | 普通动态列表 |
| top_list | []Trend | 置顶动态列表 |
| last_id | int | 本页最后一条动态ID |
| last_time | int | 本页最后一条动态时间 |

**请求示例**

```bash
curl -X GET "http://localhost:8080/v1/user/trends?target_user_id=10001" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "list": [
      {
        "id": 50001,
        "user_id": { "id": "10001", "nickname": "小明", "sex": 1, "avatar": "https://cdn.hichat.com/avatar/10001.jpg" },
        "type": 2,
        "content": "今天天气真好",
        "scope": 3,
        "create_time": 1712620800,
        "reply_count": 5,
        "agree_count": 12,
        "resources": ["https://cdn.hichat.com/img/001.jpg"],
        "open_reply": 1,
        "is_top": 0
      }
    ],
    "top_list": [],
    "last_id": 50001,
    "last_time": 1712620800
  }
}
```

---

### 3.1.7 获取动态列表（高级筛选）

- **接口**: `GET /v1/trends`
- **认证**: 需要
- **描述**: 获取动态列表，支持按类型、排序方式、用户筛选。

> ⚠️ **当前未实现**：该接口为预留的高级筛选入口，logic 尚未实现，恒返回空结果。前端动态流请使用 [3.1.5 获取最新动态列表](#315-获取最新动态列表)（`GET /v1/trends/latest`）与 [3.1.6 获取用户动态列表](#316-获取用户动态列表)。

**请求参数 (Query/JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| last_id | int | 是 | 上一页最后一条动态ID |
| last_time | int | 是 | 上一页最后一条动态时间 |
| types | []string | 是 | 动态类型过滤列表 |
| sort_column | string | 是 | 排序字段 |
| sort_type | int | 是 | 排序方式 (0-降序, 1-升序) |
| user_ids | []string | 是 | 指定用户ID列表 |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| trends | []string | 动态数据列表（JSON 序列化字符串） |
| last_id | int | 本页最后一条动态ID |
| last_time | int | 本页最后一条动态时间 |

**请求示例**

```bash
curl -X GET "http://localhost:8080/v1/trends?last_id=0&last_time=0&sort_type=0" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "trends": [],
    "last_id": 50001,
    "last_time": 1712620800
  }
}
```

---

## 3.2 评论模块

### 3.2.1 发表评论

- **接口**: `POST /v1/trend/comment`
- **认证**: 需要
- **描述**: 对动态发表评论或回复评论

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| trend_id | uint64 | 是 | 动态ID |
| father | uint64 | 否 | 父评论ID，一级评论不传此字段 |
| user_id | uint64 | 是 | 被评论/被回复的用户ID |
| content | string | 是 | 评论内容 |
| at_users | []uint64 | 否 | 被 @ 的用户ID列表 |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/trend/comment \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "trend_id": 50001,
    "user_id": 10001,
    "content": "照片拍得真好看！"
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 3.2.2 获取评论树

- **接口**: `GET /v1/trend/comment/tree`
- **认证**: 需要
- **描述**: 获取动态的评论树结构（含嵌套子评论）

**请求参数 (Query/JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| trend_id | []uint64 | 是 | 动态ID列表 |
| last_id | int | 否 | 上一页最后一条评论ID |
| last_time | int | 否 | 上一页最后一条评论时间 |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| discuss | object | 评论树数据 |
| last_id | int | 本页最后一条评论ID |
| last_time | int | 本页最后一条评论时间 |

**discuss 内部结构**

| 字段 | 类型 | 说明 |
|------|------|------|
| trend_discuss | map[uint64]TrendDiscusses | 按动态ID分组的评论，key 为动态ID |

**TrendDiscusses 内部结构**

| 字段 | 类型 | 说明 |
|------|------|------|
| discusses | []Discuss | 评论列表 |

**Discuss 字段明细**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 评论ID |
| trend_id | int64 | 所属动态ID |
| root_id | int64 | 根评论ID |
| father | int64 | 父评论ID (0 表示一级评论) |
| replyer | User | 评论者用户信息 |
| user_id | User | 被评论/被回复的用户信息 |
| at_user_ids | []User | 被 @ 的用户列表 |
| level | int | 评论层级 (1-一级, 2-二级...) |
| content | string | 评论内容 |
| agree_count | int64 | 点赞数 |
| discuss_count | int64 | 子评论数量 |
| state | int | 状态 (0-已删除, 1-正常) |
| read | bool | 是否已读 |
| create_time | int64 | 创建时间（Unix 时间戳） |
| update_time | int64 | 更新时间（Unix 时间戳） |
| children | []Discuss | 子评论列表（递归结构） |

**请求示例**

```bash
curl -X GET "http://localhost:8080/v1/trend/comment/tree?trend_id=50001" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "discuss": {
      "trend_discuss": {
        "50001": {
          "discusses": [
            {
              "id": 60001,
              "trend_id": 50001,
              "root_id": 0,
              "father": 0,
              "replyer": { "id": "10002", "nickname": "小红", "sex": 2, "avatar": "https://cdn.hichat.com/avatar/10002.jpg" },
              "user_id": { "id": "10001", "nickname": "小明", "sex": 1, "avatar": "https://cdn.hichat.com/avatar/10001.jpg" },
              "at_user_ids": [],
              "level": 1,
              "content": "照片拍得真好看！",
              "agree_count": 3,
              "discuss_count": 1,
              "state": 1,
              "read": true,
              "create_time": 1712624400,
              "update_time": 1712624400,
              "children": [
                {
                  "id": 60002,
                  "trend_id": 50001,
                  "root_id": 60001,
                  "father": 60001,
                  "replyer": { "id": "10001", "nickname": "小明", "sex": 1, "avatar": "https://cdn.hichat.com/avatar/10001.jpg" },
                  "user_id": { "id": "10002", "nickname": "小红", "sex": 2, "avatar": "https://cdn.hichat.com/avatar/10002.jpg" },
                  "level": 2,
                  "content": "谢谢！",
                  "agree_count": 0,
                  "discuss_count": 0,
                  "state": 1,
                  "read": false,
                  "create_time": 1712625000,
                  "update_time": 1712625000,
                  "children": []
                }
              ]
            }
          ]
        }
      }
    },
    "last_id": 60001,
    "last_time": 1712624400
  }
}
```

---

### 3.2.3 获取根评论列表

- **接口**: `GET /v1/trend/comment/root`
- **认证**: 需要
- **描述**: 获取动态下的一级评论列表（不含子评论）

**请求参数 (Query/JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| trend_id | uint64 | 是 | 动态ID |
| last_id | int | 否 | 上一页最后一条评论ID |
| last_time | int | 否 | 上一页最后一条评论时间 |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| list | []Discuss | 评论列表，Discuss 结构见 [3.2.2](#322-获取评论树) |
| last_id | int | 本页最后一条评论ID |
| last_time | int | 本页最后一条评论时间 |
| total | int | 总评论数 |

**请求示例**

```bash
curl -X GET "http://localhost:8080/v1/trend/comment/root?trend_id=50001" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "list": [
      {
        "id": 60001,
        "trend_id": 50001,
        "root_id": 0,
        "father": 0,
        "replyer": { "id": "10002", "nickname": "小红", "sex": 2, "avatar": "https://cdn.hichat.com/avatar/10002.jpg" },
        "user_id": { "id": "10001", "nickname": "小明", "sex": 1, "avatar": "https://cdn.hichat.com/avatar/10001.jpg" },
        "level": 1,
        "content": "照片拍得真好看！",
        "agree_count": 3,
        "discuss_count": 1,
        "state": 1,
        "create_time": 1712624400,
        "update_time": 1712624400
      }
    ],
    "last_id": 60001,
    "last_time": 1712624400,
    "total": 5
  }
}
```

---

### 3.2.4 获取子评论列表

- **接口**: `GET /v1/trend/comment/children`
- **认证**: 需要
- **描述**: 获取指定评论的子评论列表

**请求参数 (Query/JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| father | uint64 | 是 | 父评论ID |
| last_id | int | 否 | 上一页最后一条评论ID |
| last_time | uint64 | 否 | 上一页最后一条评论时间 |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| list | []Discuss | 子评论列表 |
| last_id | int | 本页最后一条评论ID |
| last_time | int | 本页最后一条评论时间 |
| total | int | 总评论数 |

**请求示例**

```bash
curl -X GET "http://localhost:8080/v1/trend/comment/children?father=60001" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "list": [
      {
        "id": 60002,
        "trend_id": 50001,
        "root_id": 60001,
        "father": 60001,
        "replyer": { "id": "10001", "nickname": "小明", "sex": 1, "avatar": "https://cdn.hichat.com/avatar/10001.jpg" },
        "user_id": { "id": "10002", "nickname": "小红", "sex": 2, "avatar": "https://cdn.hichat.com/avatar/10002.jpg" },
        "level": 2,
        "content": "谢谢！",
        "agree_count": 0,
        "discuss_count": 0,
        "state": 1,
        "create_time": 1712625000,
        "update_time": 1712625000
      }
    ],
    "last_id": 60002,
    "last_time": 1712625000,
    "total": 1
  }
}
```

---

### 3.2.5 删除评论

- **接口**: `DELETE /v1/trend/comment`
- **认证**: 需要
- **描述**: 删除指定评论

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint64 | 是 | 评论ID |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| success | bool | 是否删除成功 |

**请求示例**

```bash
curl -X DELETE http://localhost:8080/v1/trend/comment \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "id": 60002
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "success": true
  }
}
```

---

### 3.2.6 获取未读回复

- **接口**: `GET /v1/trend/unread`
- **认证**: 需要
- **描述**: 获取未读的评论回复和点赞通知

**请求参数 (Query/JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| like_last_id | int | 否 | 点赞上一页最后ID |
| discuss_last_id | int | 否 | 评论上一页最后ID |
| last_time | int | 否 | 上一页最后时间 |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| replies | []Discuss | 未读评论回复列表 |
| likes | []Like | 未读点赞列表 |
| like_last_id | int | 点赞本页最后ID |
| discuss_last_id | int | 评论本页最后ID |
| last_time | int | 本页最后时间 |

**Like 字段明细**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int | 点赞记录ID |
| user | User | 点赞用户信息 |
| trend_id | uint64 | 被点赞的动态ID |
| like_time | uint64 | 点赞时间（Unix 时间戳） |

**请求示例**

```bash
curl -X GET "http://localhost:8080/v1/trend/unread" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "replies": [
      {
        "id": 60003,
        "trend_id": 50001,
        "father": 0,
        "replyer": { "id": "10003", "nickname": "小王", "sex": 1, "avatar": "" },
        "user_id": { "id": "10001", "nickname": "小明", "sex": 1, "avatar": "" },
        "content": "真不错",
        "read": false,
        "create_time": 1712628000
      }
    ],
    "likes": [
      {
        "id": 1,
        "user": { "id": "10004", "nickname": "小李", "sex": 1, "avatar": "" },
        "trend_id": 50001,
        "like_time": 1712627000
      }
    ],
    "like_last_id": 1,
    "discuss_last_id": 60003,
    "last_time": 1712628000
  }
}
```

---

### 3.2.7 标记评论已读

- **接口**: `PUT /v1/trend/comment/mark-read`
- **认证**: 需要
- **描述**: 批量标记评论为已读

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| dis_ids | []uint64 | 是 | 评论ID列表 |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X PUT http://localhost:8080/v1/trend/comment/mark-read \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "dis_ids": [60001, 60002, 60003]
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

## 3.3 点赞模块

### 3.3.1 点赞/取消点赞

- **接口**: `POST /v1/trend/like`
- **认证**: 需要
- **描述**: 对动态进行点赞或取消点赞（切换操作）

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| trend_id | uint32 | 是 | 动态ID |
| author_id | uint32 | 是 | 动态作者用户ID |
| like_type | int | 是 | 点赞类型：`1`-点赞，`0`-取消点赞（服务端按 `like_type > 0` 判定为点赞，否则取消） |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/trend/like \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "trend_id": 50001,
    "author_id": 10001,
    "like_type": 1
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 3.3.2 获取动态点赞摘要

- **接口**: `GET /v1/trend/like/summary`
- **认证**: 需要
- **描述**: 获取指定动态的点赞摘要信息

> ⚠️ **当前未实现**：该接口 logic 为空桩，恒返回空 `summary_json`。批量获取点赞用户请使用 [3.3.3 批量获取动态点赞摘要](#333-批量获取动态点赞摘要)（`GET /v1/trend/like/batch-summary`）。

**请求参数 (Query/JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | string | 是 | 当前用户ID |
| trend_id | string | 是 | 动态ID |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| summary_json | string | 点赞摘要的 JSON 序列化字符串 |

**请求示例**

```bash
curl -X GET "http://localhost:8080/v1/trend/like/summary?user_id=10001&trend_id=50001" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "summary_json": "{\"total\":12,\"is_liked\":true}"
  }
}
```

---

### 3.3.3 批量获取动态点赞摘要

- **接口**: `GET /v1/trend/like/batch-summary`
- **认证**: 需要
- **描述**: 批量获取多个动态的点赞用户列表

**请求参数 (Query/JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| trend_id | []uint64 | 是 | 动态ID列表 |
| last_id | int | 否 | 上一页最后ID |
| last_time | int | 否 | 上一页最后时间 |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| trend_likes | map[uint64][]User | 按动态ID分组的点赞用户列表 |

**请求示例**

```bash
curl -X GET "http://localhost:8080/v1/trend/like/batch-summary?trend_id=50001&trend_id=50002" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "trend_likes": {
      "50001": [
        { "id": "10002", "nickname": "小红", "sex": 2, "avatar": "https://cdn.hichat.com/avatar/10002.jpg" },
        { "id": "10003", "nickname": "小王", "sex": 1, "avatar": "" }
      ],
      "50002": [
        { "id": "10001", "nickname": "小明", "sex": 1, "avatar": "https://cdn.hichat.com/avatar/10001.jpg" }
      ]
    }
  }
}
```

---

### 3.3.4 获取点赞用户列表

- **接口**: `GET /v1/trend/like/users`
- **认证**: 需要
- **描述**: 获取指定动态的点赞用户列表（游标分页）

**请求参数 (Query/JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| trend_id | uint32 | 是 | 动态ID |
| cursor | uint32 | 否 | 游标位置 |
| limit | uint32 | 否 | 每页数量 |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| users | []User | 点赞用户列表 |
| next_cursor | int | 下一页游标 |
| has_more | bool | 是否还有更多数据 |
| total | int | 总点赞数 |

**请求示例**

```bash
curl -X GET "http://localhost:8080/v1/trend/like/users?trend_id=50001&limit=20" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "users": [
      { "id": "10002", "nickname": "小红", "sex": 2, "avatar": "https://cdn.hichat.com/avatar/10002.jpg" },
      { "id": "10003", "nickname": "小王", "sex": 1, "avatar": "" }
    ],
    "next_cursor": 2,
    "has_more": true,
    "total": 12
  }
}
```

---

### 3.3.5 获取未读点赞

- **接口**: `GET /v1/trend/like/unread`
- **认证**: 需要
- **描述**: 获取用户未读的点赞通知

> ⚠️ **当前未实现**：该独立接口 logic 为空桩。未读点赞已合并进 [3.2.6 获取未读回复](#326-获取未读回复)（`GET /v1/trend/unread`）的 `likes` 字段一并返回，前端从该接口读取未读点赞。

**请求参数 (Query/JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | string | 是 | 用户ID |
| last_id | int | 是 | 上一页最后点赞ID |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| likes_json | string | 未读点赞数据的 JSON 序列化字符串 |
| total_unread | int | 未读总数 |
| last_id | int | 本页最后ID |

**请求示例**

```bash
curl -X GET "http://localhost:8080/v1/trend/like/unread?user_id=10001&last_id=0" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "likes_json": "[{\"user_id\":\"10002\",\"trend_id\":50001}]",
    "total_unread": 3,
    "last_id": 5
  }
}
```

---

### 3.3.6 标记点赞已读

- **接口**: `PUT /v1/trend/like/mark-read`
- **认证**: 需要
- **描述**: 批量标记点赞通知为已读

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| like_ids | []uint64 | 是 | 点赞记录ID列表 |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X PUT http://localhost:8080/v1/trend/like/mark-read \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "like_ids": [1, 2, 3]
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

# 4. 即时通讯服务 (IM Service)

> REST 基础路径: `/v1/im`  
> WebSocket 路径: `/ws`  
> 所有接口均需要认证

## 4.1 聊天记录模块 (REST)

### 4.1.1 获取聊天记录

- **接口**: `GET /v1/im/chatlog`
- **认证**: 需要
- **描述**: 获取指定会话的聊天记录列表

**请求参数 (Query/JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| msgId | string | 是 | 消息ID（游标） |
| conversationId | string | 是 | 会话ID |
| startSendTime | int64 | 否 | 查询开始时间（Unix 时间戳） |
| endSendTime | int64 | 否 | 查询结束时间（Unix 时间戳） |
| count | int64 | 否 | 查询数量 |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| list | []ChatLog | 聊天记录列表 |

**ChatLog 字段明细**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 消息唯一ID |
| conversationId | string | 所属会话ID |
| sendId | string | 发送者用户ID |
| recvId | string | 接收者ID（私聊为用户ID，群聊为群ID） |
| msgType | int32 | 消息类型 (1-文本, 2-文件, 3-语音, 4-图片, 5-表情包, 8-视频) |
| msgContent | string | 消息内容 |
| chatType | int32 | 聊天类型 (1-私聊, 2-群聊) |
| sendTime | int64 | 发送时间（Unix 时间戳） |

**请求示例**

```bash
curl -X GET "http://localhost:8080/v1/im/chatlog?conversationId=conv_10001_10002&msgId=&count=20" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "list": [
      {
        "id": "msg_001",
        "conversationId": "conv_10001_10002",
        "sendId": "10001",
        "recvId": "10002",
        "msgType": 1,
        "msgContent": "你好，在吗？",
        "chatType": 1,
        "sendTime": 1712620800
      },
      {
        "id": "msg_002",
        "conversationId": "conv_10001_10002",
        "sendId": "10002",
        "recvId": "10001",
        "msgType": 1,
        "msgContent": "在的，什么事？",
        "chatType": 1,
        "sendTime": 1712620860
      }
    ]
  }
}
```

---

### 4.1.2 获取消息已读记录

- **接口**: `GET /v1/im/chatlog/readRecords`
- **认证**: 需要
- **描述**: 获取指定消息的已读/未读用户列表（群聊场景）

**请求参数 (Query/JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| msgId | string | 是 | 消息ID |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| reads | []string | 已读用户ID列表 |
| unReads | []string | 未读用户ID列表 |

**请求示例**

```bash
curl -X GET "http://localhost:8080/v1/im/chatlog/readRecords?msgId=msg_001" \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "reads": ["10002", "10003"],
    "unReads": ["10004", "10005"]
  }
}
```

---

## 4.2 会话管理模块 (REST)

### 4.2.1 建立会话

- **接口**: `POST /v1/im/setup/conversation`
- **认证**: 需要
- **描述**: 创建一个新的聊天会话（私聊/群聊）

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sendId | string | 否 | 发起者用户ID |
| recvId | string | 否 | 接收者ID（私聊为用户ID，群聊为群ID） |
| chatType | int32 | 否 | 聊天类型 (1-私聊, 2-群聊) |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X POST http://localhost:8080/v1/im/setup/conversation \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "sendId": "10001",
    "recvId": "10002",
    "chatType": 1
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

### 4.2.2 获取会话列表

- **接口**: `GET /v1/im/conversation`
- **认证**: 需要
- **描述**: 获取当前用户的所有会话列表

**请求参数**

无

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| conversationList | map[string]Conversation | 会话列表，key 为会话ID |

**Conversation 字段明细**

| 字段 | 类型 | 说明 |
|------|------|------|
| conversationId | string | 会话ID |
| chatType | int32 | 聊天类型 (1-私聊, 2-群聊) |
| isShow | bool | 是否在会话列表中显示 |
| seq | int64 | 消息序列号 |
| read | int32 | 已读消息数 |
| message | ChatLog | 最后一条消息，结构见 [4.1.1](#411-获取聊天记录) |

**请求示例**

```bash
curl -X GET http://localhost:8080/v1/im/conversation \
  -H "Authorization: Bearer eyJhbGci..."
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "conversationList": {
      "conv_10001_10002": {
        "conversationId": "conv_10001_10002",
        "chatType": 1,
        "isShow": true,
        "seq": 100,
        "read": 98,
        "message": {
          "id": "msg_002",
          "conversationId": "conv_10001_10002",
          "sendId": "10002",
          "recvId": "10001",
          "msgType": 1,
          "msgContent": "在的，什么事？",
          "chatType": 1,
          "sendTime": 1712620860
        }
      }
    }
  }
}
```

---

### 4.2.3 更新会话

- **接口**: `PUT /v1/im/conversation`
- **认证**: 需要
- **描述**: 批量更新会话状态（显示/隐藏、已读数等）

**请求参数 (Body - JSON)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| conversationList | map[string]Conversation | 是 | 要更新的会话列表，key 为会话ID |

**Conversation 字段明细**

| 字段 | 类型 | 说明 |
|------|------|------|
| conversationId | string | 会话ID |
| chatType | int32 | 聊天类型 (1-私聊, 2-群聊) |
| isShow | bool | 是否显示 |
| seq | int64 | 消息序列号 |
| read | int32 | 已读消息数 |
| message | ChatLog | 最后一条消息 |

**响应参数 (data)**

无额外字段（空对象）

**请求示例**

```bash
curl -X PUT http://localhost:8080/v1/im/conversation \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "conversationList": {
      "conv_10001_10002": {
        "conversationId": "conv_10001_10002",
        "chatType": 1,
        "isShow": true,
        "read": 100
      }
    }
  }'
```

**响应示例**

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

---

## 4.3 WebSocket 通讯模块

> WebSocket 连接地址: `ws://localhost:8080/ws`  
> 认证方式: 连接时通过请求参数传递用户身份信息

### 通用消息帧结构 (Message)

所有 WebSocket 消息均使用统一的 JSON 格式:

```json
{
  "id": "消息ID",
  "ackSeq": 0,
  "frameType": 0,
  "method": "处理方法名",
  "userId": "接收者ID",
  "formId": "发送者ID",
  "data": {}
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 消息唯一ID |
| ackSeq | int | ACK 确认序号（用于消息可靠投递） |
| frameType | int | 帧类型，见下方枚举 |
| method | string | 业务处理方法名 |
| userId | string | 消息接收者用户ID |
| formId | string | 消息发送者用户ID |
| data | object | 消息业务数据（结构因 method 而异） |

**frameType 枚举**

| 值 | 名称 | 说明 |
|------|------|------|
| 0x0 | FrameData | 普通业务消息 |
| 0x1 | FramePing | 心跳检测消息 |
| 0x2 | FrameErr | 错误消息 |
| 0x3 | FrameAck | ACK 应答消息 |
| 0x4 | FrameNoAck | 无需应答的消息 |
| 0x5 | FrameCAck | 后续应答消息 |

**ACK 机制说明**

| 模式 | 说明 |
|------|------|
| NoAck (0) | 不进行 ACK 确认 |
| OnlyAck (1) | 只回模式 - 客户端发送消息，服务端 ACK 后处理业务（2次通信） |
| RigorAck (2) | 严格模式 - 客户端发送，服务端 ACK，客户端确认，服务端处理业务（3次通信） |

---

### 4.3.1 心跳检测

- **Method**: `chat.ping`
- **描述**: 客户端发送心跳保活

**请求消息**

```json
{
  "id": "ping_001",
  "frameType": 0,
  "method": "chat.ping",
  "data": {}
}
```

**响应消息**

```json
{
  "frameType": 0,
  "data": "pong"
}
```

---

### 4.3.2 获取在线用户

- **Method**: `user.online`
- **描述**: 获取当前在线的用户列表

**请求消息**

```json
{
  "id": "online_001",
  "frameType": 0,
  "method": "user.online",
  "data": {}
}
```

**响应消息**

```json
{
  "frameType": 0,
  "formId": "10001",
  "data": ["10001", "10002", "10004"]
}
```

---

### 4.3.3 发送聊天消息

- **Method**: `chat.user`
- **描述**: 发送聊天消息（私聊/群聊），消息通过 MQ 异步投递

**请求消息 data 结构 (Chat)**

| 字段 | 类型 | 说明 |
|------|------|------|
| conversationId | string | 会话ID |
| chatType | int | 聊天类型 (1-私聊, 2-群聊) |
| sendId | string | 发送者用户ID |
| recvId | string | 接收者ID（私聊为用户ID，群聊为群ID） |
| sendTime | int64 | 发送时间（Unix 时间戳） |
| msg | Msg | 消息内容 |

**Msg 结构**

| 字段 | 类型 | 说明 |
|------|------|------|
| mType | int | 消息类型 (1-文本, 2-文件, 3-语音, 4-图片, 5-表情包, 8-视频) |
| content | string | 消息内容 |
| readRecords | map[string]string | 已读记录，key 为消息ID，value 为状态 |

**请求消息示例**

```json
{
  "id": "chat_001",
  "frameType": 0,
  "method": "chat.user",
  "data": {
    "conversationId": "conv_10001_10002",
    "chatType": 1,
    "sendId": "10001",
    "recvId": "10002",
    "sendTime": 1712620800,
    "msg": {
      "mType": 1,
      "content": "你好，在吗？",
      "readRecords": {}
    }
  }
}
```

**说明**: 消息发送后不会立即收到响应，而是通过 `push` 方法异步推送给接收方。

---

### 4.3.4 消息推送 (服务端 -> 客户端)

- **Method**: `push`
- **描述**: 服务端通过 MQ 消费后将消息推送给在线客户端

**推送消息 data 结构 (Chat)**

| 字段 | 类型 | 说明 |
|------|------|------|
| conversationId | string | 会话ID |
| chatType | int | 聊天类型 (1-私聊, 2-群聊) |
| sendId | string | 发送者用户ID |
| recvId | string | 接收者用户ID |
| sendTime | int64 | 发送时间（Unix 时间戳） |
| msg | Msg | 消息内容，结构见 [4.3.3](#433-发送聊天消息) |

**推送消息示例**

```json
{
  "id": "push_001",
  "frameType": 0,
  "method": "push",
  "formId": "10001",
  "data": {
    "conversationId": "conv_10001_10002",
    "chatType": 1,
    "sendId": "10001",
    "recvId": "10002",
    "sendTime": 1712620800,
    "msg": {
      "mType": 1,
      "content": "你好，在吗？",
      "readRecords": {}
    }
  }
}
```

---

### 4.3.5 标记消息已读

- **Method**: `chat.markChat`
- **描述**: 标记指定消息为已读状态

**请求消息 data 结构 (MarkRead)**

| 字段 | 类型 | 说明 |
|------|------|------|
| chatType | int | 聊天类型 (1-私聊, 2-群聊) |
| recvId | string | 接收者ID |
| conversationId | string | 会话ID |
| sendId | string | 发送者（消息阅读者）用户ID |
| msgIds | []string | 已被阅读的消息ID列表 |
| readRecords | map[string]string | 消息阅读状态，key 为消息ID，value 为状态 |

**请求消息示例**

```json
{
  "id": "mark_001",
  "frameType": 0,
  "method": "chat.markChat",
  "data": {
    "chatType": 1,
    "recvId": "10001",
    "conversationId": "conv_10001_10002",
    "sendId": "10002",
    "msgIds": ["msg_001", "msg_002"],
    "readRecords": {
      "msg_001": "1",
      "msg_002": "1"
    }
  }
}
```

---

## 4.4 富媒体上传模块 (REST)

### 4.4.1 上传富媒体文件

- **接口**: `POST /v1/im/upload`
- **认证**: 需要
- **描述**: 上传图片/视频/文件/语音，返回访问 URL 与媒体类型。单文件上限 100MB，本地存储按类型归档到 `im/<image|video|voice|file>`。

**请求 (multipart/form-data)**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | file | 是 | 待上传的文件（表单字段名固定为 `file`） |

**响应参数 (data)**

| 字段 | 类型 | 说明 |
|------|------|------|
| url | string | 文件访问 URL |
| name | string | 原始文件名 |
| size | int64 | 文件大小（字节） |
| fileType | string | 媒体类型：image / video / voice / file |

**说明**：发送富媒体消息时，前端把元数据（url、缩略图、宽高、时长等）序列化为 JSON 字符串放入 WebSocket 消息的 `content`，并设置对应 `mType`（见消息类型枚举）。发送链路对 `content` 原样透传。

**请求示例**

```bash
curl -X POST "http://localhost:8890/v1/im/upload" \
  -H "Authorization: Bearer eyJhbGci..." \
  -F "file=@/path/to/photo.jpg"
```

**响应示例**

```json
{
  "url": "http://localhost:8887/static/im/image/1717300000_photo.jpg",
  "name": "photo.jpg",
  "size": 204800,
  "fileType": "image"
}
```

---

# 附录: 公共数据结构

## User Service - User

```json
{
  "id": "string - 用户唯一ID",
  "mobile": "string - 手机号",
  "nickname": "string - 昵称",
  "sex": "int - 性别 (0-未知, 1-男, 2-女)",
  "avatar": "string - 头像URL",
  "lastLogin": "string - 最后登录时间",
  "introduction": "string - 个性签名/简介",
  "email": "string - 邮箱地址",
  "region": "string - 所在地区",
  "occupation": "string - 职业",
  "tags": "string - 个人标签（JSON 数组字符串）"
}
```

## Social/Trend Service - User (简化)

```json
{
  "id": "string - 用户ID",
  "nickname": "string - 昵称",
  "sex": "int - 性别 (0-未知, 1-男, 2-女)",
  "avatar": "string - 头像URL"
}
```

## 消息类型枚举 (MType)

| 值 | 名称 | 说明 |
|----|------|------|
| 1 | TextMType | 文本消息 |
| 2 | FileMType | 文件消息 |
| 3 | VoiceMType | 语音消息 |
| 4 | ImageMType | 图片消息 |
| 5 | MemesMType | 表情包消息 |
| 8 | VideoMType | 视频消息（追加在 6/7 控制类型之后） |

## 聊天类型枚举 (ChatType)

| 值 | 名称 | 说明 |
|----|------|------|
| 1 | SingleChatType | 私聊 |
| 2 | GroupChatType | 群聊 |

## 动态类型枚举

| 值 | 说明 |
|----|------|
| 1 | 纯文本 |
| 2 | 图文混合 |
| 3 | 文章 |
| 4 | 分享 |
| 5 | 视频 |
| 6 | 广告 |

## 好友申请处理结果

| 值 | 说明 |
|----|------|
| 0 | 待处理 |
| 1 | 已同意 |
| 2 | 已拒绝 |
| 3 | 已忽略 |

## 群成员角色

| 值 | 说明 |
|----|------|
| 1 | 普通成员 |
| 2 | 管理员 |
| 3 | 群主 |
