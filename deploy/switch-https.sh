#!/bin/bash
# ============================================================
# umniai.ru — 切换 Nginx 到 HTTPS 配置
# 在 SSL 证书获取后执行
# ============================================================

set -euo pipefail

COMPOSE_FILE="docker-compose.prod.yml"

echo "切换 Nginx 到 HTTPS 配置..."
# Nginx 的 conf.d 目录中，active.conf 会被加载
# 初始时使用 umniai.ru.initial.conf，获取证书后切换到 umniai.ru.conf

# 检查证书是否存在
CERT_DIR=$(docker compose -f "${COMPOSE_FILE}" exec -T nginx ls /etc/letsencrypt/live/umniai.ru/ 2>/dev/null || echo "")
if [ -z "${CERT_DIR}" ]; then
    echo "❌ 未找到 SSL 证书，请先运行 init-ssl.sh"
    exit 1
fi

echo "✅ 检测到 SSL 证书"
echo "重新加载 Nginx 配置..."
docker compose -f "${COMPOSE_FILE}" exec -T nginx nginx -s reload
echo "✅ Nginx 已重新加载 HTTPS 配置"
