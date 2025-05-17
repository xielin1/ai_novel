# 网文大纲续写工具 API 文档

## 基本信息

- 基础URL: `/api`
- 所有请求/响应数据格式: JSON
- 认证方式: Bearer Token
- 响应状态码:
  - 200: 成功
  - 400: 请求参数错误
  - 401: 未授权
  - 403: 禁止访问
  - 404: 资源不存在
  - 500: 服务器内部错误

## 通用响应格式

```json
{
  "success": true/false,
  "message": "状态描述信息",
  "data": {
    // 响应数据
  }
}
```

## 一、用户与账户管理

### 1. Token相关 API

#### 1.1 获取Token余额

- **URL**: `/user/tokens`
- **方法**: `GET`
- **描述**: 获取当前用户的token余额和使用记录
- **请求头**: `Authorization: Bearer <token>`
- **请求参数**:
  - `page`: 页码(可选，默认1)
  - `limit`: 每页条数(可选，默认10)
- **响应**:
  ```json
  {
    "success": true,
    "data": {
      "balance": 850,
      "records": [
        {
          "id": 123,
          "amount": -150,
          "balance": 850,
          "record_type": 3,  // 3表示续写消费
          "description": "AI续写消费",
          "created_at": "2023-05-24T14:30:00Z"
        },
        {
          "id": 120,
          "amount": 1000,
          "balance": 1000,
          "record_type": 1,  // 1表示套餐赠送
          "description": "基础套餐月度赠送",
          "created_at": "2023-05-01T00:00:00Z"
        }
        // ...更多记录
      ],
      "pagination": {
        "total": 15,
        "page": 1,
        "limit": 10,
        "pages": 2
      }
    }
  }
  ```

### 2. 套餐管理 API

#### 2.1 获取套餐列表

- **URL**: `/packages`
- **方法**: `GET`
- **描述**: 获取所有可用的套餐信息
- **请求头**: `Authorization: Bearer <token>`
- **响应**:
  ```json
  {
    "success": true,
    "data": [
      {
        "id": 1,
        "name": "基础版",
        "description": "适合轻度使用的创作者",
        "price": 19.9,
        "monthly_tokens": 5000,
        "duration": "monthly",
        "features": ["基础AI续写", "历史版本保存"]
      },
      {
        "id": 2,
        "name": "升级版",
        "description": "适合中度创作需求",
        "price": 49.9,
        "monthly_tokens": 15000,
        "duration": "monthly",
        "features": ["高级AI续写", "历史版本保存", "优先客服支持"]
      },
      {
        "id": 3,
        "name": "永久版会员",
        "description": "适合专业创作者",
        "price": 199.9,
        "monthly_tokens": 50000,
        "duration": "permanent",
        "features": ["高级AI续写", "无限历史版本", "专属客服", "高级导出格式"]
      }
    ]
  }
  ```

#### 2.2 购买/订阅套餐

- **URL**: `/packages/subscribe`
- **方法**: `POST`
- **描述**: 购买或订阅指定的套餐
- **请求头**: `Authorization: Bearer <token>`
- **请求体**:
  ```json
  {
    "package_id": 2,
    "payment_method": "alipay"  // 支持的支付方式: alipay, wechat, creditcard
  }
  ```
- **响应**:
  ```json
  {
    "success": true,
    "message": "订阅成功",
    "data": {
      "order_id": "ORD20230605001",
      "package_name": "升级版",
      "amount": 49.9,
      "payment_status": "completed",
      "valid_until": "2023-07-05T23:59:59Z",
      "tokens_awarded": 15000
    }
  }
  ```

#### 2.3 获取当前套餐信息

