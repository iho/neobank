package dev.horobets.neobanknative.core.networking

import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.put

/**
 * `[String: Any?]`-style request body builder: null values are omitted
 * entirely, mirroring `body.compactMapValues { $0 }` on the iOS client.
 */
fun jsonBodyOf(vararg pairs: Pair<String, Any?>): JsonObject = buildJsonObject {
    for ((key, value) in pairs) {
        when (value) {
            null -> Unit
            is String -> put(key, value)
            is Boolean -> put(key, value)
            is Int -> put(key, value)
            is Long -> put(key, value)
            is Double -> put(key, value)
            else -> put(key, value.toString())
        }
    }
}
