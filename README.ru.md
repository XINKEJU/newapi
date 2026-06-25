<div align="center">

![new-api](/web/default/public/logo.png)

# New API

🍥 **Шлюз API для больших языковых моделей нового поколения и система управления AI-активами**

<p align="center">
  <a href="./README.md">中文</a> |
  <a href="./README.en.md">English</a> |
  <strong>Русский</strong> |
  <a href="./README.fr.md">Français</a> |
  <a href="./README.ja.md">日本語</a>
</p>

<p align="center">
  <a href="https://raw.githubusercontent.com/Calcium-Ion/new-api/main/LICENSE">
    <img src="https://img.shields.io/github/license/Calcium-Ion/new-api?color=brightgreen" alt="license">
  </a>
  <a href="https://github.com/Calcium-Ion/new-api/releases/latest">
    <img src="https://img.shields.io/github/v/release/Calcium-Ion/new-api?color=brightgreen&include_prereleases" alt="release">
  </a>
  <a href="https://github.com/users/Calcium-Ion/packages/container/package/new-api">
    <img src="https://img.shields.io/badge/docker-ghcr.io-blue" alt="docker">
  </a>
  <a href="https://hub.docker.com/r/CalciumIon/new-api">
    <img src="https://img.shields.io/badge/docker-dockerHub-blue" alt="docker">
  </a>
  <a href="https://goreportcard.com/report/github.com/Calcium-Ion/new-api">
    <img src="https://goreportcard.com/badge/github.com/Calcium-Ion/new-api" alt="GoReportCard">
  </a>
</p>

<p align="center">
  <a href="https://trendshift.io/repositories/8227" target="_blank">
    <img src="https://trendshift.io/api/badge/repositories/8227" alt="Calcium-Ion%2Fnew-api | Trendshift" style="width: 250px; height: 55px;" width="250" height="55"/>
  </a>
</p>

<p align="center">
  <a href="#-быстрый-запуск">Быстрый запуск</a> •
  <a href="#-ключевые-возможности">Возможности</a> •
  <a href="#-развёртывание">Развёртывание</a> •
  <a href="#-документация">Документация</a> •
  <a href="#-поддержка">Поддержка</a>
</p>

</div>

## 📝 Описание проекта

