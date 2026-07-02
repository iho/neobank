import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/card_repository.dart';
import '../domain/card_models.dart';

final authorizationsControllerProvider =
    AsyncNotifierProvider<AuthorizationsController, List<CardAuthorization>>(
  AuthorizationsController.new,
);

class AuthorizationsController extends AsyncNotifier<List<CardAuthorization>> {
  @override
  Future<List<CardAuthorization>> build() =>
      ref.read(cardRepositoryProvider).listAuthorizations();

  Future<void> refresh() async {
    state =
        await AsyncValue.guard(() => ref.read(cardRepositoryProvider).listAuthorizations());
  }

  Future<void> capture(String id) async {
    final updated = await ref.read(cardRepositoryProvider).captureAuthorization(id);
    final current = state.valueOrNull ?? [];
    state = AsyncData([
      for (final a in current) if (a.id == updated.id) updated else a,
    ]);
  }
}
