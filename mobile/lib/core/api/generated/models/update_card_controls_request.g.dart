// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'update_card_controls_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

UpdateCardControlsRequest _$UpdateCardControlsRequestFromJson(
  Map<String, dynamic> json,
) => UpdateCardControlsRequest(
  dailyLimit: json['daily_limit'] as String?,
  onlineOnly: json['online_only'] as bool?,
);

Map<String, dynamic> _$UpdateCardControlsRequestToJson(
  UpdateCardControlsRequest instance,
) => <String, dynamic>{
  'daily_limit': instance.dailyLimit,
  'online_only': instance.onlineOnly,
};
