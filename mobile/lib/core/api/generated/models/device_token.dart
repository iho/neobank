// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'device_token.g.dart';

@JsonSerializable()
class DeviceToken {
  const DeviceToken({
    required this.id,
    required this.userId,
    required this.platform,
    required this.token,
    required this.createdAt,
  });
  
  factory DeviceToken.fromJson(Map<String, Object?> json) => _$DeviceTokenFromJson(json);
  
  final String id;
  @JsonKey(name: 'user_id')
  final String userId;
  final String platform;
  final String token;
  @JsonKey(name: 'created_at')
  final DateTime createdAt;

  Map<String, Object?> toJson() => _$DeviceTokenToJson(this);
}
