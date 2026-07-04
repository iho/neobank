import Foundation
import Observation

@MainActor
@Observable
final class KycController {
    enum LoadState {
        case loading
        case loaded(KycStatusInfo)
        case failed(APIError)
    }

    private(set) var state: LoadState = .loading
    private(set) var hasSubmittedThisSession = false

    private let repository: KycRepository

    init(repository: KycRepository) {
        self.repository = repository
    }

    func load() async {
        state = .loading
        hasSubmittedThisSession = false
        do {
            state = .loaded(try await repository.status())
        } catch let error as APIError {
            state = .failed(error)
        } catch {
            state = .failed(.decoding())
        }
    }

    func submit(
        fullName: String,
        dateOfBirth: String,
        countryCode: String,
        documentType: String?,
        documentNumber: String?
    ) async throws {
        let result = try await repository.submit(
            fullName: fullName,
            dateOfBirth: dateOfBirth,
            countryCode: countryCode,
            documentType: documentType,
            documentNumber: documentNumber
        )
        hasSubmittedThisSession = true
        state = .loaded(KycStatusInfo(status: result.status, rejectionReason: result.rejectionReason))
    }
}

#if DEBUG
extension KycController {
    static var preview: KycController {
        KycController(repository: KycRepository(client: APIClient(baseURL: AppConfig.apiBaseURL)))
    }
}
#endif
