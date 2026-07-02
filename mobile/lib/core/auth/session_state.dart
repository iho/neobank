import 'package:flutter_riverpod/flutter_riverpod.dart';

/// `unknown` while session restore is in flight (see
/// `features/auth/presentation/auth_controller.dart`'s `build()`); the router
/// shows a splash screen for that state instead of redirecting.
enum SessionStatus { unknown, authenticated, unauthenticated }

/// Single source of truth for "is there a logged-in session right now".
/// Lives in `core` (not `features/auth`) because both the router and the
/// network layer's forced-logout-on-refresh-failure need to read/write it
/// without depending on the auth feature.
final sessionStatusProvider = StateProvider<SessionStatus>(
  (ref) => SessionStatus.unknown,
);
