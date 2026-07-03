import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/error/api_exception.dart';
import '../../../core/theme/app_theme.dart';
import '../../auth/presentation/auth_controller.dart';
import 'kyc_controller.dart';

/// Shown by the home shell whenever KYC status isn't `approved`. Handles both
/// the never-submitted case (plain form) and the rejected case (banner +
/// resubmit) — the backend reports both as ordinary KYC statuses, there's no
/// separate "not started" signal.
///
/// This screen replaces the home shell's body in place rather than being
/// pushed onto the navigator, so there's no natural "back" destination — the
/// leading button instead signs the user out, which is the only sensible
/// escape hatch from an incomplete-KYC state.
class KycFormScreen extends ConsumerStatefulWidget {
  const KycFormScreen({required this.rejectionReason, super.key});

  final String? rejectionReason;

  @override
  ConsumerState<KycFormScreen> createState() => _KycFormScreenState();
}

class _KycFormScreenState extends ConsumerState<KycFormScreen> {
  final _formKey = GlobalKey<FormState>();
  final _fullNameController = TextEditingController();
  final _countryCodeController = TextEditingController();
  final _documentTypeController = TextEditingController();
  final _documentNumberController = TextEditingController();
  DateTime? _dateOfBirth;
  bool _submitting = false;

  @override
  void dispose() {
    _fullNameController.dispose();
    _countryCodeController.dispose();
    _documentTypeController.dispose();
    _documentNumberController.dispose();
    super.dispose();
  }

  Future<void> _pickDateOfBirth() async {
    final now = DateTime.now();
    final picked = await showDatePicker(
      context: context,
      initialDate: DateTime(now.year - 18, now.month, now.day),
      firstDate: DateTime(now.year - 120),
      lastDate: DateTime(now.year - 13, now.month, now.day),
    );
    if (picked != null) setState(() => _dateOfBirth = picked);
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate() || _dateOfBirth == null) {
      if (_dateOfBirth == null) {
        ScaffoldMessenger.of(context)
            .showSnackBar(const SnackBar(content: Text('Select your date of birth')));
      }
      return;
    }
    setState(() => _submitting = true);
    try {
      await ref.read(kycControllerProvider.notifier).submit(
            fullName: _fullNameController.text.trim(),
            dateOfBirth: _dateOfBirth!.toIso8601String().split('T').first,
            countryCode: _countryCodeController.text.trim().toUpperCase(),
            documentType: _documentTypeController.text.trim(),
            documentNumber: _documentNumberController.text.trim(),
          );
    } on ApiException catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(e.message)));
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;

    return Scaffold(
      appBar: AppBar(
        leading: _BackToLoginButton(),
        title: const Text('Verify your identity'),
      ),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.fromLTRB(
            AppSpacing.lg,
            AppSpacing.sm,
            AppSpacing.lg,
            AppSpacing.lg,
          ),
          child: Form(
            key: _formKey,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Container(
                  width: 72,
                  height: 72,
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    color: scheme.primary.withValues(alpha: 0.12),
                    borderRadius: BorderRadius.circular(AppTheme.tileRadius),
                  ),
                  child: Icon(
                    Icons.verified_user_rounded,
                    size: 36,
                    color: scheme.primary,
                  ),
                ),
                const SizedBox(height: AppSpacing.md),
                Text(
                  'A few quick details',
                  style: Theme.of(context).textTheme.headlineMedium,
                ),
                const SizedBox(height: AppSpacing.xs),
                Text(
                  "We're required to verify your identity before your wallet can go live.",
                  style: Theme.of(context).textTheme.bodyLarge,
                ),
                const SizedBox(height: AppSpacing.lg),
                if (widget.rejectionReason != null) ...[
                  _RejectionBanner(reason: widget.rejectionReason!),
                  const SizedBox(height: AppSpacing.lg),
                ],
                TextFormField(
                  controller: _fullNameController,
                  textCapitalization: TextCapitalization.words,
                  decoration: const InputDecoration(
                    labelText: 'Full legal name',
                    prefixIcon: Icon(Icons.badge_outlined),
                  ),
                  validator: (v) => (v == null || v.trim().isEmpty) ? 'Required' : null,
                ),
                const SizedBox(height: AppSpacing.md),
                InkWell(
                  borderRadius: BorderRadius.circular(16),
                  onTap: _pickDateOfBirth,
                  child: InputDecorator(
                    decoration: const InputDecoration(
                      labelText: 'Date of birth',
                      prefixIcon: Icon(Icons.cake_outlined),
                    ),
                    child: Text(
                      _dateOfBirth == null
                          ? 'Select date'
                          : _dateOfBirth!.toIso8601String().split('T').first,
                    ),
                  ),
                ),
                const SizedBox(height: AppSpacing.md),
                TextFormField(
                  controller: _countryCodeController,
                  maxLength: 2,
                  textCapitalization: TextCapitalization.characters,
                  decoration: const InputDecoration(
                    labelText: 'Country code (ISO-2, e.g. US)',
                    prefixIcon: Icon(Icons.public_outlined),
                    counterText: '',
                  ),
                  validator: (v) =>
                      (v == null || v.trim().length != 2) ? 'Enter a 2-letter code' : null,
                ),
                const SizedBox(height: AppSpacing.md),
                TextFormField(
                  controller: _documentTypeController,
                  decoration: const InputDecoration(
                    labelText: 'Document type (optional)',
                    prefixIcon: Icon(Icons.description_outlined),
                  ),
                ),
                const SizedBox(height: AppSpacing.md),
                TextFormField(
                  controller: _documentNumberController,
                  decoration: const InputDecoration(
                    labelText: 'Document number (optional)',
                    prefixIcon: Icon(Icons.numbers_outlined),
                  ),
                ),
                const SizedBox(height: AppSpacing.xl),
                FilledButton(
                  onPressed: _submitting ? null : _submit,
                  child: _submitting
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Text('Submit'),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _RejectionBanner extends StatelessWidget {
  const _RejectionBanner({required this.reason});

  final String reason;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return Container(
      padding: const EdgeInsets.all(AppSpacing.md),
      decoration: BoxDecoration(
        color: scheme.errorContainer,
        borderRadius: BorderRadius.circular(16),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(Icons.error_outline_rounded, color: scheme.onErrorContainer),
          const SizedBox(width: AppSpacing.sm),
          Expanded(
            child: Text(
              'Previous submission was rejected: $reason',
              style: TextStyle(color: scheme.onErrorContainer),
            ),
          ),
        ],
      ),
    );
  }
}

