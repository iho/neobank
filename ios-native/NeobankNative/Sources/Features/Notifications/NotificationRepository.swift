import Foundation

struct NotificationRepository: Sendable {
    let client: APIClient

    private struct MarkAllReadResponse: Decodable, Sendable {
        let markedCount: Int
    }

    func list(cursor: String? = nil, limit: Int = 20) async throws -> NotificationPage {
        try await client.send("/v1/notifications", query: ["limit": String(limit), "cursor": cursor])
    }

    func markRead(_ id: String) async throws -> AppNotification {
        try await client.send("/v1/notifications/\(id)/read", method: .post)
    }

    @discardableResult
    func markAllRead() async throws -> Int {
        let response: MarkAllReadResponse = try await client.send("/v1/notifications/read-all", method: .post)
        return response.markedCount
    }
}
