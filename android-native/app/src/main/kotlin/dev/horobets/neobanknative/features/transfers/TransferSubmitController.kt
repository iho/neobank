package dev.horobets.neobanknative.features.transfers

import androidx.compose.runtime.staticCompositionLocalOf
import dev.horobets.neobanknative.core.networking.ApiError
import java.util.UUID
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

sealed interface TransferSubmitState {
    data object Idle : TransferSubmitState
    data object Submitting : TransferSubmitState
    data class Result(val transfer: Transfer) : TransferSubmitState
    data class Failed(val error: ApiError) : TransferSubmitState
}

class TransferSubmitController(private val repository: TransferRepository) {
    private val _state = MutableStateFlow<TransferSubmitState>(TransferSubmitState.Idle)
    val state: StateFlow<TransferSubmitState> = _state.asStateFlow()

    /**
     * Stable across retries so a network hiccup can't double-send — only
     * rotated once a submission reaches a definitive server-side outcome (a
     * [Transfer] back, completed or declined).
     */
    private var idempotencyKey = UUID.randomUUID().toString()

    fun reset() {
        _state.value = TransferSubmitState.Idle
        idempotencyKey = UUID.randomUUID().toString()
    }

    suspend fun submit(amount: String, recipientPhone: String?, recipientEmail: String?, memo: String?) {
        if (_state.value is TransferSubmitState.Result) {
            idempotencyKey = UUID.randomUUID().toString()
        }
        _state.value = TransferSubmitState.Submitting
        _state.value = try {
            val transfer = repository.createTransfer(
                idempotencyKey = idempotencyKey,
                amount = amount,
                recipientPhone = recipientPhone,
                recipientEmail = recipientEmail,
                recipientUserId = null,
                currency = null,
                memo = memo,
            )
            TransferSubmitState.Result(transfer)
        } catch (e: ApiError) {
            TransferSubmitState.Failed(e)
        } catch (e: Exception) {
            TransferSubmitState.Failed(ApiError.decoding())
        }
    }
}

val LocalTransferSubmitController = staticCompositionLocalOf<TransferSubmitController> {
    error("No TransferSubmitController provided — wrap the composition in AppEnvironment's providers")
}
