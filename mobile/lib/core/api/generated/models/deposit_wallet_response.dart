// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'deposit_wallet_response.g.dart';

@JsonSerializable()
class DepositWalletResponse {
  const DepositWalletResponse({
    required this.id,
    required this.walletId,
    required this.amount,
    required this.currency,
    required this.status,
    this.ledgerTransferId,
    this.createdAt,
  });
  
  factory DepositWalletResponse.fromJson(Map<String, Object?> json) => _$DepositWalletResponseFromJson(json);
  
  final String id;
  @JsonKey(name: 'wallet_id')
  final String walletId;
  final String amount;
  final String currency;
  @JsonKey(name: 'ledger_transfer_id')
  final String? ledgerTransferId;
  final String status;
  @JsonKey(name: 'created_at')
  final DateTime? createdAt;

  Map<String, Object?> toJson() => _$DepositWalletResponseToJson(this);
}
