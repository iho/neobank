// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'authorize_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

AuthorizeRequest _$AuthorizeRequestFromJson(Map<String, dynamic> json) =>
    AuthorizeRequest(
      amount: json['amount'] as String,
      currency: json['currency'] as String?,
      merchantName: json['merchant_name'] as String?,
      merchantCategoryCode: json['merchant_category_code'] as String?,
      channel: json['channel'] == null
          ? null
          : AuthorizeRequestChannel.fromJson(json['channel'] as String),
    );

Map<String, dynamic> _$AuthorizeRequestToJson(AuthorizeRequest instance) =>
    <String, dynamic>{
      'amount': instance.amount,
      'currency': instance.currency,
      'merchant_name': instance.merchantName,
      'merchant_category_code': instance.merchantCategoryCode,
      'channel': _$AuthorizeRequestChannelEnumMap[instance.channel],
    };

const _$AuthorizeRequestChannelEnumMap = {
  AuthorizeRequestChannel.pos: 'pos',
  AuthorizeRequestChannel.online: 'online',
  AuthorizeRequestChannel.$unknown: r'$unknown',
};
