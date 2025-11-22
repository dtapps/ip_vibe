# 基础镜像
FROM docker.cnb.cool/dtapp/alpine:latest

# 拷贝文件到容器中
COPY ../../bin/backend-linux-amd64  /app/backend-linux-amd64 

# 工作区
WORKDIR /app

# 设置权限
RUN chmod +x /app/backend-linux-amd64 
RUN mkdir -p /app/data/
RUN mkdir -p /app/logs/

# 设置时区为上海
ENV TZ=Asia/Shanghai

# 设置编码
ENV LANG C.UTF-8

# 使用 ARG 声明构建时参数
ARG DEBUG
ARG HTTP_PORT
ARG BASIC_AUTH_USERNAME
ARG BASIC_AUTH_PASSWORD

# 使用 ENV 指令设置环境变量，并且可以利用上面声明的 ARG 的值（如果未传入则为空）
ENV DEBUG=${DEBUG}
ENV HTTP_PORT=${HTTP_PORT}
ENV BASIC_AUTH_USERNAME=${BASIC_AUTH_USERNAME}
ENV BASIC_AUTH_PASSWORD=${BASIC_AUTH_PASSWORD}

# 启动服务
ENTRYPOINT ["/app/backend-linux-amd64 "]