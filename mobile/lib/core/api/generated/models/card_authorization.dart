// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'card_authorization.g.dart';

@JsonSerializable()
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
  
  factory CardAuthorization.fromJson(Map<String, Object?> json) => _$CardAuthorizationFromJson(json);
  
  final String id;
  @JsonKey(name: 'card_id')
  final String cardId;
  @JsonKey(name: 'user_id')
  final String userId;
  @JsonKey(name: 'merchant_name')
  final String? merchantName;
  @JsonKey(name: 'merchant_category_code')
  final String? merchantCategoryCode;
  final String amount;
  final String currency;
  final String status;
  @JsonKey(name: 'ledger_hold_id')
  final String? ledgerHoldId;
  @JsonKey(name: 'ledger_transfer_id')
  final String? ledgerTransferId;
  @JsonKey(name: 'failure_reason')
  final String? failureReason;
  @JsonKey(name: 'created_at')
  final DateTime? createdAt;
  @JsonKey(name: 'captured_at')
  final DateTime? capturedAt;

  Map<String, Object?> toJson() => _$CardAuthorizationToJson(this);
}
