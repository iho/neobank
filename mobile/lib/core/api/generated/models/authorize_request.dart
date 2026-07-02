// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

import 'authorize_request_channel.dart';

part 'authorize_request.g.dart';

@JsonSerializable()
class AuthorizeRequest {
  const AuthorizeRequest({
    required this.amount,
    this.currency,
    this.merchantName,
    this.merchantCategoryCode,
    this.channel,
  });
  
  factory AuthorizeRequest.fromJson(Map<String, Object?> json) => _$AuthorizeRequestFromJson(json);
  
  final String amount;
  final String? currency;
  @JsonKey(name: 'merchant_name')
  final String? merchantName;
  @JsonKey(name: 'merchant_category_code')
  final String? merchantCategoryCode;
  final AuthorizeRequestChannel? channel;

  Map<String, Object?> toJson() => _$AuthorizeRequestToJson(this);
}
