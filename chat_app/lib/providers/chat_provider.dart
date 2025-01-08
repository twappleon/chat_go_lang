import 'package:flutter/foundation.dart';
import '../model/chat_message.dart';
import '../service/chat_service.dart';

class ChatProvider with ChangeNotifier {
  final ChatService _apiService = ChatService();
  List<ChatMessage> _messages = [];
  bool _isLoading = false;
  String? _error;

  List<ChatMessage> get messages => _messages;
  bool get isLoading => _isLoading;
  String? get error => _error;

  Future<void> loadChatHistory(String userId) async {
    try {
      _isLoading = true;
      _error = null;
      notifyListeners();

      _messages = await _apiService.getChatHistory(userId);
    } catch (e) {
      _error = e.toString();
    } finally {
      _isLoading = false;
      notifyListeners();
    }
  }

  // Future<void> sendMessage(ChatMessage message) async {
  //   try {
  //     await _apiService.sendMessage(message);
  //     _messages.add(message);
  //     notifyListeners();
  //   } catch (e) {
  //     _error = e.toString();
  //     notifyListeners();
  //   }
  // }
}