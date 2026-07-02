// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'issue_card_request.g.dart';

@JsonSerializable()
class IssueCardRequest {
  const IssueCardRequest({
    required this.cardholderName,
    this.walletId,
    this.dailyLimit,
    this.onlineOnly,
  });
  
  factory IssueCardRequest.fromJson(Map<String, Object?> json) => _$IssueCardRequestFromJson(json);
  
  @JsonKey(name: 'wallet_id')
  final String? walletId;
  @JsonKey(name: 'cardholder_name')
  final String cardholderName;
  @JsonKey(name: 'daily_limit')
  final String? dailyLimit;
  @JsonKey(name: 'online_only')
  final bool? onlineOnly;

  Map<String, Object?> toJson() => _$IssueCardRequestToJson(this);
}
