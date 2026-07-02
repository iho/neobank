// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'notification.g.dart';

@JsonSerializable()
class Notification {
  const Notification({
    required this.id,
    required this.userId,
    required this.eventType,
    required this.title,
    required this.body,
    required this.read,
    required this.createdAt,
  });
  
  factory Notification.fromJson(Map<String, Object?> json) => _$NotificationFromJson(json);
  
  final String id;
  @JsonKey(name: 'user_id')
  final String userId;
  @JsonKey(name: 'event_type')
  final String eventType;
  final String title;
  final String body;
  final bool read;
  @JsonKey(name: 'created_at')
  final DateTime createdAt;

  Map<String, Object?> toJson() => _$NotificationToJson(this);
}
