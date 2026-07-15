package dev.horobets.neobanknative.features.auth

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AccountBalanceWallet
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import dev.horobets.neobanknative.core.design.BrandBackground
import dev.horobets.neobanknative.core.design.BrandTextField
import dev.horobets.neobanknative.core.design.GlowIcon
import dev.horobets.neobanknative.core.design.PrimaryButton
import dev.horobets.neobanknative.core.networking.ApiError
import kotlinx.coroutines.launch

@Composable
fun LoginScreen(onNavigateToRegister: () -> Unit) {
    val authController = LocalAuthController.current
    var email by remember { mutableStateOf("") }
    var password by remember { mutableStateOf("") }
    var isSubmitting by remember { mutableStateOf(false) }
    var errorMessage by remember { mutableStateOf<String?>(null) }
    val passwordFocus = remember { FocusRequester() }
    val scope = rememberCoroutineScope()

    val canSubmit = email.contains("@") && password.isNotEmpty()

    fun submit() {
        if (!canSubmit || isSubmitting) return
        errorMessage = null
        isSubmitting = true
        scope.launch {
            try {
                authController.login(email.trim(), password)
            } catch (e: ApiError) {
                errorMessage = e.message
            } catch (e: Exception) {
                errorMessage = "Something went wrong. Please try again."
            } finally {
                isSubmitting = false
            }
        }
    }

    Box(modifier = Modifier.fillMaxSize()) {
        BrandBackground()

        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(PaddingValues(bottom = 24.dp)),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Column(
                modifier = Modifier.padding(top = 56.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(16.dp),
            ) {
                GlowIcon(icon = Icons.Filled.AccountBalanceWallet)
                Text("Neobank", fontSize = MaterialTheme.typography.headlineMedium.fontSize, fontWeight = FontWeight.Bold)
            }

            Column(
                modifier = Modifier.fillMaxWidth().padding(horizontal = 24.dp, vertical = 32.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                BrandTextField(
                    value = email,
                    onValueChange = { email = it },
                    placeholder = "Email",
                    modifier = Modifier.fillMaxWidth(),
                    keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Email, imeAction = ImeAction.Next),
                    keyboardActions = KeyboardActions(onNext = { passwordFocus.requestFocus() }),
                )

                BrandTextField(
                    value = password,
                    onValueChange = { password = it },
                    placeholder = "Password",
                    modifier = Modifier.fillMaxWidth().focusRequester(passwordFocus),
                    isPassword = true,
                    keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Password, imeAction = ImeAction.Go),
                    keyboardActions = KeyboardActions(onGo = { submit() }),
                )

                errorMessage?.let {
                    Text(it, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodySmall)
                }
            }

            Column(
                modifier = Modifier.fillMaxWidth().padding(horizontal = 24.dp),
                verticalArrangement = Arrangement.spacedBy(16.dp),
            ) {
                PrimaryButton(
                    onClick = ::submit,
                    modifier = Modifier.fillMaxWidth(),
                    enabled = canSubmit,
                    isLoading = isSubmitting,
                ) {
                    Text("Log in", fontWeight = FontWeight.SemiBold)
                }

                TextButton(onClick = onNavigateToRegister, modifier = Modifier.fillMaxWidth()) {
                    Text("Don't have an account? Register", style = MaterialTheme.typography.bodySmall)
                }
            }
        }
    }
}
