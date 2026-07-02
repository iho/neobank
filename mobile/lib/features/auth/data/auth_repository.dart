import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/dio_client.dart';
import '../../../core/network/error_mapper.dart';
import '../domain/auth_models.dart';

final authRepositoryProvider = Provider<AuthRepository>((ref) {
  return AuthRepository(
    authDio: ref.watch(authDioProvider),
    dio: ref.watch(dioProvider),
  );
});

class AuthRepository {
  AuthRepository({required this.authDio, required this.dio});

  /// Unauthenticated client — login/register never carry a bearer token.
  final Dio authDio;

  /// Authenticated client — for calls that need the current session
  /// (profile, change password).
  final Dio dio;

  Future<AuthTokens> login({
    required String email,
    required String password,
  }) async {
    try {
      final response = await authDio.post<Map<String, dynamic>>(
        '/v1/auth/login',
        data: {'email': email, 'password': password},
      );
      return AuthTokens.fromJson(response.data!);
    } on DioException catch (e) {
      throw mapDioException(e);
    }
  }

  Future<AuthTokens> register({
    required String email,
    required String password,
    String? phone,
    String? inviteCode,
  }) async {
    try {
      final response = await authDio.post<Map<String, dynamic>>(
        '/v1/auth/register',
        data: {
          'email': email,
          'password': password,
          if (phone != null && phone.isNotEmpty) 'phone': phone,
          if (inviteCode != null && inviteCode.isNotEmpty) 'invite_code': inviteCode,
        },
      );
      return AuthTokens.fromJson(response.data!);
    } on DioException catch (e) {
      throw mapDioException(e);
    }
  }

  Future<Profile> getProfile() async {
    try {
      final response = await dio.get<Map<String, dynamic>>('/v1/me');
      return Profile.fromJson(response.data!);
    } on DioException catch (e) {
      throw mapDioException(e);
    }
  }

  Future<void> changePassword({
    required String currentPassword,
    required String newPassword,
  }) async {
    try {
      await dio.post<void>(
        '/v1/auth/change-password',
        data: {
          'current_password': currentPassword,
          'new_password': newPassword,
        },
      );
    } on DioException catch (e) {
      throw mapDioException(e);
    }
  }
}
