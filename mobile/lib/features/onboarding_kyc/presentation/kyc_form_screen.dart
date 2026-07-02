import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/error/api_exception.dart';
import 'kyc_controller.dart';

/// Shown by the home shell whenever KYC status isn't `approved`. Handles both
/// the never-submitted case (plain form) and the rejected case (banner +
/// resubmit) — the backend reports both as ordinary KYC statuses, there's no
/// separate "not started" signal.
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
    return Scaffold(
      appBar: AppBar(title: const Text('Verify your identity')),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Form(
            key: _formKey,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                if (widget.rejectionReason != null) ...[
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: Theme.of(context).colorScheme.errorContainer,
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Text(
                      'Previous submission was rejected: ${widget.rejectionReason}',
                      style: TextStyle(
                        color: Theme.of(context).colorScheme.onErrorContainer,
                      ),
                    ),
                  ),
                  const SizedBox(height: 16),
                ],
                const Text('We need a few details before you can use your wallet.'),
                const SizedBox(height: 16),
                TextFormField(
                  controller: _fullNameController,
                  decoration: const InputDecoration(labelText: 'Full legal name'),
                  validator: (v) => (v == null || v.trim().isEmpty) ? 'Required' : null,
                ),
                const SizedBox(height: 16),
                InkWell(
                  onTap: _pickDateOfBirth,
                  child: InputDecorator(
                    decoration: const InputDecoration(labelText: 'Date of birth'),
                    child: Text(
                      _dateOfBirth == null
                          ? 'Select date'
                          : _dateOfBirth!.toIso8601String().split('T').first,
                    ),
                  ),
                ),
                const SizedBox(height: 16),
                TextFormField(
                  controller: _countryCodeController,
                  maxLength: 2,
                  textCapitalization: TextCapitalization.characters,
                  decoration: const InputDecoration(
                    labelText: 'Country code (ISO-2, e.g. US)',
                    counterText: '',
                  ),
                  validator: (v) =>
                      (v == null || v.trim().length != 2) ? 'Enter a 2-letter code' : null,
                ),
                const SizedBox(height: 16),
                TextFormField(
                  controller: _documentTypeController,
                  decoration: const InputDecoration(labelText: 'Document type (optional)'),
                ),
                const SizedBox(height: 16),
                TextFormField(
                  controller: _documentNumberController,
                  decoration: const InputDecoration(labelText: 'Document number (optional)'),
                ),
                const SizedBox(height: 24),
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

/// Rare with MVP auto-approve (submission usually resolves synchronously),
/// but the contract allows an async `pending` outcome — surfaced here with a
/// manual refresh in case a future backend makes KYC actually asynchronous.
class KycPendingScreen extends ConsumerWidget {
  const KycPendingScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Scaffold(
      appBar: AppBar(title: const Text('Verification pending')),
      body: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.hourglass_top, size: 48),
            const SizedBox(height: 16),
            const Text("We're reviewing your details."),
            const SizedBox(height: 16),
            OutlinedButton(
              onPressed: () => ref.read(kycControllerProvider.notifier).refresh(),
              child: const Text('Check status'),
            ),
          ],
        ),
      ),
    );
  }
}
