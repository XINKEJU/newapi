#!/bin/bash
# ============================================================
# umniai.ru — SSL 证书自动续期脚本
# Let's Encrypt 证书有效期 90 天，建议每 60 天续期一次
#
# 添加到 crontab:
#   0 3 1 * * cd /path/to/newapi && bash deploy/renew-ssl.sh
# ============================================================

set -euo pipefail

COMPOSE_FILE="docker-compose.prod.yml"

echo "[$(date)] 开始 SSL 证书续期检查..."

# 尝试续期
docker compose -f "${COMPOSE_FILE}" --profile certbot run --rm certbot \
    renew --quiet

# 重新加载 Nginx 以使用新证书
echo "重新加载 Nginx..."
docker compose -f "${COMPOSE_FILE}" exec -T nginx nginx -s reload

echo "[$(date)] SSL 证书续期完成"
