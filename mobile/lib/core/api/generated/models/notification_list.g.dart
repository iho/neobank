// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'notification_list.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

NotificationList _$NotificationListFromJson(Map<String, dynamic> json) =>
    NotificationList(
      notifications: (json['notifications'] as List<dynamic>)
          .map((e) => Notification.fromJson(e as Map<String, dynamic>))
          .toList(),
      unreadCount: (json['unread_count'] as num).toInt(),
      nextCursor: json['next_cursor'] as String?,
    );

Map<String, dynamic> _$NotificationListToJson(NotificationList instance) =>
    <String, dynamic>{
      'notifications': instance.notifications,
      'unread_count': instance.unreadCount,
      'next_cursor': instance.nextCursor,
    };
