import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/notification_repository.dart';
import '../domain/notification_models.dart';

final notificationsControllerProvider =
    AsyncNotifierProvider<NotificationsController, NotificationPage>(
  NotificationsController.new,
);

/// Polls every 30s while this provider has a listener (i.e. the
/// notifications tab is visible or mounted in the tree). Push notifications
/// are Phase 4 — see mobile/TODO.md — this is the interim "good enough" inbox.
class NotificationsController extends AsyncNotifier<NotificationPage> {
  Timer? _pollTimer;

  @override
  Future<NotificationPage> build() async {
    _pollTimer?.cancel();
    _pollTimer = Timer.periodic(const Duration(seconds: 30), (_) => refresh());
    ref.onDispose(() => _pollTimer?.cancel());
    return ref.read(notificationRepositoryProvider).list();
  }

  Future<void> refresh() async {
    final result = await AsyncValue.guard(() => ref.read(notificationRepositoryProvider).list());
    // A background poll shouldn't blow away a good list with a transient
    // error — only apply it if we don't already have data, or it succeeded.
    if (result.hasValue || !state.hasValue) {
      state = result;
    }
  }

  Future<void> markRead(String id) async {
    final current = state.valueOrNull;
    if (current == null) return;
    final updated = await ref.read(notificationRepositoryProvider).markRead(id);
    state = AsyncData(
      NotificationPage(
        notifications: [
          for (final n in current.notifications) if (n.id == id) updated else n,
        ],
        unreadCount: (current.unreadCount - (updated.read ? 1 : 0)).clamp(0, 1 << 31),
        nextCursor: current.nextCursor,
      ),
    );
  }

  Future<void> markAllRead() async {
    final current = state.valueOrNull;
    if (current == null) return;
    await ref.read(notificationRepositoryProvider).markAllRead();
    state = AsyncData(
      NotificationPage(
        notifications: [for (final n in current.notifications) n.copyWithRead(true)],
        unreadCount: 0,
        nextCursor: current.nextCursor,
      ),
    );
  }
}
