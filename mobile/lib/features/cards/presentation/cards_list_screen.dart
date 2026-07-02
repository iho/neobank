import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../domain/card_models.dart';
import 'cards_controller.dart';

class CardsListScreen extends ConsumerWidget {
  const CardsListScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final cards = ref.watch(cardsControllerProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Cards'),
        actions: [
          IconButton(
            icon: const Icon(Icons.receipt_long),
            tooltip: 'Authorizations',
            onPressed: () => context.push('/authorizations'),
          ),
          IconButton(
            icon: const Icon(Icons.add),
            onPressed: () => context.push('/cards/issue'),
          ),
        ],
      ),
      body: cards.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (error, _) => Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text('$error'),
              const SizedBox(height: 12),
              OutlinedButton(
                onPressed: () => ref.read(cardsControllerProvider.notifier).refresh(),
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
        data: (list) => list.isEmpty
            ? Center(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Text('No cards yet'),
                    const SizedBox(height: 12),
                    FilledButton(
                      onPressed: () => context.push('/cards/issue'),
                      child: const Text('Issue a card'),
                    ),
                  ],
                ),
              )
            : RefreshIndicator(
                onRefresh: () => ref.read(cardsControllerProvider.notifier).refresh(),
                child: ListView.builder(
                  padding: const EdgeInsets.all(16),
                  itemCount: list.length,
                  itemBuilder: (context, index) => _CardTile(card: list[index]),
                ),
              ),
      ),
    );
  }
}

class _CardTile extends StatelessWidget {
  const _CardTile({required this.card});

  final BankCard card;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: ListTile(
        leading: Icon(card.isFrozen ? Icons.ac_unit : Icons.credit_card),
        title: Text('•••• ${card.lastFour}'),
        subtitle: Text(
          '${card.status} · exp ${card.expiryMonth.toString().padLeft(2, '0')}/${card.expiryYear}',
        ),
        trailing: const Icon(Icons.chevron_right),
        onTap: () => context.push('/cards/${card.id}', extra: card),
      ),
    );
  }
}
