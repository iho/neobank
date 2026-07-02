import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../auth/presentation/auth_controller.dart';
import '../domain/wallet_models.dart';
import 'wallet_home_controller.dart';

class WalletHomeScreen extends ConsumerStatefulWidget {
  const WalletHomeScreen({super.key});

  @override
  ConsumerState<WalletHomeScreen> createState() => _WalletHomeScreenState();
}

class _WalletHomeScreenState extends ConsumerState<WalletHomeScreen> {
  final _scrollController = ScrollController();

  @override
  void initState() {
    super.initState();
    _scrollController.addListener(() {
      if (_scrollController.position.pixels >
          _scrollController.position.maxScrollExtent - 200) {
        ref.read(walletHomeControllerProvider.notifier).loadMore();
      }
    });
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(walletHomeControllerProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Wallet'),
        actions: [
          IconButton(
            icon: const Icon(Icons.logout),
            tooltip: 'Log out',
            onPressed: () => ref.read(authControllerProvider.notifier).logout(),
          ),
        ],
      ),
      body: state.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (error, _) => _ErrorView(
          message: '$error',
          onRetry: () => ref.read(walletHomeControllerProvider.notifier).refresh(),
        ),
        data: (data) => RefreshIndicator(
          onRefresh: () => ref.read(walletHomeControllerProvider.notifier).refresh(),
          child: ListView(
            controller: _scrollController,
            padding: const EdgeInsets.all(16),
            children: [
              _BalanceCard(balance: data.balance),
              const SizedBox(height: 24),
              Text('Transactions', style: Theme.of(context).textTheme.titleMedium),
              const SizedBox(height: 8),
              if (data.transactions.isEmpty)
                const Padding(
                  padding: EdgeInsets.symmetric(vertical: 32),
                  child: Center(child: Text('No transactions yet')),
                )
              else
                ...data.transactions.map((tx) => _TransactionTile(tx: tx)),
              if (data.isLoadingMore)
                const Padding(
                  padding: EdgeInsets.symmetric(vertical: 16),
                  child: Center(child: CircularProgressIndicator()),
                ),
            ],
          ),
        ),
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => context.push('/transfer'),
        icon: const Icon(Icons.send),
        label: const Text('Send'),
      ),
    );
  }
}

class _BalanceCard extends StatelessWidget {
  const _BalanceCard({required this.balance});

  final WalletBalance balance;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Available balance', style: Theme.of(context).textTheme.bodyMedium),
            const SizedBox(height: 8),
            Text(
              '${balance.availableBalance} ${balance.currency}',
              style: Theme.of(context).textTheme.headlineMedium,
            ),
            if (balance.encumberedBalance != null) ...[
              const SizedBox(height: 4),
              Text(
                'Held: ${balance.encumberedBalance} ${balance.currency}',
                style: Theme.of(context).textTheme.bodySmall,
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _TransactionTile extends StatelessWidget {
  const _TransactionTile({required this.tx});

  final WalletTransaction tx;

  @override
  Widget build(BuildContext context) {
    final sign = tx.isCredit ? '+' : '-';
    final color = tx.isCredit ? Colors.green : null;
    return ListTile(
      contentPadding: EdgeInsets.zero,
      leading: Icon(tx.isCredit ? Icons.arrow_downward : Icons.arrow_upward),
      title: Text(tx.counterparty ?? tx.type),
      subtitle: Text('${tx.status} · ${tx.createdAt.toLocal()}'.split('.').first),
      trailing: Text(
        '$sign${tx.amount} ${tx.currency}',
        style: TextStyle(color: color, fontWeight: FontWeight.bold),
      ),
    );
  }
}

class _ErrorView extends StatelessWidget {
  const _ErrorView({required this.message, required this.onRetry});

  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(message, textAlign: TextAlign.center),
          const SizedBox(height: 16),
          OutlinedButton(onPressed: onRetry, child: const Text('Retry')),
        ],
      ),
    );
  }
}
