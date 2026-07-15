package dev.horobets.neobanknative.root

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.horobets.neobanknative.core.auth.LocalSessionStore
import dev.horobets.neobanknative.core.networking.ApiError
import dev.horobets.neobanknative.features.kyc.KycFormScreen
import dev.horobets.neobanknative.features.kyc.KycLoadState
import dev.horobets.neobanknative.features.kyc.KycPendingScreen
import dev.horobets.neobanknative.features.kyc.KycStatus
import dev.horobets.neobanknative.features.kyc.LocalKycController
import kotlinx.coroutines.launch

/**
 * Sits between the session gate and the main tabs: nothing past this point
 * should be reachable with KYC outstanding, so this is where that check
 * lives — not sprinkled across each destination.
 */
@Composable
fun HomeGateScreen() {
    val sessionStore = LocalSessionStore.current
    val kycController = LocalKycController.current
    val generation by sessionStore.generation.collectAsStateWithLifecycle()
    val state by kycController.state.collectAsStateWithLifecycle()
    val hasSubmittedThisSession by kycController.hasSubmittedThisSession.collectAsStateWithLifecycle()
    val scope = rememberCoroutineScope()

    LaunchedEffect(generation) { kycController.load() }

    when (val current = state) {
        is KycLoadState.Loading -> Box(modifier = Modifier.fillMaxSize()) {
            CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
        }
        is KycLoadState.Failed -> KycLoadErrorView(current.error, onRetry = { scope.launch { kycController.load() } })
        is KycLoadState.Loaded -> when (current.info.status) {
            KycStatus.APPROVED -> MainTabsScreen()
            KycStatus.REJECTED -> KycFormScreen(rejectionReason = current.info.rejectionReason)
            KycStatus.PENDING, KycStatus.MANUAL_REVIEW -> if (hasSubmittedThisSession) {
                KycPendingScreen()
            } else {
                KycFormScreen(rejectionReason = null)
            }
        }
    }
}

@Composable
private fun KycLoadErrorView(error: ApiError, onRetry: () -> Unit) {
    Column(
        modifier = Modifier.fillMaxSize().padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
    ) {
        Text(error.message ?: "Something went wrong.", textAlign = TextAlign.Center)
        OutlinedButton(onClick = onRetry, modifier = Modifier.padding(top = 12.dp)) { Text("Retry") }
    }
}
