---
title: go-hichat-api
language_tabs:
  - shell: Shell
  - http: HTTP
  - go: Go
toc_footers: []
includes: []
search: true
code_clipboard: true
highlight_theme: darkula
headingLevel: 2
generator: "@tarslib/widdershins v4.0.30"

---

# go-hichat-api

Base URLs:

* <a href="http://127.0.0.1:8888">测试环境: http://127.0.0.1:8888</a>

* <a href="http://127.0.0.1:8889">social-api: http://127.0.0.1:8889</a>

# Authentication

# user

## POST 注册

POST /api/v1/user/register

> Body 请求参数

```json
{
  "nickname": "范丞丞",
  "phone": "17585710989",
  "password": "dsafasf",
  "sex": 0,
  "avatar": "http://127.0.0.1:8082/avatar/101.png"
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 是 |none|
|body|body|object| 否 |none|
|» nickname|body|string| 是 |none|
|» phone|body|string| 是 |none|
|» password|body|string| 是 |none|
|» sex|body|integer| 是 |none|
|» avatar|body|string| 是 |none|

> 返回示例

> 200 Response

```json
{
  "code": 200,
  "msg": "",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTU2MDI2MDgsImhpY2hhdDIuY29tIjoiMCIsImlhdCI6MTc0Njk2MjYwOH0.y9X7swqjETlzG-8FmcHJNupCv8WMqB-7wDXVW-4QOPg",
    "expire": 1755602608
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|string|

## POST 登录

POST /api/v2/user/login

> Body 请求参数

```json
{
  "phone": "17585710998",
  "password": "dsafasf",
  "nickname": "17585710992",
  "sex": 1,
  "avatar": "http://hichat2.com/user/avatar/101.png"
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 是 |none|
|body|body|object| 否 |none|
|» phone|body|string| 是 |none|
|» password|body|string| 是 |none|
|» nickname|body|string| 是 |none|
|» sex|body|integer| 是 |none|
|» avatar|body|string| 是 |none|

> 返回示例

> 200 Response

```json
{
  "code": 200,
  "msg": "",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTU5NjMyNzQsImhpY2hhdDIuY29tIjoiMTEiLCJpYXQiOjE3NDczMjMyNzR9.c857dKgjh7_mQ6U4WlblgOpq22oNh8oEdVw0lB9RIUs",
    "expire": 1755963274
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» code|integer|true|none||none|
|» msg|string|true|none||none|
|» data|object|true|none||none|
|»» token|string|true|none||none|
|»» expire|integer|true|none||none|

## GET 获取用户信息

GET /api/v1/user/detail

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 是 |none|

> 返回示例

> 200 Response

```json
{
  "code": 200,
  "msg": "",
  "data": {
    "info": {
      "id": "9",
      "mobile": "17585710997",
      "nickname": "17585710992",
      "sex": 1,
      "avatar": "http://hichat2.com/user/avatar/101.png",
      "lastLogin": "1747321585",
      "Introduction": "",
      "email": ""
    }
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» code|integer|true|none||none|
|» msg|string|true|none||none|
|» data|object|true|none||none|
|»» info|object|true|none||none|
|»»» id|string|true|none||none|
|»»» mobile|string|true|none||none|
|»»» nickname|string|true|none||none|
|»»» sex|integer|true|none||none|
|»»» avatar|string|true|none||none|
|»»» lastLogin|string|true|none||none|
|»»» Introduction|string|true|none||none|
|»»» email|string|true|none||none|

# social

## POST 好友-申请

POST /friend/putIn

> Body 请求参数

```json
{
  "req_msg": "我是你大学同学",
  "user_uid": "9",
  "req_time": 0
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 是 |none|
|body|body|object| 否 |none|
|» req_msg|body|string| 是 |none|
|» user_uid|body|string| 是 |none|
|» req_time|body|integer| 是 |none|

> 返回示例

> 200 Response

```json
{}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

## PUT 好友-处理好友申请

PUT /v1/social/friend/putIn

> Body 请求参数

```json
{
  "friend_req_id": 7,
  "handle_result": 1
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 是 |none|
|body|body|object| 否 |none|
|» friend_req_id|body|integer| 是 |none|
|» handle_result|body|integer| 是 |none|

> 返回示例

> 200 Response

```json
{}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

## GET 好友-获取申请列表

GET /v1/socialfriend/putIns

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|type|query|string| 否 |0：未处理，1：通过，2：拒绝|
|class|query|string| 否 |1： 我的申请， 2：请求我为好友的|
|Authorization|header|string| 是 |none|

> 返回示例

> 200 Response

```json
{
  "list": [
    {
      "id": 6,
      "user_id": "10",
      "req_uid": "9",
      "req_msg": "我是你大学同学",
      "req_time": 1747205051
    },
    {
      "id": 7,
      "user_id": "12",
      "req_uid": "9",
      "req_msg": "我是你大学同学",
      "req_time": 1747205167,
      "handle_result": 1
    }
  ]
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» list|[object]|true|none||none|
|»» id|integer|true|none||none|
|»» user_id|string|true|none||none|
|»» req_uid|string|true|none||none|
|»» req_msg|string|true|none||none|
|»» req_time|integer|true|none||none|
|»» handle_result|integer|false|none||none|

## GET 好友-列表

GET /v1/social/friends

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 是 |none|

> 返回示例

> 200 Response

```json
{
  "list": [
    {
      "id": 7,
      "friend_uid": "10",
      "nickname": "17585710992",
      "avatar": "http://hichat2.com/user/avatar/101.png",
      "remark": "17585710992"
    },
    {
      "id": 12,
      "friend_uid": "12",
      "nickname": "范丞丞",
      "avatar": "http://127.0.0.1:8082/avatar/101.png",
      "remark": "范丞丞"
    }
  ]
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» list|[object]|true|none||none|
|»» id|integer|true|none||none|
|»» friend_uid|string|true|none||none|
|»» nickname|string|true|none||none|
|»» avatar|string|true|none||none|
|»» remark|string|true|none||none|

## POST 群聊-创建群聊

POST /v1/social/group

> Body 请求参数

```json
{
  "name": "rust开发交流群",
  "icon": "http://127.0.0.1:8000/hichat/social/group/102.png"
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 是 |none|
|body|body|object| 否 |none|
|» name|body|string| 是 |none|
|» icon|body|string| 是 |none|

> 返回示例

> 200 Response

```json
{}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

## GET 群聊-用户群列表

GET /v1/socail/groups

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 是 |none|

> 返回示例

> 200 Response

```json
{
  "list": [
    {
      "id": "12",
      "name": "BTC梭哈",
      "icon": "http://127.0.0.1:8000/hichat/social/group/101.png",
      "create_uid": "9",
      "group_type": 1,
      "is_verify": true,
      "notification_uid": "0"
    },
    {
      "id": "13",
      "name": "rust开发交流群",
      "icon": "http://127.0.0.1:8000/hichat/social/group/102.png",
      "create_uid": "9",
      "group_type": 1,
      "is_verify": true,
      "notification_uid": "0"
    }
  ]
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» list|[object]|true|none||none|
|»» id|string|true|none||none|
|»» name|string|true|none||none|
|»» icon|string|true|none||none|
|»» create_uid|string|true|none||none|
|»» group_type|integer|true|none||none|
|»» is_verify|boolean|true|none||none|
|»» notification_uid|string|true|none||none|

## POST 群聊-申请加群

POST /v1/social/group/putIn

> Body 请求参数

```json
{
  "group_id": "13",
  "req_msg": "我是一个php程序员，目前正在学习rust",
  "req_time": 0,
  "join_source": 1,
  "inviter_uid": "0"
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 是 |none|
|body|body|object| 否 |none|
|» group_id|body|string| 是 |none|
|» req_msg|body|string| 是 |none|
|» req_time|body|integer| 是 |none|
|» join_source|body|integer| 是 |none|
|» inviter_uid|body|string| 是 |none|

> 返回示例

> 200 Response

```json
{
  "group_id": 13
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» group_id|integer|true|none||none|

## PUT 群聊-处理加群申请

PUT /v1/social/group/putIn

> Body 请求参数

```json
{
  "group_req_id": 8,
  "group_id": "13",
  "handle_result": 2
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 是 |none|
|body|body|object| 否 |none|
|» group_req_id|body|integer| 是 |none|
|» group_id|body|string| 是 |none|
|» handle_result|body|integer| 是 |none|

> 返回示例

> 200 Response

```json
{}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

## GET 群聊-获取加群申请

GET /v1/social/group/putIns

> Body 请求参数

```json
{
  "group_id": "13",
  "type": [],
  "class": 2
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 是 |none|
|body|body|object| 否 |none|
|» group_id|body|string| 是 |none|
|» type|body|[string]| 是 |none|
|» class|body|integer| 是 |none|

> 返回示例

> 200 Response

```json
{
  "list": [
    {
      "id": 1,
      "user": {
        "id": "",
        "nickname": "",
        "sex": 0,
        "avatar": "",
        "Introduction": ""
      },
      "group": {
        "id": "13",
        "name": "rust开发交流群",
        "icon": "http://127.0.0.1:8000/hichat/social/group/102.png",
        "create_uid": "9"
      },
      "req_msg": "我是来学习的",
      "req_time": 1342343,
      "join_source": 1,
      "inviter_user_id": "10",
      "handle_user_id": "10",
      "handle_time": 1747149535,
      "handle_result": 2
    },
    {
      "id": 2,
      "user": {
        "id": "9",
        "nickname": "17585710992",
        "sex": 1,
        "avatar": "http://hichat2.com/user/avatar/101.png",
        "Introduction": ""
      },
      "group": {
        "id": "13",
        "name": "rust开发交流群",
        "icon": "http://127.0.0.1:8000/hichat/social/group/102.png",
        "create_uid": "9"
      },
      "req_msg": "我是大户",
      "req_time": 1747323719,
      "join_source": 1,
      "inviter_user_id": "9",
      "handle_user_id": "9",
      "handle_time": 1747323719,
      "handle_result": 1
    },
    {
      "id": 8,
      "user": {
        "id": "",
        "nickname": "",
        "sex": 0,
        "avatar": "",
        "Introduction": ""
      },
      "group": {
        "id": "13",
        "name": "rust开发交流群",
        "icon": "http://127.0.0.1:8000/hichat/social/group/102.png",
        "create_uid": "9"
      },
      "req_msg": "我是一个php程序员，目前正在学习rust",
      "req_time": 1747405404,
      "join_source": 1,
      "inviter_user_id": "0",
      "handle_time": 1747405404
    }
  ]
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» list|[object]|true|none||none|
|»» id|integer|true|none||none|
|»» user|object|true|none||none|
|»»» id|string|true|none||none|
|»»» nickname|string|true|none||none|
|»»» sex|integer|true|none||none|
|»»» avatar|string|true|none||none|
|»»» Introduction|string|true|none||none|
|»» group|object|true|none||none|
|»»» id|string|true|none||none|
|»»» name|string|true|none||none|
|»»» icon|string|true|none||none|
|»»» create_uid|string|true|none||none|
|»» req_msg|string|true|none||none|
|»» req_time|integer|true|none||none|
|»» join_source|integer|true|none||none|
|»» inviter_user_id|string|true|none||none|
|»» handle_user_id|string|true|none||none|
|»» handle_time|integer|true|none||none|
|»» handle_result|integer|true|none||none|

## GET 群聊-获取群成员

GET /v/social/group/users

> Body 请求参数

```json
{
  "group_id": "13"
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 是 |none|
|body|body|object| 否 |none|
|» group_id|body|string| 是 |none|

> 返回示例

> 200 Response

```json
{
  "List": [
    {
      "id": 6,
      "group_id": "13",
      "user": {
        "id": "9",
        "nickname": "17585710992",
        "sex": 1,
        "avatar": "http://hichat2.com/user/avatar/101.png",
        "Introduction": "",
        "is_current_user": 1
      },
      "role_level": 2,
      "inviter_uid": "9"
    },
    {
      "id": 4,
      "group_id": "13",
      "user": {
        "id": "10",
        "nickname": "17585710992",
        "sex": 1,
        "avatar": "http://hichat2.com/user/avatar/101.png",
        "Introduction": "",
        "is_current_user": 0
      },
      "role_level": 3,
      "inviter_uid": "10"
    }
  ]
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» List|[object]|true|none||none|
|»» id|integer|true|none||none|
|»» group_id|string|true|none||none|
|»» user|object|true|none||none|
|»»» id|string|true|none||none|
|»»» nickname|string|true|none||none|
|»»» sex|integer|true|none||none|
|»»» avatar|string|true|none||none|
|»»» Introduction|string|true|none||none|
|»»» is_current_user|integer|true|none||none|
|»» role_level|integer|true|none||none|
|»» inviter_uid|string|true|none||none|

# 数据模型

