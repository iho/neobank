package dev.horobets.neobanknative.features.notifications

import dev.horobets.neobanknative.core.networking.ApiClient
import dev.horobets.neobanknative.core.networking.HttpMethod
import kotlinx.serialization.Serializable

class NotificationRepository(private val client: ApiClient) {
    @Serializable
    private data class MarkAllReadResponse(val markedCount: Int)

    suspend fun list(cursor: String? = null, limit: Int = 20): NotificationPage = client.send(
        "/v1/notifications",
        query = mapOf("limit" to limit.toString(), "cursor" to cursor),
    )

    suspend fun markRead(id: String): AppNotification = client.send("/v1/notifications/$id/read", method = HttpMethod.POST)

    suspend fun markAllRead(): Int =
        client.send<MarkAllReadResponse>("/v1/notifications/read-all", method = HttpMethod.POST).markedCount
}
