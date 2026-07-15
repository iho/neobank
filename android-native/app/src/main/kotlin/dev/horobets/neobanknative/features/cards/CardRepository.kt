package dev.horobets.neobanknative.features.cards

import dev.horobets.neobanknative.core.networking.ApiClient
import dev.horobets.neobanknative.core.networking.HttpMethod
import dev.horobets.neobanknative.core.networking.jsonBodyOf
import kotlinx.serialization.Serializable

class CardRepository(private val client: ApiClient) {
    @Serializable
    private data class CardsResponse(val cards: List<BankCard>)

    suspend fun list(): List<BankCard> = client.send<CardsResponse>("/v1/cards").cards

    suspend fun issue(cardholderName: String, dailyLimit: String?, onlineOnly: Boolean?): BankCard = client.send(
        "/v1/cards",
        method = HttpMethod.POST,
        body = jsonBodyOf(
            "cardholder_name" to cardholderName,
            "daily_limit" to dailyLimit?.ifEmpty { null },
            "online_only" to onlineOnly,
        ),
    )

    suspend fun freeze(id: String): BankCard = client.send("/v1/cards/$id/freeze", method = HttpMethod.POST)

    suspend fun unfreeze(id: String): BankCard = client.send("/v1/cards/$id/unfreeze", method = HttpMethod.POST)
}
