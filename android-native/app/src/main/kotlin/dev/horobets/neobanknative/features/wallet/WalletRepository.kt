package dev.horobets.neobanknative.features.wallet

import dev.horobets.neobanknative.core.networking.ApiClient

class WalletRepository(private val client: ApiClient) {
    suspend fun balance(currency: String = "USD"): WalletBalance =
        client.send("/v1/wallet", query = mapOf("currency" to currency))

    suspend fun transactions(cursor: String? = null, limit: Int = 20): WalletTransactionPage = client.send(
        "/v1/wallet/transactions",
        query = mapOf("limit" to limit.toString(), "cursor" to cursor),
    )
}
