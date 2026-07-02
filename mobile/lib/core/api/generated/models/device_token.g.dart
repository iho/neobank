// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'device_token.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

DeviceToken _$DeviceTokenFromJson(Map<String, dynamic> json) => DeviceToken(
  id: json['id'] as String,
  userId: json['user_id'] as String,
  platform: json['platform'] as String,
  token: json['token'] as String,
  createdAt: DateTime.parse(json['created_at'] as String),
);

Map<String, dynamic> _$DeviceTokenToJson(DeviceToken instance) =>
    <String, dynamic>{
      'id': instance.id,
      'user_id': instance.userId,
      'platform': instance.platform,
      'token': instance.token,
      'created_at': instance.createdAt.toIso8601String(),
    };
