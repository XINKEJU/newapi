# newapi 项目记忆

## 仓库信息
- 来源：https://github.com/QuantumNous/new-api.git
- 本地路径：C:\Users\CORGI\IdeaProjects\newapi

## 项目定位
下一代 LLM API 网关 + AI 资产管理系统。基于 One-API 扩展，支持 38+ AI 提供商，含计费/订阅/支付体系。

## 技术栈
- 后端：Go 1.25 + Gin + GORM（MySQL/PG/SQLite 三选一）
- 缓存：Redis
- 前端：React 19 + TanStack Router + TailwindCSS 4 + Rsbuild（双主题 default/classic）
- 部署：Docker 多阶段构建，默认端口 3000

## 关键目录
- `relay/channel/` — AI 适配器（38+）
- `model/` — 数据库模型（24 张表，AutoMigrate）
- `web/default/` — 主前端主题
- `web/classic/` — 经典主题

## 支持的 AI 提供商（部分）
OpenAI, Claude, Gemini, DeepSeek, Baidu, Ali (Qwen), Zhipu, Tencent, Mistral, AWS Bedrock, Vertex AI, Ollama, Coze, Replicate, Suno, Kling, Jimeng, Vidu, Sora, MiniMax

## 支付系统
Stripe, Epay (易支付), Waffo Pancake
