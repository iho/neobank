// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'register_device_token_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

RegisterDeviceTokenRequest _$RegisterDeviceTokenRequestFromJson(
  Map<String, dynamic> json,
) => RegisterDeviceTokenRequest(
  platform: RegisterDeviceTokenRequestPlatform.fromJson(
    json['platform'] as String,
  ),
  token: json['token'] as String,
);

Map<String, dynamic> _$RegisterDeviceTokenRequestToJson(
  RegisterDeviceTokenRequest instance,
) => <String, dynamic>{
  'platform': _$RegisterDeviceTokenRequestPlatformEnumMap[instance.platform]!,
  'token': instance.token,
};

const _$RegisterDeviceTokenRequestPlatformEnumMap = {
  RegisterDeviceTokenRequestPlatform.ios: 'ios',
  RegisterDeviceTokenRequestPlatform.android: 'android',
  RegisterDeviceTokenRequestPlatform.web: 'web',
  RegisterDeviceTokenRequestPlatform.$unknown: r'$unknown',
};
