import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/error/api_exception.dart';
import '../../wallet/presentation/wallet_home_controller.dart';
import '../domain/transfer_models.dart';
import 'transfer_submit_controller.dart';

enum _RecipientType { phone, email }

class TransferFlowScreen extends ConsumerStatefulWidget {
  const TransferFlowScreen({super.key});

  @override
  ConsumerState<TransferFlowScreen> createState() => _TransferFlowScreenState();
}

class _TransferFlowScreenState extends ConsumerState<TransferFlowScreen> {
  int _step = 0;
  _RecipientType _recipientType = _RecipientType.phone;
  final _recipientController = TextEditingController();
  final _amountController = TextEditingController();
  final _memoController = TextEditingController();

  @override
  void dispose() {
    _recipientController.dispose();
    _amountController.dispose();
    _memoController.dispose();
    super.dispose();
  }

  bool get _recipientValid => _recipientController.text.trim().isNotEmpty;

  bool get _amountValid =>
      RegExp(r'^\d+(\.\d{1,2})?$').hasMatch(_amountController.text.trim()) &&
      double.parse(_amountController.text.trim()) > 0;

  void _submit() {
    ref.read(transferSubmitControllerProvider.notifier).submit(
          amount: _amountController.text.trim(),
          recipientPhone: _recipientType == _RecipientType.phone
              ? _recipientController.text.trim()
              : null,
          recipientEmail: _recipientType == _RecipientType.email
              ? _recipientController.text.trim()
              : null,
          memo: _memoController.text.trim(),
        );
  }

  @override
  Widget build(BuildContext context) {
    final submission = ref.watch(transferSubmitControllerProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Send money')),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: switch (_step) {
            0 => _RecipientStep(
                recipientType: _recipientType,
                controller: _recipientController,
                onTypeChanged: (t) => setState(() => _recipientType = t),
                onChanged: () => setState(() {}),
                onNext: _recipientValid ? () => setState(() => _step = 1) : null,
              ),
            1 => _AmountStep(
                amountController: _amountController,
                memoController: _memoController,
                onChanged: () => setState(() {}),
                onBack: () => setState(() => _step = 0),
                onNext: _amountValid ? () => setState(() => _step = 2) : null,
              ),
            _ => _ConfirmAndResultStep(
                recipient: _recipientController.text.trim(),
                amount: _amountController.text.trim(),
                memo: _memoController.text.trim(),
                submission: submission,
                onConfirm: _submit,
                onRetry: _submit,
                onEditDetails: () => setState(() => _step = 1),
              ),
          },
        ),
      ),
    );
  }
}

class _RecipientStep extends StatelessWidget {
  const _RecipientStep({
    required this.recipientType,
    required this.controller,
    required this.onTypeChanged,
    required this.onChanged,
    required this.onNext,
  });

  final _RecipientType recipientType;
  final TextEditingController controller;
  final ValueChanged<_RecipientType> onTypeChanged;
  final VoidCallback onChanged;
  final VoidCallback? onNext;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text('Who are you sending to?', style: Theme.of(context).textTheme.titleLarge),
        const SizedBox(height: 16),
        SegmentedButton<_RecipientType>(
          segments: const [
            ButtonSegment(value: _RecipientType.phone, label: Text('Phone')),
            ButtonSegment(value: _RecipientType.email, label: Text('Email')),
          ],
          selected: {recipientType},
          onSelectionChanged: (s) => onTypeChanged(s.first),
        ),
        const SizedBox(height: 16),
        TextField(
          controller: controller,
          keyboardType: recipientType == _RecipientType.phone
              ? TextInputType.phone
              : TextInputType.emailAddress,
          onChanged: (_) => onChanged(),
          decoration: InputDecoration(
            labelText: recipientType == _RecipientType.phone ? 'Recipient phone' : 'Recipient email',
          ),
        ),
        const Spacer(),
        FilledButton(onPressed: onNext, child: const Text('Next')),
      ],
    );
  }
}

class _AmountStep extends StatelessWidget {
  const _AmountStep({
    required this.amountController,
    required this.memoController,
    required this.onChanged,
    required this.onBack,
    required this.onNext,
  });

