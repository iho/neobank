// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'transfer.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

Transfer _$TransferFromJson(Map<String, dynamic> json) => Transfer(
  id: json['id'] as String?,
  status: json['status'] as String?,
  senderUserId: json['sender_user_id'] as String?,
  recipientUserId: json['recipient_user_id'] as String?,
  amount: json['amount'] as String?,
  currency: json['currency'] as String?,
  ledgerTransferId: json['ledger_transfer_id'] as String?,
  failureReason: json['failure_reason'] as String?,
  memo: json['memo'] as String?,
  createdAt: json['created_at'] == null
      ? null
      : DateTime.parse(json['created_at'] as String),
  completedAt: json['completed_at'] == null
      ? null
      : DateTime.parse(json['completed_at'] as String),
);

Map<String, dynamic> _$TransferToJson(Transfer instance) => <String, dynamic>{
  'id': instance.id,
  'status': instance.status,
  'sender_user_id': instance.senderUserId,
  'recipient_user_id': instance.recipientUserId,
  'amount': instance.amount,
  'currency': instance.currency,
  'ledger_transfer_id': instance.ledgerTransferId,
  'failure_reason': instance.failureReason,
  'memo': instance.memo,
  'created_at': instance.createdAt?.toIso8601String(),
  'completed_at': instance.completedAt?.toIso8601String(),
};
