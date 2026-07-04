import Foundation
import Observation

enum SessionStatus: Sendable {
    case unknown
    case authenticated
    case unauthenticated
}

@MainActor
@Observable
final class SessionStore: Sendable {
    private(set) var status: SessionStatus = .unknown

    func setStatus(_ status: SessionStatus) {
        self.status = status
    }
}
