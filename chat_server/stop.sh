#!/bin/bash

# 停止並移除容器
docker-compose down

# 可選：如果要清除數據卷，取消下面的註釋
# docker-compose down -v