  final TextEditingController amountController;
  final TextEditingController memoController;
  final VoidCallback onChanged;
  final VoidCallback onBack;
  final VoidCallback? onNext;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text('How much?', style: Theme.of(context).textTheme.titleLarge),
        const SizedBox(height: 16),
        TextField(
          controller: amountController,
          keyboardType: const TextInputType.numberWithOptions(decimal: true),
          onChanged: (_) => onChanged(),
          decoration: const InputDecoration(labelText: 'Amount (USD)', prefixText: '\$'),
        ),
        const SizedBox(height: 16),
        TextField(
          controller: memoController,
          decoration: const InputDecoration(labelText: 'Memo (optional)'),
        ),
        const Spacer(),
        Row(
          children: [
            Expanded(child: OutlinedButton(onPressed: onBack, child: const Text('Back'))),
            const SizedBox(width: 12),
            Expanded(child: FilledButton(onPressed: onNext, child: const Text('Next'))),
          ],
        ),
      ],
    );
  }
}

class _ConfirmAndResultStep extends ConsumerWidget {
  const _ConfirmAndResultStep({
    required this.recipient,
    required this.amount,
    required this.memo,
    required this.submission,
    required this.onConfirm,
    required this.onRetry,
    required this.onEditDetails,
  });

  final String recipient;
  final String amount;
  final String memo;
  final AsyncValue<Transfer?> submission;
  final VoidCallback onConfirm;
  final VoidCallback onRetry;
  final VoidCallback onEditDetails;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return submission.when(
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (error, _) => _ResultView(
        icon: Icons.error_outline,
        color: Theme.of(context).colorScheme.error,
        title: 'Something went wrong',
        message: error is ApiException ? error.message : '$error',
        primaryLabel: 'Retry',
        onPrimary: onRetry,
        secondaryLabel: 'Cancel',
        onSecondary: () => context.pop(),
      ),
      data: (transfer) {
        if (transfer == null) {
          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text('Confirm transfer', style: Theme.of(context).textTheme.titleLarge),
              const SizedBox(height: 24),
              _SummaryRow(label: 'To', value: recipient),
              _SummaryRow(label: 'Amount', value: '\$$amount'),
              if (memo.isNotEmpty) _SummaryRow(label: 'Memo', value: memo),
              const Spacer(),
              FilledButton(onPressed: onConfirm, child: const Text('Confirm & Send')),
              TextButton(onPressed: onEditDetails, child: const Text('Edit')),
            ],
          );
        }

        if (transfer.isCompleted) {
          return _ResultView(
            icon: Icons.check_circle,
            color: Colors.green,
            title: 'Sent!',
            message: '\$${transfer.amount ?? amount} to $recipient',
            primaryLabel: 'Done',
            onPrimary: () {
              ref.read(walletHomeControllerProvider.notifier).refresh();
              context.pop();
            },
          );
        }

        if (transfer.isFailed) {
          return _ResultView(
            icon: Icons.cancel,
            color: Theme.of(context).colorScheme.error,
            title: 'Transfer declined',
            message: transfer.failureReason ?? 'The transfer could not be completed.',
            primaryLabel: 'Retry',
            onPrimary: onRetry,
            secondaryLabel: 'Cancel',
            onSecondary: () => context.pop(),
          );
        }

        // Any other status (e.g. still processing) — same retry affordance,
        // safe because the Idempotency-Key is stable for this flow.
        return _ResultView(
          icon: Icons.hourglass_top,
          color: Theme.of(context).colorScheme.primary,
          title: 'Processing',
          message: 'Status: ${transfer.status ?? 'unknown'}',
          primaryLabel: 'Check again',
          onPrimary: onRetry,
        );
      },
    );
  }
}

class _SummaryRow extends StatelessWidget {
  const _SummaryRow({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [Text(label), Text(value, style: const TextStyle(fontWeight: FontWeight.bold))],
      ),
    );
  }
}

class _ResultView extends StatelessWidget {
  const _ResultView({
    required this.icon,
    required this.color,
    required this.title,
    required this.message,
    required this.primaryLabel,
    required this.onPrimary,
    this.secondaryLabel,
    this.onSecondary,
  });

  final IconData icon;
  final Color color;
  final String title;
  final String message;
  final String primaryLabel;
  final VoidCallback onPrimary;
  final String? secondaryLabel;
  final VoidCallback? onSecondary;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, color: color, size: 64),
          const SizedBox(height: 16),
          Text(title, style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: 8),
          Text(message, textAlign: TextAlign.center),
          const SizedBox(height: 24),
          FilledButton(onPressed: onPrimary, child: Text(primaryLabel)),
          if (secondaryLabel != null)
            TextButton(onPressed: onSecondary, child: Text(secondaryLabel!)),
        ],
      ),
    );
  }
}
