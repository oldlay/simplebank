# SimpleBank 银行系统 — 需求分析文档

> **版本**: 1.2  
> **作者**: oldlay  
> **生成日期**: 2026-06-22  
> **项目仓库**: https://github.com/oldlay/simplebank

---

## 目录

1. [系统概述](#1-系统概述)
2. [用户角色与权限模型](#2-用户角色与权限模型)
3. [功能模块概览](#3-功能模块概览)
4. [接口详细说明](#4-接口详细说明)
   - [4.1 用户模块](#41-用户模块)
   - [4.2 认证模块](#42-认证模块)
   - [4.3 账户模块](#43-账户模块)
   - [4.4 转账模块](#44-转账模块)
   - [4.5 邮箱验证模块](#45-邮箱验证模块)
5. [数据模型](#5-数据模型)
6. [业务流程](#6-业务流程)
7. [安全设计](#7-安全设计)
8. [技术架构概览](#8-技术架构概览)
9. [附录：接口汇总](#附录-a接口汇总)

---

## 1. 系统概述

SimpleBank 是一个**简化版银行系统**，提供用户注册/登录、账户管理、货币转账等核心银行功能。系统采用前后端分离架构，后端同时支持 RESTful API（Gin）和 gRPC 两种通信协议，前端使用 Vue.js 构建单页应用。

### 1.1 核心目标

- 提供安全可靠的多币种账户管理服务
- 支持用户间资金转账，保证事务一致性
- 通过邮箱验证确保用户身份真实性
- 基于角色的访问控制（RBAC），区分普通用户与银行职员权限

### 1.2 系统边界

| 维度 | 说明 |
|------|------|
| **支持币种** | USD（美元）、EUR（欧元）、CAD（加元） |
| **用户交互** | Web 前端 + REST API / gRPC API |
| **数据存储** | PostgreSQL（业务数据）+ Redis（异步任务队列） |
| **异步任务** | Redis + Asynq 处理邮件发送等后台任务 |
| **认证方式** | PASETO Token（Access Token + Refresh Token） |

---

## 2. 用户角色与权限模型

系统定义两种角色，采用 RBAC 模型控制接口访问权限：

| 角色 | 标识 | 权限范围 |
|------|------|----------|
| **存款用户** | `depositor` | 注册/登录、管理本人账户、发起转账、查看本人信息、更新本人信息 |
| **银行职员** | `banker` | 继承存款用户全部权限 + 可更新任意用户的信息 |

### 权限矩阵

| 操作 | 游客 | 存款用户 | 银行职员 |
|------|:---:|:------:|:------:|
| 注册用户 | ✅ | — | — |
| 登录 | ✅ | — | — |
| 邮箱验证 | ✅ | — | — |
| 刷新令牌 | ✅ | — | — |
| 创建账户 | — | ✅ | ✅ |
| 查看本人账户 | — | ✅ | ✅ |
| 查看账户列表 | — | ✅ | ✅ |
| 更新账户余额 | — | ✅ | ✅ |
| 删除账户 | — | ✅ | ✅ |
| 转账 | — | ✅ | ✅ |
| 更新本人信息 | — | ✅ | ✅ |
| 更新他人信息 | — | — | ✅ |

---

## 3. 功能模块概览

```
SimpleBank 银行系统
├── 用户管理模块
│   ├── 用户注册（含邮箱异步验证）
│   ├── 用户登录
│   ├── 用户信息更新（本人 / 管理员）
│   └── 邮箱验证
├── 认证授权模块
│   ├── PASETO Token 签发
│   ├── Access Token 续期
│   ├── Session 管理
│   └── Bearer Token 中间件
├── 账户管理模块
│   ├── 创建银行账户（指定币种）
│   ├── 查看单个账户
│   ├── 分页列出本人账户
│   ├── 更新账户余额
│   └── 删除账户
├── 转账交易模块
│   ├── 创建转账（含余额校验 + 币种校验）
│   └── 交易事务（原子性：transfer + entries + balance update）
└── 前端界面
    ├── 首页（HomeView）
    ├── 登录页（LoginView）
    ├── 注册页（RegisterView）
    ├── 账户管理页（AccountsView）
    ├── 转账页（TransferView）
    └── 个人中心页（ProfileView）
```

---

## 4. 接口详细说明

### 4.1 用户模块

#### 4.1.1 创建用户（注册）

| 属性 | 说明 |
|------|------|
| **REST 接口** | `POST /users` |
| **gRPC 接口** | `CreateUser` → `POST /v1/create_user` |
| **认证要求** | 无需认证 |
| **功能描述** | 注册新的银行用户，系统自动发送邮箱验证邮件 |

**请求参数**:

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
|------|------|:---:|------|------|
| `username` | string | ✅ | 字母数字组合 | 用户名，全局唯一 |
| `password` | string | ✅ | 最少 6 位 | 登录密码（bcrypt 加密存储） |
| `full_name` | string | ✅ | 非空 | 用户全名 |
| `email` | string | ✅ | 合法邮箱格式 | 电子邮箱，全局唯一 |

**响应**:

```json
{
  "username": "john_doe",
  "full_name": "John Doe",
  "email": "john@example.com",
  "password_changed_at": "2026-06-22T00:00:00Z",
  "created_at": "2026-06-22T10:00:00Z"
}
```

> **注意**: 响应中不包含 `hashed_password` 字段，密码已做脱敏处理。

**功能效果**:
1. 参数校验通过后，密码经 bcrypt 哈希存储
2. 在一个数据库事务中创建用户记录
3. 事务提交后，异步向 Redis 队列投递"发送验证邮件"任务
4. 验证邮件包含带 `secret_code` 的验证链接，15 分钟内有效
5. 若用户名或邮箱已存在，返回 403 Forbidden

---

#### 4.1.2 更新用户信息

| 属性 | 说明 |
|------|------|
| **REST 接口** | 无（仅 gRPC） |
| **gRPC 接口** | `UpdateUser` → `PATCH /v1/update_user` |
| **认证要求** | 需要认证 |
| **功能描述** | 更新用户的全名、邮箱、密码（部分更新） |

**请求参数**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| `username` | string | ✅ | 要更新的目标用户名 |
| `full_name` | string | — | 更新后的全名（可选） |
| `email` | string | — | 更新后的邮箱（可选） |
| `password` | string | — | 更新后的密码（可选） |

**功能效果**:
1. **存款用户**只能更新自己的信息（`username` 必须与 Token 中的一致），否则返回 `PermissionDenied`
2. **银行职员**可以更新任意用户的信息
3. 若提供 `password`，系统自动更新 `password_changed_at` 时间戳
4. 字段采用**部分更新**模式（仅更新请求中提供的字段）

---

### 4.2 认证模块

#### 4.2.1 用户登录

| 属性 | 说明 |
|------|------|
| **REST 接口** | `POST /users/login` |
| **gRPC 接口** | `LoginUser` → `POST /v1/login_user` |
| **认证要求** | 无需认证 |
| **功能描述** | 验证用户名密码，签发 Access Token 和 Refresh Token |

**请求参数**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| `username` | string | ✅ | 用户名 |
| `password` | string | ✅ | 明文密码（最少 6 位） |

**响应**:

```json
{
  "session_id": "uuid-string",
  "access_token": "v2.local.xxx...",
  "access_token_expires_at": "2026-06-22T10:30:00Z",
  "refresh_token": "v2.local.yyy...",
  "refresh_token_expires_at": "2026-06-23T10:00:00Z",
  "user": {
    "username": "john_doe",
    "full_name": "John Doe",
    "email": "john@example.com",
    "password_changed_at": "2026-06-22T00:00:00Z",
    "created_at": "2026-06-22T10:00:00Z"
  }
}
```

**功能效果**:
1. 验证用户名是否存在、密码是否匹配
2. 签发 Access Token（短期，携带 `role` + `token_type: access_token`）
3. 签发 Refresh Token（长期，`token_type: refresh_token`）
4. 创建 Session 记录（关联 `refresh_token`、`user_agent`、`client_ip`、过期时间）
5. 前端将 Token 存储在 `localStorage`，后续请求通过 `Authorization: Bearer <token>` 携带

---

#### 4.2.2 刷新 Access Token

| 属性 | 说明 |
|------|------|
| **REST 接口** | `POST /tokens/renew_access` |
| **认证要求** | 无需认证（使用 Refresh Token 作为凭证） |
| **功能描述** | 使用 Refresh Token 换取新的 Access Token |

**请求参数**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| `refresh_token` | string | ✅ | 有效的 Refresh Token |

**响应**:

```json
{
  "access_token": "v2.local.zzz...",
  "access_token_expires_at": "2026-06-22T11:00:00Z"
}
```

**功能效果**:
1. 验证 Refresh Token 签名和有效期
2. 检查对应 Session 是否存在、未被冻结（`is_blocked = false`）、未过期、令牌匹配
3. 签发新的 Access Token
4. **不返回**新的 Refresh Token（Refresh Token 过期后需重新登录）

---

### 4.3 账户模块

> 所有账户接口均需认证。接口从 Token 中提取当前用户身份，确保每用户只能操作本人的账户。

#### 4.3.1 创建账户

| 属性 | 说明 |
|------|------|
| **接口** | `POST /accounts` |
| **认证要求** | 需要认证 |
| **功能描述** | 为当前登录用户创建一个指定币种的银行账户 |

**请求参数**:

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
|------|------|:---:|------|------|
| `currency` | string | ✅ | USD / EUR / CAD | 账户币种 |

**响应**: 返回完整的 Account 对象，初始余额为 0。

**功能效果**:
1. 校验币种有效性（仅支持 USD、EUR、CAD）
2. 初始余额为 `0`
3. 账户所有者自动设置为当前认证用户
4. 同一用户、同一币种只能有一个账户（`UNIQUE(owner, currency)` 约束）

---

#### 4.3.2 查看单个账户

| 属性 | 说明 |
|------|------|
| **接口** | `GET /accounts/:id` |
| **认证要求** | 需要认证 |
| **功能描述** | 根据账户 ID 获取账户详情 |

**路径参数**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| `id` | int64 | ✅ (≥ 1) | 账户 ID |

**功能效果**:
1. 从数据库查询账户信息
2. 验证该账户所有者是否为当前认证用户（**防止水平越权**）
3. 账户不存在返回 404，不属于当前用户返回 401

---

#### 4.3.3 列出账户列表

| 属性 | 说明 |
|------|------|
| **接口** | `GET /accounts` |
| **认证要求** | 需要认证 |
| **功能描述** | 分页查询当前登录用户的所有账户 |

**查询参数**:

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
|------|------|:---:|------|------|
| `page_id` | int32 | ✅ | ≥ 1 | 页码（从 1 开始） |
| `page_size` | int32 | ✅ | 5 ~ 10 | 每页条数 |

**功能效果**:
1. 仅返回 `owner = 当前用户` 的账户
2. 分页偏移量为 `(page_id - 1) * page_size`

---

#### 4.3.4 更新账户余额

| 属性 | 说明 |
|------|------|
| **接口** | `PATCH /accounts/:id` |
| **认证要求** | 需要认证 |
| **功能描述** | 更新指定账户的余额（直接设置，非增量） |

**请求参数**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| `balance` | decimal | ✅ | 新的账户余额 |

**功能效果**:
1. 直接更新数据库中的余额字段
2. ⚠️ **注意**：此接口为管理用途设计，正常业务中余额变更应通过「转账」接口间接实现以保证审计追踪

---

#### 4.3.5 删除账户

| 属性 | 说明 |
|------|------|
| **接口** | `DELETE /accounts/:id` |
| **认证要求** | 需要认证 |
| **功能描述** | 删除指定的银行账户 |

**功能效果**:
1. 删除数据库中的账户记录
2. 若账户不存在，返回 404 Not Found

---

### 4.4 转账模块

#### 4.4.1 创建转账

| 属性 | 说明 |
|------|------|
| **接口** | `POST /transfers` |
| **认证要求** | 需要认证 |
| **功能描述** | 从本人账户向另一个账户转账指定金额 |

**请求参数**:

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
|------|------|:---:|------|------|
| `from_account_id` | int64 | ✅ | ≥ 1 | 转出账户 ID |
| `to_account_id` | int64 | ✅ | ≥ 1 | 转入账户 ID |
| `amount` | decimal | ✅ | — | 转账金额 |
| `currency` | string | ✅ | USD / EUR / CAD | 币种 |

**响应**:

```json
{
  "transfer": { "id": 1, "from_account_id": 1, "to_account_id": 2, "amount": "100.00", "created_at": "..." },
  "from_account": { "id": 1, "owner": "alice", "balance": "900.00", "currency": "USD" },
  "to_account": { "id": 2, "owner": "bob", "balance": "1100.00", "currency": "USD" },
  "from_entry": { "id": 1, "account_id": 1, "amount": "-100.00" },
  "to_entry": { "id": 2, "account_id": 2, "amount": "100.00" }
}
```

**功能效果**（在一个数据库事务内完成）：

| 步骤 | 操作 | 说明 |
|:--:|------|------|
| 1 | 校验转出账户 | 存在性 + `owner == 当前用户` + 币种匹配 |
| 2 | 校验转入账户 | 存在性 + 币种匹配 |
| 3 | 校验余额 | 转出账户余额 ≥ 转账金额 |
| 4 | 创建 Transfer 记录 | 记录转账本身 |
| 5 | 创建 Entry 记录（出） | 转出账户流水（负金额） |
| 6 | 创建 Entry 记录（入） | 转入账户流水（正金额） |
| 7 | 更新转出账户余额 | 扣减金额 |
| 8 | 更新转入账户余额 | 增加金额 |

> **事务安全**: 使用数据库事务保证原子性。为避免死锁，按账户 ID 升序更新余额。

**错误场景**:

| 错误 | HTTP 状态码 | 说明 |
|------|:---:|------|
| 转出账户不存在 | 404 | 无效的 from_account_id |
| 转出账户不属于当前用户 | 401 | 不可转出他人资金 |
| 转入账户不存在 | 404 | 无效的 to_account_id |
| 币种不匹配 | 400 | 账户币种与请求币种不一致 |
| 余额不足 | 404 | 转出余额 < 转账金额 |
| 数据库异常 | 500 | 事务执行失败 |

---

### 4.5 邮箱验证模块

#### 4.5.1 验证邮箱

| 属性 | 说明 |
|------|------|
| **gRPC 接口** | `VerifyEmail` → `GET /v1/verify_email` |
| **认证要求** | 无需认证 |
| **功能描述** | 通过邮件中的链接验证邮箱，完成注册 |

**查询参数**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| `email_id` | int64 | ✅ | 验证邮件记录 ID |
| `secret_code` | string | ✅ | 32 位随机密钥（邮件中携带） |

**功能效果**:
1. 核验 `secret_code` 是否匹配、是否在 15 分钟有效期内、是否已被使用
2. 更新 `verify_emails.is_used = true`（防重放）
3. 更新 `users.is_email_verified = true`
4. 返回验证状态

---

## 5. 数据模型

### 5.1 实体关系图（ER 概要）

```
users ──1:N──> accounts ──1:N──> entries
  │              │
  │              ├──1:N──> transfers (from_account_id)
  │              └──1:N──> transfers (to_account_id)
  │
  ├──1:N──> verify_emails
  └──1:N──> sessions
```

### 5.2 核心表结构

#### users（用户表）

| 列名 | 类型 | 约束 | 说明 |
|------|------|------|------|
| `username` | varchar | PK | 用户名 |
| `role` | varchar | NOT NULL, DEFAULT 'depositor' | 角色 |
| `hashed_password` | varchar | NOT NULL | bcrypt 哈希密码 |
| `full_name` | varchar | NOT NULL | 全名 |
| `email` | varchar | UNIQUE, NOT NULL | 邮箱 |
| `is_email_verified` | bool | DEFAULT false | 邮箱已验证 |
| `password_changed_at` | timestamptz | NOT NULL | 密码修改时间 |
| `created_at` | timestamptz | NOT NULL, DEFAULT now() | 创建时间 |

#### accounts（账户表）

| 列名 | 类型 | 约束 | 说明 |
|------|------|------|------|
| `id` | bigserial | PK | 账户 ID |
| `owner` | varchar | FK → users, NOT NULL | 所有者 |
| `balance` | numeric | NOT NULL | 余额（精确小数） |
| `currency` | varchar | NOT NULL | 币种 |
| `created_at` | timestamptz | NOT NULL, DEFAULT now() | 创建时间 |

> **索引**: `(owner)` 普通索引；`(owner, currency)` 唯一索引

#### entries（流水记录表）

| 列名 | 类型 | 约束 | 说明 |
|------|------|------|------|
| `id` | bigserial | PK | 流水 ID |
| `account_id` | bigint | FK → accounts, NOT NULL | 关联账户 |
| `amount` | numeric | NOT NULL | 金额（正=入账 / 负=出账） |
| `created_at` | timestamptz | NOT NULL, DEFAULT now() | 创建时间 |

#### transfers（转账记录表）

| 列名 | 类型 | 约束 | 说明 |
|------|------|------|------|
| `id` | bigserial | PK | 转账 ID |
| `from_account_id` | bigint | FK → accounts, NOT NULL | 转出账户 |
| `to_account_id` | bigint | FK → accounts, NOT NULL | 转入账户 |
| `amount` | numeric | NOT NULL | 金额（始终为正数） |
| `created_at` | timestamptz | NOT NULL, DEFAULT now() | 创建时间 |

> **索引**: `(from_account_id)`、`(to_account_id)`、`(from_account_id, to_account_id)` 组合索引

#### verify_emails（邮箱验证表）

| 列名 | 类型 | 约束 | 说明 |
|------|------|------|------|
| `id` | bigserial | PK | 验证记录 ID |
| `username` | varchar | FK → users, NOT NULL | 关联用户 |
| `email` | varchar | NOT NULL | 待验证邮箱 |
| `secret_code` | varchar | NOT NULL | 32 位随机验证码 |
| `is_used` | bool | DEFAULT false | 是否已使用（防重放） |
| `created_at` | timestamptz | NOT NULL, DEFAULT now() | 创建时间 |
| `expired_at` | timestamptz | NOT NULL, DEFAULT now()+15min | 过期时间 |

#### sessions（会话表）

| 列名 | 类型 | 约束 | 说明 |
|------|------|------|------|
| `id` | uuid | PK | 与 Refresh Token Payload ID 一致 |
| `username` | varchar | NOT NULL | 关联用户 |
| `refresh_token` | varchar | NOT NULL | 刷新令牌 |
| `user_agent` | varchar | NOT NULL | 用户代理 |
| `client_ip` | varchar | NOT NULL | 客户端 IP |
| `is_blocked` | bool | NOT NULL, DEFAULT false | 是否被冻结 |
| `expires_at` | timestamptz | NOT NULL | 过期时间 |
| `created_at` | timestamptz | NOT NULL, DEFAULT now() | 创建时间 |

---

## 6. 业务流程

### 6.1 用户注册与验证流程

```
用户填写注册表单
       │
       ▼
┌─────────────────┐
│  POST /users    │ ── 参数校验（username/password/full_name/email）
└──────┬──────────┘
       │ 校验通过
       ▼
┌─────────────────┐
│ 密码 bcrypt 哈希 │
└──────┬──────────┘
       │
       ▼
┌═════════════════╗
║  数据库事务      ║ ── 原子写入 user 记录
║  CreateUserTx   ║
╚═══════┬═════════╝
       │ 事务提交成功
       ▼
┌──────────────────────┐
│ Redis 异步任务投递    │ ── 发送验证邮件
│ (Asynq: QueueCritical)│     10s 延迟，最多 10 次重试
└──────┬───────────────┘
       │
       ▼
┌─────────────────┐
│ 生成验证码 + 记录 │ ── 写入 verify_emails 表
│ RandomString(32) │     expired_at = now + 15min
└──────┬──────────┘
       │
       ▼
┌─────────────────┐
│ 发送验证邮件      │ ── SMTP (Gmail)
│ 含验证链接        │     /v1/verify_email?email_id=X&secret_code=Y
└──────┬──────────┘
       │
       ▼
┌─────────────────┐
│ 用户点击链接      │
│ GET /v1/verify_  │ ── 核验 secret_code + 有效期 + is_used
│      email       │     更新 is_email_verified = true
└─────────────────┘
```

### 6.2 转账流程

```
用户发起转账请求
       │
       ▼
┌────────────────────┐
│ 1. 校验转出账户      │ ── 存在？所有者是当前用户？币种匹配？
└──────┬─────────────┘
       │ ✓
       ▼
┌────────────────────┐
│ 2. 校验转入账户      │ ── 存在？币种匹配？
└──────┬─────────────┘
       │ ✓
       ▼
┌────────────────────┐
│ 3. 校验余额          │ ── from_account.balance >= amount
└──────┬─────────────┘
       │ ✓
       ▼
┌════════════════════════════════════════┐
║         数据库事务（TransferTx）         ║
║                                        ║
║  a. INSERT INTO transfers              ║
║  b. INSERT INTO entries (from, -amt)  ║
║  c. INSERT INTO entries (to, +amt)    ║
║  d. UPDATE accounts SET balance       ║
║     = balance - amount (from)         ║
║  e. UPDATE accounts SET balance       ║
║     = balance + amount (to)          ║
║                                        ║
║  更新顺序按 account_id 升序防死锁       ║
╚════════════════════════════════════════╝
       │ 事务提交
       ▼
┌────────────────────┐
│ 返回完整交易结果      │ ── transfer + entries + updated accounts
└────────────────────┘
```

### 6.3 Token 认证流程

```
                    ┌──────────────┐
                    │   客户端      │
                    └──────┬───────┘
                           │ POST /users/login
                           ▼
                    ┌──────────────┐
                    │ 签发 Token    │
                    │ Access (短)   │
                    │ Refresh (长)  │
                    └──────┬───────┘
                           │ 返回 token pair
                           ▼
              ┌──────────────────────┐
              │ 后续请求携带           │
              │ Authorization:        │
              │ Bearer <access_token> │
              └──────┬───────────────┘
                     │
                     ▼
              ┌──────────────┐
              │ Auth 中间件    │
              │ VerifyToken() │
              └──┬──────────┬──┘
                 │ 有效      │ 过期
                 ▼           ▼
            ┌────────┐  ┌──────────────────┐
            │ 注入    │  │ 401 → 客户端调用   │
            │ Payload │  │ POST /tokens/     │
            │ 继续处理 │  │ renew_access →    │
            └────────┘  │ 获取新 Access Token│
                        └──────────────────┘
```

---

## 7. 安全设计

| 安全层面 | 实现方式 |
|------|------|
| **密码存储** | bcrypt 哈希，不存储明文 |
| **Token 签名** | PASETO v2.local（对称加密，32 字节密钥） |
| **Token 类型区分** | `access_token` vs `refresh_token`，防止 Token 混用 |
| **水平越权防护** | 账户/转账操作校验 `owner == 当前认证用户` |
| **垂直越权防护** | `UpdateUser` 中非本人操作需 `role = banker` |
| **转账权限** | 只能从本人账户转出资金 |
| **会话管理** | Session 支持 `is_blocked` 冻结，可主动终止异常会话 |
| **邮箱验证** | 32 位随机 `secret_code`，15 分钟过期，一次性使用 |
| **防死锁** | 转账事务按主键排序更新余额 |
| **CORS** | 配置白名单 `AllowedOrigins`，限制方法和头部 |

---

## 8. 技术架构概览

```
┌───────────────────────────────────────────────────────┐
│                      Client                           │
│            Vue.js SPA (Vite + PrimeVue)               │
└─────────────┬────────────────────┬────────────────────┘
              │ HTTP/REST          │ gRPC
              ▼                    ▼
┌─────────────────────┐  ┌─────────────────────┐
│   Gin HTTP Server   │  │   gRPC Server       │
│   :8080             │  │   :9090             │
│                     │  │                     │
│ Auth Middleware     │  │ Unary Interceptor   │
│ (PASETO Verify)     │  │ (Logger + Auth)     │
└──────────┬──────────┘  └──────────┬──────────┘
           │                        │
           └────────┬───────────────┘
                    │
        ┌───────────▼───────────┐
        │   HTTP Gateway Server │
        │   gRPC-Gateway        │
        │   :8080               │
        │   + Swagger Docs      │
        │   + CORS Middleware   │
        └───────────┬───────────┘
                    │
        ┌───────────▼───────────┐
        │    Service Layer      │
        │  db.Store (SQLC)      │
        │  + Transaction Support│
        └───────────┬───────────┘
                    │
    ┌───────────────┼───────────────┐
    ▼               ▼               ▼
┌────────┐   ┌──────────┐   ┌──────────┐
│PostgreSQL│  │  Redis   │   │  Gmail   │
│ (数据)  │   │ (Asynq)  │   │ (SMTP)   │
└────────┘   └──────────┘   └──────────┘
```

| 层级 | 技术选型 | 说明 |
|------|------|------|
| **前端** | Vue 3 + PrimeVue + Axios | 类型安全单页应用，localStorage 持久化会话 |
| **HTTP API** | Gin (Go) | RESTful 风格，json binding + validator v10 |
| **gRPC API** | Protocol Buffers + gRPC-Gateway | gRPC 原生 + HTTP/JSON（Gateway 自动转换） |
| **数据库** | PostgreSQL + golang-migrate | 关系型存储，精确数值（numeric） |
| **SQL 生成** | sqlc | 类型安全的 SQL 代码生成 |
| **认证** | PASETO v2 | 现代化 Token 方案 |
| **任务队列** | Asynq (Redis) | 异步邮件，支持重试、优先级、延时 |
| **邮件** | Gmail SMTP | 发送欢迎邮件和验证链接 |
| **日志** | zerolog | 结构化、高性能日志 |
| **文档** | gRPC reflection + Swagger | API 自动文档 |

---

## 附录 A: 接口汇总

### REST API (Gin, :8080)

| 方法 | 路径 | 认证 | 描述 |
|:----:|------|:----:|------|
| POST | `/users` | — | 创建用户（注册） |
| POST | `/users/login` | — | 用户登录，获取 Token |
| POST | `/tokens/renew_access` | — | 刷新 Access Token |
| POST | `/accounts` | ✅ | 创建银行账户 |
| GET | `/accounts/:id` | ✅ | 查看单个账户详情 |
| GET | `/accounts` | ✅ | 分页列出本人账户 |
| PATCH | `/accounts/:id` | ✅ | 更新账户余额 |
| DELETE | `/accounts/:id` | ✅ | 删除账户 |
| POST | `/transfers` | ✅ | 创建转账交易 |

### gRPC 服务 (SimpleBank, :9090 / HTTP Gateway)

| 方法 | 路径（Gateway） | 认证 | 描述 |
|:----:|------|:----:|------|
| CreateUser | `POST /v1/create_user` | — | 创建用户 |
| LoginUser | `POST /v1/login_user` | — | 用户登录 |
| UpdateUser | `PATCH /v1/update_user` | ✅ | 更新用户信息（本人/管理员） |
| VerifyEmail | `GET /v1/verify_email` | — | 邮箱验证 |

### 前端路由 (Vue Router)

| 路径 | 页面 | 认证 | 描述 |
|:----:|------|:----:|------|
| `/` | HomeView | ✅ | 首页 |
| `/login` | LoginView | 游客 | 登录页 |
| `/register` | RegisterView | 游客 | 注册页 |
| `/accounts` | AccountsView | ✅ | 账户管理页 |
| `/transfer` | TransferView | ✅ | 转账页 |
| `/profile` | ProfileView | ✅ | 个人中心页 |

---

> 🤖 Generated with [Claude Code](https://claude.com/claude-code) — 基于项目源码 [simplebank](https://github.com/oldlay/simplebank) 分析生成。
