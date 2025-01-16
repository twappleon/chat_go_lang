import 'package:dio/dio.dart';

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

  // 可以根据需要添加更多方法，如 PUT、DELETE 等
} 