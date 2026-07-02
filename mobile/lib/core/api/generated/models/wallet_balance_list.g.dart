// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'wallet_balance_list.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

WalletBalanceList _$WalletBalanceListFromJson(Map<String, dynamic> json) =>
    WalletBalanceList(
      wallets: (json['wallets'] as List<dynamic>)
          .map((e) => WalletBalance.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$WalletBalanceListToJson(WalletBalanceList instance) =>
    <String, dynamic>{'wallets': instance.wallets};
