// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'notification_preferences.g.dart';

@JsonSerializable()
class NotificationPreferences {
  const NotificationPreferences({
    required this.transfers,
    required this.cards,
    required this.kyc,
    required this.push,
    required this.email,
  });
  
  factory NotificationPreferences.fromJson(Map<String, Object?> json) => _$NotificationPreferencesFromJson(json);
  
  final bool transfers;
  final bool cards;
  final bool kyc;
  final bool push;
  final bool email;

  Map<String, Object?> toJson() => _$NotificationPreferencesToJson(this);
}
