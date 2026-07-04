import Foundation

/// Coalesces concurrent 401-triggered refresh attempts into a single in-flight
/// request, mirroring the Flutter app's `_refreshFuture` memoization — without
/// it, N parallel requests failing with 401 would each kick off their own
/// refresh call and race to persist tokens.
actor RefreshCoordinator {
    private var inFlight: Task<String?, Error>?

    func refresh(_ operation: @escaping @Sendable () async throws -> String?) async throws -> String? {
        if let inFlight {
            return try await inFlight.value
        }
        let task = Task { try await operation() }
        inFlight = task
        defer { inFlight = nil }
        return try await task.value
    }
}
