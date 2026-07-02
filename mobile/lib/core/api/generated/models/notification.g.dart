// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'notification.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

Notification _$NotificationFromJson(Map<String, dynamic> json) => Notification(
  id: json['id'] as String,
  userId: json['user_id'] as String,
  eventType: json['event_type'] as String,
  title: json['title'] as String,
  body: json['body'] as String,
  read: json['read'] as bool,
  createdAt: DateTime.parse(json['created_at'] as String),
);

Map<String, dynamic> _$NotificationToJson(Notification instance) =>
    <String, dynamic>{
      'id': instance.id,
      'user_id': instance.userId,
      'event_type': instance.eventType,
      'title': instance.title,
      'body': instance.body,
      'read': instance.read,
      'created_at': instance.createdAt.toIso8601String(),
    };
