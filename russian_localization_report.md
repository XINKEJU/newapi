# new-api 项目俄罗斯本地化综合调研报告

> 调研范围：`C:\Users\CORGI\IdeaProjects\newapi`  
> 调研日期：2026-06-12  
> 性质：纯调研，未做任何代码修改

---

## 1. 前端 i18n 俄语翻译完整性

### 文件概况

| 文件 | 大小 | 键数 |
|------|------|------|
| `web/default/src/i18n/locales/ru.json` | ~0.4 MB | 4539 |
| `web/default/src/i18n/locales/en.json` | ~0.3 MB | 4539 |

### 键值对齐

- 缺失键：**0**（ru.json 与 en.json 完全对齐）
- 多余键：**0**
- 结构：扁平键，`translation` 命名空间下

### 未翻译键分析

共 **145 个键**的 ru.json 值与 en.json 完全相同（英文原文），分类如下：

| 类别 | 数量 | 说明 |
|------|------|------|
| 品牌名/产品名 | ~31 | 如 "OpenAI", "Anthropic", "Claude" 等无需翻译 |
| 技术配置项 | ~9 | 如 "OAuth Client Secret", "Top Up Link" 等技术术语 |
| 有意义未翻译 UI 文本 | ~26 | 需要人工翻译的 UI 字符串 |
| 其他英文原文保留 | ~79 | 部分为通用 UI 术语 |

### 典型未翻译 UI 文本示例

- `"AI Proxy"` — 产品/功能名
- `"WeChat Pay"` — 支付方式名
- `"Waffo Pancake Dashboard"` — 第三方面板名
- `"TTFT P50/P95/P99"` — 性能指标缩写
- `"OAuth Client Secret"` — 技术配置项

### 中文残留

- **1 个键**仍含中文文本，未从中文源翻译为俄语

### 总体评价

翻译完整度极高（4539/4539 键对齐），未翻译的 145 个键中约 80% 为品牌名或技术术语，实际需翻译的 UI 文本约 26 个。整体质量优秀。

---

## 2. 后端 i18n

### 文件概况

| 文件 | 行数 |
|------|------|
| `i18n/locales/ru.yaml` | 279 |
| `i18n/locales/en.yaml` | 279 |

### 键值对齐

- ru.yaml 与 en.yaml **1:1 完全对齐**，0 缺失键
- 覆盖领域：common, auth, token, redemption, user, quota, subscription, payment, topup, channel, model, vendor, group, checkin, passkey, 2FA, rate_limit, setting, deployment, performance, ability, oauth, distributor, custom_oauth

### 基础设施 (`i18n/i18n.go`, 236 行)

- 库：`github.com/nicksnyder/go-i18n/v2/i18n`
- 加载方式：Go `embed.FS` 嵌入，YAML 反序列化
- 支持 4 种语言：`zh-CN`, `zh-TW`, `en`, `ru`

### ⚠️ 关键发现：默认语言为俄语

```go
// i18n/i18n.go
DefaultLang = LangRu  // 默认语言设为俄语
```

语言检测优先级：用户设置 → DB 懒加载 → 中间件上下文 → Accept-Language 头 → **DefaultLang (ru)**

这意味着未配置语言的请求将默认返回俄语消息。

### 键常量 (`i18n/keys.go`, 333 行)

- 所有 i18n 消息键常量均有对应 ru.yaml 条目
- 无遗漏

### 翻译质量

俄语翻译语法正确，术语专业，如：
- "Подписка" (订阅), "Пополнение" (充值), "Канал" (渠道), "Пользователь" (用户)

---

## 3. 文档

### README 多语言情况

| 文件 | 行数 | 状态 |
|------|------|------|
| `README.md` | 489 | 简体中文 |
| `README.en.md` | 463 | English |
| `README.fr.md` | ~480 | Français |
| `README.ja.md` | ~480 | 日本語 |
| `README.ru.md` | — | ❌ 不存在 |

### 语言链接缺失

README.md 顶部语言切换链接：简体中文、繁體中文、English、Français、日本語 — **无 Русский 链接**

### 其他文档

- 无俄语 API 文档
- 无俄语部署指南
- 无俄语用户手册
- 代码注释中仍保留大量中文注释（如 `service/yoomoney.go`）

---

## 4. 支付/货币 RUB 支持

### 货币显示系统

**后端** (`setting/operation_setting/general_setting.go`, 97 行)：

| 显示类型 | 符号 | 说明 |
|----------|------|------|
| USD | $ | 美元 |
| CNY | ¥ | 人民币 |
| RUB | ₽ | 俄罗斯卢布 |
| TOKENS | — | Token 计量 |
| CUSTOM | — | 自定义 |

