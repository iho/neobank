package dev.horobets.neobanknative.features.kyc

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.HourglassEmpty
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import dev.horobets.neobanknative.core.design.BrandBackground
import dev.horobets.neobanknative.core.design.GlowIcon
import dev.horobets.neobanknative.features.auth.LocalAuthController
import kotlinx.coroutines.launch

/**
 * Rare with MVP auto-approve (submission usually resolves synchronously),
 * but the contract allows an async `pending` outcome — surfaced here with a
 * manual refresh in case a future backend makes KYC actually asynchronous.
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun KycPendingScreen() {
    val authController = LocalAuthController.current
    val kycController = LocalKycController.current
    var showLogoutConfirmation by remember { mutableStateOf(false) }
    val scope = rememberCoroutineScope()

    Scaffold(
        topBar = {
            TopAppBar(
                title = {},
                navigationIcon = {
                    IconButton(onClick = { showLogoutConfirmation = true }) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Log out")
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = Color.Transparent),
            )
        },
    ) { padding ->
        Box(modifier = Modifier.fillMaxSize().padding(padding)) {
            BrandBackground()

            Column(
                modifier = Modifier.fillMaxSize().padding(16.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.Center,
            ) {
                GlowIcon(icon = Icons.Filled.HourglassEmpty, diameter = 72.dp, iconSize = 32.dp)
                androidx.compose.foundation.layout.Spacer(modifier = Modifier.padding(8.dp))
                Text("We're reviewing your details", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                Text(
                    "This usually only takes a moment.",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f),
                )
                androidx.compose.foundation.layout.Spacer(modifier = Modifier.padding(8.dp))
                OutlinedButton(onClick = { scope.launch { kycController.load() } }) {
                    Text("Check status")
                }
            }
        }
    }

    if (showLogoutConfirmation) {
        AlertDialog(
            onDismissRequest = { showLogoutConfirmation = false },
            title = { Text("Log out?") },
            confirmButton = {
                TextButton(onClick = {
                    showLogoutConfirmation = false
                    authController.logout()
                }) { Text("Log out", color = MaterialTheme.colorScheme.error) }
            },
            dismissButton = { TextButton(onClick = { showLogoutConfirmation = false }) { Text("Cancel") } },
        )
    }
}
