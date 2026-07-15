package dev.horobets.neobanknative.features.transfers

import dev.horobets.neobanknative.core.networking.ApiClient
import dev.horobets.neobanknative.core.networking.ApiError
import dev.horobets.neobanknative.core.networking.HttpMethod
import dev.horobets.neobanknative.core.networking.jsonBodyOf
import kotlinx.serialization.decodeFromString

class TransferRepository(private val client: ApiClient) {
    /**
     * A 422 here means "declined" — the gateway returns a full transfer
     * body (with `status`/`failure_reason` set), not an error envelope. The
     * generic error path doesn't have this body, so decode it from
     * [ApiError.responseBody] instead of surfacing a bare failure.
     */
    suspend fun createTransfer(
        idempotencyKey: String,
        amount: String,
        recipientPhone: String?,
        recipientEmail: String?,
        recipientUserId: String?,
        currency: String?,
        memo: String?,
    ): Transfer = try {
        client.send(
            "/v1/transfers",
            method = HttpMethod.POST,
            body = jsonBodyOf(
                "amount" to amount,
                "recipient_phone" to recipientPhone?.ifEmpty { null },
                "recipient_email" to recipientEmail?.ifEmpty { null },
                "recipient_user_id" to recipientUserId?.ifEmpty { null },
                "currency" to currency?.ifEmpty { null },
                "memo" to memo?.ifEmpty { null },
            ),
            idempotencyKey = idempotencyKey,
        )
    } catch (e: ApiError) {
        val body = e.responseBody
        if (e.statusCode == 422 && body != null) {
            try {
                ApiClient.json.decodeFromString<Transfer>(body)
            } catch (decodeError: Exception) {
                throw e
            }
        } else {
            throw e
        }
    }
}
