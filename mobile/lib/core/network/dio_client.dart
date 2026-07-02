import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../auth/session_state.dart';
import '../config/app_config.dart';
import '../storage/token_storage.dart';
import 'auth_interceptor.dart';
import 'correlation_id_interceptor.dart';
import 'idempotency_interceptor.dart';

/// Unauthenticated client for `/v1/auth/*` — must not carry the
/// [AuthInterceptor] or a 401 during login/refresh would recurse into
/// another refresh attempt.
final authDioProvider = Provider<Dio>((ref) {
  return Dio(BaseOptions(baseUrl: AppConfig.current.apiBaseUrl))
    ..interceptors.addAll([
      CorrelationIdInterceptor(),
      IdempotencyInterceptor(),
    ]);
});

/// Client for all authenticated gateway calls. Attaches the bearer token and
/// transparently refreshes it on 401 via [authDioProvider].
final dioProvider = Provider<Dio>((ref) {
  final dio = Dio(BaseOptions(baseUrl: AppConfig.current.apiBaseUrl));
  final tokenStorage = ref.watch(tokenStorageProvider);
  final authDio = ref.watch(authDioProvider);

  dio.interceptors.addAll([
    CorrelationIdInterceptor(),
    IdempotencyInterceptor(),
    AuthInterceptor(
      tokenStorage: tokenStorage,
      dio: dio,
      refresh: (refreshToken) async {
        final response = await authDio.post<Map<String, dynamic>>(
          '/v1/auth/refresh',
          data: {'refresh_token': refreshToken},
        );
        final accessToken = response.data?['access_token'] as String?;
        final newRefreshToken = response.data?['refresh_token'] as String?;
        if (accessToken == null || newRefreshToken == null) return null;
        await tokenStorage.saveTokens(
          accessToken: accessToken,
          refreshToken: newRefreshToken,
        );
        return accessToken;
      },
      onSessionExpired: () async {
        await tokenStorage.clear();
        ref.read(sessionStatusProvider.notifier).state = SessionStatus.unauthenticated;
      },
    ),
  ]);

  return dio;
});