### ⚠️ 关键发现：后端默认配额显示类型为 RUB

```go
// setting/operation_setting/general_setting.go
generalSetting.QuotaDisplayType = QuotaDisplayTypeRUB  // 默认 RUB
```

但前端默认为 USD：

```typescript
// web/default/src/stores/system-config-store.ts
quotaDisplayType: 'USD'
```

前后端默认值不一致。

### ⚠️ RUB 汇率 Bug

```go
// general_setting.go - GetUsdToCurrencyRate()
case QuotaDisplayTypeRUB:
    return usdToCny  // 返回的是 CNY 汇率，非 RUB 汇率！
```

RUB 货币转换使用了 CNY 汇率，会导致 RUB 金额计算错误（除非管理员手动将 CNY 汇率设为 RUB 汇率）。

### YooMoney 支付集成

**配置** (`setting/payment_yoomoney.go`, 86 行)：

| 配置项 | 默认值 |
|--------|--------|
| `YOOMONEY_ENABLED` | `true` |
| `YOOMONEY_CURRENCY` | `RUB` |
| `YOOMONEY_MIN_TOPUP` | `50` (卢布) |

**服务实现** (`service/yoomoney.go`, 152 行)：

- 支付方式：PC（钱包）、AC（银行卡）、MC（手机）、SB（SberPay/联邦储蓄银行在线）
- SHA1 签名验证用于 webhook
- 路由：`/api/yoomoney/topup`、`/api/yoomoney/subscription`、`/api/yoomoney/notify`、`/api/yoomoney/return`

### 前端货币格式化 (`web/default/src/lib/currency.ts`, 583 行)

- 完整 RUB 支持：`Intl.NumberFormat` + `style: 'currency'` + `currencyCode: 'RUB'`
- RUB 显示元数据：`{ kind: 'currency', symbol: '₽', currencyCode: 'RUB' }`
- `getCurrencyLabel()` 对 RUB 返回 `'RUB'`

---

## 5. 俄罗斯 OAuth 服务

### VK (VKontakte) — `oauth/vk.go`, 210 行

| 项目 | 详情 |
|------|------|
| OAuth 端点 | `oauth.vk.com/access_token` |
| API 版本 | 5.131 |
| 注册名 | `vk` |
| 用户标识 | `vk_<userID>`，存储于 `user.VkId` 字段 |
| 启用条件 | `common.VKOAuthEnabled` |

### Yandex ID — `oauth/yandex.go`, 200 行

| 项目 | 详情 |
|------|------|
| Token 端点 | `oauth.yandex.ru/token` |
| 用户信息 | `login.yandex.ru/info` |
| 注册名 | `yandex` |
| 用户标识 | `ya_<userID>` 或 login，存储于 `user.YandexId` 字段 |
| 启用条件 | `common.YandexOAuthEnabled` |

### 环境变量 (`common/init.go`, 197 行)

| 变量 | 默认值 |
|------|--------|
| `VK_OAUTH_ENABLED` | `false` |
| `VK_CLIENT_ID` | — |
| `VK_CLIENT_SECRET` | — |
| `YANDEX_OAUTH_ENABLED` | `false` |
| `YANDEX_CLIENT_ID` | — |
| `YANDEX_CLIENT_SECRET` | — |

### OAuth 注册表 (`oauth/registry.go`, 135 行)

- 支持动态加载/卸载自定义 OAuth 提供商（从数据库）
- VK 和 Yandex 为内置提供者

---

## 6. 俄罗斯 AI 服务商

### 渠道类型定义 (`constant/channel.go`)

| 渠道 | 类型 ID | Base URL |
|------|---------|----------|
| YandexGPT | 58 | `https://llm.api.cloud.yandex.net` |
| GigaChat | 59 | `https://gigachat.devices.sberbank.ru/api/v1` |

### ⚠️ 关键发现：无 Relay 适配器实现

在 `relay/channel/` 目录下：
- 无 `yandexgpt/` 目录
- 无 `gigachat/` 目录
- 无任何对应适配器文件

这意味着虽然渠道类型已定义，但**无法实际使用** YandexGPT 和 GigaChat 服务。需要实现完整的 relay 适配器才能对接这两个俄罗斯 AI 服务。

---

## 7. 前端 UI 硬编码字符串与日期/数字格式

### 硬编码英文字符串

**`web/default/src/lib/format.ts`, 第 121 行**：

```typescript
return 'Never'  // 硬编码英文，未使用 t() 包裹
```

此字符串在"永不过期"等时间显示场景中出现，未通过 i18n 翻译函数处理。

### 日期/时间格式化

