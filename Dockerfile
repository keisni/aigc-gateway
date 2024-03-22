FROM registry.cn-hangzhou.aliyuncs.com/acs/aigc-gateway:v1.1.0

WORKDIR /app
COPY ./bin/aigc-gateway .

# Run
CMD ["./aigc-gateway"]