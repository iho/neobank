// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

import 'register_device_token_request_platform.dart';

part 'register_device_token_request.g.dart';

@JsonSerializable()
class RegisterDeviceTokenRequest {
  const RegisterDeviceTokenRequest({
    required this.platform,
    required this.token,
  });
  
  factory RegisterDeviceTokenRequest.fromJson(Map<String, Object?> json) => _$RegisterDeviceTokenRequestFromJson(json);
  
  final RegisterDeviceTokenRequestPlatform platform;
  final String token;

  Map<String, Object?> toJson() => _$RegisterDeviceTokenRequestToJson(this);
}
