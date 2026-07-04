import Foundation

struct KycRepository: Sendable {
    let client: APIClient

    func status() async throws -> KycStatusInfo {
        try await client.send("/v1/kyc/status")
    }

    func submit(
        fullName: String,
        dateOfBirth: String,
        countryCode: String,
        documentType: String?,
        documentNumber: String?
    ) async throws -> KycSubmitResult {
        try await client.send(
            "/v1/kyc",
            method: .post,
            body: [
                "full_name": fullName,
                "date_of_birth": dateOfBirth,
                "country_code": countryCode,
                "document_type": documentType?.isEmpty == false ? documentType : nil,
                "document_number": documentNumber?.isEmpty == false ? documentNumber : nil,
            ]
        )
    }
}