> [!NOTE]
> Это проект с открытым исходным кодом, основанный на [One API](https://github.com/songquanpeng/one-api)

> [!IMPORTANT]
> - Данный проект предназначен исключительно для законного и авторизованного использования в качестве AI API-шлюза, аутентификации на уровне организации, управления мультимоделями, аналитики использования, учёта затрат и частного развёртывания.
> - Пользователи обязаны законно получать ключи API, учётные записи, доступ к моделям и интерфейсам от провайдеров и соблюдать условия обслуживания и применимое законодательство.
> - При предоставлении генеративных AI-сервисов общественности пользователи должны соблюдать требования регулирующих органов, включая регистрацию, лицензирование, модерацию контента, верификацию пользователей, хранение логов, налогообложение и авторизацию провайдеров.

---

## 🤝 Доверенные партнёры

<p align="center">
  <em>В произвольном порядке</em>
</p>

<p align="center">
  <a href="https://www.cherry-ai.com/" target="_blank">
    <img src="./docs/images/cherry-studio.png" alt="Cherry Studio" height="80" />
  </a>
  <a href="https://bda.pku.edu.cn/" target="_blank">
    <img src="./docs/images/pku.png" alt="Peking University" height="80" />
  </a>
  <a href="https://www.compshare.cn/?ytag=GPU_yy_gh_newapi" target="_blank">
    <img src="./docs/images/ucloud.png" alt="UCloud" height="80" />
  </a>
  <a href="https://www.aliyun.com/" target="_blank">
    <img src="./docs/images/aliyun.png" alt="Alibaba Cloud" height="80" />
  </a>
  <a href="https://io.net/" target="_blank">
    <img src="./docs/images/io-net.png" alt="IO.NET" height="80" />
  </a>
</p>

---

## 🙏 Особая благодарность

<p align="center">
  <a href="https://www.jetbrains.com/?from=new-api" target="_blank">
    <img src="https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.png" alt="JetBrains Logo" width="120" />
  </a>
</p>

<p align="center">
  <strong>Благодарим <a href="https://www.jetbrains.com/?from=new-api">JetBrains</a> за предоставление бесплатной лицензии для разработки open-source проекта</strong>
</p>

---

## 🚀 Быстрый запуск

### Docker Compose (рекомендуется)

```bash
# Клонирование репозитория
git clone https://github.com/QuantumNous/new-api.git
cd new-api

# Настройка docker-compose.yml
nano docker-compose.yml

# Запуск сервиса
docker-compose up -d
```

<details>
<summary><strong>Запуск через Docker команды</strong></summary>

```bash
# Получение последнего образа
docker pull calciumion/new-api:latest

# Использование SQLite (по умолчанию)
docker run --name new-api -d --restart always \
  -p 3000:3000 \
  -e TZ=Europe/Moscow \
  -v ./data:/data \
  calciumion/new-api:latest

# Использование MySQL
docker run --name new-api -d --restart always \
  -p 3000:3000 \
  -e SQL_DSN="root:123456@tcp(localhost:3306)/oneapi" \
  -e TZ=Europe/Moscow \
  -v ./data:/data \
  calciumion/new-api:latest
```

> **💡 Подсказка:** `-v ./data:/data` сохранит данные в папку `data` текущей директории. Можно указать абсолютный путь: `-v /your/custom/path:/data`

</details>

---

🎉 После завершения развёртывания откройте `http://localhost:3000` для начала работы!

> [!WARNING]
> При эксплуатации данного проекта в качестве публичного сервиса генеративного AI или сервиса перепродажи API, пользователи должны предварительно выполнить все требования по регистрации, лицензированию, безопасности контента, верификации пользователей, хранению логов, налогам, платежам и авторизации провайдеров.

📖 Подробнее о способах развёртывания: [Руководство по установке](https://docs.newapi.pro/en/docs/installation)

---

## 📚 Документация

<div align="center">

### 📖 [Официальная документация](https://docs.newapi.pro/en/docs) | [![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/QuantumNous/new-api)

</div>

**Быстрая навигация:**

| Категория | Ссылка |
|------|------|
| 🚀 Установка | [Документация по установке](https://docs.newapi.pro/en/docs/installation) |
| ⚙️ Конфигурация | [Переменные окружения](https://docs.newapi.pro/en/docs/installation/config-maintenance/environment-variables) |
| 📡 API | [Документация API](https://docs.newapi.pro/en/docs/api) |
| ❓ FAQ | [Часто задаваемые вопросы](https://docs.newapi.pro/en/docs/support/faq) |
| 💬 Сообщество | [Каналы связи](https://docs.newapi.pro/en/docs/support/community-interaction) |

---

## ✨ Ключевые возможности

> Подробнее: [Обзор функций](https://docs.newapi.pro/en/docs/guide/wiki/basic-concepts/features-introduction)

### 🎨 Основные функции

| Функция | Описание |
|------|------|
| 🎨 Новый интерфейс | Современный дизайн пользовательского интерфейса |
| 🌍 Мультиязычность | Поддержка русского, китайского, английского, французского, японского |
| 🇷🇺 Локализация для РФ | Рубль (₽) по умолчанию, часовой пояс Europe/Moscow, провайдеры YandexGPT и GigaChat |
| 💳 Платежи для РФ | YooMoney (ЮMoney), SberPay, Stripe |
| 🔐 OAuth для РФ | Yandex ID, VK (ВКонтакте), GitHub, Discord, Telegram, LinuxDO, OIDC |
| 🔄 Совместимость данных | Полная совместимость с базой данных One API |
| 📈 Панель аналитики | Визуальная консоль и статистический анализ |
| 🔒 Управление доступом | Группы токенов, ограничения моделей, управление пользователями |

### 💰 Учёт и биллинг

- ✅ Внутреннее пополнение и распределение квот (YooMoney, SberPay, Stripe, EPay)
- ✅ Посессионный и модельный учёт затрат, кэширование
- ✅ Статистика биллинга кэша для OpenAI, Azure, DeepSeek, Claude, Qwen и других
- ✅ Гибкая биллинговая политика для внутреннего управления и корпоративных клиентов

### 🔐 Авторизация и безопасность

- 😈 Авторизация через Discord
- 🤖 Авторизация через LinuxDO
- 📱 Авторизация через Telegram
- 🇷🇺 Авторизация через Yandex ID
- 🇷🇺 Авторизация через VK (ВКонтакте)
- 🔑 Единая аутентификация OIDC

### 🤖 Российские AI-провайдеры

| Провайдер | Описание | Статус |
|-----------|----------|--------|
| 🇷🇺 YandexGPT | Нейросеть Яндекса (Yandex Cloud) | ✅ Поддерживается |
| 🇷🇺 GigaChat | Нейросеть Сбера | ✅ Поддерживается |

### 🚀 Продвинутые возможности

**Поддержка форматов API:**
- ⚡ [OpenAI Responses](https://docs.newapi.pro/en/docs/api/ai-model/chat/openai/create-response)
- ⚡ [OpenAI Realtime API](https://docs.newapi.pro/en/docs/api/ai-model/realtime/create-realtime-session) (включая Azure)
- ⚡ [Claude Messages](https://docs.newapi.pro/en/docs/api/ai-model/chat/create-message)
- ⚡ [Google Gemini](https://doc.newapi.pro/en/api/google-gemini-chat)
- 🔄 [Rerank модели](https://docs.newapi.pro/en/docs/api/ai-model/rerank/create-rerank) (Cohere, Jina)

**Интеллектуальная маршрутизация:**
- ⚖️ Взвешенная случайная маршрутизация каналов
- 🔄 Автоматический повтор при сбое
- 🚦 Ограничение частоты запросов на уровне пользователя

**Конвертация форматов:**
- 🔄 **Совместимость с OpenAI ⇄ Claude Messages**
- 🔄 **Совместимость с OpenAI → Google Gemini**
- 🔄 **Google Gemini → Совместимость с OpenAI** — Только текст, вызов функций пока не поддерживается
- 🚧 **Совместимость с OpenAI ⇄ OpenAI Responses** — В разработке

---

## 🤖 Поддержка моделей

> Подробнее: [Документация API - Интерфейс шлюза](https://docs.newapi.pro/en/docs/api)

| Тип модели | Описание | Документация |
|---------|------|------|
| 🤖 OpenAI GPTs | gpt-4-gizmo-* серия | - |
| 🇷🇺 YandexGPT | Модели Yandex Cloud | - |
| 🇷🇺 GigaChat | Модели Сбера | - |
| 🎨 Midjourney-Proxy | [Midjourney-Proxy(Plus)](https://github.com/novicezk/midjourney-proxy) | [Документация](https://doc.newapi.pro/en/api/midjourney-proxy-image) |
| 🎵 Suno-API | [Suno API](https://github.com/Suno-API/Suno-API) | [Документация](https://doc.newapi.pro/en/api/suno-music) |
| 🔄 Rerank | Cohere, Jina | [Документация](https://docs.newapi.pro/en/docs/api/ai-model/rerank/create-rerank) |
| 💬 Claude | Формат Messages | [Документация](https://docs.newapi.pro/en/docs/api/ai-model/chat/create-message) |
| 🌐 Gemini | Формат Google Gemini | [Документация](https://doc.newapi.pro/en/api/google-gemini-chat) |
| 🔧 Dify | Режим ChatFlow | - |
| 🎯 Пользовательский | Поддержка настройки авторизованных эндпоинтов | - |

---

## 🚢 Развёртывание

> [!TIP]
> **Последний Docker образ:** `calciumion/new-api:latest`

### 📋 Требования

| Компонент | Требование |
|------|------|
| **Локальная БД** | SQLite (Docker должен монтировать `/data`)|
| **Внешняя БД** | MySQL ≥ 5.7.8 или PostgreSQL ≥ 9.6 |
| **Контейнеризация** | Docker / Docker Compose |

### ⚙️ Переменные окружения

<details>
<summary>Основные переменные окружения</summary>

| Переменная | Описание | По умолчанию |
|--------|------|--------|
| `SESSION_SECRET` | Секрет сессии (обязателен при многомашинном развёртывании) | - |
| `CRYPTO_SECRET` | Секрет шифрования (обязателен при использовании Redis) | - |
| `SQL_DSN` | Строка подключения к БД | - |
| `REDIS_CONN_STRING` | Строка подключения к Redis | - |
| `STREAMING_TIMEOUT` | Таймаут потоковой передачи (сек) | `300` |
| `STREAM_SCANNER_MAX_BUFFER_MB` | Макс. буфер строки (МБ) для потокового сканера | `64` |
| `MAX_REQUEST_BODY_MB` | Макс. размер тела запроса (МБ, после декомпрессии) | `32` |
| `ERROR_LOG_ENABLED` | Включение журнала ошибок | `false` |
| `YANDEX_OAUTH_ENABLED` | Включить OAuth через Yandex ID | `false` |
| `YANDEX_CLIENT_ID` | Yandex OAuth Client ID | - |
| `YANDEX_CLIENT_SECRET` | Yandex OAuth Client Secret | - |
| `VK_OAUTH_ENABLED` | Включить OAuth через VK | `false` |
| `VK_CLIENT_ID` | VK OAuth Client ID | - |
| `VK_CLIENT_SECRET` | VK OAuth Client Secret | - |

📖 **Полная конфигурация:** [Документация переменных окружения](https://docs.newapi.pro/en/docs/installation/config-maintenance/environment-variables)

</details>

### ⚠️ Особенности многомашинного развёртывания

> [!WARNING]
> - **Обязательно задайте** `SESSION_SECRET` — иначе статус входа будет несогласованным
> - **При использовании общего Redis задайте** `CRYPTO_SECRET` — иначе данные не расшифруются

---

## 🔗 Связанные проекты

### Исходные проекты

| Проект | Описание |
|------|------|
| [One API](https://github.com/songquanpeng/one-api) | Основа проекта |
| [Midjourney-Proxy](https://github.com/novicezk/midjourney-proxy) | Поддержка Midjourney |

### Инструменты

| Проект | Описание |
|------|------|
| [new-api-key-tool](https://github.com/Calcium-Ion/new-api-key-tool) | Инструмент проверки квоты ключей |
| [new-api-horizon](https://github.com/Calcium-Ion/new-api-horizon) | Высокопроизводительная версия New API |

---

## 💬 Поддержка

### 📖 Ресурсы документации

| Ресурс | Ссылка |
|------|------|
| 📘 FAQ | [Часто задаваемые вопросы](https://docs.newapi.pro/en/docs/support/faq) |
| 💬 Сообщество | [Каналы связи](https://docs.newapi.pro/en/docs/support/community-interaction) |
| 🐛 Баг-репорты | [Сообщить о проблеме](https://docs.newapi.pro/en/docs/support/feedback-issues) |
| 📚 Полная документация | [Официальная документация](https://docs.newapi.pro/en/docs) |

### 🤝 Участие в разработке

Приветствуются любые формы участия!

- 🐛 Сообщения об ошибках
- 💡 Предложения новых функций
- 📝 Улучшение документации
- 🔧 Pull request'ы с кодом

---

## 🌟 История звёзд

<div align="center">

[![Star History Chart](https://api.star-history.com/svg?repos=Calcium-Ion/new-api&type=Date)](https://star-history.com/#Calcium-Ion/new-api&Date)

</div>

---

<div align="center">

### 💖 Спасибо за использование New API

Если этот проект оказался полезным, поставьте ⭐️ Star!

**[Официальная документация](https://docs.newapi.pro/en/docs)** • **[Баг-репорты](https://github.com/Calcium-Ion/new-api/issues)** • **[Релизы](https://github.com/Calcium-Ion/new-api/releases)**

<sub>Создано с ❤️ командой QuantumNous</sub>

</div>