**`web/default/src/lib/format.ts`, 237 行**：
- 日期格式硬编码为 `YYYY-MM-DD HH:mm:ss`（ISO 格式，语言无关）
- 使用 `dayjs` 库

**`web/default/src/lib/dayjs.ts`, 24 行**：

```typescript
// 仅导入 relativeTime 插件，未导入俄语 locale
import relativeTime from 'dayjs/plugin/relativeTime'
// 缺失：import 'dayjs/locale/ru'
```

**影响**：相对时间文本（如 "5 minutes ago"、"in 3 days"）将始终显示英文，即使界面语言切换为俄语。

### 数字格式化

- 货币使用 `Intl.NumberFormat`（浏览器原生国际化 API），**可正确处理俄语格式**
- 其他数字显示未使用 `Intl.NumberFormat`，无千位分隔符本地化

---

## 8. 法律/合规

### 调查结果

- 在整个项目代码库中**未发现**以下任何合规相关内容：
  - 俄罗斯联邦 ФЗ-152《个人数据法》
  - GDPR（通用数据保护条例）
  - 任何数据隐私政策模板
  - 用户数据处理同意机制
  - 数据保留策略

### 缺失的合规要素

| 合规要求 | 状态 |
|----------|------|
| ФЗ-152 个人数据处理声明 | ❌ 不存在 |
| 隐私政策页面 | ❌ 不存在 |
| 用户数据处理同意 checkbox | ❌ 不存在 |
| Cookie 同意机制 | ❌ 不存在 |
| 数据导出/删除功能 | ❌ 不存在 |
| 数据保留策略 | ❌ 不存在 |

如果项目面向俄罗斯用户运营，需补充 ФЗ-152 合规要素。

---

## 9. Docker/部署

### Dockerfile (53 行)

- 标准 Go 多阶段构建：Bun (前端) → Go (后端) → Debian slim
- **无俄罗斯特定部署配置**
- 无时区默认设置（可考虑 `TZ=Europe/Moscow`）
- 无俄语 locale 安装

### 部署建议（仅建议，未实施）

- 如需面向俄罗斯用户，可考虑添加 `TZ=Europe/Moscow` 环境变量
- 考虑添加俄语 locale 包（`locales ru_RU.UTF-8`）用于服务端日志时间格式等

---

## 10. 错误消息本地化

### 后端错误消息

- **完全本地化**：通过 `i18n/i18n.go` 基础设施，所有后端返回的错误消息均支持俄语
- 错误消息键定义在 `i18n/keys.go`，翻译在 `i18n/locales/ru.yaml`
- 由于默认语言为俄语，未指定语言的请求将收到俄语错误消息

### 前端错误消息

- 使用 i18n 键引用，可正常翻译
- 唯一例外：`format.ts` 中的硬编码 `'Never'` 字符串

### YooMoney 服务注释

`service/yoomoney.go` 中仍保留中文注释，未本地化（功能不受影响，但影响代码可维护性）。

---

## 汇总：关键问题与建议

### 🔴 需修复的问题

| # | 问题 | 文件 | 严重度 |
|---|------|------|--------|
| 1 | RUB 汇率使用 CNY 值（汇率 Bug） | `setting/operation_setting/general_setting.go` | 高 |
| 2 | YandexGPT/GigaChat 无 relay 适配器 | `relay/channel/` 缺失 | 高 |
| 3 | 前后端默认货币不一致（RUB vs USD） | `general_setting.go` vs `system-config-store.ts` | 中 |

### 🟡 建议改进

| # | 问题 | 文件 | 优先级 |
|---|------|------|--------|
| 4 | 硬编码 'Never' 字符串 | `web/default/src/lib/format.ts:121` | 中 |
| 5 | dayjs 缺少俄语 locale | `web/default/src/lib/dayjs.ts` | 中 |
| 6 | 无 README.ru.md | 项目根目录 | 低 |
| 7 | ru.json 中 26 个未翻译 UI 文本 | `web/default/src/i18n/locales/ru.json` | 低 |
| 8 | ru.json 中 1 个中文残留 | `web/default/src/i18n/locales/ru.json` | 低 |
| 9 | 无 ФЗ-152 合规内容 | 全项目 | 视运营需求 |
| 10 | YooMoney 代码中文注释 | `service/yoomoney.go` | 低 |

### ✅ 已完善的部分

- 后端 i18n 完整（279 键 1:1 对齐，高质量翻译）
- 前端 i18n 近乎完整（4539 键对齐，仅 26 个 UI 文本待翻译）
- VK/Yandex OAuth 完整实现
- YooMoney 支付完整集成（含 SberPay）
- RUB 货币格式化完整（`Intl.NumberFormat`）
- 后端默认语言为俄语
