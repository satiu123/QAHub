# 使用官方的 Nginx 镜像作为基础
FROM nginx:1.25-alpine

# 删除默认的 Nginx 配置文件
RUN rm /etc/nginx/conf.d/default.conf

# 将我们自定义的配置文件复制到容器中
COPY nginx.conf /etc/nginx/conf.d/
