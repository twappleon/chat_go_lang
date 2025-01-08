import 'package:flutter/material.dart';
import '../model/chat_message.dart';
import '../providers/chat_provider.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'dart:convert';

class ChatPage extends StatefulWidget {
  @override
  _ChatPageState createState() => _ChatPageState();
}

class _ChatPageState extends State<ChatPage> {
  final TextEditingController _messageController = TextEditingController();
  late WebSocketChannel _channel;
  late String _currentUserId;

  @override
  void initState() {
    super.initState();
    _currentUserId = 'user123'; // 實際應用中應從用戶認證獲取
    _initializeChat();
  }

  void _initializeChat() {
    _connectWebSocket();
    _loadChatHistory();
  }

  void _connectWebSocket() {
    _channel = WebSocketChannel.connect(
      Uri.parse('ws://localhost:8080/ws?userId=$_currentUserId'),
    );

    _channel.stream.listen(
          (message) {
        final chatMessage = ChatMessage.fromJson(jsonDecode(message));
        context.read<ChatProvider>().messages.add(chatMessage);
      },
      onError: (error) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('連接錯誤: $error')),
        );
      },
      onDone: () {
        // 可以在這裡實現重連邏輯
      },
    );
  }

  Future<void> _loadChatHistory() async {
    await context.read<ChatProvider>().loadChatHistory(_currentUserId);
  }

  void _sendMessage() {
    if (_messageController.text.isEmpty) return;

    final message = ChatMessage(
      id: DateTime.now().toString(), // 實際應用中應由服務器生成
      type: 'chat',
      sender: _currentUserId,
      receiver: 'receiver123', // 實際應用中應為實際接收者ID
      message: _messageController.text,
      timestamp: DateTime.now(),
    );

    context.read<ChatProvider>().sendMessage(message);
    _messageController.clear();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('聊天'),
      ),
      body: Consumer<ChatProvider>(
        builder: (context, chatProvider, child) {
          if (chatProvider.isLoading) {
            return Center(child: CircularProgressIndicator());
          }

          if (chatProvider.error != null) {
            return Center(child: Text(chatProvider.error!));
          }

          return Column(
            children: [
              Expanded(
                child: ListView.builder(
                  reverse: true,
                  itemCount: chatProvider.messages.length,
                  itemBuilder: (context, index) {
                    final message = chatProvider.messages[index];
                    return MessageBubble(
                      message: message,
                      isMe: message.sender == _currentUserId,
                    );
                  },
                ),
              ),
              _buildMessageInput(),
            ],
          );
        },
      ),
    );
  }

  Widget _buildMessageInput() {
    return Padding(
      padding: const EdgeInsets.all(8.0),
      child: Row(
        children: [
          Expanded(
            child: TextField(
              controller: _messageController,
              decoration: InputDecoration(
                hintText: '輸入訊息...',
                border: OutlineInputBorder(),
              ),
            ),
          ),
          SizedBox(width: 8),
          IconButton(
            icon: Icon(Icons.send),
            onPressed: _sendMessage,
          ),
        ],
      ),
    );
  }

  @override
  void dispose() {
    _messageController.dispose();
    _channel.sink.close();
    super.dispose();
  }
}

class MessageBubble extends StatelessWidget {
  final ChatMessage message;
  final bool isMe;

  const MessageBubble({
    Key? key,
    required this.message,
    required this.isMe,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Align(
      alignment: isMe ? Alignment.centerRight : Alignment.centerLeft,
      child: Container(
        margin: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
        padding: EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: isMe ? Colors.blue : Colors.grey[300],
          borderRadius: BorderRadius.circular(12),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              message.message,
              style: TextStyle(
                color: isMe ? Colors.white : Colors.black,
              ),
            ),
            Text(
              message.timestamp.toLocal().toString().substring(11, 16),
              style: TextStyle(
                fontSize: 12,
                color: isMe ? Colors.white70 : Colors.black54,
              ),
            ),
          ],
        ),
      ),
    );
  }
}