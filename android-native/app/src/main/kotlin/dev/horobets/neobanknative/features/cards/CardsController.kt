package dev.horobets.neobanknative.features.cards

import androidx.compose.runtime.staticCompositionLocalOf
import dev.horobets.neobanknative.core.networking.ApiError
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

sealed interface CardsLoadState {
    data object Loading : CardsLoadState
    data class Loaded(val cards: List<BankCard>) : CardsLoadState
    data class Failed(val error: ApiError) : CardsLoadState
}

class CardsController(private val repository: CardRepository) {
    private val _state = MutableStateFlow<CardsLoadState>(CardsLoadState.Loading)
    val state: StateFlow<CardsLoadState> = _state.asStateFlow()

    suspend fun load() {
        _state.value = CardsLoadState.Loading
        _state.value = try {
            CardsLoadState.Loaded(repository.list())
        } catch (e: ApiError) {
            CardsLoadState.Failed(e)
        } catch (e: Exception) {
            CardsLoadState.Failed(ApiError.decoding())
        }
    }

    suspend fun issueCard(cardholderName: String, dailyLimit: String?, onlineOnly: Boolean?) {
        repository.issue(cardholderName, dailyLimit, onlineOnly)
        load()
    }

    suspend fun toggleFreeze(card: BankCard) {
        val updated = if (card.isFrozen) repository.unfreeze(card.id) else repository.freeze(card.id)
        val loaded = _state.value as? CardsLoadState.Loaded ?: return
        _state.value = CardsLoadState.Loaded(loaded.cards.map { if (it.id == updated.id) updated else it })
    }
}

val LocalCardsController = staticCompositionLocalOf<CardsController> {
    error("No CardsController provided — wrap the composition in AppEnvironment's providers")
}
