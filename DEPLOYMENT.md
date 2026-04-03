# Timer SaaS 部署指南

## 架构

```
外部用户
  │
  ├── /api/* (用户管理) → one-api (:3000) 直接处理
  │     登录/注册/token/日志/OAuth
  │
  ├── /timer/api/v1/forecast (简单预测)
  │     → one-api [SessionTokenAuth, 计费] → timer-rest-service (:10810)
  │
  └── /timer/api/v1/predictions/* (复杂业务)
        → one-api [SessionTokenAuth, 计费] → saas-backend (:8080)
                                                  │ 内部直连
                                                  └── timer-rest-service (:10810)
```

## 认证方式

- **网页用户**: session cookie → SessionTokenAuth 自动查找 system token 计费
- **SDK/API 用户**: Bearer token (sk-xxx) → 标准 TokenAuth
- **saas-backend 验证**: HMAC 签名 (X-OneApi-Signature)

## 注册时自动创建的 Token

| Token 名称 | 模型限制 | 用途 |
|-----------|---------|------|
| default | 无 | 用户自用 API key |
| system | 无 | 网页请求计费（子网限制，用户不可删） |

## 环境变量

### one-api
```bash
SYSTEM_SAAS_TOKEN_SUBNET=127.0.0.1/32  # system token 访问子网限制
CRITICAL_RATE_LIMIT_NUM=100            # 关键接口限流
SAAS_BACKEND_SECRET=your-secret        # HMAC 签名密钥（与 saas-backend 一致）
```

### saas-backend (application.yml)
```yaml
timer:
  api:
    gateway-type: DIRECT
    base-url: http://127.0.0.1:10810
  one-api:
    shared-secret: your-secret  # 与 one-api 的 SAAS_BACKEND_SECRET 一致
```

## 启动顺序

```bash
# 1. timer-rest-service (AI 推理)
tmux new-session -d -s timer-rest -c /path/to/timer-rest-service \
  'poetry run python -m iotdb.ainode.core.script start'

# 2. one-api (API 网关)
tmux new-session -d -s one-api -c /path/to/one-api \
  'SYSTEM_SAAS_TOKEN_SUBNET=127.0.0.1/32 SAAS_BACKEND_SECRET=your-secret ./one-api --port 3000'

# 3. saas-backend (业务后端)
tmux new-session -d -s saas-backend -c /path/to/timer-saas \
  'java -jar backend/target/timer-saas-1.0.0.jar'

# 4. saas-frontend (开发模式)
tmux new-session -d -s saas-frontend -c /path/to/timer-saas/frontend \
  'npm run dev'
```

## 渠道配置

在 one-api 管理后台创建两个渠道：

1. **Timer 时序大模型** (type 52): Base URL = `http://127.0.0.1:10810`, 模型 = sundial,chronos2,timer,timer_xl,moirai2
2. **Timer SaaS Backend** (type 53): Base URL = `http://127.0.0.1:8080/api`, 模型 = saas-backend

## 构建

```bash
# Go
cd one-api-for-tslm && go build -o one-api

# Java (JDK 26)
export JAVA_HOME=/usr/local/jdk-26
cd timer-saas/backend && mvn package -DskipTests

# Frontend (one-api)
cd one-api-for-tslm/web/default && DISABLE_ESLINT_PLUGIN=true npx react-scripts build && mv build ../build/default
```
