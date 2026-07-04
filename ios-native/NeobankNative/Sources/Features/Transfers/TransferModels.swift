import Foundation

struct Transfer: Decodable, Sendable, Equatable {
    let id: String?
    let status: String?
    let senderUserId: String?
    let recipientUserId: String?
    let amount: String?
    let currency: String?
    let failureReason: String?
    let memo: String?
    let createdAt: Date?
    let completedAt: Date?

    var isCompleted: Bool { status == "completed" }
    var isFailed: Bool { status == "failed" || status == "declined" }
}
