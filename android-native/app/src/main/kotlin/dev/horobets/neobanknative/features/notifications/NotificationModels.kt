package dev.horobets.neobanknative.features.notifications

import kotlinx.serialization.Serializable

@Serializable
data class AppNotification(
    val id: String,
    val userId: String,
    val eventType: String,
    val title: String,
    val body: String,
    val read: Boolean,
    val createdAt: String,
)

@Serializable
data class NotificationPage(
    val notifications: List<AppNotification>,
    val unreadCount: Int,
    val nextCursor: String? = null,
)
