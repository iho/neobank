// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'payee.g.dart';

@JsonSerializable()
class Payee {
  const Payee({
    required this.id,
    required this.payeeUserId,
    required this.lastUsedAt,
    required this.createdAt,
    this.nickname,
    this.payeeEmail,
    this.payeePhone,
  });
  
  factory Payee.fromJson(Map<String, Object?> json) => _$PayeeFromJson(json);
  
  final String id;
  @JsonKey(name: 'payee_user_id')
  final String payeeUserId;
  final String? nickname;
  @JsonKey(name: 'payee_email')
  final String? payeeEmail;
  @JsonKey(name: 'payee_phone')
  final String? payeePhone;
  @JsonKey(name: 'last_used_at')
  final DateTime lastUsedAt;
  @JsonKey(name: 'created_at')
  final DateTime createdAt;

  Map<String, Object?> toJson() => _$PayeeToJson(this);
}
