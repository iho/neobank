import Foundation
import Observation

@MainActor
@Observable
final class CardsController {
    enum LoadState {
        case loading
        case loaded([BankCard])
        case failed(APIError)
    }

    private(set) var state: LoadState = .loading

    private let repository: CardRepository

    init(repository: CardRepository) {
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

    func issueCard(cardholderName: String, dailyLimit: String?, onlineOnly: Bool?) async throws {
        _ = try await repository.issue(cardholderName: cardholderName, dailyLimit: dailyLimit, onlineOnly: onlineOnly)
        await load()
    }

    func toggleFreeze(_ card: BankCard) async throws {
        let updated = card.isFrozen ? try await repository.unfreeze(card.id) : try await repository.freeze(card.id)
        guard case .loaded(var cards) = state, let index = cards.firstIndex(where: { $0.id == updated.id }) else { return }
        cards[index] = updated
        state = .loaded(cards)
    }
}

#if DEBUG
extension CardsController {
    static var preview: CardsController {
        CardsController(repository: CardRepository(client: APIClient(baseURL: AppConfig.apiBaseURL)))
    }
}
#endif
