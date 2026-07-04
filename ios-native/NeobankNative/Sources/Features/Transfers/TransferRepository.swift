import Foundation

struct TransferRepository: Sendable {
    let client: APIClient

    /// A 422 here means "declined" — the gateway returns a full transfer
    /// body (with `status`/`failure_reason` set), not an error envelope. The
    /// generic error path doesn't have this body, so decode it from
    /// `APIError.responseData` instead of surfacing a bare failure.
    func createTransfer(
        idempotencyKey: String,
        amount: String,
        recipientPhone: String?,
        recipientEmail: String?,
        recipientUserId: String?,
        currency: String?,
        memo: String?
    ) async throws -> Transfer {
        do {
            return try await client.send(
                "/v1/transfers",
                method: .post,
                body: [
                    "amount": amount,
                    "recipient_phone": recipientPhone?.isEmpty == false ? recipientPhone : nil,
                    "recipient_email": recipientEmail?.isEmpty == false ? recipientEmail : nil,
                    "recipient_user_id": recipientUserId?.isEmpty == false ? recipientUserId : nil,
                    "currency": currency?.isEmpty == false ? currency : nil,
                    "memo": memo?.isEmpty == false ? memo : nil,
                ],
                idempotencyKey: idempotencyKey
            )
        } catch let error as APIError where error.statusCode == 422 {
            if let data = error.responseData, let transfer = try? APIClient.jsonDecoder.decode(Transfer.self, from: data) {
                return transfer
            }
            throw error
        }
    }
}
