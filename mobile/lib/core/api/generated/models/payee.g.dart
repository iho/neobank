// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'payee.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

Payee _$PayeeFromJson(Map<String, dynamic> json) => Payee(
  id: json['id'] as String,
  payeeUserId: json['payee_user_id'] as String,
  lastUsedAt: DateTime.parse(json['last_used_at'] as String),
  createdAt: DateTime.parse(json['created_at'] as String),
  nickname: json['nickname'] as String?,
  payeeEmail: json['payee_email'] as String?,
  payeePhone: json['payee_phone'] as String?,
);

Map<String, dynamic> _$PayeeToJson(Payee instance) => <String, dynamic>{
  'id': instance.id,
  'payee_user_id': instance.payeeUserId,
  'nickname': instance.nickname,
  'payee_email': instance.payeeEmail,
  'payee_phone': instance.payeePhone,
  'last_used_at': instance.lastUsedAt.toIso8601String(),
  'created_at': instance.createdAt.toIso8601String(),
};
