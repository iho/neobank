import 'package:dio/dio.dart';
import 'package:uuid/uuid.dart';

const _mutatingMethods = {'POST', 'PUT', 'PATCH', 'DELETE'};

/// Adds `Idempotency-Key` to every mutating request that doesn't already
/// specify one. The gateway 400s mutations without this header.
///
/// Callers that need to retry the *same* logical operation (e.g. a transfer
/// after a client timeout) must set the header explicitly on the request
/// options so the key is reused rather than regenerated here.
class IdempotencyInterceptor extends Interceptor {
  IdempotencyInterceptor({Uuid? uuid}) : _uuid = uuid ?? const Uuid();

  final Uuid _uuid;

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    if (_mutatingMethods.contains(options.method.toUpperCase()) &&
        !options.headers.containsKey('Idempotency-Key')) {
      options.headers['Idempotency-Key'] = _uuid.v4();
    }
    handler.next(options);
  }
}
