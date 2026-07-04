import Foundation
import Observation

@MainActor
@Observable
final class NotificationsController {
    enum LoadState {
        case loading
        case loaded(NotificationPage)
        case failed(APIError)
    }

    private(set) var state: LoadState = .loading

    private let repository: NotificationRepository

    init(repository: NotificationRepository) {
        self.repository = repository
    }

    func load() async {
        state = .loading
        do {
            state = .loaded(try await repository.list())
        } catch let error as APIError {
            state = .failed(error)
        } catch {
            state = .failed(.decoding())
        }
    }

    /// Used by the background poll loop: keeps the last good page on screen
    /// instead of flashing a spinner or error every 30 seconds on a blip.
    func refresh() async {
        guard let page = try? await repository.list() else { return }
        state = .loaded(page)
    }

    func markRead(_ id: String) async throws {
        guard case .loaded(var page) = state else { return }
        let updated = try await repository.markRead(id)
        if let index = page.notifications.firstIndex(where: { $0.id == id }) {
            page.notifications[index] = updated
        }
        page.unreadCount = max(0, page.unreadCount - (updated.read ? 1 : 0))
        state = .loaded(page)
    }

    func markAllRead() async throws {
        guard case .loaded(var page) = state else { return }
        try await repository.markAllRead()
        page.notifications = page.notifications.map { $0.withRead(true) }
        page.unreadCount = 0
        state = .loaded(page)
    }
}

#if DEBUG
extension NotificationsController {
    static var preview: NotificationsController {
        NotificationsController(repository: NotificationRepository(client: APIClient(baseURL: AppConfig.apiBaseURL)))
    }
}
#endif
