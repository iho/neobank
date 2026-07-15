package dev.horobets.neobanknative.features.kyc

import dev.horobets.neobanknative.core.networking.ApiClient
import dev.horobets.neobanknative.core.networking.HttpMethod
import dev.horobets.neobanknative.core.networking.jsonBodyOf

class KycRepository(private val client: ApiClient) {
    suspend fun status(): KycStatusInfo = client.send("/v1/kyc/status")

    suspend fun submit(
        fullName: String,
        dateOfBirth: String,
        countryCode: String,
        documentType: String?,
        documentNumber: String?,
    ): KycSubmitResult = client.send(
        "/v1/kyc",
        method = HttpMethod.POST,
        body = jsonBodyOf(
            "full_name" to fullName,
            "date_of_birth" to dateOfBirth,
            "country_code" to countryCode,
            "document_type" to documentType?.ifEmpty { null },
            "document_number" to documentNumber?.ifEmpty { null },
        ),
    )
}
