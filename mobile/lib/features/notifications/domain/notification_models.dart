/// Named `AppNotification` (not `Notification`) to avoid colliding with
/// `package:flutter`'s `Notification`/`NotificationListener` widgets.
class AppNotification {
  const AppNotification({
    required this.id,
    required this.userId,
    required this.eventType,
    required this.title,
    required this.body,
    required this.read,
    required this.createdAt,
  });

  factory AppNotification.fromJson(Map<String, dynamic> json) => AppNotification(
        id: json['id'] as String,
        userId: json['user_id'] as String,
        eventType: json['event_type'] as String,
        title: json['title'] as String,
        body: json['body'] as String,
        read: json['read'] as bool,
        createdAt: DateTime.parse(json['created_at'] as String),
      );

  final String id;
  final String userId;
  final String eventType;
  final String title;
  final String body;
  final bool read;
  final DateTime createdAt;

  AppNotification copyWithRead(bool read) => AppNotification(
        id: id,
        userId: userId,
        eventType: eventType,
        title: title,
        body: body,
        read: read,
        createdAt: createdAt,
      );
}

class NotificationPage {
  const NotificationPage({
    required this.notifications,
    required this.unreadCount,
    this.nextCursor,
  });

  factory NotificationPage.fromJson(Map<String, dynamic> json) => NotificationPage(
        notifications: (json['notifications'] as List<dynamic>)
            .map((e) => AppNotification.fromJson(e as Map<String, dynamic>))
            .toList(),
        unreadCount: json['unread_count'] as int,
        nextCursor: json['next_cursor'] as String?,
      );

  final List<AppNotification> notifications;
  final int unreadCount;
  final String? nextCursor;
}
