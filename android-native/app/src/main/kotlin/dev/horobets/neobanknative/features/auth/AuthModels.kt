package dev.horobets.neobanknative.features.auth

import kotlinx.serialization.Serializable

@Serializable
data class AuthTokens(
    val userId: String,
    val accessToken: String,
    val refreshToken: String,
)

@Serializable
data class Profile(
    val userId: String,
    val email: String,
    val phone: String,
    val status: String,
    val kycStatus: String,
    val createdAt: String,
    val fullName: String? = null,
    val dateOfBirth: String? = null,
    val countryCode: String? = null,
)
