// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'wallet_transaction.g.dart';

@JsonSerializable()
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
  
  factory WalletTransaction.fromJson(Map<String, Object?> json) => _$WalletTransactionFromJson(json);
  
  final String id;

  /// p2p_out, p2p_in, card_purchase, card_hold
  final String type;
  final String amount;
  final String currency;

  /// debit or credit
  final String direction;
  final String status;
  final String? counterparty;
  final String? memo;
  @JsonKey(name: 'reference_id')
  final String? referenceId;
  @JsonKey(name: 'created_at')
  final DateTime createdAt;

  Map<String, Object?> toJson() => _$WalletTransactionToJson(this);
}
