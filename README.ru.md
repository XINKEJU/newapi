<div align="center">

![new-api](/web/default/public/logo.png)

# New API

🍥 **API-шлюз нового поколения для больших языковых моделей и система управления AI-ресурсами**

<p align="center">
  <a href="./README.md">中文</a> | 
  <a href="./README.en.md">English</a> | 
  <a href="./README.fr.md">Français</a> | 
  <a href="./README.ja.md">日本語</a> | 
  <strong>Русский</strong>
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
  <a href="#-быстрый-старт">Быстрый старт</a> •
  <a href="#-ключевые-возможности">Возможности</a> •
  <a href="#-развёртывание">Развёртывание</a> •
  <a href="#-документация">Документация</a> •
  <a href="#-поддержка">Поддержка</a>
</p>

</div>

## 📝 Описание проекта

> [!NOTE]  
> Проект является open-source решением, разработанным на основе [One API](https://github.com/songquanpeng/one-api)

> [!IMPORTANT]  
> - Проект предназначен исключительно для законных сценариев: API-шлюз, корпоративная аутентификация, управление несколькими моделями, аналитика использования, учёт затрат и частное развёртывание.
> - Пользователи обязаны законно получить ключи API, учётные записи, модели и права на использование у провайдеров, а также соблюдать условия обслуживания провайдеров и применимые законы.
> - При предоставлении генеративных AI-сервисов публично необходимо соблюдать все регуляторные требования юрисдикции.

---

## 🤝 Доверенные партнёры

<p align="center">
  <em>В произвольном порядке</em>
</p>

<p align="center">
  <a href="https://www.cherry-ai.com/" target="_blank">
    <img src="./docs/images/cherry-studio.png" alt="Cherry Studio" height="80" />
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

## 🚀 Быстрый старт

### Использование Docker Compose (рекомендуется)

```bash
# Клонирование репозитория
git clone https://github.com/QuantumNous/new-api.git
cd new-api

# Редактирование конфигурации
nano docker-compose.yml

# Запуск сервиса
docker-compose up -d
```

<details>
<summary><strong>Использование команд Docker</strong></summary>

```bash
# Скачать последний образ
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

> **💡 Подсказка:** `-v ./data:/data` сохраняет данные в папке `data` текущей директории. Можно также указать абсолютный путь: `-v /your/custom/path:/data`

</details>

---

🎉 После развёртывания откройте `http://localhost:3000` и начните работу!

> [!WARNING]
> При использовании проекта в качестве публичного AI-сервиса или API-посредника необходимо выполнить все требования регулятора: регистрация, лицензирование, проверка личности, хранение журналов, уплата налогов и авторизация от провайдеров.

📖 Подробнее о способах развёртывания: [Руководство по установке](https://docs.newapi.pro/en/docs/installation)

---

## 📚 Документация

<div align="center">

### 📖 [Официальная документация](https://docs.newapi.pro/en/docs) | [![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/QuantumNous/new-api)

</div>

**Быстрая навигация:**

| Категория | Ссылка |
|-----------|--------|
| 🚀 Руководство по установке | [Документация по установке](https://docs.newapi.pro/en/docs/installation) |
| ⚙️ Переменные окружения | [Переменные окружения](https://docs.newapi.pro/en/docs/installation/config-maintenance/environment-variables) |
| 📡 API | [Документация API](https://docs.newapi.pro/en/docs/api) |
| ❓ FAQ | [Часто задаваемые вопросы](https://docs.newapi.pro/en/docs/support/faq) |
| 💬 Сообщество | [Каналы общения](https://docs.newapi.pro/en/docs/support/community-interaction) |

---

## ✨ Ключевые возможности

> Подробное описание: [Введение в возможности](https://docs.newapi.pro/en/docs/guide/wiki/basic-concepts/features-introduction)

### 🎨 Основные функции

| Функция | Описание |
|---------|----------|
| 🎨 Новый интерфейс | Современный дизайн пользовательского интерфейса |
| 🌍 Мультиязычность | Поддержка китайского, английского, французского, японского, **русского** |
| 🔄 Совместимость данных | Полная совместимость с базой данных оригинального One API |
| 📈 Дашборд данных | Визуальная консоль и статистический анализ |
| 🔒 Управление доступом | Группировка токенов, ограничение моделей, управление пользователями |

### 💰 Биллинг и учёт использования

- ✅ Пополнение баланса и распределение квот (EPay, Stripe, YooMoney)
- ✅ Учёт затрат на уровне запроса, использования и попадания в кеш
- ✅ Статистика кеш-биллинга для OpenAI, Azure, DeepSeek, Claude, Qwen
- ✅ Гибкие политики биллинга для внутреннего управления или корпоративных клиентов

### 🔐 Авторизация и безопасность

- 😈 Авторизация через Discord
- 🤖 Авторизация через LinuxDO
- 📱 Авторизация через Telegram
- 🔑 OIDC единая аутентификация

### 🤖 Поддерживаемые провайдеры (включая российские)

| Провайдер | Описание |
|-----------|----------|
| YandexGPT | Яндекс Foundation Models (`gpt://folder-id/yandexgpt/latest`) |
| GigaChat (Sber) | GigaChat-Plus, GigaChat-Pro, GigaChat-Max |
| OpenAI | GPT-4, GPT-4o, o1, o3 и др. |
| Claude (Anthropic) | Claude 3.5, Claude 3 Opus и др. |
| Gemini (Google) | Gemini 2.5 Pro/Flash и др. |
| DeepSeek | deepseek-chat, deepseek-reasoner |
| + 30 других провайдеров | Baidu, Ali, Zhipu, MiniMax, Mistral и др. |

### 🚀 Расширенные возможности

**Поддерживаемые форматы API:**
- ⚡ OpenAI Responses API
- ⚡ OpenAI Realtime API (включая Azure)
- ⚡ Claude Messages
- ⚡ Google Gemini
- 🔄 Rerank Models (Cohere, Jina)

**Интеллектуальная маршрутизация:**
- ⚖️ Взвешенный случайный выбор канала
- 🔄 Автоматические повторные попытки при сбоях
- 🚦 Ограничение частоты запросов на уровне пользователя

**Преобразование форматов:**
- 🔄 OpenAI ⇄ Claude Messages
- 🔄 OpenAI → Google Gemini
- 🔄 Google Gemini → OpenAI

---

## 🚢 Развёртывание

> [!TIP]
> **Последний Docker-образ:** `calciumion/new-api:latest`

### 📋 Требования к развёртыванию

| Компонент | Требование |
|-----------|-----------|
| **Локальная БД** | SQLite (Docker обязательно монтирует директорию `/data`) |
| **Удалённая БД** | MySQL ≥ 5.7.8 или PostgreSQL ≥ 9.6 |
| **Контейнеризация** | Docker / Docker Compose |

### ⚙️ Ключевые переменные окружения

<details>
<summary>Основные переменные конфигурации</summary>

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `SESSION_SECRET` | Секрет сессии (обязателен для мультисерверного развёртывания) | — |
| `CRYPTO_SECRET` | Ключ шифрования (обязателен при использовании Redis) | — |
| `SQL_DSN` | Строка подключения к БД | — |
| `REDIS_CONN_STRING` | Строка подключения к Redis | — |
| `STREAMING_TIMEOUT` | Таймаут потоковой передачи (сек) | `300` |
| `TZ` | Временная зона | — |

📖 **Полный список:** [Документация по переменным окружения](https://docs.newapi.pro/en/docs/installation/config-maintenance/environment-variables)

</details>

### 🔧 Способы развёртывания

<details>
<summary><strong>Способ 1: Docker Compose (рекомендуется)</strong></summary>

```bash
git clone https://github.com/QuantumNous/new-api.git
cd new-api
nano docker-compose.yml
docker-compose up -d
```

</details>

<details>
<summary><strong>Способ 2: Docker команды</strong></summary>

**SQLite:**
```bash
docker run --name new-api -d --restart always \
  -p 3000:3000 \
  -e TZ=Europe/Moscow \
  -v ./data:/data \
  calciumion/new-api:latest
```

**MySQL:**
```bash
docker run --name new-api -d --restart always \
  -p 3000:3000 \
  -e SQL_DSN="root:password@tcp(localhost:3306)/oneapi" \
  -e TZ=Europe/Moscow \
  -v ./data:/data \
  calciumion/new-api:latest
```

</details>

### ⚠️ Особенности мультисерверного развёртывания

> [!WARNING]
> - **Обязательно установите** `SESSION_SECRET` — иначе состояние входа будет несогласованным
> - **При общем Redis установите** `CRYPTO_SECRET` — иначе данные не смогут быть расшифрованы

---

## 🔗 Связанные проекты

### Вышестоящие проекты

| Проект | Описание |
|--------|----------|
| [One API](https://github.com/songquanpeng/one-api) | Оригинальный проект-основа |
| [Midjourney-Proxy](https://github.com/novicezk/midjourney-proxy) | Поддержка интерфейса Midjourney |

### Вспомогательные инструменты

| Проект | Описание |
|--------|----------|
| [new-api-key-tool](https://github.com/Calcium-Ion/new-api-key-tool) | Инструмент запроса квоты ключей |
| [new-api-horizon](https://github.com/Calcium-Ion/new-api-horizon) | Высокопроизводительная оптимизированная версия New API |

---

## 💬 Поддержка

### 📖 Документация

| Ресурс | Ссылка |
|--------|--------|
| 📘 FAQ | [Часто задаваемые вопросы](https://docs.newapi.pro/en/docs/support/faq) |
| 💬 Сообщество | [Каналы общения](https://docs.newapi.pro/en/docs/support/community-interaction) |
| 🐛 Отчёты об ошибках | [Обратная связь](https://docs.newapi.pro/en/docs/support/feedback-issues) |
| 📚 Полная документация | [Официальная документация](https://docs.newapi.pro/en/docs) |

### 🤝 Руководство по участию

Приветствуются любые формы участия!

- 🐛 Сообщения об ошибках
- 💡 Предложение новых функций
- 📝 Улучшение документации
- 🔧 Отправка кода

---

## 🌟 История звёзд

<div align="center">

[![Star History Chart](https://api.star-history.com/svg?repos=Calcium-Ion/new-api&type=Date)](https://star-history.com/#Calcium-Ion/new-api&Date)

</div>

---

<div align="center">

### 💖 Спасибо за использование New API

Если проект оказался полезным, поставьте ⭐️ Star!

**[Официальная документация](https://docs.newapi.pro/en/docs)** • **[Обратная связь](https://github.com/Calcium-Ion/new-api/issues)** • **[Последние релизы](https://github.com/Calcium-Ion/new-api/releases)**

</div>
