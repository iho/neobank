import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/auth/session_state.dart';
import '../../../core/storage/token_storage.dart';
import '../data/auth_repository.dart';

final authControllerProvider = AsyncNotifierProvider<AuthController, void>(AuthController.new);

/// Owns the auth *actions* (login/register/logout); session *status* itself
/// lives in [sessionStatusProvider] so the router and network layer can read
/// it without depending on this feature. A failed login/register throws
/// straight to the calling screen rather than putting this notifier into an
/// error state — one bad password shouldn't affect anything else that reads
/// this provider.
class AuthController extends AsyncNotifier<void> {
  @override
  Future<void> build() async {
    final token = await ref.read(tokenStorageProvider).readAccessToken();
    ref.read(sessionStatusProvider.notifier).state =
        token == null ? SessionStatus.unauthenticated : SessionStatus.authenticated;
  }

  Future<void> login({required String email, required String password}) async {
    final tokens = await ref.read(authRepositoryProvider).login(
          email: email,
          password: password,
        );
    await ref.read(tokenStorageProvider).saveTokens(
          accessToken: tokens.accessToken,
          refreshToken: tokens.refreshToken,
        );
    ref.read(sessionStatusProvider.notifier).state = SessionStatus.authenticated;
  }

  Future<void> register({
    required String email,
    required String password,
    String? phone,
    String? inviteCode,
  }) async {
    final tokens = await ref.read(authRepositoryProvider).register(
          email: email,
          password: password,
          phone: phone,
          inviteCode: inviteCode,
        );
    await ref.read(tokenStorageProvider).saveTokens(
          accessToken: tokens.accessToken,
          refreshToken: tokens.refreshToken,
        );
    ref.read(sessionStatusProvider.notifier).state = SessionStatus.authenticated;
  }

  Future<void> logout() async {
    await ref.read(tokenStorageProvider).clear();
    ref.read(sessionStatusProvider.notifier).state = SessionStatus.unauthenticated;
  }
}
