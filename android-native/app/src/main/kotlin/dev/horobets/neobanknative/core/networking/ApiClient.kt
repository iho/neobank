package dev.horobets.neobanknative.core.networking

import java.io.IOException
import java.net.SocketTimeoutException
import java.util.UUID
import java.util.concurrent.TimeUnit
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.ExperimentalSerializationApi
import kotlinx.serialization.Serializable
import kotlinx.serialization.decodeFromString
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonNamingStrategy
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.contentOrNull
import okhttp3.HttpUrl.Companion.toHttpUrl
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import okhttp3.Response

enum class HttpMethod(val rawValue: String) {
    GET("GET"),
    POST("POST"),
    PUT("PUT"),
    PATCH("PATCH"),
    DELETE("DELETE"),
}

@Serializable
object EmptyResponse

/**
 * Talks to the gateway BFF. Two instances are wired up in practice: one with
 * `auth == null` for `/v1/auth/login` and `/v1/auth/register` (which must
 * not try to attach or refresh a bearer token), and one with `auth` set for
 * every other call.
 */
@OptIn(ExperimentalSerializationApi::class)
class ApiClient(
    private val baseUrl: String,
    private val httpClient: OkHttpClient = sharedHttpClient,
    private val auth: AuthContext? = null,
) {
    class AuthContext(
        val tokenProvider: () -> String?,
        val refresh: suspend () -> String?,
        val onSessionExpired: suspend () -> Unit,
    )

    private val refreshCoordinator = RefreshCoordinator()

    suspend inline fun <reified T> send(
        path: String,
        method: HttpMethod = HttpMethod.GET,
        query: Map<String, String?>? = null,
        body: JsonObject? = null,
        idempotencyKey: String? = null,
    ): T {
        val responseBody = executeRequest(path, method, query, body, idempotencyKey)
        if (T::class == EmptyResponse::class) {
            @Suppress("UNCHECKED_CAST")
            return EmptyResponse as T
        }
        return try {
            json.decodeFromString(responseBody)
        } catch (e: Exception) {
            throw ApiError.decoding()
        }
    }

    suspend fun executeRequest(
        path: String,
        method: HttpMethod,
        query: Map<String, String?>?,
        body: JsonObject?,
        idempotencyKey: String?,
    ): String = executeRequest(path, method, query, body, idempotencyKey, isRetryAfterRefresh = false)

    private suspend fun executeRequest(
        path: String,
        method: HttpMethod,
        query: Map<String, String?>?,
        body: JsonObject?,
        idempotencyKey: String?,
        isRetryAfterRefresh: Boolean,
    ): String = withContext(Dispatchers.IO) {
        val request = buildRequest(path, method, query, body, idempotencyKey)

        val response: Response = try {
            httpClient.newCall(request).execute()
        } catch (e: SocketTimeoutException) {
            throw ApiError.timeout()
        } catch (e: IOException) {
            throw ApiError.network()
        }

        response.use { resp ->
            val bodyString = resp.body?.string().orEmpty()

            if (resp.code == 401 && auth != null && !path.contains("/v1/auth/") && !isRetryAfterRefresh) {
                val refreshed = try {
                    refreshCoordinator.refresh(auth.refresh)
                } catch (e: Exception) {
                    null
                }
                if (refreshed == null) {
                    auth.onSessionExpired()
                    throw ApiError.unauthenticated()
                }
                return@withContext executeRequest(path, method, query, body, idempotencyKey, isRetryAfterRefresh = true)
            }

            if (resp.code !in 200..299) {
                throw mapFailure(resp.code, bodyString, resp)
            }

            bodyString
        }
    }

    private fun buildRequest(
        path: String,
        method: HttpMethod,
        query: Map<String, String?>?,
        body: JsonObject?,
        idempotencyKey: String?,
    ): Request {
        val urlBuilder = baseUrl.toHttpUrl().newBuilder()
        path.trim('/').split("/").forEach { segment ->
            if (segment.isNotEmpty()) urlBuilder.addPathSegment(segment)
        }
        query?.forEach { (key, value) -> if (value != null) urlBuilder.addQueryParameter(key, value) }

        val requestBuilder = Request.Builder().url(urlBuilder.build())
        requestBuilder.addHeader("Content-Type", "application/json")
        requestBuilder.addHeader("X-Correlation-Id", UUID.randomUUID().toString())
        if (method in mutatingMethods) {
            requestBuilder.addHeader("Idempotency-Key", idempotencyKey ?: UUID.randomUUID().toString())
        }
        auth?.tokenProvider?.invoke()?.let { token -> requestBuilder.addHeader("Authorization", "Bearer $token") }

        val requestBody = if (method == HttpMethod.GET) {
            null
        } else {
            val payload = body ?: JsonObject(emptyMap())
            json.encodeToString(JsonObject.serializer(), payload).toRequestBody(jsonMediaType)
        }
        requestBuilder.method(method.rawValue, requestBody)
        return requestBuilder.build()
    }

    private fun mapFailure(statusCode: Int, bodyString: String, response: Response): ApiError {
        if (statusCode == 401) return ApiError.unauthenticated()

        var message = "Something went wrong. Please try again."
        try {
            val root = json.parseToJsonElement(bodyString)
            val apiMessage = (root as? JsonObject)?.get("error") as? JsonPrimitive
            apiMessage?.contentOrNull?.let { message = it }
        } catch (e: Exception) {
            // Non-JSON or unexpected body shape — fall back to the generic message.
        }
        val correlationId = response.header("x-correlation-id")
        return ApiError(message = message, statusCode = statusCode, correlationId = correlationId, responseBody = bodyString)
    }

    companion object {
        private val mutatingMethods = setOf(HttpMethod.POST, HttpMethod.PUT, HttpMethod.PATCH, HttpMethod.DELETE)
        private val jsonMediaType = "application/json".toMediaType()

        /** convertFromSnakeCase equivalent — the gateway's field naming maps directly, no per-model CodingKeys/@SerialName. */
        val json: Json = Json {
            ignoreUnknownKeys = true
            namingStrategy = JsonNamingStrategy.SnakeCase
        }

        val sharedHttpClient: OkHttpClient by lazy {
            OkHttpClient.Builder()
                .connectTimeout(15, TimeUnit.SECONDS)
                .readTimeout(30, TimeUnit.SECONDS)
                .writeTimeout(30, TimeUnit.SECONDS)
                .build()
        }
    }
}
