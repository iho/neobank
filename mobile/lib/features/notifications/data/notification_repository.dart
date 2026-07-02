import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/dio_client.dart';
import '../../../core/network/error_mapper.dart';
import '../domain/notification_models.dart';

final notificationRepositoryProvider = Provider<NotificationRepository>((ref) {
  return NotificationRepository(ref.watch(dioProvider));
});

class NotificationRepository {
  NotificationRepository(this._dio);

  final Dio _dio;

  Future<NotificationPage> list({String? cursor, int limit = 20}) async {
    try {
      final response = await _dio.get<Map<String, dynamic>>(
        '/v1/notifications',
        queryParameters: {
          'limit': limit,
          'cursor': ?cursor,
        },
      );
      return NotificationPage.fromJson(response.data!);
    } on DioException catch (e) {
      throw mapDioException(e);
    }
  }

  Future<AppNotification> markRead(String id) async {
    try {
      final response = await _dio.post<Map<String, dynamic>>('/v1/notifications/$id/read');
      return AppNotification.fromJson(response.data!);
    } on DioException catch (e) {
      throw mapDioException(e);
    }
  }

  Future<int> markAllRead() async {
    try {
      final response = await _dio.post<Map<String, dynamic>>('/v1/notifications/read-all');
      return response.data!['marked_count'] as int;
    } on DioException catch (e) {
      throw mapDioException(e);
    }
  }
}
