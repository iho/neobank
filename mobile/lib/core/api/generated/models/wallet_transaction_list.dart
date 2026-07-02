// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

import 'wallet_transaction.dart';

part 'wallet_transaction_list.g.dart';

@JsonSerializable()
class WalletTransactionList {
  const WalletTransactionList({
    required this.transactions,
    this.nextCursor,
  });
  
  factory WalletTransactionList.fromJson(Map<String, Object?> json) => _$WalletTransactionListFromJson(json);
  
  final List<WalletTransaction> transactions;
  @JsonKey(name: 'next_cursor')
  final String? nextCursor;

  Map<String, Object?> toJson() => _$WalletTransactionListToJson(this);
}
