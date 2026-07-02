/// Typed failure mapped from the gateway's error envelope (or transport
/// failures that never reach the gateway). Presentation layers should catch
/// this and never show raw Dio/network errors to the user.
class ApiException implements Exception {
  const ApiException({
    required this.message,
    this.statusCode,
    this.code,
    this.correlationId,
  });

  factory ApiException.network() => const ApiException(
        message: 'Could not reach the server. Check your connection.',
      );

  factory ApiException.timeout() => const ApiException(
        message: 'The request timed out. Please try again.',
      );

  factory ApiException.unauthenticated() => const ApiException(
        message: 'Your session has expired. Please log in again.',
        statusCode: 401,
      );

  /// Human-readable message safe to show to the user.
  final String message;

  /// HTTP status code, if the request reached the gateway.
  final int? statusCode;

  /// Gateway error code (from the error envelope body), if present.
  final String? code;

  /// `X-Correlation-Id` echoed back, for support/debugging.
  final String? correlationId;

  @override
  String toString() => 'ApiException($statusCode, $code, $message)';
}
