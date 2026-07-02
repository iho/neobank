import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/wallet_repository.dart';
import '../domain/wallet_models.dart';

final walletHomeControllerProvider =
    AsyncNotifierProvider<WalletHomeController, WalletHomeState>(WalletHomeController.new);

class WalletHomeState {
  const WalletHomeState({
    required this.balance,
    required this.transactions,
    this.nextCursor,
    this.isLoadingMore = false,
  });

  final WalletBalance balance;
  final List<WalletTransaction> transactions;
  final String? nextCursor;
  final bool isLoadingMore;

  bool get hasMore => nextCursor != null;

  WalletHomeState copyWith({
    WalletBalance? balance,
    List<WalletTransaction>? transactions,
    String? nextCursor,
    bool clearNextCursor = false,
    bool? isLoadingMore,
  }) {
    return WalletHomeState(
      balance: balance ?? this.balance,
      transactions: transactions ?? this.transactions,
      nextCursor: clearNextCursor ? null : (nextCursor ?? this.nextCursor),
      isLoadingMore: isLoadingMore ?? this.isLoadingMore,
    );
  }
}

class WalletHomeController extends AsyncNotifier<WalletHomeState> {
  @override
  Future<WalletHomeState> build() => _load();

  Future<WalletHomeState> _load() async {
    final repo = ref.read(walletRepositoryProvider);
    final balance = await repo.getBalance();
    final page = await repo.listTransactions();
    return WalletHomeState(
      balance: balance,
      transactions: page.transactions,
      nextCursor: page.nextCursor,
    );
  }

  Future<void> refresh() async {
    state = await AsyncValue.guard(_load);
  }

  Future<void> loadMore() async {
    final current = state.valueOrNull;
    if (current == null || !current.hasMore || current.isLoadingMore) return;

    state = AsyncData(current.copyWith(isLoadingMore: true));
    try {
      final page = await ref
          .read(walletRepositoryProvider)
          .listTransactions(cursor: current.nextCursor);
      state = AsyncData(
        current.copyWith(
          transactions: [...current.transactions, ...page.transactions],
          nextCursor: page.nextCursor,
          clearNextCursor: page.nextCursor == null,
          isLoadingMore: false,
        ),
      );
    } catch (_) {
      state = AsyncData(current.copyWith(isLoadingMore: false));
      rethrow;
    }
  }
}
