// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'update_notification_preferences_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

UpdateNotificationPreferencesRequest
_$UpdateNotificationPreferencesRequestFromJson(Map<String, dynamic> json) =>
    UpdateNotificationPreferencesRequest(
      transfers: json['transfers'] as bool?,
      cards: json['cards'] as bool?,
      kyc: json['kyc'] as bool?,
      push: json['push'] as bool?,
      email: json['email'] as bool?,
    );

Map<String, dynamic> _$UpdateNotificationPreferencesRequestToJson(
  UpdateNotificationPreferencesRequest instance,
) => <String, dynamic>{
  'transfers': instance.transfers,
  'cards': instance.cards,
  'kyc': instance.kyc,
  'push': instance.push,
  'email': instance.email,
};
