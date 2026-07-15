package dev.horobets.neobanknative.core.networking

import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Deferred
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.async
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock

/**
 * Coalesces concurrent 401-triggered refresh attempts into a single
 * in-flight request, mirroring the Flutter app's `_refreshFuture`
 * memoization (and ios-native's `actor RefreshCoordinator`) — without it, N
 * parallel requests failing with 401 would each kick off their own refresh
 * call and race to persist tokens.
 */
class RefreshCoordinator {
    private val scope = CoroutineScope(SupervisorJob())
    private val mutex = Mutex()
    private var inFlight: Deferred<String?>? = null

    suspend fun refresh(operation: suspend () -> String?): String? {
        val deferred = mutex.withLock {
            inFlight ?: scope.async { operation() }.also { inFlight = it }
        }
        try {
            return deferred.await()
        } finally {
            mutex.withLock { if (inFlight === deferred) inFlight = null }
        }
    }
}
