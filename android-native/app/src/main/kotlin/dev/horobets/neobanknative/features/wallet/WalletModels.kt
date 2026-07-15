package dev.horobets.neobanknative.features.wallet

import kotlinx.serialization.Serializable

/**
 * Amounts are decimal strings straight from the ledger (never [Double], to
 * avoid floating-point drift) — formatted for display, never parsed and
 * re-computed on-device.
 */
@Serializable
data class WalletBalance(
    val walletId: String,
    val currency: String,
    val balance: String,
    val availableBalance: String,
    val encumberedBalance: String? = null,
)

@Serializable
data class WalletTransaction(
    val id: String,
    val type: String,
    val amount: String,
    val currency: String,
    val direction: String,
    val status: String,
    val createdAt: String,
    val counterparty: String? = null,
    val memo: String? = null,
    val referenceId: String? = null,
) {
    val isCredit: Boolean get() = direction == "credit"
}

@Serializable
data class WalletTransactionPage(
    val transactions: List<WalletTransaction>,
    val nextCursor: String? = null,
)
