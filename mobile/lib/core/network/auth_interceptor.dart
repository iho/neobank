import 'package:dio/dio.dart';

import '../storage/token_storage.dart';

/// Injects the bearer token on every request and transparently refreshes it
/// on a 401, retrying the original request exactly once.
///
/// Refresh is single-flighted: concurrent 401s share one refresh call so we
/// never fire two refresh requests (which would race on the backend's
/// refresh-token rotation).
class AuthInterceptor extends Interceptor {
  AuthInterceptor({
    required this.tokenStorage,
    required this.dio,
    required this.refresh,
    required this.onSessionExpired,
  });

  final TokenStorage tokenStorage;

  /// The `Dio` instance to use for retrying the original request after a
  /// successful refresh. Must be the same instance this interceptor is
  /// attached to.
  final Dio dio;

  /// Calls `POST /v1/auth/refresh` and persists the new tokens. Returns the
  /// new access token, or `null` if the refresh token is invalid/expired.
  final Future<String?> Function(String refreshToken) refresh;

  /// Called when refresh fails or no refresh token exists — the caller
  /// should route to the login screen.
  final void Function() onSessionExpired;

  Future<String?>? _refreshFuture;

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) async {
    if (!options.headers.containsKey('Authorization')) {
      final token = await tokenStorage.readAccessToken();
      if (token != null) {
        options.headers['Authorization'] = 'Bearer $token';
      }
    }
    handler.next(options);
  }

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) async {
    final isAuthEndpoint = err.requestOptions.path.contains('/v1/auth/');
    if (err.response?.statusCode != 401 ||
        isAuthEndpoint ||
        err.requestOptions.extra['retried_after_refresh'] == true) {
      handler.next(err);
      return;
    }

    final newToken = await _refreshAccessToken();
    if (newToken == null) {
      onSessionExpired();
      handler.next(err);
      return;
    }

    final retryOptions = err.requestOptions
      ..headers['Authorization'] = 'Bearer $newToken'
      ..extra['retried_after_refresh'] = true;

    try {
      final response = await dio.fetch<dynamic>(retryOptions);
      handler.resolve(response);
    } on DioException catch (retryError) {
      handler.next(retryError);
    }
  }

  Future<String?> _refreshAccessToken() {
    return _refreshFuture ??= _doRefresh().whenComplete(() {
      _refreshFuture = null;
    });
  }

  Future<String?> _doRefresh() async {
    final refreshToken = await tokenStorage.readRefreshToken();
    if (refreshToken == null) return null;
    try {
      return await refresh(refreshToken);
    } catch (_) {
      await tokenStorage.clear();
      return null;
    }
  }
}