- **URL**: `/user/package`
- **方法**: `GET`
- **描述**: 获取当前用户的套餐订阅信息
- **请求头**: `Authorization: Bearer <token>`
- **响应**:
  ```json
  {
    "success": true,
    "data": {
      "package": {
        "id": 2,
        "name": "升级版",
        "monthly_tokens": 15000
      },
      "subscription_status": "active",
      "start_date": "2023-06-05T14:30:00Z",
      "expiry_date": "2023-07-05T23:59:59Z",
      "auto_renew": true,
      "next_renewal_date": "2023-07-05T00:00:00Z"
    }
  }
  ```

#### 2.4 取消自动续费

- **URL**: `/packages/cancel-renewal`
- **方法**: `POST`
- **描述**: 取消当前套餐的自动续费
- **请求头**: `Authorization: Bearer <token>`
- **响应**:
  ```json
  {
    "success": true,
    "message": "自动续费已取消",
    "data": {
      "package_name": "升级版",
      "expiry_date": "2023-07-05T23:59:59Z",
      "auto_renew": false
    }
  }
  ```

### 3. 推荐码相关 API

#### 3.1 获取个人推荐码

- **URL**: `/user/referral-code`
- **方法**: `GET`
- **描述**: 获取当前用户的推荐码
- **请求头**: `Authorization: Bearer <token>`
- **响应**:
  ```json
  {
    "success": true,
    "data": {
      "referral_code": "RF7XYZ9",
      "total_referred": 5,
      "total_tokens_earned": 1000,
      "share_url": "https://example.com/register?ref=RF7XYZ9"
    }
  }
  ```

#### 3.2 获取推荐记录

- **URL**: `/user/referrals`
- **方法**: `GET`
- **描述**: 获取用户推荐的历史记录
- **请求头**: `Authorization: Bearer <token>`
- **请求参数**:
  - `page`: 页码(可选，默认1)
  - `limit`: 每页条数(可选，默认10)
- **响应**:
  ```json
  {
    "success": true,
    "data": {
      "referrals": [
        {
          "id": 34,
          "user_id": 156,
          "username": "user***56",
          "registered_at": "2023-05-10T09:25:00Z",
          "tokens_rewarded": 200
        },
        // ...更多记录
      ],
      "statistics": {
        "total_referred": 5,
        "total_tokens_earned": 1000
      },
      "pagination": {
        "total": 5,
        "page": 1,
        "limit": 10,
        "pages": 1
      }
    }
  }
  ```

#### 3.3 生成新的推荐码

- **URL**: `/user/generate-referral-code`
- **方法**: `POST`
- **描述**: 重新生成个人推荐码(旧的推荐码将失效)
- **请求头**: `Authorization: Bearer <token>`
- **响应**:
  ```json
  {
    "success": true,
    "message": "推荐码已重新生成",
    "data": {
      "previous_code": "RF7XYZ9",
      "new_code": "RF8ABC3",
      "share_url": "https://example.com/register?ref=RF8ABC3"
    }
  }
  ```

#### 3.4 使用推荐码

- **URL**: `/user/referral`
- **方法**: `POST`
- **描述**: 使用他人的推荐码获得奖励
- **请求头**: `Authorization: Bearer <token>`
- **请求体**:
  ```json
  {
    "referralCode": "ABC123"
  }
  ```
- **响应**:
  ```json
  {
    "success": true,
    "message": "推荐码使用成功",
    "data": {
      "tokens_rewarded": 200,
      "new_balance": 1050
    }
  }
  ```

## 二、内容创作管理

### 1. 项目管理 API

#### 1.1 获取项目列表

- **URL**: `/projects`
- **方法**: `GET`
- **描述**: 获取当前用户的所有项目
- **请求头**: `Authorization: Bearer <token>`
- **请求参数**:
  - `page`: 页码(可选，默认1)
  - `limit`: 每页条数(可选，默认10)
