# Neuro 阅读 App API 文档

> Android 客户端 API 接口文档
> 版本：v1.3
> 日期：2026-05-05

---

## 目录

1. [通用规范](#通用规范)
2. [认证模块](#认证模块)
3. [用户模块](#用户模块)
4. [书架模块](#书架模块)
5. [文章模块](#文章模块)
6. [评论模块](#评论模块)
7. [段评模块](#段评模块)
8. [作者模块](#作者模块)
9. [动态/关注流模块](#动态关注流模块)
10. [创作中心模块](#创作中心模块)
11. [数据模型汇总](#数据模型汇总)

---

## 通用规范

### Base URL

```
http://47.118.22.220:9091/api/v1
```

### 请求格式

- Content-Type: `application/json`
- 所有请求参数均为 JSON 格式
- 请求体中不要包含值为 `null` 的字段（可选字段未填写时直接省略）

### 响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| code | int | 业务状态码，0=成功，其他=错误 |
| message | string | 提示信息 |
| data | object | 响应数据，失败时可能为 null |

### 分页响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [],
    "total": 100,
    "page": 1,
    "pageSize": 20,
    "hasMore": true
  }
}
```

### 认证方式

除登录相关接口外，所有接口需在请求头中携带 Token：

```
Authorization: Bearer {access_token}
```

### 状态码

| Code | 含义 |
|------|------|
| 0 | 成功 |
| 1001 | 参数错误 |
| 1002 | 未授权/Token过期 |
| 1003 | 禁止访问 |
| 1004 | 资源不存在 |
| 1005 | 服务器内部错误 |
| 2001 | 账号已存在 |
| 2002 | 账号或密码错误 |
| 2003 | 验证码错误/过期 |
| 2004 | 密码不一致 |
| 2005 | 账号不存在 |

---

## 认证模块

### 1. 发送验证码

```http
POST /api/v1/auth/send-code
```

**请求体：**

```json
{
  "account": "123456789@qq.com",
  "type": "email"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| account | string | 是 | QQ邮箱地址 |
| type | string | 是 | 固定传 `email`（通过QQ邮箱SMTP发送） |

**响应：**

```json
{
  "code": 0,
  "message": "验证码已发送",
  "data": null
}
```

**业务说明：**
- 验证码有效期 5 分钟
- 同一账号 60 秒内只能发送一次
- 验证码通过 `3949110906@qq.com` 的 QQ邮箱 SMTP 服务发送到目标邮箱
- 前端应引导用户输入QQ邮箱地址接收验证码

---

### 2. 登录

```http
POST /api/v1/auth/login
```

**请求体：**

```json
{
  "account": "123456789",
  "password": "e10adc3949ba59abbe56e057f20f883e"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| account | string | 是 | QQ号或邮箱（同注册时的账号） |
| password | string | 是 | MD5加密后的密码（32位小写） |

**响应：**

```json
{
  "code": 0,
  "message": "登录成功",
  "data": {
    "userId": "u_123456",
    "account": "123456789",
    "nickname": "用户123456",
    "avatar": "https://example.com/avatar.jpg",
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIs...",
    "expiresIn": 604800
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| userId | string | 用户唯一标识（重要：用于替代account作为用户标识） |
| account | string | 登录账号 |
| nickname | string | 用户昵称 |
| avatar | string | 头像URL |
| token | string | Access Token |
| refreshToken | string | Refresh Token |
| expiresIn | long | Token过期时间（秒），7天=604800 |

**客户端处理：**
- 登录成功后，客户端需保存 `userId`、`nickname`、`avatar`、`token`、`refreshToken`
- 后续所有需要用户标识的地方使用 `userId` 而非 `account`

---

### 3. 注册

```http
POST /api/v1/auth/register
```

**请求体：**

```json
{
  "account": "123456789",
  "code": "123456",
  "password": "e10adc3949ba59abbe56e057f20f883e",
  "confirmPassword": "e10adc3949ba59abbe56e057f20f883e",
  "nickname": "书虫小明"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| account | string | 是 | QQ号或邮箱（用作登录账号） |
| code | string | 是 | 6位验证码 |
| password | string | 是 | MD5加密后的密码 |
| confirmPassword | string | 是 | 确认密码，需与password一致 |
| nickname | string | 否 | 用户昵称，不传则自动生成 |

**响应：** 同登录接口（注册成功后自动登录）

**业务说明：**
- `account` 字段即登录用的账号，与发送验证码时填写的 `account` 一致
- `nickname` 可选，不传则自动生成默认昵称：`用户{account后6位}`
- 注册成功后直接返回登录凭证，客户端无需再次调用登录接口

---

### 4. 忘记密码

```http
POST /api/v1/auth/forgot-password
```

**请求体：**

```json
{
  "account": "123456789",
  "code": "123456",
  "newPassword": "e10adc3949ba59abbe56e057f20f883e"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| account | string | 是 | QQ号或邮箱 |
| code | string | 是 | 6位验证码 |
| newPassword | string | 是 | MD5加密后的新密码 |

**响应：**

```json
{
  "code": 0,
  "message": "密码重置成功",
  "data": null
}
```

---

### 5. 刷新Token

```http
POST /api/v1/auth/refresh-token
```

**请求体：**

```json
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIs..."
}
```

**响应：** 同登录接口

**业务说明：**
- Access Token 有效期 7 天，Refresh Token 有效期 30 天
- 当 Access Token 过期时（code=1002），客户端应自动使用 Refresh Token 调用此接口获取新 Token
- 如果 Refresh Token 也过期，需要重新登录

---

## 用户模块

### 6. 获取用户资料

```http
GET /api/v1/user/profile
```

**请求头：** `Authorization: Bearer {token}`

**响应：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "userId": "u_123456",
    "account": "123456789",
    "nickname": "书虫小明",
    "avatar": "https://example.com/avatar.jpg",
    "bio": "热爱阅读",
    "gender": 1,
    "followingCount": 12,
    "bookshelfCount": 8,
    "readDuration": 3600
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| userId | string | 用户唯一标识 |
| gender | int | 0=保密, 1=男, 2=女 |
| readDuration | long | 累计阅读时长（分钟） |

---

### 7. 更新用户资料

```http
PUT /api/v1/user/profile
```

**请求头：** `Authorization: Bearer {token}`

**请求体：**

```json
{
  "nickname": "新昵称",
  "avatar": "https://example.com/new_avatar.jpg",
  "bio": "新的简介",
  "gender": 1
}
```

**说明：** 所有字段均为可选，只传需要修改的字段

**响应：** 同获取用户资料

---

### 8. 上传头像

```http
POST /api/v1/upload/avatar
```

**请求头：**
- `Authorization: Bearer {token}`
- `Content-Type: multipart/form-data`

**请求体：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | File | 是 | 头像图片文件，支持 jpg/png/gif，最大 5MB |

**响应：**

```json
{
  "code": 0,
  "message": "上传成功",
  "data": {
    "filename": "1714723456_1234.jpg",
    "url": "http://47.118.22.220:9091/uploads/1714723456_1234.jpg"
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| filename | string | 文件名 |
| url | string | 完整访问 URL |

**业务说明：**
- 上传成功后，将返回的 `url` 传给 `PUT /api/v1/user/profile` 更新头像
- 头像会自动压缩为 256x256，JPEG 质量 80%
- 上传新头像后会自动删除旧头像

---

### 9. 关注作者

```http
POST /api/v1/user/follow/{authorId}
```

**请求头：** `Authorization: Bearer {token}`

**响应：**

```json
{
  "code": 0,
  "message": "关注成功",
  "data": null
}
```

---

### 10. 取消关注

```http
DELETE /api/v1/user/follow/{authorId}
```

**请求头：** `Authorization: Bearer {token}`

**响应：** 同上

---

### 11. 获取关注列表

```http
GET /api/v1/user/following?page=1&pageSize=20
```

**请求头：** `Authorization: Bearer {token}`

**响应：** 分页数据，data.list 为 AuthorResponse 数组

---

## 书架模块

### 12. 获取书架列表

```http
GET /api/v1/bookshelf?page=1&pageSize=20
```

**请求头：** `Authorization: Bearer {token}`

**响应：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "articleId": "ar_001",
        "title": "仿生之心",
        "author": "未知作者",
        "cover": "",
        "lastReadChapter": "正文",
        "lastReadTime": "2024-01-15 14:30:00",
        "progress": 65,
        "isFinished": false
      }
    ],
    "total": 8,
    "page": 1,
    "pageSize": 20,
    "hasMore": false
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| lastReadChapter | string | 上次阅读到的章节标题 |
| lastReadTime | string | 上次阅读时间，格式：yyyy-MM-dd HH:mm:ss |
| progress | int | 阅读进度 0-100 |
| isFinished | boolean | 是否已读完 |

---

### 13. 加入书架

```http
POST /api/v1/bookshelf/{articleId}
```

**请求头：** `Authorization: Bearer {token}`

**响应：**

```json
{
  "code": 0,
  "message": "已加入书架",
  "data": null
}
```

---

### 14. 移出书架

```http
DELETE /api/v1/bookshelf/{articleId}
```

**请求头：** `Authorization: Bearer {token}`

**响应：** 同上

---

### 15. 更新阅读进度

```http
PUT /api/v1/bookshelf/{articleId}/progress
```

**请求头：** `Authorization: Bearer {token}`

**请求体：**

```json
{
  "chapterIndex": 0,
  "progress": 65,
  "position": 1250
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| chapterIndex | int | 是 | 当前章节索引（从0开始） |
| progress | int | 是 | 阅读进度 0-100 |
| position | int | 是 | 当前阅读位置（字符偏移） |

**响应：** 同上

**业务说明：**
- 客户端应在用户阅读时定期（如每30秒）调用此接口同步进度
- 切换章节时也应调用

---

## 文章模块

### 16. 获取文章列表

```http
GET /api/v1/articles?page=1&pageSize=10
```

**响应：** 分页数据，data.list 为 ArticleIndex 数组

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "articleId": "ar_1234567890",
        "title": "仿生之心",
        "author": "未知作者",
        "summary": "Neuro在仿生人坟场醒来，发现自己身处一个陌生的世界...",
        "wordCount": 15000,
        "chapterCount": 1,
        "tags": ["NE文", "刀子"],
        "lastUpdateTime": "2026-05-04 10:00:00"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 10,
    "hasMore": false
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| articleId | string | 文章唯一标识 |
| title | string | 文章标题 |
| author | string | 作者名（字符串） |
| summary | string | 文章简介/摘要 |
| wordCount | int | 总字数 |
| chapterCount | int | 章节数 |
| tags | array | 标签数组 |
| lastUpdateTime | string | 最后更新时间 |

---

### 17. 搜索文章

```http
GET /api/v1/articles/search?keyword=仿生
```

**响应：** data 为 ArticleIndex 数组（不分页）

---

### 18. 获取文章详情

```http
GET /api/v1/articles/{articleId}
```

**响应：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "articleId": "ar_1234567890",
    "title": "仿生之心",
    "author": "未知作者",
    "summary": "Neuro在仿生人坟场醒来...",
    "tags": ["NE文", "刀子"],
    "wordCount": 15000,
    "chapterCount": 1,
    "status": "published",
    "publishTime": "2026-05-04 10:00:00",
    "lastUpdateTime": "2026-05-04 10:00:00",
    "chapters": [
      {
        "index": 0,
        "chapterId": "c_1234567890",
        "title": "正文",
        "wordCount": 15000
      }
    ]
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| articleId | string | 文章ID |
| title | string | 标题 |
| author | string | 作者名 |
| summary | string | 摘要 |
| tags | array | 标签 |
| wordCount | int | 总字数 |
| chapterCount | int | 章节数 |
| status | string | published/draft |
| publishTime | string | 发布时间 |
| lastUpdateTime | string | 最后更新时间 |
| chapters | array | 章节列表 |

**章节对象：**

| 字段 | 类型 | 说明 |
|------|------|------|
| index | int | 章节索引（从0开始） |
| chapterId | string | 章节ID |
| title | string | 章节标题 |
| wordCount | int | 章节字数 |

---

### 19. 获取章节内容

```http
GET /api/v1/articles/{articleId}/chapters/{chapterIndex}
```

**参数说明：**
- `articleId`: 文章ID
- `chapterIndex`: 章节索引（从0开始的数字，如 0, 1, 2...）

**响应：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "chapterId": "c_1234567890",
    "bookId": "ar_1234567890",
    "title": "正文",
    "content": "“咳呃………”\n\n声带嘶哑地扯出几个音节...",
    "paragraphs": ["段落1", "段落2", "段落3..."],
    "prevChapterId": null,
    "nextChapterId": null,
    "paragraphComments": {}
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| chapterId | string | 章节ID |
| bookId | string | 文章ID（字段名保持兼容） |
| title | string | 章节标题 |
| content | string | 完整正文内容 |
| paragraphs | array | 按段落分割的数组，前端直接渲染 |
| prevChapterId | int | 上一章索引，第一章为 null |
| nextChapterId | int | 下一章索引，最后一章为 null |
| paragraphComments | object | 段落索引 -> 段评数量 |

**业务说明：**
- `paragraphs` 数组前端直接按段落渲染
- 下拉加载下一章时，用 `nextChapterId` 作为新的 `chapterIndex` 请求
- `prevChapterId` 和 `nextChapterId` 是 **int 类型**（章节索引），不是字符串

---

### 20. 上传文章（管理接口）

```http
POST /api/v1/articles/upload
```

**请求头：**
- `Authorization: Bearer {token}`
- `Content-Type: multipart/form-data`

**请求体：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | File | 是 | 文章文件，支持 .txt / .md，最大 10MB |
| title | string | 是 | 文章标题 |
| author | string | 否 | 作者名 |
| summary | string | 否 | 文章简介 |
| tags | string | 否 | 标签，逗号分隔，如 "NE文,刀子,连载" |

**响应：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "articleId": "ar_1234567890",
    "title": "仿生之心",
    "chapters": 3
  }
}
```

**分章规则：**
- `.txt` 文件：识别 `第X章`、`第X节`、`序` 开头的行作为章节标题
- `.md` 文件：识别 `# `、`## ` 开头的行作为章节标题
- 无章节标记的整篇作为一章，标题为"正文"

---

### 21. 删除文章（管理接口）

```http
DELETE /api/v1/articles/{articleId}
```

**请求头：** `Authorization: Bearer {token}`

**响应：**

```json
{
  "code": 0,
  "message": "删除成功",
  "data": null
}
```

---

## 评论模块

### 22. 获取文章评论

```http
GET /api/v1/articles/{articleId}/comments?page=1&pageSize=20&sort=hot
```

**请求头：** `Authorization: Bearer {token}`

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sort | string | 否 | `hot`=热门, `new`=最新, `author`=作者回复 |

**响应：** 分页数据，data.list 为 CommentResponse 数组

---

### 23. 发表评论

```http
POST /api/v1/articles/{articleId}/comments
```

**请求头：** `Authorization: Bearer {token}`

**请求体：**

```json
{
  "content": "写得真好！",
  "parentId": null
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| content | string | 是 | 评论内容，1-1000字 |
| parentId | string | 否 | 回复某条评论时填写，一级评论不传/null |

**响应：** data 为 CommentResponse

---

### 24. 点赞评论

```http
POST /api/v1/comments/{commentId}/like
```

**请求头：** `Authorization: Bearer {token}`

**响应：**

```json
{
  "code": 0,
  "message": "点赞成功",
  "data": null
}
```

---

### 25. 取消点赞

```http
DELETE /api/v1/comments/{commentId}/like
```

**请求头：** `Authorization: Bearer {token}`

**响应：** 同上

---

## 段评模块

### 26. 获取段评列表

```http
GET /api/v1/chapters/{chapterId}/paragraph-comments?paragraphIndex=5
```

**请求头：** `Authorization: Bearer {token}`

**响应：**

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "commentId": "pc_001",
      "chapterId": "c_001",
      "paragraphIndex": 5,
      "userId": "u_001",
      "userName": "书虫小明",
      "userAvatar": "https://example.com/avatar.jpg",
      "content": "这段描写太精彩了！",
      "createTime": "2024-01-15 14:30:00",
      "likeCount": 23,
      "isLiked": false
    }
  ]
}
```

---

### 27. 发表段评

```http
POST /api/v1/chapters/{chapterId}/paragraph-comments
```

**请求头：** `Authorization: Bearer {token}`

**请求体：**

```json
{
  "paragraphIndex": 5,
  "content": "这段描写太精彩了！"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| paragraphIndex | int | 是 | 段落索引（从0开始） |
| content | string | 是 | 段评内容，1-500字 |

**响应：** data 为 ParagraphCommentResponse

---

### 28. 点赞段评

```http
POST /api/v1/paragraph-comments/{commentId}/like
```

**请求头：** `Authorization: Bearer {token}`

**响应：**

```json
{
  "code": 0,
  "message": "点赞成功",
  "data": null
}
```

---

### 29. 取消点赞段评

```http
DELETE /api/v1/paragraph-comments/{commentId}/like
```

**请求头：** `Authorization: Bearer {token}`

**响应：** 同上

---

## 作者模块

### 30. 获取作者主页

```http
GET /api/v1/authors/{authorId}
```

**响应：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "authorId": "a_001",
    "name": "天蚕土豆",
    "avatar": "https://example.com/avatar.jpg",
    "description": "白金作家，代表作《斗破苍穹》《大主宰》",
    "worksCount": 5,
    "followersCount": 1280000,
    "totalWords": 32000000,
    "isFollowing": false
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| worksCount | int | 作品数量 |
| followersCount | int | 粉丝数 |
| totalWords | long | 累计字数 |
| isFollowing | boolean | 当前用户是否已关注 |

---

### 31. 获取作者作品

```http
GET /api/v1/authors/{authorId}/works?page=1&pageSize=20
```

**响应：** 分页数据，data.list 为 ArticleIndex 数组

---

### 32. 获取作者动态

```http
GET /api/v1/authors/{authorId}/activities?page=1&pageSize=20
```

**响应：** 分页数据，data.list 为 AuthorActivityResponse 数组

---

## 动态/关注流模块

### 33. 获取关注流

```http
GET /api/v1/feed?page=1&pageSize=20
```

**请求头：** `Authorization: Bearer {token}`

**响应：** 分页数据，data.list 为 FeedActivityResponse 数组

---

### 34. 点赞动态

```http
POST /api/v1/feed/{feedId}/like
```

**请求头：** `Authorization: Bearer {token}`

**响应：**

```json
{
  "code": 0,
  "message": "点赞成功",
  "data": null
}
```

---

### 35. 取消点赞动态

```http
DELETE /api/v1/feed/{feedId}/like
```

**请求头：** `Authorization: Bearer {token}`

**响应：** 同上

---

## 创作中心模块

### 36. 创作者注册

```http
POST /api/v1/creator/register
```

**请求体：**

```json
{
  "account": "writer001",
  "password": "123456",
  "confirmPassword": "123456",
  "name": "作家小明",
  "email": "writer@example.com"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| account | string | 是 | 账号，3-32字符 |
| password | string | 是 | 密码，6-32字符 |
| confirmPassword | string | 是 | 确认密码 |
| name | string | 是 | 笔名，2-32字符 |
| email | string | 是 | 邮箱 |

**响应：**

```json
{
  "code": 0,
  "message": "注册成功",
  "data": {
    "creatorId": "cr_1234567890",
    "account": "writer001",
    "name": "作家小明",
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

---

### 37. 创作者登录

```http
POST /api/v1/creator/login
```

**请求体：**

```json
{
  "account": "writer001",
  "password": "123456"
}
```

**响应：** 同注册

---

### 38. 获取创作者资料

```http
GET /api/v1/creator/profile
```

**请求头：** `Authorization: Bearer {token}`

**响应：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "creatorId": "cr_1234567890",
    "account": "writer001",
    "name": "作家小明",
    "avatar": "https://example.com/avatar.jpg",
    "description": "专注玄幻小说创作",
    "email": "writer@example.com",
    "articleCount": 5,
    "createTime": "2026-05-04 10:00:00",
    "lastLoginTime": "2026-05-04 15:30:00"
  }
}
```

---

### 39. 更新创作者资料

```http
PUT /api/v1/creator/profile
```

**请求头：** `Authorization: Bearer {token}`

**请求体：**

```json
{
  "name": "新笔名",
  "avatar": "https://example.com/new_avatar.jpg",
  "description": "新的简介"
}
```

**说明：** 所有字段均为可选，只传需要修改的字段

**响应：**

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 40. 创建作品

```http
POST /api/v1/creator/works
```

**请求头：** `Authorization: Bearer {token}`

**请求体：**

```json
{
  "title": "我的小说",
  "summary": "这是一个关于...",
  "tags": ["玄幻", "热血"],
  "cover": "https://example.com/cover.jpg"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | 是 | 作品标题，1-50字 |
| summary | string | 否 | 作品简介，0-500字 |
| tags | array | 否 | 标签数组，最多5个 |
| cover | string | 否 | 封面图片URL |

**响应：**

```json
{
  "code": 0,
  "message": "创建成功",
  "data": {
    "articleId": "ar_1234567890",
    "title": "我的小说",
    "status": "draft"
  }
}
```

---

### 41. 获取我的作品列表

```http
GET /api/v1/creator/works?page=1&pageSize=20&status=all
```

**请求头：** `Authorization: Bearer {token}`

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| status | string | 否 | `all`=全部, `draft`=草稿, `published`=已发布 |

**响应：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "articleId": "ar_1234567890",
        "creatorId": "cr_1234567890",
        "title": "我的小说",
        "summary": "这是一个关于...",
        "cover": "https://example.com/cover.jpg",
        "status": "draft",
        "chapterCount": 5,
        "wordCount": 50000,
        "lastUpdateTime": "2026-05-04 15:30:00"
      }
    ],
    "total": 3,
    "page": 1,
    "pageSize": 20,
    "hasMore": false
  }
}
```

---

### 42. 获取作品详情（编辑用）

```http
GET /api/v1/creator/works/{workId}
```

**请求头：** `Authorization: Bearer {token}`

**响应：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "articleId": "ar_1234567890",
    "creatorId": "cr_1234567890",
    "title": "我的小说",
    "summary": "这是一个关于...",
    "tags": ["玄幻", "热血"],
    "cover": "https://example.com/cover.jpg",
    "status": "draft",
    "chapters": [
      {
        "index": 0,
        "chapterId": "c_001",
        "title": "第一章 开始",
        "wordCount": 3000
      }
    ],
    "wordCount": 50000,
    "publishTime": "2026-05-04 10:00:00",
    "lastUpdateTime": "2026-05-04 15:30:00"
  }
}
```

---

### 43. 更新作品信息

```http
PUT /api/v1/creator/works/{workId}
```

**请求头：** `Authorization: Bearer {token}`

**请求体：**

```json
{
  "title": "新标题",
  "summary": "新的简介",
  "tags": ["玄幻", "热血", "升级流"],
  "cover": "https://example.com/new_cover.jpg"
}
```

**说明：** 所有字段均为可选，只传需要修改的字段

**响应：**

```json
{
  "code": 0,
  "message": "更新成功",
  "data": null
}
```

---

### 44. 删除作品

```http
DELETE /api/v1/creator/works/{workId}
```

**请求头：** `Authorization: Bearer {token}`

**响应：**

```json
{
  "code": 0,
  "message": "删除成功",
  "data": null
}
```

---

### 45. 发布作品

```http
POST /api/v1/creator/works/{workId}/publish
```

**请求头：** `Authorization: Bearer {token}`

**业务说明：**
- 发布前需至少有一个章节
- 发布后作品状态变为 `published`，其他用户可见

**响应：**

```json
{
  "code": 0,
  "message": "发布成功",
  "data": {
    "articleId": "ar_1234567890",
    "status": "published"
  }
}
```

---

### 46. 创建章节

```http
POST /api/v1/creator/works/{workId}/chapters
```

**请求头：** `Authorization: Bearer {token}`

**请求体：**

```json
{
  "title": "第一章 开始",
  "content": "正文内容..."
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | 是 | 章节标题，1-50字 |
| content | string | 是 | 章节内容 |

**响应：**

```json
{
  "code": 0,
  "message": "创建成功",
  "data": {
    "chapterId": "c_1234567890",
    "index": 0,
    "title": "第一章 开始",
    "wordCount": 3000
  }
}
```

---

### 47. 获取章节内容（编辑用）

```http
GET /api/v1/creator/works/{workId}/chapters/{chapterIndex}
```

**请求头：** `Authorization: Bearer {token}`

**参数说明：**
- `workId`: 作品ID（articleId）
- `chapterIndex`: 章节索引（从0开始的数字，如 0, 1, 2...）

**响应：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "chapterId": "c_1234567890",
    "index": 0,
    "title": "第一章 开始",
    "content": "正文内容...",
    "wordCount": 3000
  }
}
```

---

### 48. 更新章节

```http
PUT /api/v1/creator/works/{workId}/chapters/{chapterIndex}
```

**请求头：** `Authorization: Bearer {token}`

**参数说明：**
- `chapterIndex`: 章节索引（从0开始的数字）

**请求体：**

```json
{
  "title": "新的章节标题",
  "content": "新的章节内容..."
}
```

**说明：** 所有字段均为可选，只传需要修改的字段

**响应：**

```json
{
  "code": 0,
  "message": "更新成功",
  "data": null
}
```

---

### 49. 删除章节

```http
DELETE /api/v1/creator/works/{workId}/chapters/{chapterIndex}
```

**请求头：** `Authorization: Bearer {token}`

**参数说明：**
- `chapterIndex`: 章节索引（从0开始的数字）

**响应：**

```json
{
  "code": 0,
  "message": "删除成功",
  "data": null
}
```

---

### 50. 上传 docx 文档

```http
POST /api/v1/creator/works/upload/docx
```

**请求头：**
- `Authorization: Bearer {token}`
- `Content-Type: multipart/form-data`

**请求参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | File | 是 | docx 文件 |
| title | string | 否 | 作品标题，不传则使用文件名 |

**响应：**

```json
{
  "code": 0,
  "message": "上传成功",
  "data": {
    "articleId": "ar_1234567890",
    "title": "我的小说",
    "chapterCount": 1,
    "wordCount": 5000
  }
}
```

**业务说明：**
- 上传后自动创建作品，状态为 `draft`
- 文件内容作为第一章，标题为作品标题
- 自动生成摘要（取前200字）

---

## 数据模型汇总

### AuthorResponse

```json
{
  "authorId": "string",
  "name": "string",
  "avatar": "string",
  "description": "string"
}
```

### ArticleIndex

```json
{
  "articleId": "string",
  "title": "string",
  "author": "string",
  "summary": "string",
  "wordCount": 0,
  "chapterCount": 0,
  "tags": ["string"],
  "lastUpdateTime": "string"
}
```

### ArticleMeta

```json
{
  "articleId": "string",
  "title": "string",
  "author": "string",
  "summary": "string",
  "tags": ["string"],
  "wordCount": 0,
  "chapterCount": 0,
  "status": "string",
  "publishTime": "string",
  "lastUpdateTime": "string",
  "chapters": [
    {
      "index": 0,
      "chapterId": "string",
      "title": "string",
      "wordCount": 0
    }
  ]
}
```

### ChapterContentResponse

```json
{
  "chapterId": "string",
  "bookId": "string",
  "title": "string",
  "content": "string",
  "paragraphs": ["string"],
  "prevChapterId": 0,
  "nextChapterId": 1,
  "paragraphComments": {
    "0": 0
  }
}
```

**重要变更：**
- `prevChapterId` / `nextChapterId` 从 **string** 改为 **int**（章节索引）
- 新增 `paragraphs` 数组，前端直接按段落渲染

### CommentResponse

```json
{
  "commentId": "string",
  "articleId": "string",
  "userId": "string",
  "userName": "string",
  "userAvatar": "string",
  "content": "string",
  "createTime": "string",
  "likeCount": 0,
  "isLiked": false,
  "replyCount": 0,
  "replies": ["CommentReplyResponse"]
}
```

### CommentReplyResponse

```json
{
  "replyId": "string",
  "userId": "string",
  "userName": "string",
  "userAvatar": "string",
  "content": "string",
  "createTime": "string",
  "toUserName": "string"
}
```

### ParagraphCommentResponse

```json
{
  "commentId": "string",
  "chapterId": "string",
  "paragraphIndex": 0,
  "userId": "string",
  "userName": "string",
  "userAvatar": "string",
  "content": "string",
  "createTime": "string",
  "likeCount": 0,
  "isLiked": false
}
```

### AuthorProfileResponse

```json
{
  "authorId": "string",
  "name": "string",
  "avatar": "string",
  "description": "string",
  "worksCount": 0,
  "followersCount": 0,
  "totalWords": 0,
  "isFollowing": false
}
```

### AuthorActivityResponse

```json
{
  "activityId": "string",
  "authorId": "string",
  "authorName": "string",
  "authorAvatar": "string",
  "type": "string",
  "content": "string",
  "articleId": "string",
  "articleTitle": "string",
  "chapterTitle": "string",
  "chapterPreview": "string",
  "readHeat": "string",
  "createTime": "string",
  "likeCount": 0,
  "commentCount": 0
}
```

### FeedActivityResponse

```json
{
  "feedId": "string",
  "authorId": "string",
  "authorName": "string",
  "authorAvatar": "string",
  "publishTime": "string",
  "activityContent": "string",
  "articleId": "string",
  "articleCover": "string",
  "chapterPreview": "string",
  "readHeat": "string",
  "likeCount": 0,
  "commentCount": 0,
  "isLiked": false
}
```

### UserProfileResponse

```json
{
  "userId": "string",
  "account": "string",
  "nickname": "string",
  "avatar": "string",
  "bio": "string",
  "gender": 0,
  "followingCount": 0,
  "bookshelfCount": 0,
  "readDuration": 0
}
```

### BookshelfItemResponse

```json
{
  "articleId": "string",
  "title": "string",
  "author": "string",
  "cover": "string",
  "lastReadChapter": "string",
  "lastReadTime": "string",
  "progress": 0,
  "isFinished": false
}
```

### LoginResponse

```json
{
  "userId": "string",
  "account": "string",
  "nickname": "string",
  "avatar": "string",
  "token": "string",
  "refreshToken": "string",
  "expiresIn": 0
}
```

### CreatorProfileResponse

```json
{
  "creatorId": "string",
  "account": "string",
  "name": "string",
  "avatar": "string",
  "description": "string",
  "email": "string",
  "articleCount": 0,
  "createTime": "string",
  "lastLoginTime": "string"
}
```

### CreatorLoginResponse

```json
{
  "creatorId": "string",
  "account": "string",
  "name": "string",
  "token": "string"
}
```

### WorkListItem

```json
{
  "articleId": "string",
  "creatorId": "string",
  "title": "string",
  "summary": "string",
  "cover": "string",
  "status": "string",
  "chapterCount": 0,
  "wordCount": 0,
  "lastUpdateTime": "string"
}
```

---

## 附录

### 密码加密说明

前端在传输密码前需进行 MD5 加密：

```kotlin
import java.security.MessageDigest

fun md5(input: String): String {
    val md = MessageDigest.getInstance("MD5")
    val digest = md.digest(input.toByteArray())
    return digest.joinToString("") { "%02x".format(it) }
}

// 使用示例
val md5Password = md5("123456")
// 结果：e10adc3949ba59abbe56e057f20f883e
```

### Token 刷新机制

1. Access Token 有效期：7天（expiresIn=604800秒）
2. Refresh Token 有效期：30天
3. 当 Access Token 过期时（code=1002），客户端应自动使用 Refresh Token 调用 `/auth/refresh-token` 获取新的 Token
4. 如果 Refresh Token 也过期，需要重新登录

### 验证码说明

- 验证码通过 QQ邮箱 SMTP 服务发送
- 发送邮箱：`3949110906@qq.com`
- account 字段传目标QQ邮箱地址
- type 字段固定传 `email`
- 同一邮箱 60 秒内只能发送一次
- 验证码 5 分钟内有效

### 图片上传

#### 头像上传

```http
POST /api/v1/upload/avatar
```

**请求头：**
- `Authorization: Bearer {token}`
- `Content-Type: multipart/form-data`

**请求参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | File | 是 | 头像图片文件，支持 jpg/png/gif，最大 5MB |

**响应：**

```json
{
  "code": 0,
  "message": "上传成功",
  "data": {
    "filename": "1714723456_1234.jpg",
    "url": "http://47.118.22.220:9091/uploads/1714723456_1234.jpg"
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| filename | string | 文件名 |
| url | string | 完整访问 URL |

**业务说明：**
- 上传成功后，将返回的 `url` 传给 `PUT /api/v1/user/profile` 更新头像
- 头像会自动压缩为 256x256，JPEG 质量 80%
- 上传新头像后会自动删除旧头像
- 头像 URL 添加 30 天缓存头

### 用户ID系统说明

**重要变更：**
- 用户唯一标识从 `account` 改为 `userId`
- `userId` 格式：`u_` + 数字/字母组合，如 `u_123456`
- 所有涉及用户标识的接口和数据模型均使用 `userId`
- `account` 仅用于登录时的账号输入

**客户端存储：**
- 登录成功后保存：`userId`、`nickname`、`avatar`、`token`、`refreshToken`
- 使用 `userId` 作为用户唯一标识

### v1.2 重要变更（文章系统）

**书籍模块 -> 文章模块：**

| 旧接口 | 新接口 | 说明 |
|--------|--------|------|
| `/books` | `/articles` | 文章列表 |
| `/books/{bookId}` | `/articles/{articleId}` | 文章详情 |
| `/books/{bookId}/chapters` | `/articles/{articleId}` | 章节列表在 meta 里 |
| `/books/{bookId}/chapters/{chapterId}` | `/articles/{articleId}/chapters/{chapterIndex}` | **chapterIndex 是数字** |

**数据结构变更：**

| 字段 | 旧 | 新 |
|------|-----|-----|
| ID | `bookId` | `articleId` |
| 作者 | `author: {name, avatar}` | `author: "字符串"` |
| 上一章 | `prevChapterId: string` | `prevChapterId: int` |
| 下一章 | `nextChapterId: string` | `nextChapterId: int` |
| 段落 | `content` 字符串 | 新增 `paragraphs` 数组 |

**前端适配要点：**
1. 书架列表用 `articleId` 替代 `bookId`
2. 作者字段从对象改为字符串
3. 章节内容接口传数字索引（0, 1, 2...）而非字符串 chapterId
4. 阅读页用 `paragraphs` 数组渲染段落
5. 下拉加载下一章用 `nextChapterId`（int）作为索引请求

### 接口汇总表

| 序号 | 接口 | 方法 | 说明 | 需登录 |
|------|------|------|------|--------|
| 1 | /auth/send-code | POST | 发送验证码 | 否 |
| 2 | /auth/login | POST | 登录 | 否 |
| 3 | /auth/register | POST | 注册 | 否 |
| 4 | /auth/forgot-password | POST | 忘记密码 | 否 |
| 5 | /auth/refresh-token | POST | 刷新Token | 否 |
| 6 | /user/profile | GET | 获取用户资料 | 是 |
| 7 | /user/profile | PUT | 更新用户资料 | 是 |
| 8 | /upload/avatar | POST | 上传头像 | 是 |
| 9 | /user/follow/{authorId} | POST | 关注作者 | 是 |
| 10 | /user/follow/{authorId} | DELETE | 取消关注 | 是 |
| 11 | /user/following | GET | 获取关注列表 | 是 |
| 12 | /bookshelf | GET | 获取书架列表 | 是 |
| 13 | /bookshelf/{articleId} | POST | 加入书架 | 是 |
| 14 | /bookshelf/{articleId} | DELETE | 移出书架 | 是 |
| 15 | /bookshelf/{articleId}/progress | PUT | 更新阅读进度 | 是 |
| 16 | /articles | GET | 获取文章列表 | 否 |
| 17 | /articles/search | GET | 搜索文章 | 否 |
| 18 | /articles/{articleId} | GET | 获取文章详情 | 否 |
| 19 | /articles/{articleId}/chapters/{chapterIndex} | GET | 获取章节内容 | 否 |
| 20 | /articles/upload | POST | 上传文章 | 是 |
| 21 | /articles/{articleId} | DELETE | 删除文章 | 是 |
| 22 | /articles/{articleId}/comments | GET | 获取文章评论 | 是 |
| 23 | /articles/{articleId}/comments | POST | 发表评论 | 是 |
| 24 | /comments/{commentId}/like | POST | 点赞评论 | 是 |
| 25 | /comments/{commentId}/like | DELETE | 取消点赞评论 | 是 |
| 26 | /chapters/{chapterId}/paragraph-comments | GET | 获取段评列表 | 是 |
| 27 | /chapters/{chapterId}/paragraph-comments | POST | 发表段评 | 是 |
| 28 | /paragraph-comments/{commentId}/like | POST | 点赞段评 | 是 |
| 29 | /paragraph-comments/{commentId}/like | DELETE | 取消点赞段评 | 是 |
| 30 | /authors/{authorId} | GET | 获取作者主页 | 否 |
| 31 | /authors/{authorId}/works | GET | 获取作者作品 | 否 |
| 32 | /authors/{authorId}/activities | GET | 获取作者动态 | 否 |
| 33 | /feed | GET | 获取关注流 | 是 |
| 34 | /feed/{feedId}/like | POST | 点赞动态 | 是 |
| 35 | /feed/{feedId}/like | DELETE | 取消点赞动态 | 是 |
| 36 | /creator/register | POST | 创作者注册 | 否 |
| 37 | /creator/login | POST | 创作者登录 | 否 |
| 38 | /creator/profile | GET | 获取创作者资料 | 是 |
| 39 | /creator/profile | PUT | 更新创作者资料 | 是 |
| 40 | /creator/works | POST | 创建作品 | 是 |
| 41 | /creator/works | GET | 获取我的作品列表 | 是 |
| 42 | /creator/works/{workId} | GET | 获取作品详情 | 是 |
| 43 | /creator/works/{workId} | PUT | 更新作品信息 | 是 |
| 44 | /creator/works/{workId} | DELETE | 删除作品 | 是 |
| 45 | /creator/works/{workId}/publish | POST | 发布作品 | 是 |
| 46 | /creator/works/{workId}/chapters | POST | 创建章节 | 是 |
| 47 | /creator/works/{workId}/chapters/{chapterIndex} | GET | 获取章节内容 | 是 |
| 48 | /creator/works/{workId}/chapters/{chapterIndex} | PUT | 更新章节 | 是 |
| 49 | /creator/works/{workId}/chapters/{chapterIndex} | DELETE | 删除章节 | 是 |
| 50 | /creator/works/upload/docx | POST | 上传docx文档 | 是 |
