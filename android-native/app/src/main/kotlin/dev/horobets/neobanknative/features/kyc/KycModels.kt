package dev.horobets.neobanknative.features.kyc

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

enum class KycStatus {
    PENDING,
    APPROVED,
    REJECTED,
    MANUAL_REVIEW,
    ;

    companion object {
        fun fromRaw(raw: String): KycStatus = when (raw) {
            "approved" -> APPROVED
            "rejected" -> REJECTED
            "manual_review" -> MANUAL_REVIEW
            else -> PENDING
        }
    }
}

@Serializable
data class KycStatusInfo(
    @SerialName("status") val rawStatus: String,
    val rejectionReason: String? = null,
) {
    val status: KycStatus get() = KycStatus.fromRaw(rawStatus)
}

@Serializable
data class KycSubmitResult(
    val kycCaseId: String,
    @SerialName("status") val rawStatus: String,
    val walletId: String? = null,
    val rejectionReason: String? = null,
) {
    val status: KycStatus get() = KycStatus.fromRaw(rawStatus)
}
