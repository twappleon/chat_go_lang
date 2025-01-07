import 'dart:convert';
import 'package:http/http.dart' as http;
import '../model/chat_message.dart';

class ChatService {
  static const String baseUrl = 'http://localhost:8080';

  Future<List<ChatMessage>> getChatHistory(String userId) async {
    final response = await http.get(
      Uri.parse('$baseUrl/api/chat/history?userId=$userId'),
    );

    if (response.statusCode == 200) {
      final List<dynamic> data = json.decode(response.body);
      return data.map((json) => ChatMessage.fromJson(json)).toList();
    } else {
      throw Exception('Failed to load chat history');
    }
  }
}