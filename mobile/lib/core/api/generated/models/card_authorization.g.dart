// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'card_authorization.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

CardAuthorization _$CardAuthorizationFromJson(Map<String, dynamic> json) =>
    CardAuthorization(
      id: json['id'] as String,
      cardId: json['card_id'] as String,
      userId: json['user_id'] as String,
      amount: json['amount'] as String,
      currency: json['currency'] as String,
      status: json['status'] as String,
      merchantName: json['merchant_name'] as String?,
      merchantCategoryCode: json['merchant_category_code'] as String?,
      ledgerHoldId: json['ledger_hold_id'] as String?,
      ledgerTransferId: json['ledger_transfer_id'] as String?,
      failureReason: json['failure_reason'] as String?,
      createdAt: json['created_at'] == null
          ? null
          : DateTime.parse(json['created_at'] as String),
      capturedAt: json['captured_at'] == null
          ? null
          : DateTime.parse(json['captured_at'] as String),
    );

Map<String, dynamic> _$CardAuthorizationToJson(CardAuthorization instance) =>
    <String, dynamic>{
      'id': instance.id,
      'card_id': instance.cardId,
      'user_id': instance.userId,
      'merchant_name': instance.merchantName,
      'merchant_category_code': instance.merchantCategoryCode,
      'amount': instance.amount,
      'currency': instance.currency,
      'status': instance.status,
      'ledger_hold_id': instance.ledgerHoldId,
      'ledger_transfer_id': instance.ledgerTransferId,
      'failure_reason': instance.failureReason,
      'created_at': instance.createdAt?.toIso8601String(),
      'captured_at': instance.capturedAt?.toIso8601String(),
    };
