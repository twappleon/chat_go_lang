#!/bin/bash

# 構建和啟動容器
docker-compose up --build -d

# 等待服務啟動
echo "Waiting for services to start..."
sleep 5

# 顯示容器狀態
docker-compose ps

# 顯示日誌
echo "Showing logs..."
docker-compose logs -f