// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'wallet_transaction_list.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

WalletTransactionList _$WalletTransactionListFromJson(
  Map<String, dynamic> json,
) => WalletTransactionList(
  transactions: (json['transactions'] as List<dynamic>)
      .map((e) => WalletTransaction.fromJson(e as Map<String, dynamic>))
      .toList(),
  nextCursor: json['next_cursor'] as String?,
);

Map<String, dynamic> _$WalletTransactionListToJson(
  WalletTransactionList instance,
) => <String, dynamic>{
  'transactions': instance.transactions,
  'next_cursor': instance.nextCursor,
};
