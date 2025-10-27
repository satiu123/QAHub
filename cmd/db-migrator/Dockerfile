# 使用官方的 migrate/migrate 镜像作为基础
FROM migrate/migrate:v4.17.0

# 将本地迁移脚本复制到镜像中的 /migrations 目录
# 这是 migrate CLI 默认会查找脚本的路径之一
COPY scripts/migrations/all /migrations