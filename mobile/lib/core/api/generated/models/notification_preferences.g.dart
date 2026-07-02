// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'notification_preferences.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

NotificationPreferences _$NotificationPreferencesFromJson(
  Map<String, dynamic> json,
) => NotificationPreferences(
  transfers: json['transfers'] as bool,
  cards: json['cards'] as bool,
  kyc: json['kyc'] as bool,
  push: json['push'] as bool,
  email: json['email'] as bool,
);

Map<String, dynamic> _$NotificationPreferencesToJson(
  NotificationPreferences instance,
) => <String, dynamic>{
  'transfers': instance.transfers,
  'cards': instance.cards,
  'kyc': instance.kyc,
  'push': instance.push,
  'email': instance.email,
};
