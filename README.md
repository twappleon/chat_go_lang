# P2P 即時通訊應用

一個基於 WebSocket 的即時通訊應用，支持一對一聊天、視訊通話和照片配對功能。

## 功能特點

- 📱 即時文字通訊
- 🎥 WebRTC 視訊通話
- 👥 照片配對功能
- 💾 MongoDB 訊息儲存
- 🔄 歷史訊息同步
- 📱 響應式設計

## 技術棧

### 後端
- Golang
- WebSocket
- MongoDB
- Docker

### 前端
- Flutter
- Dio
- Provider
- WebRTC

## 系統要求

- Docker 20.10.0 或更高
- Docker Compose 2.0.0 或更高
- Flutter SDK 3.0.0 或更高
- Android Studio / VS Code
- MongoDB 4.4 或更高

## 快速開始

### 後端部署

1. 克隆倉庫：
```bash
git clone https://github.com/yourusername/p2p-chat-app.git
cd p2p-chat-app
```
2. 啟動 Docker 服務：
```
chmod +x start.sh
./start.sh
```

服務將在以下端口啟動：

##### 後端 API: http://localhost:8080
##### MongoDB: http://localhost:27017
##### Mongo Express: http://localhost:8081

# 前端開發
安裝依賴：
bash

複製
cd frontend
flutter pub get
運行應用：
bash

複製
flutter run

### 項目結構
```
p2p_chat_app/
├── backend/
│   ├── config/
│   │   └── config.go
│   ├── database/
│   │   └── mongodb.go
│   ├── models/
│   │   └── message.go
│   ├── Dockerfile
│   ├── main.go
│   └── go.mod
├── frontend/
│   ├── lib/
│   │   ├── config/
│   │   ├── models/
│   │   ├── pages/
│   │   ├── providers/
│   │   ├── services/
│   │   └── widgets/
│   └── pubspec.yaml
├── docker-compose.yml
├── start.sh
├── stop.sh
└── README.md
```
