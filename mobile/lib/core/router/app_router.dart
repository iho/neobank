import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../features/auth/presentation/auth_controller.dart';
import '../../features/auth/presentation/login_screen.dart';
import '../../features/auth/presentation/register_screen.dart';
import '../../features/cards/domain/card_models.dart';
import '../../features/cards/presentation/authorization_detail_screen.dart';
import '../../features/cards/presentation/authorizations_list_screen.dart';
import '../../features/cards/presentation/card_detail_screen.dart';
import '../../features/cards/presentation/issue_card_screen.dart';
import '../../features/home/presentation/home_shell_screen.dart';
import '../../features/transfers/presentation/transfer_flow_screen.dart';
import '../auth/session_state.dart';

final appRouterProvider = Provider<GoRouter>((ref) {
  final refreshNotifier = _RouterRefreshNotifier(ref);
  ref.onDispose(refreshNotifier.dispose);

  return GoRouter(
    initialLocation: '/splash',
    refreshListenable: refreshNotifier,
    redirect: (context, state) => _redirect(ref, state),
    routes: [
      GoRoute(path: '/splash', builder: (context, state) => const _SplashScreen()),
      GoRoute(path: '/login', builder: (context, state) => const LoginScreen()),
      GoRoute(path: '/register', builder: (context, state) => const RegisterScreen()),
      GoRoute(path: '/', builder: (context, state) => const HomeShellScreen()),
      GoRoute(path: '/transfer', builder: (context, state) => const TransferFlowScreen()),
      GoRoute(path: '/cards/issue', builder: (context, state) => const IssueCardScreen()),
      GoRoute(
        path: '/cards/:id',
        builder: (context, state) => CardDetailScreen(
          cardId: state.pathParameters['id']!,
          initialCard: state.extra as BankCard?,
        ),
      ),
      GoRoute(
        path: '/authorizations',
        builder: (context, state) => const AuthorizationsListScreen(),
      ),
      GoRoute(
        path: '/authorizations/:id',
        builder: (context, state) => AuthorizationDetailScreen(
          authorizationId: state.pathParameters['id']!,
          initialAuthorization: state.extra as CardAuthorization?,
        ),
      ),
    ],
  );
});

String? _redirect(Ref ref, GoRouterState state) {
  final status = ref.read(sessionStatusProvider);
  final loc = state.matchedLocation;
  final loggingIn = loc == '/login' || loc == '/register';

  if (status == SessionStatus.unknown) {
    return loc == '/splash' ? null : '/splash';
  }
  if (status == SessionStatus.unauthenticated) {
    return loggingIn ? null : '/login';
  }
  // authenticated
  return (loggingIn || loc == '/splash') ? '/' : null;
}

/// Bridges Riverpod state changes into go_router's imperative
/// `refreshListenable`, and — by listening at all — forces
/// [authControllerProvider] to build (and thus restore the session) as soon
/// as the router is first created.
class _RouterRefreshNotifier extends ChangeNotifier {
  _RouterRefreshNotifier(Ref ref) {
    ref.listen(sessionStatusProvider, (_, _) => notifyListeners());
    ref.listen(authControllerProvider, (_, _) {});
  }
}

class _SplashScreen extends StatelessWidget {
  const _SplashScreen();

  @override
  Widget build(BuildContext context) {
    return const Scaffold(body: Center(child: CircularProgressIndicator()));
  }
}
