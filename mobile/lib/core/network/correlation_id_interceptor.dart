import 'package:dio/dio.dart';
import 'package:uuid/uuid.dart';

/// Adds a fresh `X-Correlation-Id` to every outbound request so failures can
/// be traced end-to-end through gateway -> services -> outbox (see
/// `pkg/reqctx` on the backend).
class CorrelationIdInterceptor extends Interceptor {
  CorrelationIdInterceptor({Uuid? uuid}) : _uuid = uuid ?? const Uuid();

  final Uuid _uuid;

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    options.headers.putIfAbsent('X-Correlation-Id', () => _uuid.v4());
    handler.next(options);
  }
}
