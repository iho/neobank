// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'update_notification_preferences_request.g.dart';

@JsonSerializable()
class UpdateNotificationPreferencesRequest {
  const UpdateNotificationPreferencesRequest({
    this.transfers,
    this.cards,
    this.kyc,
    this.push,
    this.email,
  });
  
  factory UpdateNotificationPreferencesRequest.fromJson(Map<String, Object?> json) => _$UpdateNotificationPreferencesRequestFromJson(json);
  
  final bool? transfers;
  final bool? cards;
  final bool? kyc;
  final bool? push;
  final bool? email;

  Map<String, Object?> toJson() => _$UpdateNotificationPreferencesRequestToJson(this);
}
