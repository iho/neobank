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

    /// Bumped on every transition. User-scoped feature controllers key their
    /// `.task(id:)` reloads off this instead of `status` directly — `status`
    /// stays `.authenticated` across a logout+login-as-someone-else cycle, so
    /// it wouldn't by itself signal that cached data needs to be dropped.
    private(set) var generation = UUID()

    func setStatus(_ status: SessionStatus) {
        self.status = status
        generation = UUID()
    }
}
