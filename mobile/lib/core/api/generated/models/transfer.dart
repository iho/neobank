// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'transfer.g.dart';

@JsonSerializable()
class Transfer {
  const Transfer({
    this.id,
    this.status,
    this.senderUserId,
    this.recipientUserId,
    this.amount,
    this.currency,
    this.ledgerTransferId,
    this.failureReason,
    this.memo,
    this.createdAt,
    this.completedAt,
  });
  
  factory Transfer.fromJson(Map<String, Object?> json) => _$TransferFromJson(json);
  
  final String? id;
  final String? status;
  @JsonKey(name: 'sender_user_id')
  final String? senderUserId;
  @JsonKey(name: 'recipient_user_id')
  final String? recipientUserId;
  final String? amount;
  final String? currency;
  @JsonKey(name: 'ledger_transfer_id')
  final String? ledgerTransferId;
  @JsonKey(name: 'failure_reason')
  final String? failureReason;
  final String? memo;
  @JsonKey(name: 'created_at')
  final DateTime? createdAt;
  @JsonKey(name: 'completed_at')
  final DateTime? completedAt;

  Map<String, Object?> toJson() => _$TransferToJson(this);
}
