package dev.horobets.neobanknative.features.auth

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import dev.horobets.neobanknative.core.design.BrandBackground
import dev.horobets.neobanknative.core.design.BrandTextField
import dev.horobets.neobanknative.core.design.PrimaryButton
import dev.horobets.neobanknative.core.networking.ApiError
import kotlinx.coroutines.launch

@Composable
fun RegisterScreen() {
    val authController = LocalAuthController.current
    var email by remember { mutableStateOf("") }
    var phone by remember { mutableStateOf("") }
    var password by remember { mutableStateOf("") }
    var inviteCode by remember { mutableStateOf("") }
    var isSubmitting by remember { mutableStateOf(false) }
    var errorMessage by remember { mutableStateOf<String?>(null) }
    val scope = rememberCoroutineScope()

    val canSubmit = email.contains("@") && password.length >= 8

    fun submit() {
        if (!canSubmit || isSubmitting) return
        errorMessage = null
        isSubmitting = true
        scope.launch {
            try {
                authController.register(email.trim(), password, phone.trim(), inviteCode.trim())
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
                .padding(horizontal = 24.dp)
                .padding(bottom = 24.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(24.dp),
        ) {
            Column(
                modifier = Modifier.padding(top = 24.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                Text("Create account", fontSize = MaterialTheme.typography.headlineSmall.fontSize, fontWeight = FontWeight.Bold)
                Text(
                    "Set up your Neobank wallet",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f),
                )
            }

            Column(modifier = Modifier.fillMaxWidth(), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                BrandTextField(
                    value = email,
                    onValueChange = { email = it },
                    placeholder = "Email",
                    modifier = Modifier.fillMaxWidth(),
                    keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Email),
                )

                BrandTextField(
                    value = phone,
                    onValueChange = { phone = it },
                    placeholder = "Phone (optional)",
                    modifier = Modifier.fillMaxWidth(),
                    keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Phone),
                )

                BrandTextField(
                    value = password,
                    onValueChange = { password = it },
                    placeholder = "Password",
                    modifier = Modifier.fillMaxWidth(),
                    isPassword = true,
                    keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Password),
                )

                if (password.isNotEmpty() && password.length < 8) {
                    Text("At least 8 characters", color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodySmall)
                }

                BrandTextField(
                    value = inviteCode,
                    onValueChange = { inviteCode = it.uppercase() },
                    placeholder = "Invite code (optional)",
                    modifier = Modifier.fillMaxWidth(),
                )
            }

            errorMessage?.let {
                Text(it, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodySmall)
            }

            PrimaryButton(
                onClick = ::submit,
                modifier = Modifier.fillMaxWidth(),
                enabled = canSubmit,
                isLoading = isSubmitting,
            ) {
                Text("Create account", fontWeight = FontWeight.SemiBold)
            }
        }
    }
}
