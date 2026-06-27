# umniai.ru 生产环境部署指南

**域名**: umniai.ru  
**服务器 IP**: 170.168.89.127  
**应用端口**: 3000 (内部) → 80/443 (Nginx 反向代理)

---

## 1. 前置条件

| 项目 | 要求 |
|------|------|
| 操作系统 | Linux (Ubuntu 22.04+ / Debian 12+) |
| Docker | 24.0+ |
| Docker Compose | v2.20+ |
| 域名 DNS | `umniai.ru` A 记录 → `170.168.89.127` |
| 域名 DNS | `www.umniai.ru` A 记录 → `170.168.89.127` |
| 防火墙 | 开放 80 (HTTP) 和 443 (HTTPS) 端口 |

## 2. 文件结构

```
newapi/
├── docker-compose.prod.yml    # 生产环境编排配置
├── nginx/
│   ├── umniai.ru.conf         # HTTPS 完整配置 (Nginx 反向代理 + SSL)
│   └── umniai.ru.initial.conf # HTTP-only 初始配置 (用于获取证书前)
├── deploy/
│   ├── init-ssl.sh            # 首次获取 SSL 证书
│   ├── switch-https.sh        # 切换到 HTTPS
│   └── renew-ssl.sh           # 证书自动续期
└── data/                      # 应用数据 (自动创建)
    └── logs/                  # 应用日志
```

## 3. 部署步骤

### 3.1 服务器准备

```bash
# 克隆项目到服务器
git clone <repo-url> /opt/newapi
cd /opt/newapi
```

### 3.2 修改密码

编辑 `docker-compose.prod.yml`，将所有 `CHANGE_THIS_PASSWORD` 替换为强密码：

```bash
# 生成随机密码
openssl rand -hex 24
```

需要修改的位置：
- `SQL_DSN` 中的 PostgreSQL 密码
- `REDIS_CONN_STRING` 中的 Redis 密码
- `redis` 服务的 `--requirepass`
- `postgres` 服务的 `POSTGRES_PASSWORD`

> **注意**: SQL_DSN、Redis 连接串和对应服务中的密码必须一致。

### 3.3 获取 SSL 证书并启动

```bash
# 一键获取 SSL 证书并启动 HTTPS 服务
bash deploy/init-ssl.sh admin@umniai.ru
```

此脚本会：
1. 使用 HTTP-only 配置启动 Nginx
2. 验证 HTTP 访问
3. 通过 Let's Encrypt 获取 SSL 证书
4. 切换到 HTTPS 配置

### 3.4 手动启动（如已有证书）

```bash
docker compose -f docker-compose.prod.yml up -d
```

## 4. 配置 ServerAddress

服务启动后，需要在前端管理后台配置服务器地址：

1. 访问 `https://umniai.ru`
2. 完成初始设置向导（创建管理员账户）
3. 进入 **设置 → 系统设置**
4. 将 **服务器地址** 设置为：`https://umniai.ru`

> 此地址用于：
> - OAuth 登录回调（Yandex、VK、GitHub 等）
> - 支付回调（YooMoney、Stripe、Epay）
> - 密码重置邮件链接
> - Midjourney 图片 URL
> - Passkey 认证 Origin
> - 视频任务内容 URL

## 5. SSL 证书续期

Let's Encrypt 证书有效期为 90 天，添加 cron 定时续期：

```bash
# 添加到 crontab（每月 1 号凌晨 3 点续期）
crontab -e
# 添加以下行：
0 3 1 * * cd /opt/newapi && bash deploy/renew-ssl.sh >> /var/log/certbot-renew.log 2>&1
```

## 6. Nginx 架构

```
客户端请求
    │
    ▼
┌──────────────────────────┐
│  Nginx (80/443)          │
│  umniai.ru               │
│                          │
│  80 → 301 → 443          │  HTTP 重定向到 HTTPS
│  443 → proxy_pass        │  反向代理到后端
│      new-api:3000        │
│                          │
│  SSL: Let's Encrypt      │
│  SSE: proxy_buffering off│  流式响应支持
│  WS:  Upgrade 头          │  WebSocket 支持
└──────────┬───────────────┘
           │
           ▼
┌──────────────────────────┐
│  new-api (3000)          │
│  Go + Gin                │
│  前端静态文件 + API       │
└──────────┬───────────────┘
           │
     ┌─────┴─────┐
     ▼           ▼
┌─────────┐ ┌─────────┐
│ Redis   │ │PostgreSQL│
│ (cache) │ │ (data)   │
└─────────┘ └─────────┘
```

## 7. 常用运维命令

```bash
# 查看服务状态
docker compose -f docker-compose.prod.yml ps

# 查看日志
docker compose -f docker-compose.prod.yml logs -f new-api
docker compose -f docker-compose.prod.yml logs -f nginx

# 重启服务
docker compose -f docker-compose.prod.yml restart new-api
docker compose -f docker-compose.prod.yml restart nginx

# 更新镜像
docker compose -f docker-compose.prod.yml pull new-api
docker compose -f docker-compose.prod.yml up -d new-api

# 备份数据库
docker compose -f docker-compose.prod.yml exec postgres \
    pg_dump -U newapi newapi > backup_$(date +%Y%m%d).sql

# 恢复数据库
docker compose -f docker-compose.prod.yml exec -T postgres \
    psql -U newapi newapi < backup_YYYYMMDD.sql
```

## 8. 故障排查

| 问题 | 排查方法 |
|------|----------|
| HTTPS 访问失败 | 检查证书: `docker compose -f docker-compose.prod.yml exec nginx ls /etc/letsencrypt/live/umniai.ru/` |
| 502 Bad Gateway | 检查 new-api 是否运行: `docker compose -f docker-compose.prod.yml logs new-api` |
| 流式响应中断 | 确认 Nginx 配置中 `proxy_buffering off` 已生效 |
| OAuth 回调失败 | 确认管理后台 ServerAddress 已设置为 `https://umniai.ru` |
| 支付回调失败 | 确认 `TRUSTED_REDIRECT_DOMAINS=umniai.ru` 已设置 |
| DNS 未生效 | `dig umniai.ru` 或 `nslookup umniai.ru` 确认解析到 170.168.89.127 |
```
