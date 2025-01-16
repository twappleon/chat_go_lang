import 'package:dio/dio.dart';

class TokenInterceptor extends Interceptor {
  final String token;

  TokenInterceptor(this.token);

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    // Add the token to the request headers
    options.headers['Authorization'] = 'Bearer $token';
    super.onRequest(options, handler);
  }
}

class ApiService {
  final Dio _dio;

  ApiService(this._dio);

  Future<Response> get(String url, {Map<String, dynamic>? queryParameters}) async {
    try {
      return await _dio.get(url, queryParameters: queryParameters);
    } catch (e) {
      throw Exception('GET请求失败: $e');
    }
  }

  Future<Response> post(String url, {Map<String, dynamic>? data}) async {
    try {
      return await _dio.post(url, data: data);
    } catch (e) {
      throw Exception('POST请求失败: $e');
    }
  }

  Future<Response> put(String url, {Map<String, dynamic>? data}) async {
    try {
      return await _dio.put(url, data: data);
    } catch (e) {
      throw Exception('PUT请求失败: $e');
    }
  }

  Future<Response> delete(String url, {Map<String, dynamic>? data}) async {
    try {
      return await _dio.delete(url, data: data);
    } catch (e) {
      throw Exception('DELETE请求失败: $e');
    }
  }

  // 可以根据需要添加更多方法，如 PUT、DELETE 等

  Future<ChatMessage> getChatHistory() async{
    final response = await _dio.get(chat_message);
    return ChatMessage.fromJson(response.data);
  }
}

class ApiPaths {
  static const String baseUrl = 'http://localhost:8888'; // 基本 URL
  static const String getUser = '$baseUrl/user'; // 獲取用戶
  static const String createUser = '$baseUrl/user/create'; // 創建用戶
  static const String updateUser = '$baseUrl/user/update'; // 更新用戶
  static const String deleteUser = '$baseUrl/user/delete'; // 刪除用戶
  // 可以根據需要添加更多 API 路徑

  static const String chat_message = '$baseUrl/api/chat/history?userId=$userId'
} 