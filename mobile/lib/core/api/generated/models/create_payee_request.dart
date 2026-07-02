// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'create_payee_request.g.dart';

@JsonSerializable()
class CreatePayeeRequest {
  const CreatePayeeRequest({
    required this.payeeUserId,
    this.nickname,
  });
  
  factory CreatePayeeRequest.fromJson(Map<String, Object?> json) => _$CreatePayeeRequestFromJson(json);
  
  @JsonKey(name: 'payee_user_id')
  final String payeeUserId;
  final String? nickname;

  Map<String, Object?> toJson() => _$CreatePayeeRequestToJson(this);
}
