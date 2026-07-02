// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'register_response.g.dart';

@JsonSerializable()
class RegisterResponse {
  const RegisterResponse({
    required this.userId,
    required this.accessToken,
    required this.refreshToken,
  });
  
  factory RegisterResponse.fromJson(Map<String, Object?> json) => _$RegisterResponseFromJson(json);
  
  @JsonKey(name: 'user_id')
  final String userId;
  @JsonKey(name: 'access_token')
  final String accessToken;
  @JsonKey(name: 'refresh_token')
  final String refreshToken;

  Map<String, Object?> toJson() => _$RegisterResponseToJson(this);
}
