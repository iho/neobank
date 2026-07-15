package dev.horobets.neobanknative.features.transfers

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.Error
import androidx.compose.material.icons.filled.HourglassEmpty
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SegmentedButton
import androidx.compose.material3.SegmentedButtonDefaults
import androidx.compose.material3.SingleChoiceSegmentedButtonRow
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.horobets.neobanknative.core.design.BrandBackground
import dev.horobets.neobanknative.core.design.BrandTextField
import dev.horobets.neobanknative.core.design.PrimaryButton
import dev.horobets.neobanknative.core.design.surfaceCard
import dev.horobets.neobanknative.core.formatting.LedgerAmount
import dev.horobets.neobanknative.features.wallet.LocalWalletController
import kotlinx.coroutines.launch

private enum class TransferStep { RECIPIENT, AMOUNT, CONFIRM_AND_RESULT }
private enum class RecipientType(val label: String) { PHONE("Phone"), EMAIL("Email") }

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TransferFlowScreen(onDismiss: () -> Unit) {
    val walletController = LocalWalletController.current
    val controller = LocalTransferSubmitController.current
    val submitState by controller.state.collectAsStateWithLifecycle()

    var step by remember { mutableStateOf(TransferStep.RECIPIENT) }
    var recipientType by remember { mutableStateOf(RecipientType.PHONE) }
    var recipient by remember { mutableStateOf("") }
    var amount by remember { mutableStateOf("") }
    var memo by remember { mutableStateOf("") }
    val scope = rememberCoroutineScope()

    LaunchedEffect(Unit) { controller.reset() }

    val recipientValid = recipient.trim().isNotEmpty()
    val amountValid = amount.trim().let { it.matches(Regex("^\\d+(\\.\\d{1,2})?$")) && (it.toDoubleOrNull() ?: 0.0) > 0 }

    fun submit() {
        scope.launch {
            controller.submit(
                amount = amount.trim(),
                recipientPhone = if (recipientType == RecipientType.PHONE) recipient.trim() else null,
                recipientEmail = if (recipientType == RecipientType.EMAIL) recipient.trim() else null,
                memo = memo.trim(),
            )
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Send money") },
                navigationIcon = { TextButton(onClick = onDismiss) { Text("Cancel") } },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = Color.Transparent),
            )
        },
    ) { padding ->
        Box(modifier = Modifier.fillMaxSize().padding(padding)) {
            BrandBackground()

            Box(modifier = Modifier.fillMaxSize().padding(24.dp)) {
                when (step) {
                    TransferStep.RECIPIENT -> RecipientStep(
                        recipientType = recipientType,
                        onRecipientTypeChange = { recipientType = it },
                        recipient = recipient,
                        onRecipientChange = { recipient = it },
                        canContinue = recipientValid,
                        onNext = { step = TransferStep.AMOUNT },
                    )
                    TransferStep.AMOUNT -> AmountStep(
                        amount = amount,
                        onAmountChange = { amount = it },
                        memo = memo,
                        onMemoChange = { memo = it },
                        canContinue = amountValid,
                        onNext = { step = TransferStep.CONFIRM_AND_RESULT },
                    )
                    TransferStep.CONFIRM_AND_RESULT -> ConfirmAndResultStep(
                        submitState = submitState,
                        recipient = recipient,
                        amount = amount,
                        memo = memo,
                        onSubmit = ::submit,
                        onEdit = { step = TransferStep.AMOUNT },
                        onDone = {
                            scope.launch { walletController.load() }
                            onDismiss()
                        },
                        onCancel = onDismiss,
                    )
                }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun RecipientStep(
    recipientType: RecipientType,
    onRecipientTypeChange: (RecipientType) -> Unit,
    recipient: String,
    onRecipientChange: (String) -> Unit,
    canContinue: Boolean,
    onNext: () -> Unit,
) {
    Column(modifier = Modifier.fillMaxSize()) {
        Text("Who are you sending to?", style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold)
        Spacer(Modifier.height(20.dp))

        SingleChoiceSegmentedButtonRow(modifier = Modifier.fillMaxWidth()) {
            RecipientType.entries.forEachIndexed { index, option ->
                SegmentedButton(
                    selected = recipientType == option,
                    onClick = { onRecipientTypeChange(option) },
                    shape = SegmentedButtonDefaults.itemShape(index = index, count = RecipientType.entries.size),
                ) { Text(option.label) }
            }
        }
        Spacer(Modifier.height(20.dp))

        BrandTextField(
            value = recipient,
            onValueChange = onRecipientChange,
            placeholder = if (recipientType == RecipientType.PHONE) "Recipient phone" else "Recipient email",
            modifier = Modifier.fillMaxWidth(),
            keyboardOptions = KeyboardOptions(
                keyboardType = if (recipientType == RecipientType.PHONE) KeyboardType.Phone else KeyboardType.Email,
            ),
        )

        Spacer(Modifier.weight(1f))

        PrimaryButton(onClick = onNext, modifier = Modifier.fillMaxWidth(), enabled = canContinue) {
            Text("Next", fontWeight = FontWeight.SemiBold)
        }
    }
}

@Composable
private fun AmountStep(
    amount: String,
    onAmountChange: (String) -> Unit,
    memo: String,
    onMemoChange: (String) -> Unit,
    canContinue: Boolean,
    onNext: () -> Unit,
) {
    Column(modifier = Modifier.fillMaxSize()) {
        Text("How much?", style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold)
        Spacer(Modifier.height(20.dp))

        Row(
            modifier = Modifier.fillMaxWidth().surfaceCard(cornerRadius = 12.dp).padding(14.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text("$", color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f))
            Spacer(Modifier.width(4.dp))
            androidx.compose.foundation.text.BasicTextField(
                value = amount,
                onValueChange = onAmountChange,
                singleLine = true,
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Decimal),
                textStyle = androidx.compose.material3.LocalTextStyle.current.copy(color = MaterialTheme.colorScheme.onSurface),
                modifier = Modifier.weight(1f),
            )
        }
        Spacer(Modifier.height(12.dp))

        BrandTextField(
            value = memo,
            onValueChange = onMemoChange,
            placeholder = "Memo (optional)",
            modifier = Modifier.fillMaxWidth(),
        )

        Spacer(Modifier.weight(1f))

        PrimaryButton(onClick = onNext, modifier = Modifier.fillMaxWidth(), enabled = canContinue) {
            Text("Next", fontWeight = FontWeight.SemiBold)
        }
    }
}

@Composable
private fun ConfirmAndResultStep(
    submitState: TransferSubmitState,
    recipient: String,
    amount: String,
    memo: String,
    onSubmit: () -> Unit,
    onEdit: () -> Unit,
    onDone: () -> Unit,
    onCancel: () -> Unit,
) {
    when (submitState) {
        is TransferSubmitState.Idle -> ConfirmationView(recipient, amount, memo, onSubmit, onEdit)
        is TransferSubmitState.Submitting -> Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
            CircularProgressIndicator()
        }
        is TransferSubmitState.Result -> ResultViewForTransfer(submitState.transfer, recipient, amount, onSubmit, onDone, onCancel)
        is TransferSubmitState.Failed -> ResultView(
            icon = Icons.Filled.Error,
            tint = MaterialTheme.colorScheme.error,
            title = "Something went wrong",
            message = submitState.error.message ?: "Something went wrong. Please try again.",
            primaryLabel = "Retry",
            onPrimary = onSubmit,
            secondaryLabel = "Cancel",
            onSecondary = onCancel,
        )
    }
}

@Composable
private fun ConfirmationView(recipient: String, amount: String, memo: String, onSubmit: () -> Unit, onEdit: () -> Unit) {
    Column(modifier = Modifier.fillMaxSize()) {
        Text("Confirm transfer", style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold)
        Spacer(Modifier.height(20.dp))

        Column(
            modifier = Modifier.fillMaxWidth().surfaceCard(cornerRadius = 14.dp).padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            SummaryRow("To", recipient)
            SummaryRow("Amount", "$$amount")
            if (memo.isNotEmpty()) SummaryRow("Memo", memo)
        }

        Spacer(Modifier.weight(1f))

        Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
            PrimaryButton(onClick = onSubmit, modifier = Modifier.fillMaxWidth()) {
                Text("Confirm & Send", fontWeight = FontWeight.SemiBold)
            }
            TextButton(onClick = onEdit, modifier = Modifier.fillMaxWidth()) { Text("Edit") }
        }
    }
}

