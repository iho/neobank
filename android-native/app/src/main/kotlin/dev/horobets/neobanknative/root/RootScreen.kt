package dev.horobets.neobanknative.root

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import dev.horobets.neobanknative.core.auth.LocalSessionStore
import dev.horobets.neobanknative.core.auth.SessionStatus
import dev.horobets.neobanknative.features.auth.LocalAuthController
import dev.horobets.neobanknative.features.auth.LoginScreen
import dev.horobets.neobanknative.features.auth.RegisterScreen

@Composable
fun RootScreen() {
    val sessionStore = LocalSessionStore.current
    val authController = LocalAuthController.current
    val status by sessionStore.status.collectAsStateWithLifecycle()

    LaunchedEffect(Unit) { authController.bootstrap() }

    when (status) {
        SessionStatus.UNKNOWN -> Box(modifier = Modifier.fillMaxSize()) {
            CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
        }
        SessionStatus.UNAUTHENTICATED -> AuthNavHost()
        SessionStatus.AUTHENTICATED -> HomeGateScreen()
    }
}

@Composable
private fun AuthNavHost() {
    val navController = rememberNavController()
    NavHost(navController = navController, startDestination = "login") {
        composable("login") { LoginScreen(onNavigateToRegister = { navController.navigate("register") }) }
        composable("register") { RegisterScreen() }
    }
}
