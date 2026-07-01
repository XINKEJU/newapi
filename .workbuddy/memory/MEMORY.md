# newapi 项目记忆

## 仓库信息
- 来源：https://github.com/QuantumNous/new-api.git
- 本地路径：C:\Users\CORGI\IdeaProjects\newapi

## 项目定位
下一代 LLM API 网关 + AI 资产管理系统。基于 One-API 扩展，支持 38+ AI 提供商，含计费/订阅/支付体系。

## 技术栈
- 后端：Go 1.25 + Gin v1.10.0 + GORM v1.25.12（MySQL/PG/SQLite 三选一）
- 缓存：Redis (go-redis v9.7.3 — 2026-07-01 从 v8 迁移)
- 前端：React 19 + TanStack Router + TailwindCSS 4 + Rsbuild（双主题 default/classic）
- 部署：Docker 多阶段构建（非 root 用户 appuser），默认端口 3000

## 关键目录
- `relay/channel/` — AI 适配器（38+）
- `model/` — 数据库模型（24 张表，AutoMigrate）
- `web/default/` — 主前端主题
- `web/classic/` — 经典主题

## 支持的 AI 提供商（部分）
OpenAI, Claude, Gemini, DeepSeek, Baidu, Ali (Qwen), Zhipu, Tencent, Mistral, AWS Bedrock, Vertex AI, Ollama, Coze, Replicate, Suno, Kling, Jimeng, Vidu, Sora, MiniMax

## 支付系统
Stripe, Epay (易支付), Waffo Pancake, YooMoney

## 生产部署
- 域名：umniai.ru (IP: 170.168.89.127)
- Nginx 反向代理：80/443 → new-api:3000
- SSL：Let's Encrypt (certbot)
- 生产配置：docker-compose.prod.yml
- 部署脚本：deploy/ (init-ssl.sh, switch-https.sh, renew-ssl.sh)
- Nginx 配置：nginx/ (umniai.ru.conf, umniai.ru.initial.conf)
- ServerAddress 需在管理后台设置为 https://umniai.ru
- TRUSTED_REDIRECT_DOMAINS=umniai.ru

## 已知修复
- **Yandex OAuth state 参数为空**: main.go 中 session cookie `SameSite` 从 `Strict` 改为 `Lax`，`Secure` 改为 `true`。原因：OAuth 回调来自跨站（yandex.ru → umniai.ru），Strict 模式下浏览器不发送 cookie，导致 oauth_state 从 session 读取为 nil。此修复同时影响所有 OAuth 提供商（GitHub/Discord/OIDC/LinuxDO/VK/Yandex）。
- **全面安全加固 (2026-07-01)**: 移除 docker-compose.prod.yml 中明文密钥(改用 .env.prod env_file)；Docker 容器改用非 root 用户 appuser；Nginx 新增 rate limiting + CSP + Permissions-Policy + OCSP Stapling + server_tokens off + map 动态 Connection 头；go-redis v8→v9 迁移；gin v1.9.1→v1.10.0；gorm v1.25.2→v1.25.12；SESSION_SECRET/CRYPTO_SECRET 未设置时增加警告日志。
- **Nginx rate limit 503 修复 (2026-07-01)**: 正则 `^/(api/user|api/oauth|api/auth)` 中 `api/auth` 误匹配 `/api/authz/`（授权目录接口），导致正常 API 请求被 5r/m 登录限流误杀返回 503。改为精确路径 `^/(api/user/login|api/user/register|api/oauth/)`。
- **部署时必须重建前端**: 安全加固部署只重编译了 Go 二进制，未重建前端导致 VK/Yandex 图标修复丢失。部署流程须包含 `cd web/default && bun run build`，然后上传 dist 到服务器并重启 new-api 容器（删除重建 dist 目录后绑定挂载 inode 变化，需 restart 容器）。
