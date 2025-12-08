FROM openjdk:21-slim

LABEL maintainer="CozyInsight Team"
LABEL description="Apache Calcite Avatica Server for CozyInsight"

WORKDIR /app

# 创建必要的目录
RUN mkdir -p /app/config /app/drivers /app/logs

# 复制 Avatica Server JAR (需要手动下载放置)
# 下载地址: https://repo1.maven.org/maven2/org/apache/calcite/avatica/avatica-server/
COPY avatica-server.jar /app/

# 复制配置文件
COPY application.yml /app/config/

# 复制数据库驱动 (JDBC drivers)
# COPY drivers/*.jar /app/drivers/

# 设置环境变量
ENV JAVA_OPTS="-Xmx2g -Xms1g -XX:+UseG1GC -XX:MaxGCPauseMillis=200"

# 暴露端口
EXPOSE 8765

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --retries=3 --start-period=40s \
  CMD curl -f http://localhost:8765/ || exit 1

# 启动命令
CMD java $JAVA_OPTS \
    -jar avatica-server.jar \
    --spring.config.location=file:/app/config/application.yml
