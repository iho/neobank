import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../domain/card_models.dart';
import 'cards_controller.dart';

class CardDetailScreen extends ConsumerWidget {
  const CardDetailScreen({required this.cardId, required this.initialCard, super.key});

  final String cardId;

  /// Passed via `GoRouterState.extra` from the list screen so the detail
  /// screen has something to render before `cardsControllerProvider`
  /// (re-)resolves; once it does, the live copy from the list takes over.
  final BankCard? initialCard;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final cards = ref.watch(cardsControllerProvider).valueOrNull;
    BankCard? card = initialCard;
    if (cards != null) {
      for (final c in cards) {
        if (c.id == cardId) {
          card = c;
          break;
        }
      }
    }

    if (card == null) {
      return const Scaffold(body: Center(child: Text('Card not found')));
    }
    final resolvedCard = card;

    return Scaffold(
      appBar: AppBar(title: Text('•••• ${card.lastFour}')),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(20),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('•••• •••• •••• ${card.lastFour}',
                          style: Theme.of(context).textTheme.headlineSmall),
                      const SizedBox(height: 12),
                      Text(
                        'Expires ${card.expiryMonth.toString().padLeft(2, '0')}/${card.expiryYear}',
                      ),
                      const SizedBox(height: 4),
                      Text('Status: ${card.status}'),
                      if (card.dailyLimit != null) ...[
                        const SizedBox(height: 4),
                        Text('Daily limit: ${card.dailyLimit}'),
                      ],
                      const SizedBox(height: 4),
                      Text(card.onlineOnly ? 'Online purchases only' : 'Online + in-person'),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 24),
              FilledButton.icon(
                icon: Icon(card.isFrozen ? Icons.ac_unit_outlined : Icons.ac_unit),
                label: Text(card.isFrozen ? 'Unfreeze card' : 'Freeze card'),
                onPressed: () =>
                    ref.read(cardsControllerProvider.notifier).toggleFreeze(resolvedCard),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
