import 'package:dio/dio.dart';

import '../error/api_exception.dart';

/// Maps a transport-level [DioException] to the typed [ApiException] the UI
/// layer understands. Repositories should be the only place this is called.
ApiException mapDioException(DioException e) {
  switch (e.type) {
    case DioExceptionType.connectionError:
      return ApiException.network();
    case DioExceptionType.connectionTimeout:
    case DioExceptionType.sendTimeout:
    case DioExceptionType.receiveTimeout:
      return ApiException.timeout();
    default:
      break;
  }

  final response = e.response;
  if (response == null) return ApiException.network();
  if (response.statusCode == 401) return ApiException.unauthenticated();

  var message = 'Something went wrong. Please try again.';
  final data = response.data;
  if (data is Map && data['error'] is String) {
    message = data['error'] as String;
  }

  return ApiException(
    message: message,
    statusCode: response.statusCode,
    correlationId: response.headers.value('x-correlation-id'),
  );
}
