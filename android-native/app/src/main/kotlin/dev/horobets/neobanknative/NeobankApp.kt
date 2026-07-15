package dev.horobets.neobanknative

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.getValue
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.horobets.neobanknative.core.AppEnvironment
import dev.horobets.neobanknative.core.auth.LocalSessionStore
import dev.horobets.neobanknative.core.design.AppAppearance
import dev.horobets.neobanknative.core.design.LocalAppAppearanceStore
import dev.horobets.neobanknative.core.design.NeobankTheme
import dev.horobets.neobanknative.features.auth.LocalAuthController
import dev.horobets.neobanknative.features.cards.LocalCardsController
import dev.horobets.neobanknative.features.kyc.LocalKycController
import dev.horobets.neobanknative.features.notifications.LocalNotificationsController
import dev.horobets.neobanknative.features.transfers.LocalTransferSubmitController
import dev.horobets.neobanknative.features.wallet.LocalWalletController
import dev.horobets.neobanknative.root.RootScreen

@Composable
fun NeobankApp(environment: AppEnvironment) {
    val appearance by environment.appAppearanceStore.appearance.collectAsStateWithLifecycle()
    val systemIsDark = isSystemInDarkTheme()
    val isDark = when (appearance) {
        AppAppearance.SYSTEM -> systemIsDark
        AppAppearance.LIGHT -> false
        AppAppearance.DARK -> true
    }

    CompositionLocalProvider(
        LocalSessionStore provides environment.sessionStore,
        LocalAppAppearanceStore provides environment.appAppearanceStore,
        LocalAuthController provides environment.authController,
        LocalKycController provides environment.kycController,
        LocalWalletController provides environment.walletController,
        LocalCardsController provides environment.cardsController,
        LocalNotificationsController provides environment.notificationsController,
        LocalTransferSubmitController provides environment.transferSubmitController,
    ) {
        NeobankTheme(isDark = isDark) {
            RootScreen()
        }
    }
}
