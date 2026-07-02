import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/card_repository.dart';
import '../domain/card_models.dart';

final cardsControllerProvider =
    AsyncNotifierProvider<CardsController, List<BankCard>>(CardsController.new);

class CardsController extends AsyncNotifier<List<BankCard>> {
  @override
  Future<List<BankCard>> build() => ref.read(cardRepositoryProvider).listCards();

  Future<void> refresh() async {
    state = await AsyncValue.guard(() => ref.read(cardRepositoryProvider).listCards());
  }

  Future<void> issueCard({
    required String cardholderName,
    String? dailyLimit,
    bool? onlineOnly,
  }) async {
    await ref.read(cardRepositoryProvider).issueCard(
          cardholderName: cardholderName,
          dailyLimit: dailyLimit,
          onlineOnly: onlineOnly,
        );
    await refresh();
  }

  Future<void> toggleFreeze(BankCard card) async {
    final repo = ref.read(cardRepositoryProvider);
    final updated = card.isFrozen ? await repo.unfreeze(card.id) : await repo.freeze(card.id);
    final current = state.valueOrNull ?? [];
    state = AsyncData([
      for (final c in current) if (c.id == updated.id) updated else c,
    ]);
  }
}
