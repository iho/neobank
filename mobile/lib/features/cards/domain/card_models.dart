/// Named `BankCard` (not `Card`) to avoid colliding with `package:flutter`'s
/// Material `Card` widget, which every screen in this feature also uses.
class BankCard {
  const BankCard({
    required this.id,
    required this.userId,
    required this.walletId,
    required this.lastFour,
    required this.status,
    required this.expiryMonth,
    required this.expiryYear,
    required this.onlineOnly,
    this.dailyLimit,
  });

  factory BankCard.fromJson(Map<String, dynamic> json) => BankCard(
        id: json['id'] as String,
        userId: json['user_id'] as String,
        walletId: json['wallet_id'] as String,
        lastFour: json['last_four'] as String,
        status: json['status'] as String,
        expiryMonth: json['expiry_month'] as int,
        expiryYear: json['expiry_year'] as int,
        onlineOnly: json['online_only'] as bool,
        dailyLimit: json['daily_limit'] as String?,
      );

  final String id;
  final String userId;
  final String walletId;
  final String lastFour;
  final String status;
  final int expiryMonth;
  final int expiryYear;
  final bool onlineOnly;
  final String? dailyLimit;

  bool get isFrozen => status == 'frozen';
}

class CardAuthorization {
  const CardAuthorization({
    required this.id,
    required this.cardId,
    required this.userId,
    required this.amount,
    required this.currency,
    required this.status,
    this.merchantName,
    this.merchantCategoryCode,
    this.ledgerHoldId,
    this.ledgerTransferId,
    this.failureReason,
    this.createdAt,
    this.capturedAt,
  });

  factory CardAuthorization.fromJson(Map<String, dynamic> json) => CardAuthorization(
        id: json['id'] as String,
        cardId: json['card_id'] as String,
        userId: json['user_id'] as String,
        amount: json['amount'] as String,
        currency: json['currency'] as String,
        status: json['status'] as String,
        merchantName: json['merchant_name'] as String?,
        merchantCategoryCode: json['merchant_category_code'] as String?,
        ledgerHoldId: json['ledger_hold_id'] as String?,
        ledgerTransferId: json['ledger_transfer_id'] as String?,
        failureReason: json['failure_reason'] as String?,
        createdAt:
            json['created_at'] != null ? DateTime.parse(json['created_at'] as String) : null,
        capturedAt:
            json['captured_at'] != null ? DateTime.parse(json['captured_at'] as String) : null,
      );

  final String id;
  final String cardId;
  final String userId;
  final String amount;
  final String currency;
  final String status;
  final String? merchantName;
  final String? merchantCategoryCode;
  final String? ledgerHoldId;
  final String? ledgerTransferId;
  final String? failureReason;
  final DateTime? createdAt;
  final DateTime? capturedAt;

  bool get isCapturable => status == 'authorized';
}