@Composable
private fun SummaryRow(label: String, value: String) {
    Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
        Text(label, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f))
        Text(value, fontWeight = FontWeight.SemiBold)
    }
}

@Composable
private fun ResultViewForTransfer(
    transfer: Transfer,
    recipient: String,
    amount: String,
    onRetry: () -> Unit,
    onDone: () -> Unit,
    onCancel: () -> Unit,
) {
    when {
        transfer.isCompleted -> ResultView(
            icon = Icons.Filled.CheckCircle,
            tint = Color(0xFF22C55E),
            title = "Sent!",
            message = "\$${LedgerAmount.formatted(transfer.amount ?: amount)} to $recipient",
            primaryLabel = "Done",
            onPrimary = onDone,
        )
        transfer.isFailed -> ResultView(
            icon = Icons.Filled.Error,
            tint = MaterialTheme.colorScheme.error,
            title = "Transfer declined",
            message = transfer.failureReason ?: "The transfer could not be completed.",
            primaryLabel = "Retry",
            onPrimary = onRetry,
            secondaryLabel = "Cancel",
            onSecondary = onCancel,
        )
        else -> ResultView(
            icon = Icons.Filled.HourglassEmpty,
            tint = MaterialTheme.colorScheme.primary,
            title = "Processing",
            message = "Status: ${transfer.status ?: "unknown"}",
            primaryLabel = "Check again",
            onPrimary = onRetry,
        )
    }
}

@Composable
private fun ResultView(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    tint: Color,
    title: String,
    message: String,
    primaryLabel: String,
    onPrimary: () -> Unit,
    secondaryLabel: String? = null,
    onSecondary: (() -> Unit)? = null,
) {
    Column(modifier = Modifier.fillMaxSize(), horizontalAlignment = Alignment.CenterHorizontally) {
        Spacer(Modifier.weight(1f))
        Icon(icon, contentDescription = null, tint = tint, modifier = Modifier.height(56.dp))
        Spacer(Modifier.height(16.dp))
        Text(title, style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold)
        Spacer(Modifier.height(8.dp))
        Text(
            message,
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f),
            textAlign = TextAlign.Center,
        )
        Spacer(Modifier.weight(1f))
        PrimaryButton(onClick = onPrimary, modifier = Modifier.fillMaxWidth()) {
            Text(primaryLabel, fontWeight = FontWeight.SemiBold)
        }
        if (secondaryLabel != null && onSecondary != null) {
            TextButton(onClick = onSecondary) { Text(secondaryLabel) }
        }
    }
}
