# 基础镜像
FROM docker.cnb.cool/dtapp/alpine:latest

# 拷贝文件到容器中
COPY ../../bin/reporter-linux-amd64 /app/reporter-linux-amd64

# 工作区
WORKDIR /app

# 设置权限
RUN chmod +x /app/reporter-linux-amd64
RUN mkdir -p /app/data/
RUN mkdir -p /app/logs/

# 设置时区为上海
ENV TZ=Asia/Shanghai

# 设置编码
ENV LANG C.UTF-8

# 使用 ARG 声明构建时参数
ARG DEBUG
ARG HTTP_URL
ARG BASIC_AUTH_USERNAME
ARG BASIC_AUTH_PASSWORD
ARG DEVICE_CODE
ARG DEVICE_NAME

# 使用 ENV 指令设置环境变量，并且可以利用上面声明的 ARG 的值（如果未传入则为空）
ENV DEBUG=${DEBUG}
ENV HTTP_URL=${HTTP_URL}
ENV BASIC_AUTH_USERNAME=${BASIC_AUTH_USERNAME}
ENV BASIC_AUTH_PASSWORD=${BASIC_AUTH_PASSWORD}
ENV DEVICE_CODE=${DEVICE_CODE}
ENV DEVICE_NAME=${DEVICE_NAME}

# 启动服务
ENTRYPOINT ["/app/reporter-linux-amd64"]