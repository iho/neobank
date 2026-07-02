// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'wallet_transaction.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

WalletTransaction _$WalletTransactionFromJson(Map<String, dynamic> json) =>
    WalletTransaction(
      id: json['id'] as String,
      type: json['type'] as String,
      amount: json['amount'] as String,
      currency: json['currency'] as String,
      direction: json['direction'] as String,
      status: json['status'] as String,
      createdAt: DateTime.parse(json['created_at'] as String),
      counterparty: json['counterparty'] as String?,
      memo: json['memo'] as String?,
      referenceId: json['reference_id'] as String?,
    );

Map<String, dynamic> _$WalletTransactionToJson(WalletTransaction instance) =>
    <String, dynamic>{
      'id': instance.id,
      'type': instance.type,
      'amount': instance.amount,
      'currency': instance.currency,
      'direction': instance.direction,
      'status': instance.status,
      'counterparty': instance.counterparty,
      'memo': instance.memo,
      'reference_id': instance.referenceId,
      'created_at': instance.createdAt.toIso8601String(),
    };
