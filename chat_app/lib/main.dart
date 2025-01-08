import 'package:chat_app/service/chat_service.dart';
import 'package:flutter/material.dart';
import 'package:flutter_webrtc/flutter_webrtc.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'dart:convert';
import '../model/chat_message.dart';


void main() {
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
  final TextEditingController _controller = TextEditingController();
  late WebSocketChannel channel;
  final List<ChatMessage> messages = [];
  final ChatService chatService = ChatService();
  late String currentUserId;

  // WebRTC 相關變量
  RTCPeerConnection? _peerConnection;
  MediaStream? _localStream;
  MediaStream? _remoteStream;
  final _localRenderer = RTCVideoRenderer();
  final _remoteRenderer = RTCVideoRenderer();
  bool _inCalling = false;

  @override
  void initState() {
    super.initState();
    _initWebRTC();
    currentUserId = 'user123'; // 實際應用中應該從登錄狀態獲取
    _loadChatHistory();
    _connectWebSocket();
  }

  Future<void> _loadChatHistory() async {
    try {
      final history = await chatService.getChatHistory(currentUserId);
      setState(() {
        messages.addAll(history);
      });
    } catch (e) {
      print('Error loading chat history: $e');
    }
  }

  void _connectWebSocket() {    
    channel = WebSocketChannel.connect(
      Uri.parse('ws://localhost:8888/ws'),
    );
    channel.stream.listen((message) {
      final data = jsonDecode(message);
      if (data['type'] == 'chat') {
        final chatMessage = ChatMessage.fromJson(data);
        setState(() {
          messages.add(chatMessage);
        });
      } else if (data['type'] == 'webrtc') {
        _handleWebRTCMessage(data);
      }
    });
  }

  Future<void> _initWebRTC() async {
    await _localRenderer.initialize();
    await _remoteRenderer.initialize();

    final Map<String, dynamic> configuration = {
      "iceServers": [
        {"url": "stun:stun.l.google.com:19302"},
      ]
    };

    _peerConnection = await createPeerConnection(configuration);

    _peerConnection?.onIceCandidate = (RTCIceCandidate candidate) {
      channel.sink.add(jsonEncode({
        'type': 'webrtc',
        'action': 'ice_candidate',
        'candidate': candidate.toMap(),
      }));
    };

    _peerConnection?.onTrack = (RTCTrackEvent event) {
      if (event.streams.isNotEmpty) {
        setState(() {
          _remoteStream = event.streams[0];
          _remoteRenderer.srcObject = _remoteStream;
        });
      }
    };
  }

  Future<void> _startCall() async {
    final Map<String, dynamic> mediaConstraints = {
      'audio': true,
      'video': {
        'facingMode': 'user',
      }
    };

    try {
      _localStream = await navigator.mediaDevices.getUserMedia(mediaConstraints);
      _localRenderer.srcObject = _localStream;

      _localStream?.getTracks().forEach((track) {
        _peerConnection?.addTrack(track, _localStream!);
      });

      RTCSessionDescription offer = await _peerConnection!.createOffer();
      await _peerConnection!.setLocalDescription(offer);

      channel.sink.add(jsonEncode({
        'type': 'webrtc',
        'action': 'offer',
        'sdp': offer.toMap(),
      }));

      setState(() {
        _inCalling = true;
      });
    } catch (e) {
      print(e.toString());
    }
  }

  Future<void> _handleWebRTCMessage(Map<String, dynamic> data) async {
    switch (data['action']) {
      case 'offer':
        await _peerConnection?.setRemoteDescription(
          RTCSessionDescription(
            data['sdp']['sdp'],
            data['sdp']['type'],
          ),
        );
        RTCSessionDescription answer = await _peerConnection!.createAnswer();
        await _peerConnection?.setLocalDescription(answer);
        
        channel.sink.add(jsonEncode({
          'type': 'webrtc',
          'action': 'answer',
          'sdp': answer.toMap(),
        }));
        break;

      case 'answer':
        await _peerConnection?.setRemoteDescription(
          RTCSessionDescription(
            data['sdp']['sdp'],
            data['sdp']['type'],
          ),
        );
        break;

      case 'ice_candidate':
        await _peerConnection?.addCandidate(
          RTCIceCandidate(
            data['candidate']['candidate'],
            data['candidate']['sdpMid'],
            data['candidate']['sdpMLineIndex'],
          ),
        );
        break;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('P2P Chat & Video'),
        actions: [
          IconButton(
            icon: Icon(_inCalling ? Icons.call_end : Icons.video_call),
            onPressed: _inCalling ? _endCall : _startCall,
          ),
        ],
      ),
      body: Column(
        children: [
          if (_inCalling) Expanded(
            flex: 2,
            child: Row(
              children: [
                Expanded(
                  child: RTCVideoView(_localRenderer),
                ),
                Expanded(
                  child: RTCVideoView(_remoteRenderer),
                ),
              ],
            ),
          ),
          Expanded(
            flex: 3,
            child: ListView.builder(
              itemCount: messages.length,
              itemBuilder: (context, index) {
                return ListTile(
                  title: Text(messages[index].message),
                );
              },
            ),
          ),
          Padding(
            padding: EdgeInsets.all(8.0),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _controller,
                    decoration: InputDecoration(
                      labelText: '輸入訊息',
                    ),
                  ),
                ),
                IconButton(
                  icon: Icon(Icons.send),
                  onPressed: _sendMessage,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  void _sendMessage() {
    if (_controller.text.isNotEmpty) {
      final message = {
        'type': 'chat',
        'sender': this.currentUserId,
        'message': _controller.text,
      };
      channel.sink.add(jsonEncode(message));
      _controller.clear();
    }
  }

  Future<void> _endCall() async {
    try {
      await _localStream?.dispose();
      await _peerConnection?.close();
      _localRenderer.srcObject = null;
      _remoteRenderer.srcObject = null;
      setState(() {
        _inCalling = false;
      });
    } catch (e) {
      print(e.toString());
    }
  }

  @override
  void dispose() {
    _endCall();
    channel.sink.close();
    _controller.dispose();
    _localRenderer.dispose();
    _remoteRenderer.dispose();
    super.dispose();
  }
}