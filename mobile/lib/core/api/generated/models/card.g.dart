// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'card.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

Card _$CardFromJson(Map<String, dynamic> json) => Card(
  id: json['id'] as String,
  userId: json['user_id'] as String,
  walletId: json['wallet_id'] as String,
  lastFour: json['last_four'] as String,
  status: json['status'] as String,
  expiryMonth: (json['expiry_month'] as num).toInt(),
  expiryYear: (json['expiry_year'] as num).toInt(),
  onlineOnly: json['online_only'] as bool,
  dailyLimit: json['daily_limit'] as String?,
);

Map<String, dynamic> _$CardToJson(Card instance) => <String, dynamic>{
  'id': instance.id,
  'user_id': instance.userId,
  'wallet_id': instance.walletId,
  'last_four': instance.lastFour,
  'status': instance.status,
  'expiry_month': instance.expiryMonth,
  'expiry_year': instance.expiryYear,
  'daily_limit': instance.dailyLimit,
  'online_only': instance.onlineOnly,
};
