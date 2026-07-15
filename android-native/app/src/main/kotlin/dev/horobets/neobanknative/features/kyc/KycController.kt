package dev.horobets.neobanknative.features.kyc

import androidx.compose.runtime.staticCompositionLocalOf
import dev.horobets.neobanknative.core.networking.ApiError
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

sealed interface KycLoadState {
    data object Loading : KycLoadState
    data class Loaded(val info: KycStatusInfo) : KycLoadState
    data class Failed(val error: ApiError) : KycLoadState
}

class KycController(private val repository: KycRepository) {
    private val _state = MutableStateFlow<KycLoadState>(KycLoadState.Loading)
    val state: StateFlow<KycLoadState> = _state.asStateFlow()

    private val _hasSubmittedThisSession = MutableStateFlow(false)
    val hasSubmittedThisSession: StateFlow<Boolean> = _hasSubmittedThisSession.asStateFlow()

    suspend fun load() {
        _state.value = KycLoadState.Loading
        _hasSubmittedThisSession.value = false
        _state.value = try {
            KycLoadState.Loaded(repository.status())
        } catch (e: ApiError) {
            KycLoadState.Failed(e)
        } catch (e: Exception) {
            KycLoadState.Failed(ApiError.decoding())
        }
    }

    suspend fun submit(
        fullName: String,
        dateOfBirth: String,
        countryCode: String,
        documentType: String?,
        documentNumber: String?,
    ) {
        val result = repository.submit(fullName, dateOfBirth, countryCode, documentType, documentNumber)
        _hasSubmittedThisSession.value = true
        _state.value = KycLoadState.Loaded(KycStatusInfo(rawStatus = result.rawStatus, rejectionReason = result.rejectionReason))
    }
}

val LocalKycController = staticCompositionLocalOf<KycController> {
    error("No KycController provided — wrap the composition in AppEnvironment's providers")
}
