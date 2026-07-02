// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'login_response.g.dart';

@JsonSerializable()
class LoginResponse {
  const LoginResponse({
    required this.userId,
    required this.accessToken,
    required this.refreshToken,
  });
  
  factory LoginResponse.fromJson(Map<String, Object?> json) => _$LoginResponseFromJson(json);
  
  @JsonKey(name: 'user_id')
  final String userId;
  @JsonKey(name: 'access_token')
  final String accessToken;
  @JsonKey(name: 'refresh_token')
  final String refreshToken;

  Map<String, Object?> toJson() => _$LoginResponseToJson(this);
}
