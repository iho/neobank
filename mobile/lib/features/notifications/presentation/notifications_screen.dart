import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../domain/notification_models.dart';
import 'notifications_controller.dart';

class NotificationsScreen extends ConsumerWidget {
  const NotificationsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(notificationsControllerProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Notifications'),
        actions: [
          TextButton(
            onPressed: () => ref.read(notificationsControllerProvider.notifier).markAllRead(),
            child: const Text('Mark all read'),
          ),
        ],
      ),
      body: state.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (error, _) => Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text('$error'),
              const SizedBox(height: 12),
              OutlinedButton(
                onPressed: () => ref.read(notificationsControllerProvider.notifier).refresh(),
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
        data: (page) => page.notifications.isEmpty
            ? const Center(child: Text('No notifications yet'))
            : RefreshIndicator(
                onRefresh: () => ref.read(notificationsControllerProvider.notifier).refresh(),
                child: ListView.builder(
                  itemCount: page.notifications.length,
                  itemBuilder: (context, index) {
                    final n = page.notifications[index];
                    return _NotificationTile(notification: n);
                  },
                ),
              ),
      ),
    );
  }
}

class _NotificationTile extends ConsumerWidget {
  const _NotificationTile({required this.notification});

  final AppNotification notification;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return ListTile(
      leading: Icon(
        notification.read ? Icons.notifications_none : Icons.notifications_active,
        color: notification.read ? null : Theme.of(context).colorScheme.primary,
      ),
      title: Text(
        notification.title,
        style: TextStyle(fontWeight: notification.read ? FontWeight.normal : FontWeight.bold),
      ),
      subtitle: Text(notification.body),
      trailing: Text(notification.createdAt.toLocal().toString().split('.').first),
      onTap: notification.read
          ? null
          : () => ref.read(notificationsControllerProvider.notifier).markRead(notification.id),
    );
  }
}
