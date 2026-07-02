/// Amounts are kept as the decimal strings the gateway sends (see
/// `pkg/money` on the backend) — never parsed to `double`, to avoid binary
/// floating-point error on money. Display them as-is; arithmetic, if ever
/// needed client-side, should go through a decimal library.
class WalletBalance {
  const WalletBalance({
    required this.walletId,
    required this.currency,
    required this.balance,
    required this.availableBalance,
    this.encumberedBalance,
  });

  factory WalletBalance.fromJson(Map<String, dynamic> json) => WalletBalance(
        walletId: json['wallet_id'] as String,
        currency: json['currency'] as String,
        balance: json['balance'] as String,
        availableBalance: json['available_balance'] as String,
        encumberedBalance: json['encumbered_balance'] as String?,
      );

  final String walletId;
  final String currency;
  final String balance;
  final String availableBalance;
  final String? encumberedBalance;
}

class WalletTransaction {
  const WalletTransaction({
    required this.id,
    required this.type,
    required this.amount,
    required this.currency,
    required this.direction,
    required this.status,
    required this.createdAt,
    this.counterparty,
    this.memo,
    this.referenceId,
  });

  factory WalletTransaction.fromJson(Map<String, dynamic> json) => WalletTransaction(
        id: json['id'] as String,
        type: json['type'] as String,
        amount: json['amount'] as String,
        currency: json['currency'] as String,
        direction: json['direction'] as String,
        status: json['status'] as String,
        createdAt: DateTime.parse(json['created_at'] as String),
        counterparty: json['counterparty'] as String?,
        memo: json['memo'] as String?,
        referenceId: json['reference_id'] as String?,
      );

  final String id;
  final String type;
  final String amount;
  final String currency;

  /// `debit` or `credit`.
  final String direction;
  final String status;
  final DateTime createdAt;
  final String? counterparty;
  final String? memo;
  final String? referenceId;

  bool get isCredit => direction == 'credit';
}

class WalletTransactionPage {
  const WalletTransactionPage({required this.transactions, this.nextCursor});

  factory WalletTransactionPage.fromJson(Map<String, dynamic> json) => WalletTransactionPage(
        transactions: (json['transactions'] as List<dynamic>)
            .map((e) => WalletTransaction.fromJson(e as Map<String, dynamic>))
            .toList(),
        nextCursor: json['next_cursor'] as String?,
      );

  final List<WalletTransaction> transactions;
  final String? nextCursor;
}
