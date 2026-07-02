// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'create_transfer_request.g.dart';

@JsonSerializable()
class CreateTransferRequest {
  const CreateTransferRequest({
    required this.amount,
    this.recipientPhone,
    this.recipientEmail,
    this.recipientUserId,
    this.currency,
    this.memo,
  });
  
  factory CreateTransferRequest.fromJson(Map<String, Object?> json) => _$CreateTransferRequestFromJson(json);
  
  @JsonKey(name: 'recipient_phone')
  final String? recipientPhone;
  @JsonKey(name: 'recipient_email')
  final String? recipientEmail;
  @JsonKey(name: 'recipient_user_id')
  final String? recipientUserId;
  final String amount;
  final String? currency;
  final String? memo;

  Map<String, Object?> toJson() => _$CreateTransferRequestToJson(this);
}