- **响应**:
  ```json
  {
    "success": true,
    "data": [
      {
        "id": 1,
        "title": "我的玄幻小说",
        "description": "描述",
        "genre": "玄幻",
        "created_at": "2023-05-20T12:00:00Z",
        "updated_at": "2023-05-21T10:30:00Z",
        "last_edited_at": "2023-05-21T10:30:00Z"
      },
      // ...更多项目
    ],
    "pagination": {
      "total": 25,
      "page": 1,
      "limit": 10,
      "pages": 3
    }
  }
  ```

#### 1.2 创建新项目

- **URL**: `/projects`
- **方法**: `POST`
- **描述**: 创建一个新的写作项目
- **请求头**: `Authorization: Bearer <token>`
- **请求体**:
  ```json
  {
    "title": "项目标题",
    "description": "项目描述(可选)",
    "genre": "作品类型(可选)"
  }
  ```
- **响应**:
  ```json
  {
    "success": true,
    "message": "项目创建成功",
    "data": {
      "id": 5,
      "title": "项目标题",
      "description": "项目描述",
      "genre": "作品类型",
      "created_at": "2023-05-22T14:30:00Z",
      "updated_at": "2023-05-22T14:30:00Z"
    }
  }
  ```

#### 1.3 获取项目详情

- **URL**: `/projects/{id}`
- **方法**: `GET`
- **描述**: 获取指定项目的详细信息
- **请求头**: `Authorization: Bearer <token>`
- **路径参数**:
  - `id`: 项目ID
- **响应**:
  ```json
  {
    "success": true,
    "data": {
      "id": 5,
      "title": "项目标题",
      "description": "项目描述",
      "genre": "作品类型",
      "created_at": "2023-05-22T14:30:00Z",
      "updated_at": "2023-05-22T14:30:00Z",
      "last_edited_at": "2023-05-22T15:45:00Z"
    }
  }
  ```

#### 1.4 更新项目信息

- **URL**: `/projects/{id}`
- **方法**: `PUT`
- **描述**: 更新项目的基本信息
- **请求头**: `Authorization: Bearer <token>`
- **路径参数**:
  - `id`: 项目ID
- **请求体**:
  ```json
  {
    "title": "更新后的标题",
    "description": "更新后的描述",
    "genre": "更新后的类型"
  }
  ```
- **响应**:
  ```json
  {
    "success": true,
    "message": "项目更新成功",
    "data": {
      "id": 5,
      "title": "更新后的标题",
      "description": "更新后的描述",
      "genre": "更新后的类型",
      "updated_at": "2023-05-23T09:15:00Z"
    }
  }
  ```

#### 1.5 删除项目

- **URL**: `/projects/{id}`
- **方法**: `DELETE`
- **描述**: 删除指定的项目
- **请求头**: `Authorization: Bearer <token>`
- **路径参数**:
  - `id`: 项目ID
- **响应**:
  ```json
  {
    "success": true,
    "message": "项目删除成功"
  }
  ```

### 2. 大纲管理 API

#### 2.1 获取大纲内容

- **URL**: `/outlines/{id}`
- **方法**: `GET`
- **描述**: 获取指定项目的大纲内容
- **请求头**: `Authorization: Bearer <token>`
- **路径参数**:
  - `id`: 项目ID
- **响应**:
  ```json
  {
    "success": true,
    "data": {
      "id": 15,
      "project_id": 5,
      "content": "这里是大纲内容...",
      "current_version": 3,
      "created_at": "2023-05-22T14:35:00Z",
      "updated_at": "2023-05-23T10:20:00Z"
    }
  }
  ```

#### 2.2 保存/更新大纲内容

- **URL**: `/outlines/{id}`
- **方法**: `POST`
- **描述**: 保存或更新大纲内容，并创建新版本
- **请求头**: `Authorization: Bearer <token>`
- **路径参数**:
  - `id`: 项目ID
- **请求体**:
  ```json
  {
    "content": "更新后的大纲内容..."
  }
  ```
