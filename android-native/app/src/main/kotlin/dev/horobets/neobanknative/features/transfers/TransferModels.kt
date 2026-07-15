package dev.horobets.neobanknative.features.transfers

import kotlinx.serialization.Serializable

@Serializable
data class Transfer(
    val id: String? = null,
    val status: String? = null,
    val senderUserId: String? = null,
    val recipientUserId: String? = null,
    val amount: String? = null,
    val currency: String? = null,
    val failureReason: String? = null,
    val memo: String? = null,
    val createdAt: String? = null,
    val completedAt: String? = null,
) {
    val isCompleted: Boolean get() = status == "completed"
    val isFailed: Boolean get() = status == "failed" || status == "declined"
}
