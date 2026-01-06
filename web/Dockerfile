FROM nginx:alpine

# 删除默认配置
RUN rm /etc/nginx/conf.d/default.conf

# 安装健康检查所需工具
RUN apk --no-cache add wget

# 复制nginx配置
COPY nginx.conf /etc/nginx/conf.d/

# 复制本地构建产物
COPY dist /usr/share/nginx/html

# 暴露端口
EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:80/ || exit 1

# 启动nginx
CMD ["nginx", "-g", "daemon off;"]