- **响应**:
  ```json
  {
    "success": true,
    "message": "大纲保存成功",
    "data": {
      "id": 15,
      "project_id": 5,
      "content": "更新后的大纲内容...",
      "current_version": 4,
      "updated_at": "2023-05-23T16:40:00Z"
    }
  }
  ```

#### 2.3 获取版本历史

- **URL**: `/versions/{id}`
- **方法**: `GET`
- **描述**: 获取指定项目大纲的历史版本列表
- **请求头**: `Authorization: Bearer <token>`
- **路径参数**:
  - `id`: 项目ID
- **请求参数**:
  - `limit`: 返回版本数量(可选，默认10)
- **响应**:
  ```json
  {
    "success": true,
    "data": [
      {
        "id": 42,
        "outline_id": 15,
        "version_number": 4,
        "content": "版本4的内容...",
        "is_ai_generated": false,
        "created_at": "2023-05-23T16:40:00Z"
      },
      {
        "id": 38,
        "outline_id": 15,
        "version_number": 3,
        "content": "版本3的内容...",
        "is_ai_generated": true,
        "ai_style": "玄幻",
        "word_limit": 1000,
        "tokens_used": 150,
        "created_at": "2023-05-23T15:30:00Z"
      },
      // ...更多版本
    ]
  }
  ```

## 三、AI功能

### 1. AI续写 API

#### 1.1 AI续写

- **URL**: `/ai/generate/{id}`
- **方法**: `POST`
- **描述**: 对指定项目的大纲进行AI续写
- **请求头**: `Authorization: Bearer <token>`
- **路径参数**:
  - `id`: 项目ID
- **请求体**:
  ```json
  {
    "content": "当前大纲内容...",
    "style": "玄幻",  // 可选值: default, fantasy, scifi, urban, xianxia, history
    "wordLimit": 1000  // 生成字数限制
  }
  ```
- **响应**:
  ```json
  {
    "success": true,
    "message": "续写成功",
    "data": {
      "content": "AI生成的续写内容...",
      "tokens_used": 150,
      "token_balance": 850
    }
  }
  ```

## 四、文件操作

### 1. 文件处理 API

#### 1.1 上传大纲文件

- **URL**: `/upload/outline/{id}`
- **方法**: `POST`
- **描述**: 上传文本文件作为大纲内容
- **请求头**: 
  - `Authorization: Bearer <token>`
  - `Content-Type: multipart/form-data`
- **路径参数**:
  - `id`: 项目ID
- **请求体**:
  - `file`: 文件(支持.txt, .docx)
- **响应**:
  ```json
  {
    "success": true,
    "message": "文件上传成功",
    "data": {
      "content": "从文件解析出的内容...",
      "filename": "uploaded.txt",
      "size": 5120
    }
  }
  ```

#### 1.2 导出大纲

- **URL**: `/exports/{id}`
- **方法**: `POST`
- **描述**: 导出大纲为文件
- **请求头**: `Authorization: Bearer <token>`
- **路径参数**:
  - `id`: 项目ID
- **请求体**:
  ```json
  {
    "format": "txt"  // 支持的格式: txt, docx, pdf
  }
  ```
- **响应**:
  ```json
  {
    "success": true,
    "message": "导出成功",
    "data": {
      "file_url": "/download/outline_5_20230524.txt",
      "file_size": 10240
    }
  }
  ```

## 错误响应示例

### 未授权

```json
{
  "success": false,
  "message": "未授权，请先登录",
  "code": 401
}
```

### 参数错误

```json
{
  "success": false,
  "message": "参数错误",
  "errors": [
    {
      "field": "title",
      "message": "标题不能为空"
    }
  ],
  "code": 400
}
```

### 资源不存在

```json
{
  "success": false,
  "message": "项目不存在",
  "code": 404
}
```

### Token不足

```json
{
  "success": false,
  "message": "Token余额不足，请充值",
  "data": {
    "required": 150,
    "balance": 50
  },
  "code": 403
}
```
