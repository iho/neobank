// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'refresh_token_request.g.dart';

@JsonSerializable()
class RefreshTokenRequest {
  const RefreshTokenRequest({
    required this.refreshToken,
  });
  
  factory RefreshTokenRequest.fromJson(Map<String, Object?> json) => _$RefreshTokenRequestFromJson(json);
  
  @JsonKey(name: 'refresh_token')
  final String refreshToken;

  Map<String, Object?> toJson() => _$RefreshTokenRequestToJson(this);
}
