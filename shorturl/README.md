# ShortURL Service

基于 Golang + Gin + Redis + MySQL 构建的高并发短链服务系统。

## 功能特性

- **短链生成**：基于 Snowflake 算法生成唯一 ID，经 Base62 编码转换为短码
- **302 重定向**：短链访问接口，支持过期时间判断
- **用户认证**：JWT 无状态认证，支持邮箱密码登录、验证码登录
- **用户隔离**：每个用户只能管理自己创建的短链
- **访问统计**：异步批量写入访问日志，不阻塞 HTTP 请求
- **高并发优化**：
  - Redis 缓存短链映射关系
  - 布隆过滤器防止缓存穿透
  - 滑动窗口限流防止恶意刷接口
- **前端管理页面**：简洁的管理界面，支持短链创建、列表展示、删除操作

## 技术栈

| 分类 | 技术 | 版本 |
|------|------|------|
| 后端语言 | Go | 1.26+ |
| Web 框架 | Gin | 1.12.0 |
| ORM | GORM | 1.31.2 |
| 数据库 | MySQL | 8.0+ |
| 缓存 | Redis | 7.0+ |
| 认证 | JWT | v5 |

## 项目结构

```
shorturl/
├── cmd/main.go           # 程序入口，路由注册
├── config/               # 配置管理
│   ├── config.go         # 配置结构体与加载
│   ├── mysql.go          # MySQL 连接初始化
│   └── redis.go          # Redis 连接初始化
├── dao/                  # 数据访问层
│   ├── short_url_dao.go  # 短链数据操作
│   ├── user_dao.go       # 用户数据操作
│   └── visit_log_dao.go  # 访问日志操作
├── handler/              # 控制器层
│   ├── short_url_handler.go  # 短链 API 处理
│   └── user_handler.go       # 用户 API 处理
├── middleware/           # 中间件
│   ├── jwt.go            # JWT 认证中间件
│   └── rate_limit.go     # 限流中间件
├── model/                # 数据模型
│   ├── short_url.go      # 短链模型
│   └── user.go           # 用户模型
├── service/              # 业务逻辑层
│   ├── short_url_service.go  # 短链业务
│   ├── user_service.go       # 用户业务
│   ├── email_service.go      # 邮件服务
│   └── stats_service.go      # 统计服务
├── util/                 # 工具函数
│   ├── snowflake.go      # 雪花 ID 生成
│   ├── base62.go         # Base62 编码
│   ├── bloom_filter.go   # 布隆过滤器
│   ├── jwt.go            # JWT 工具
│   ├── captcha.go        # 验证码生成
│   └── rate_limiter.go   # 滑动窗口限流
├── static/               # 前端静态文件
│   └── admin.html        # 管理页面
├── schema.sql            # 数据库建表语句
└── go.mod                # Go 模块依赖
```

## 快速开始

### 环境要求

- Go 1.26+
- MySQL 8.0+
- Redis 7.0+

### 安装依赖

```bash
go mod download
```

### 数据库配置

1. 创建数据库：

```sql
CREATE DATABASE shorturl CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

2. 执行建表语句：

```bash
mysql -u root -p shorturl < schema.sql
```

### 配置修改

编辑 `config/config.go`，修改数据库和 Redis 配置：

```go
MySQL: MySQLConfig{
    Host:     "127.0.0.1",
    Port:     "3306",
    User:     "root",
    Password: "your_password",
    DBName:   "shorturl",
},
Redis: RedisConfig{
    Addr:     "127.0.0.1:6379",
    Password: "",
    DB:       0,
},
```

### 运行服务

```bash
go run cmd/main.go
```

或编译后运行：

```bash
go build -o shorturl.exe ./cmd/main.go
./shorturl.exe
```

服务启动后访问：`http://localhost:8080/admin`

## API 接口

### 认证接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/auth/register` | POST | 用户注册 |
| `/api/auth/login` | POST | 邮箱密码登录 |
| `/api/auth/captcha` | POST | 获取验证码 |
| `/api/auth/login/captcha` | POST | 验证码登录 |
| `/api/auth/forgot-password` | POST | 忘记密码（发送验证码） |
| `/api/auth/reset-password` | POST | 重置密码 |

### 短链接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/create` | POST | 创建短链（需登录） |
| `/api/links` | GET | 获取短链列表（需登录） |
| `/api/links/:code` | DELETE | 删除短链（需登录） |
| `/s/:code` | GET | 短链跳转（公开访问） |

### 请求示例

**创建短链**：

```bash
curl -X POST http://localhost:8080/api/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your_jwt_token" \
  -d '{"url": "https://www.example.com", "expire_at": "2026-07-02T14:50:00+08:00"}'
```

**短链跳转**：

```bash
curl http://localhost:8080/s/PhtUpcO7Au
```

## 配置说明

### 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| SERVER_PORT | 服务端口 | 8080 |
| MYSQL_HOST | MySQL 主机 | 127.0.0.1 |
| MYSQL_PORT | MySQL 端口 | 3306 |
| MYSQL_USER | MySQL 用户 | root |
| MYSQL_PASSWORD | MySQL 密码 | 123456 |
| MYSQL_DB_NAME | 数据库名 | shorturl |
| REDIS_ADDR | Redis 地址 | 127.0.0.1:6379 |
| REDIS_PASSWORD | Redis 密码 | 空 |
| REDIS_DB | Redis 数据库 | 0 |
| JWT_SECRET | JWT 密钥 | shorturl_jwt_secret_key |
| EMAIL_ENABLED | 是否启用邮件 | false |
| SMTP_HOST | SMTP 服务器 | smtp.qq.com |
| SMTP_PORT | SMTP 端口 | 587 |
| SMTP_USER | SMTP 用户 | your_email@qq.com |
| SMTP_PASSWORD | SMTP 密码 | your_email_authorization_code |
| RATE_LIMIT_WINDOW | 限流窗口（秒） | 60 |
| RATE_LIMIT_MAX | 限流最大请求数 | 100 |


## 架构设计

```
                    ┌──────────────────────────────────┐
                    │           HTTP 请求               │
                    └──────────────────┬───────────────┘
                                       │
                    ┌──────────────────▼───────────────┐
                    │        限流中间件 (IP 级别)        │
                    └──────────────────┬───────────────┘
                                       │
                    ┌──────────────────▼───────────────┐
                    │      布隆过滤器 (快速拦截)         │
                    │    不存在的短码直接返回 404        │
                    └──────────────────┬───────────────┘
                                       │ 可能存在
                    ┌──────────────────▼───────────────┐
                    │        Redis 缓存                 │
                    │    热点数据直接返回，不查库         │
                    └──────────────────┬───────────────┘
                                       │ 缓存未命中
                    ┌──────────────────▼───────────────┐
                    │        MySQL 数据库               │
                    │    冷数据回源，结果回写到缓存       │
                    └──────────────────┬───────────────┘
                                       │
                    ┌──────────────────▼───────────────┐
                    │     Channel + Goroutine          │
                    │    异步批量写入访问统计             │
                    └──────────────────────────────────┘
```

## 性能特性

- **缓存命中率**：热点短链直接走 Redis，减少数据库压力
- **缓存穿透防护**：布隆过滤器快速拦截不存在的短码
- **限流保护**：IP 级别滑动窗口限流，防止恶意刷接口
- **异步写入**：访问日志通过 Channel 异步收集，批量写入数据库
