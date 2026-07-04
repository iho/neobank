import Foundation
import Observation

@MainActor
@Observable
final class TransferSubmitController {
    enum SubmitState {
        case idle
        case submitting
        case result(Transfer)
        case failed(APIError)
    }

    private(set) var state: SubmitState = .idle

    /// Stable across retries so a network hiccup can't double-send — only
    /// rotated once a submission reaches a definitive server-side outcome
    /// (a `Transfer` back, completed or declined), mirroring the Flutter
    /// controller's `_idempotencyKey` handling.
    private var idempotencyKey = UUID().uuidString

    private let repository: TransferRepository

    init(repository: TransferRepository) {
        self.repository = repository
    }

    func reset() {
        state = .idle
        idempotencyKey = UUID().uuidString
    }

    func submit(amount: String, recipientPhone: String?, recipientEmail: String?, memo: String?) async {
        if case .result = state {
            idempotencyKey = UUID().uuidString
        }
        state = .submitting
        do {
            let transfer = try await repository.createTransfer(
                idempotencyKey: idempotencyKey,
                amount: amount,
                recipientPhone: recipientPhone,
                recipientEmail: recipientEmail,
                recipientUserId: nil,
                currency: nil,
                memo: memo
            )
            state = .result(transfer)
        } catch let error as APIError {
            state = .failed(error)
        } catch {
            state = .failed(.decoding())
        }
    }
}

#if DEBUG
extension TransferSubmitController {
    static var preview: TransferSubmitController {
        TransferSubmitController(repository: TransferRepository(client: APIClient(baseURL: AppConfig.apiBaseURL)))
    }
}
#endif
