package dev.horobets.neobanknative.features.kyc

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.VerifiedUser
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.DatePicker
import androidx.compose.material3.DatePickerDialog
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.material3.rememberDatePickerState
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
import androidx.compose.ui.text.input.KeyboardCapitalization
import androidx.compose.ui.unit.dp
import dev.horobets.neobanknative.core.design.BrandBackground
import dev.horobets.neobanknative.core.design.BrandTextField
import dev.horobets.neobanknative.core.design.GlowIcon
import dev.horobets.neobanknative.core.design.PrimaryButton
import dev.horobets.neobanknative.core.design.surfaceCard
import dev.horobets.neobanknative.core.networking.ApiError
import dev.horobets.neobanknative.features.auth.LocalAuthController
import java.time.Instant
import java.time.LocalDate
import java.time.ZoneOffset
import java.time.format.DateTimeFormatter
import kotlinx.coroutines.launch

/**
 * Shown by the home gate whenever KYC status isn't approved. Handles both
 * the never-submitted case (plain form) and the rejected case (banner +
 * resubmit) — the backend reports both as ordinary KYC statuses, there's no
 * separate "not started" signal.
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun KycFormScreen(rejectionReason: String?) {
    val authController = LocalAuthController.current
    val kycController = LocalKycController.current
    var fullName by remember { mutableStateOf("") }
    val defaultDobMillis = remember { LocalDate.now().minusYears(18).atStartOfDay(ZoneOffset.UTC).toInstant().toEpochMilli() }
    var dateOfBirthMillis by remember { mutableStateOf(defaultDobMillis) }
    var countryCode by remember { mutableStateOf("") }
    var documentType by remember { mutableStateOf("") }
    var documentNumber by remember { mutableStateOf("") }
    var isSubmitting by remember { mutableStateOf(false) }
    var errorMessage by remember { mutableStateOf<String?>(null) }
    var showLogoutConfirmation by remember { mutableStateOf(false) }
    var showDatePicker by remember { mutableStateOf(false) }
    val scope = rememberCoroutineScope()

    val canSubmit = fullName.trim().isNotEmpty() && countryCode.length == 2

    fun submit() {
        if (!canSubmit || isSubmitting) return
        errorMessage = null
        isSubmitting = true
        scope.launch {
            try {
                val dob = Instant.ofEpochMilli(dateOfBirthMillis).atZone(ZoneOffset.UTC).toLocalDate()
                kycController.submit(
                    fullName = fullName.trim(),
                    dateOfBirth = dob.format(DateTimeFormatter.ISO_LOCAL_DATE),
                    countryCode = countryCode,
                    documentType = documentType.trim(),
                    documentNumber = documentNumber.trim(),
                )
            } catch (e: ApiError) {
                errorMessage = e.message
            } catch (e: Exception) {
                errorMessage = "Something went wrong. Please try again."
            } finally {
                isSubmitting = false
            }
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Verify your identity") },
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
                modifier = Modifier
                    .fillMaxSize()
                    .verticalScroll(rememberScrollState())
                    .padding(horizontal = 24.dp)
                    .padding(bottom = 24.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(24.dp),
            ) {
                Column(
                    modifier = Modifier.padding(top = 8.dp),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.spacedBy(12.dp),
                ) {
                    GlowIcon(icon = Icons.Filled.VerifiedUser, diameter = 72.dp, iconSize = 32.dp)
                    Text("A few quick details", style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold)
                    Text(
                        "We're required to verify your identity before your wallet can go live.",
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f),
                        textAlign = androidx.compose.ui.text.style.TextAlign.Center,
                    )
                }

                if (rejectionReason != null) {
                    Row(
                        modifier = Modifier.fillMaxWidth().surfaceCard(cornerRadius = 14.dp).padding(12.dp),
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Icon(Icons.Filled.Warning, contentDescription = null, tint = MaterialTheme.colorScheme.error)
                        androidx.compose.foundation.layout.Spacer(modifier = Modifier.padding(4.dp))
                        Text(
                            "Previous submission was rejected: $rejectionReason",
                            color = MaterialTheme.colorScheme.error,
                            style = MaterialTheme.typography.bodyMedium,
                        )
                    }
                }

                Column(modifier = Modifier.fillMaxWidth(), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    BrandTextField(
                        value = fullName,
                        onValueChange = { fullName = it },
                        placeholder = "Full legal name",
                        modifier = Modifier.fillMaxWidth(),
                    )

                    Row(
                        modifier = Modifier.fillMaxWidth().surfaceCard(cornerRadius = 12.dp)
                            .then(Modifier.padding(14.dp)),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Text("Date of birth", color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f))
                        TextButton(onClick = { showDatePicker = true }) {
                            val displayDate = Instant.ofEpochMilli(dateOfBirthMillis).atZone(ZoneOffset.UTC).toLocalDate()
                            Text(displayDate.format(DateTimeFormatter.ISO_LOCAL_DATE))
                        }
                    }

                    BrandTextField(
                        value = countryCode,
                        onValueChange = { countryCode = it.uppercase().take(2) },
                        placeholder = "Country code (ISO-2, e.g. US)",
                        modifier = Modifier.fillMaxWidth(),
                        keyboardOptions = KeyboardOptions(capitalization = KeyboardCapitalization.Characters),
                    )

                    BrandTextField(
                        value = documentType,
                        onValueChange = { documentType = it },
                        placeholder = "Document type (optional)",
                        modifier = Modifier.fillMaxWidth(),
                    )

                    BrandTextField(
                        value = documentNumber,
                        onValueChange = { documentNumber = it },
                        placeholder = "Document number (optional)",
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
                    Text("Submit", fontWeight = FontWeight.SemiBold)
                }
            }
        }
    }

    if (showDatePicker) {
        val minMillis = remember { LocalDate.now().minusYears(120).atStartOfDay(ZoneOffset.UTC).toInstant().toEpochMilli() }
        val maxMillis = remember { LocalDate.now().minusYears(13).atStartOfDay(ZoneOffset.UTC).toInstant().toEpochMilli() }
        val datePickerState = rememberDatePickerState(
            initialSelectedDateMillis = dateOfBirthMillis,
            selectableDates = object : androidx.compose.material3.SelectableDates {
                override fun isSelectableDate(utcTimeMillis: Long) = utcTimeMillis in minMillis..maxMillis
            },
        )
        DatePickerDialog(
            onDismissRequest = { showDatePicker = false },
            confirmButton = {
                TextButton(onClick = {
                    datePickerState.selectedDateMillis?.let { dateOfBirthMillis = it }
                    showDatePicker = false
                }) { Text("OK") }
            },
            dismissButton = { TextButton(onClick = { showDatePicker = false }) { Text("Cancel") } },
        ) {
            DatePicker(state = datePickerState)
        }
    }

    if (showLogoutConfirmation) {
        AlertDialog(
            onDismissRequest = { showLogoutConfirmation = false },
            title = { Text("Log out?") },
            text = { Text("You'll need to sign back in to finish verifying your identity.") },
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
