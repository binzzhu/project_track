# 构建阶段
#FROM node:20-alpine AS builder
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/node:20-alpine AS builder

WORKDIR /app

# 设置npm镜像
RUN npm config set registry https://registry.npmmirror.com

# 复制依赖文件
COPY package*.json ./

# 安装依赖
RUN npm install

# 复制源码
COPY . .

# 构建应用
RUN npm run build

# 运行阶段
#FROM nginx:alpine
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/nginx:alpine

# 删除默认配置
RUN rm /etc/nginx/conf.d/default.conf

# 安装健康检查所需工具
RUN apk --no-cache add wget

# 复制nginx配置
COPY nginx.conf /etc/nginx/conf.d/

# 从构建阶段复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 暴露端口
EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:80/ || exit 1

# 启动nginx
CMD ["nginx", "-g", "daemon off;"]
