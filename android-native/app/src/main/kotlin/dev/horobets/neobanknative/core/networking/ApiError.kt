package dev.horobets.neobanknative.core.networking

class ApiError(
    message: String,
    val statusCode: Int?,
    val correlationId: String?,
    /**
     * Raw response body for non-2xx responses. Most call sites ignore this
     * — it exists for the rare case (e.g. `POST /v1/transfers` returning a
     * 422 with a full declined-transfer payload, not an error envelope)
     * where the caller needs to decode the body itself instead of treating
     * the status code as a bare failure.
     */
    val responseBody: String? = null,
) : Exception(message) {

    companion object {
        fun network() = ApiError(
            message = "No connection. Check your network and try again.",
            statusCode = null,
            correlationId = null,
        )

        fun timeout() = ApiError(
            message = "The request timed out. Please try again.",
            statusCode = null,
            correlationId = null,
        )

        fun unauthenticated() = ApiError(
            message = "Your session has expired. Please log in again.",
            statusCode = 401,
            correlationId = null,
        )

        fun decoding() = ApiError(
            message = "Something went wrong. Please try again.",
            statusCode = null,
            correlationId = null,
        )
    }
}
