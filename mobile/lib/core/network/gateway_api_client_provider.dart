import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../api/generated/gateway_api_client.dart';
import 'dio_client.dart';

/// The swagger_parser-generated Retrofit client, wrapping [dioProvider] — so
/// it inherits the same auth/idempotency/correlation-id interceptors as the
/// hand-written repositories. Not yet consumed by any feature; migrating
/// `features/*/data/*_repository.dart` onto this is tracked separately in
/// mobile/TODO.md (a larger, mechanical follow-up best done one feature at a
/// time against a running backend, not blind).
final gatewayApiClientProvider = Provider<GatewayApiClient>((ref) {
  return GatewayApiClient(ref.watch(dioProvider));
});
