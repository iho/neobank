// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'issue_card_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

IssueCardRequest _$IssueCardRequestFromJson(Map<String, dynamic> json) =>
    IssueCardRequest(
      cardholderName: json['cardholder_name'] as String,
      walletId: json['wallet_id'] as String?,
      dailyLimit: json['daily_limit'] as String?,
      onlineOnly: json['online_only'] as bool?,
    );

Map<String, dynamic> _$IssueCardRequestToJson(IssueCardRequest instance) =>
    <String, dynamic>{
      'wallet_id': instance.walletId,
      'cardholder_name': instance.cardholderName,
      'daily_limit': instance.dailyLimit,
      'online_only': instance.onlineOnly,
    };
