import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/auth/session_state.dart';
import '../../../core/storage/token_storage.dart';
import '../../cards/presentation/authorizations_controller.dart';
import '../../cards/presentation/cards_controller.dart';
import '../../notifications/presentation/notifications_controller.dart';
import '../../onboarding_kyc/presentation/kyc_controller.dart';
import '../../wallet/presentation/wallet_home_controller.dart';
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

  /// These controllers cache data for whichever account was signed in when
  /// they were first built — plain [AsyncNotifierProvider]s, not
  /// `.autoDispose`, so Riverpod has no reason to refetch on its own after a
  /// login/logout swaps the active account. Without this, switching accounts
  /// on the same app install (or in an emulator, which persists the app's
  /// storage across sessions) shows the *previous* user's wallet/cards/etc.
  void _invalidateUserScopedProviders() {
    ref.invalidate(walletHomeControllerProvider);
    ref.invalidate(kycControllerProvider);
    ref.invalidate(cardsControllerProvider);
    ref.invalidate(authorizationsControllerProvider);
    ref.invalidate(notificationsControllerProvider);
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
    _invalidateUserScopedProviders();
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
    _invalidateUserScopedProviders();
    ref.read(sessionStatusProvider.notifier).state = SessionStatus.authenticated;
  }

  Future<void> logout() async {
    await ref.read(tokenStorageProvider).clear();
    _invalidateUserScopedProviders();
    ref.read(sessionStatusProvider.notifier).state = SessionStatus.unauthenticated;
  }
}
