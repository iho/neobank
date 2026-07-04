import Foundation
import Observation

@MainActor
@Observable
final class WalletHomeController {
    struct Snapshot {
        var balance: WalletBalance
        var transactions: [WalletTransaction]
        var nextCursor: String?
        var isLoadingMore = false

        var hasMore: Bool { nextCursor != nil }
    }

    enum LoadState {
        case loading
        case loaded(Snapshot)
        case failed(APIError)
    }

    private(set) var state: LoadState = .loading

    private let repository: WalletRepository

    init(repository: WalletRepository) {
        self.repository = repository
    }

    func load() async {
        state = .loading
        do {
            async let balanceTask = repository.balance()
            async let pageTask = repository.transactions()
            let balance = try await balanceTask
            let page = try await pageTask
            state = .loaded(Snapshot(balance: balance, transactions: page.transactions, nextCursor: page.nextCursor))
        } catch let error as APIError {
            state = .failed(error)
        } catch {
            state = .failed(.decoding())
        }
    }

    /// Fetches the next page when the last visible row appears. A no-op
    /// (rather than an error) when there's nothing more or a fetch is
    /// already in flight, since callers trigger this from scroll position
    /// and can't easily guard it themselves.
    func loadMore() async {
        guard case .loaded(var snapshot) = state, snapshot.hasMore, !snapshot.isLoadingMore else { return }
        snapshot.isLoadingMore = true
        state = .loaded(snapshot)
        do {
            let page = try await repository.transactions(cursor: snapshot.nextCursor)
            snapshot.transactions.append(contentsOf: page.transactions)
            snapshot.nextCursor = page.nextCursor
        } catch {
            // Leave existing transactions in place; the user can retry by scrolling again.
        }
        snapshot.isLoadingMore = false
        state = .loaded(snapshot)
    }
}

#if DEBUG
extension WalletHomeController {
    static var preview: WalletHomeController {
        WalletHomeController(repository: WalletRepository(client: APIClient(baseURL: AppConfig.apiBaseURL)))
    }
}
#endif
