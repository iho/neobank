// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'create_transfer_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

CreateTransferRequest _$CreateTransferRequestFromJson(
  Map<String, dynamic> json,
) => CreateTransferRequest(
  amount: json['amount'] as String,
  recipientPhone: json['recipient_phone'] as String?,
  recipientEmail: json['recipient_email'] as String?,
  recipientUserId: json['recipient_user_id'] as String?,
  currency: json['currency'] as String?,
  memo: json['memo'] as String?,
);

Map<String, dynamic> _$CreateTransferRequestToJson(
  CreateTransferRequest instance,
) => <String, dynamic>{
  'recipient_phone': instance.recipientPhone,
  'recipient_email': instance.recipientEmail,
  'recipient_user_id': instance.recipientUserId,
  'amount': instance.amount,
  'currency': instance.currency,
  'memo': instance.memo,
};
