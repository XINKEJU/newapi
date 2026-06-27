#!/bin/bash
# ============================================================
# umniai.ru — 首次 SSL 证书获取脚本
# 在服务器上执行此脚本以获取 Let's Encrypt SSL 证书
#
# 前提条件：
#   1. 域名 umniai.ru 的 A 记录已指向 170.168.89.127
#   2. 服务器已安装 Docker 和 Docker Compose
#   3. 已修改 docker-compose.prod.yml 中的所有密码
#
# 用法：
#   bash deploy/init-ssl.sh your-email@example.com
# ============================================================

set -euo pipefail

DOMAIN="umniai.ru"
EMAIL="${1:-admin@umniai.ru}"
COMPOSE_FILE="docker-compose.prod.yml"

echo "=========================================="
echo "  umniai.ru SSL 证书初始化"
echo "  域名: ${DOMAIN}, www.${DOMAIN}"
echo "  邮箱: ${EMAIL}"
echo "=========================================="

# 1. 使用初始 HTTP-only 配置启动 Nginx
echo ""
echo "[1/4] 启动 Nginx (HTTP-only 模式)..."
cp nginx/umniai.ru.initial.conf nginx/umniai.ru.active.conf
docker compose -f "${COMPOSE_FILE}" up -d nginx new-api redis postgres
echo "等待服务启动..."
sleep 5

# 2. 验证 HTTP 访问
echo ""
echo "[2/4] 验证 HTTP 访问..."
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "http://${DOMAIN}/api/status" || echo "000")
if [ "${HTTP_STATUS}" != "200" ]; then
    echo "❌ HTTP 访问失败 (status: ${HTTP_STATUS})"
    echo "请检查域名 DNS 是否已指向本服务器，以及服务是否正常运行"
    exit 1
fi
echo "✅ HTTP 访问正常"

# 3. 获取 SSL 证书
echo ""
echo "[3/4] 获取 Let's Encrypt SSL 证书..."
docker compose -f "${COMPOSE_FILE}" --profile certbot run --rm certbot \
    certonly --webroot -w /var/www/certbot \
    -d "${DOMAIN}" -d "www.${DOMAIN}" \
    --email "${EMAIL}" \
    --agree-tos \
    --no-eff-email

# 4. 切换到 HTTPS 配置
echo ""
echo "[4/4] 切换到 HTTPS 配置..."
cp nginx/umniai.ru.conf nginx/umniai.ru.active.conf
docker compose -f "${COMPOSE_FILE}" restart nginx
sleep 2

# 验证 HTTPS
HTTPS_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "https://${DOMAIN}/api/status" || echo "000")
if [ "${HTTPS_STATUS}" = "200" ]; then
    echo "✅ HTTPS 访问正常"
else
    echo "⚠️  HTTPS 访问异常 (status: ${HTTPS_STATUS})，请检查证书配置"
fi

echo ""
echo "=========================================="
echo "  SSL 证书初始化完成！"
echo "  访问地址: https://${DOMAIN}"
echo ""
echo "  下一步："
echo "  1. 登录管理后台设置 ServerAddress 为 https://${DOMAIN}"
echo "  2. 配置支付回调地址等"
echo "=========================================="

# 清理初始配置文件
rm -f nginx/umniai.ru.active.conf
