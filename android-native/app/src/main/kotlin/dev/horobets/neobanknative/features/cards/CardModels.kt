package dev.horobets.neobanknative.features.cards

import kotlinx.serialization.Serializable

@Serializable
data class BankCard(
    val id: String,
    val userId: String,
    val walletId: String,
    val lastFour: String,
    val status: String,
    val expiryMonth: Int,
    val expiryYear: Int,
    val onlineOnly: Boolean,
    val dailyLimit: String? = null,
) {
    val isFrozen: Boolean get() = status == "frozen"

    /** Plain "MM/YYYY", no locale-aware grouping of the year. */
    val expiry: String get() = "%02d/%d".format(expiryMonth, expiryYear)
}
