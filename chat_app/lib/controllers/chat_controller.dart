import 'package:get/get.dart';
import 'package:chat_app/service/chat_service.dart';
import '../model/chat_message.dart';

class ChatController extends GetxController {
  var messages = <ChatMessage>[].obs; // 使用 RxList 进行响应式管理
  final ChatService chatService = ChatService();
  late String currentUserId;

  @override
  void onInit() {
    super.onInit();
    currentUserId = 'user123'; // 實際應用中應該從登錄狀態獲取
    _loadChatHistory();
  }

  Future<void> _loadChatHistory() async {
    try {
      final history = await chatService.getChatHistory(currentUserId);
      messages.addAll(history);
    } catch (e) {
      print('Error loading chat history: $e');
    }
  }

  void addMessage(ChatMessage message) {
    messages.add(message);
  }

  void removeMessage(int index) {
    messages.removeAt(index);
  }
}
