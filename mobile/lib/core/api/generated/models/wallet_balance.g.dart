// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'wallet_balance.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

WalletBalance _$WalletBalanceFromJson(Map<String, dynamic> json) =>
    WalletBalance(
      walletId: json['wallet_id'] as String,
      currency: json['currency'] as String,
      balance: json['balance'] as String,
      availableBalance: json['available_balance'] as String,
      ledgerAccountId: json['ledger_account_id'] as String?,
      encumberedBalance: json['encumbered_balance'] as String?,
    );

Map<String, dynamic> _$WalletBalanceToJson(WalletBalance instance) =>
    <String, dynamic>{
      'wallet_id': instance.walletId,
      'ledger_account_id': instance.ledgerAccountId,
      'currency': instance.currency,
      'balance': instance.balance,
      'encumbered_balance': instance.encumberedBalance,
      'available_balance': instance.availableBalance,
    };
