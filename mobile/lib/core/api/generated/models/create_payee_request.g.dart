// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'create_payee_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

CreatePayeeRequest _$CreatePayeeRequestFromJson(Map<String, dynamic> json) =>
    CreatePayeeRequest(
      payeeUserId: json['payee_user_id'] as String,
      nickname: json['nickname'] as String?,
    );

Map<String, dynamic> _$CreatePayeeRequestToJson(CreatePayeeRequest instance) =>
    <String, dynamic>{
      'payee_user_id': instance.payeeUserId,
      'nickname': instance.nickname,
    };
