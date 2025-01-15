import 'package:chat_app/service/chat_service.dart';
import 'package:flutter/material.dart';
import 'package:flutter_webrtc/flutter_webrtc.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'dart:convert';
import '../model/chat_message.dart';
import 'package:google_mobile_ads/google_mobile_ads.dart';
import 'package:get/get.dart';
import 'chat_controller.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  MobileAds.instance.initialize();
  runApp(ChatApp());
}

class ChatApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'P2P Chat & Video',
      theme: ThemeData(
        primarySwatch: Colors.blue,
      ),
      home: ChatPage(),
    );
  }
}

class ChatPage extends StatefulWidget {
  @override
  _ChatPageState createState() => _ChatPageState();
}

class _ChatPageState extends State<ChatPage> {
  final ChatController chatController = Get.put(ChatController());
  BannerAd? _bannerAd;

  @override
  void initState() {
    super.initState();
    _loadBannerAd(); // 加载广告
  }

  void _loadBannerAd() {
    _bannerAd = BannerAd(
      adUnitId: 'YOUR_AD_UNIT_ID', // 替换为您的广告单元 ID
      request: AdRequest(),
      listener: BannerAdListener(
        onAdLoaded: (_) {
          setState(() {}); // 广告加载后更新状态
        },
        onAdFailedToLoad: (ad, error) {
          ad.dispose(); // 处理广告加载失败
        },
      ),
    )..load();
  }

  @override
  void dispose() {
    _bannerAd?.dispose(); // 释放广告资源
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('P2P Chat & Video'),
      ),
      body: Column(
        children: [
          Expanded(
            flex: 3,
            child: Obx(() => ListView.builder(
              itemCount: chatController.messages.length,
              itemBuilder: (context, index) {
                return ListTile(
                  title: Text(chatController.messages[index].message),
                );
              },
            )),
          ),
          if (_bannerAd != null) // 显示广告
            Container(
              height: 50,
              child: AdWidget(ad: _bannerAd!),
            ),
          Padding(
            padding: EdgeInsets.all(8.0),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    onSubmitted: (value) {
                      if (value.isNotEmpty) {
                        final message = ChatMessage(message: value);
                        chatController.addMessage(message);
                      }
                    },
                    decoration: InputDecoration(
                      labelText: '輸入訊息',
                    ),
                  ),
                ),
                IconButton(
                  icon: Icon(Icons.send),
                  onPressed: () {
                    // 发送消息逻辑
                  },
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}