// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

import 'notification.dart';

part 'notification_list.g.dart';

@JsonSerializable()
class NotificationList {
  const NotificationList({
    required this.notifications,
    required this.unreadCount,
    this.nextCursor,
  });
  
  factory NotificationList.fromJson(Map<String, Object?> json) => _$NotificationListFromJson(json);
  
  final List<Notification> notifications;
  @JsonKey(name: 'unread_count')
  final int unreadCount;
  @JsonKey(name: 'next_cursor')
  final String? nextCursor;

  Map<String, Object?> toJson() => _$NotificationListToJson(this);
}
