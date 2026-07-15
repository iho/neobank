package dev.horobets.neobanknative.features.wallet

import androidx.compose.runtime.staticCompositionLocalOf
import dev.horobets.neobanknative.core.networking.ApiError
import kotlinx.coroutines.async
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class WalletSnapshot(
    val balance: WalletBalance,
    val transactions: List<WalletTransaction>,
    val nextCursor: String? = null,
    val isLoadingMore: Boolean = false,
) {
    val hasMore: Boolean get() = nextCursor != null
}

sealed interface WalletLoadState {
    data object Loading : WalletLoadState
    data class Loaded(val snapshot: WalletSnapshot) : WalletLoadState
    data class Failed(val error: ApiError) : WalletLoadState
}

class WalletController(private val repository: WalletRepository) {
    private val _state = MutableStateFlow<WalletLoadState>(WalletLoadState.Loading)
    val state: StateFlow<WalletLoadState> = _state.asStateFlow()

    suspend fun load() {
        _state.value = WalletLoadState.Loading
        _state.value = try {
            coroutineScope {
                val balanceDeferred = async { repository.balance() }
                val pageDeferred = async { repository.transactions() }
                val balance = balanceDeferred.await()
                val page = pageDeferred.await()
                WalletLoadState.Loaded(WalletSnapshot(balance = balance, transactions = page.transactions, nextCursor = page.nextCursor))
            }
        } catch (e: ApiError) {
            WalletLoadState.Failed(e)
        } catch (e: Exception) {
            WalletLoadState.Failed(ApiError.decoding())
        }
    }

    /**
     * Fetches the next page when the last visible row appears. A no-op
     * (rather than an error) when there's nothing more or a fetch is
     * already in flight, since callers trigger this from scroll position
     * and can't easily guard it themselves.
     */
    suspend fun loadMore() {
        val loaded = _state.value as? WalletLoadState.Loaded ?: return
        val snapshot = loaded.snapshot
        if (!snapshot.hasMore || snapshot.isLoadingMore) return

        _state.value = WalletLoadState.Loaded(snapshot.copy(isLoadingMore = true))
        val updated = try {
            val page = repository.transactions(cursor = snapshot.nextCursor)
            snapshot.copy(
                transactions = snapshot.transactions + page.transactions,
                nextCursor = page.nextCursor,
                isLoadingMore = false,
            )
        } catch (e: Exception) {
            // Leave existing transactions in place; the user can retry by scrolling again.
            snapshot.copy(isLoadingMore = false)
        }
        _state.value = WalletLoadState.Loaded(updated)
    }
}

val LocalWalletController = staticCompositionLocalOf<WalletController> {
    error("No WalletController provided — wrap the composition in AppEnvironment's providers")
}
