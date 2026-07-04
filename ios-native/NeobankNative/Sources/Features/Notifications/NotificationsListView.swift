import SwiftUI

struct NotificationsListView: View {
    @Environment(SessionStore.self) private var sessionStore
    @Environment(NotificationsController.self) private var notificationsController

    var body: some View {
        NavigationStack {
            ZStack {
                BrandBackground()

                switch notificationsController.state {
                case .loading:
                    ProgressView()
                case .failed(let error):
                    NotificationsErrorView(error: error)
                case .loaded(let page):
                    if page.notifications.isEmpty {
                        emptyState
                    } else {
                        list(page)
                    }
                }
            }
            .navigationTitle("Alerts")
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button("Mark all read") {
                        Task { try? await notificationsController.markAllRead() }
                    }
                    .disabled(unreadCount == 0)
                }
            }
        }
        .task(id: sessionStore.generation) {
            await notificationsController.load()
            while !Task.isCancelled {
                try? await Task.sleep(for: .seconds(30))
                if Task.isCancelled { break }
                await notificationsController.refresh()
            }
        }
    }

    private var unreadCount: Int {
        if case .loaded(let page) = notificationsController.state { return page.unreadCount }
        return 0
    }

    private var emptyState: some View {
        VStack(spacing: 16) {
            GlowIcon(systemName: "bell", diameter: 72, iconSize: 32)
            Text("No notifications yet")
                .font(.title3.bold())
        }
    }

    private func list(_ page: NotificationPage) -> some View {
        ScrollView {
            VStack(spacing: 12) {
                ForEach(page.notifications) { notification in
                    NotificationRow(notification: notification) {
                        Task { try? await notificationsController.markRead(notification.id) }
                    }
                }
            }
            .padding(20)
        }
        .refreshable { await notificationsController.load() }
    }
}

private struct NotificationRow: View {
    let notification: AppNotification
    let onTap: () -> Void

    var body: some View {
        Button(action: onTap) {
            HStack(alignment: .top, spacing: 14) {
                ZStack {
                    Circle()
                        .fill((notification.read ? Color.secondary : Color.blue).opacity(0.15))
                        .frame(width: 40, height: 40)
                    Image(systemName: notification.read ? "bell" : "bell.badge.fill")
                        .font(.subheadline.weight(.semibold))
                        .foregroundStyle(notification.read ? Color.secondary : Color.blue)
                }

                VStack(alignment: .leading, spacing: 4) {
                    Text(notification.title)
                        .font(notification.read ? .subheadline : .subheadline.weight(.bold))
                    Text(notification.body)
                        .font(.footnote)
                        .foregroundStyle(.secondary)
                        .lineLimit(2)
                    Text(Self.dateFormatter.string(from: notification.createdAt))
                        .font(.caption2)
                        .foregroundStyle(.secondary)
                }

                Spacer(minLength: 0)
            }
            .padding(12)
            .surfaceCard(cornerRadius: 14)
        }
        .buttonStyle(.plain)
        .disabled(notification.read)
    }

    private static let dateFormatter: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateStyle = .medium
        formatter.timeStyle = .short
        return formatter
    }()
}

private struct NotificationsErrorView: View {
    let error: APIError

    @Environment(NotificationsController.self) private var notificationsController

    var body: some View {
        VStack(spacing: 12) {
            Text(error.message).multilineTextAlignment(.center)
            Button("Retry") {
                Task { await notificationsController.load() }
            }
            .buttonStyle(.bordered)
        }
        .padding()
    }
}

#Preview {
    NotificationsListView()
        .environment(SessionStore())
        .environment(NotificationsController.preview)
}