/// Leading AppBar button for KYC screens: signs out after a confirmation,
/// since there is no previous route to pop back to.
class _BackToLoginButton extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return IconButton(
      icon: const Icon(Icons.arrow_back_rounded),
      tooltip: 'Log out',
      onPressed: () async {
        final confirmed = await showDialog<bool>(
          context: context,
          builder: (context) => AlertDialog(
            title: const Text('Log out?'),
            content: const Text(
              "You'll need to sign back in to finish verifying your identity.",
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.of(context).pop(false),
                child: const Text('Cancel'),
              ),
              FilledButton(
                onPressed: () => Navigator.of(context).pop(true),
                child: const Text('Log out'),
              ),
            ],
          ),
        );
        if (confirmed == true) {
          await ref.read(authControllerProvider.notifier).logout();
        }
      },
    );
  }
}

/// Rare with MVP auto-approve (submission usually resolves synchronously),
/// but the contract allows an async `pending` outcome — surfaced here with a
/// manual refresh in case a future backend makes KYC actually asynchronous.
class KycPendingScreen extends ConsumerWidget {
  const KycPendingScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final scheme = Theme.of(context).colorScheme;
    return Scaffold(
      appBar: AppBar(leading: _BackToLoginButton()),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.lg),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 72,
                height: 72,
                alignment: Alignment.center,
                decoration: BoxDecoration(
                  color: scheme.primary.withValues(alpha: 0.12),
                  borderRadius: BorderRadius.circular(AppTheme.tileRadius),
                ),
                child: Icon(Icons.hourglass_top_rounded, size: 36, color: scheme.primary),
              ),
              const SizedBox(height: AppSpacing.md),
              Text("We're reviewing your details", style: Theme.of(context).textTheme.titleLarge),
              const SizedBox(height: AppSpacing.xs),
              Text(
                "This usually only takes a moment.",
                style: Theme.of(context).textTheme.bodyLarge,
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: AppSpacing.lg),
              OutlinedButton(
                onPressed: () => ref.read(kycControllerProvider.notifier).refresh(),
                child: const Text('Check status'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
