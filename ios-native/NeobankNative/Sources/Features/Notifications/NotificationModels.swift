import Foundation

struct AppNotification: Decodable, Sendable, Identifiable, Equatable {
    let id: String
    let userId: String
    let eventType: String
    let title: String
    let body: String
    let read: Bool
    let createdAt: Date

    func withRead(_ read: Bool) -> AppNotification {
        AppNotification(id: id, userId: userId, eventType: eventType, title: title, body: body, read: read, createdAt: createdAt)
    }
}

struct NotificationPage: Decodable, Sendable {
    var notifications: [AppNotification]
    var unreadCount: Int
    var nextCursor: String?
}
