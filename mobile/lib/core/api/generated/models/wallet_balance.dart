// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'wallet_balance.g.dart';

@JsonSerializable()
class WalletBalance {
  const WalletBalance({
    required this.walletId,
    required this.currency,
    required this.balance,
    required this.availableBalance,
    this.ledgerAccountId,
    this.encumberedBalance,
  });
  
  factory WalletBalance.fromJson(Map<String, Object?> json) => _$WalletBalanceFromJson(json);
  
  @JsonKey(name: 'wallet_id')
  final String walletId;
  @JsonKey(name: 'ledger_account_id')
  final String? ledgerAccountId;
  final String currency;
  final String balance;
  @JsonKey(name: 'encumbered_balance')
  final String? encumberedBalance;
  @JsonKey(name: 'available_balance')
  final String availableBalance;

  Map<String, Object?> toJson() => _$WalletBalanceToJson(this);
}
