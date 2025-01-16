import 'package:chat_app/service/chat_service.dart';
import 'package:flutter/material.dart';
import 'package:flutter_webrtc/flutter_webrtc.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'dart:convert';
import './model/chat_message.dart';
import 'package:google_mobile_ads/google_mobile_ads.dart';
import 'package:get/get.dart';
import './controllers/chat_controller.dart';
import 'package:dio/dio.dart';
import './api_service.dart';

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
  final Dio _dio = Dio();
  late ApiService _apiService;

  @override
  void initState() {
    super.initState();
    _apiService = ApiService(_dio);
    _loadBannerAd();
    _fetchData();
  }

  void _fetchData() async {
    try {
      final response = await _dio.get('YOUR_API_ENDPOINT');
      print(response.data);
    } catch (e) {
      print('请求失败: $e');
    }
  }

  void _loadBannerAd() {
    _bannerAd = BannerAd(
      adUnitId: 'YOUR_AD_UNIT_ID',
      request: AdRequest(),
      size: AdSize.banner,
      listener: BannerAdListener(
        onAdLoaded: (_) {
          setState(() {});
        },
        onAdFailedToLoad: (ad, error) {
          ad.dispose();
        },
      ),
    )..load();
  }

  @override
  void dispose() {
    _bannerAd?.dispose();
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
          if (_bannerAd != null)
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
                        final message = ChatMessage(message: value, id: '1', type: 'chat', sender: 'user123', receiver: 'user456', timestamp: DateTime.now());
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