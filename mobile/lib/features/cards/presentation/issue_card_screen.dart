import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/error/api_exception.dart';
import 'cards_controller.dart';

class IssueCardScreen extends ConsumerStatefulWidget {
  const IssueCardScreen({super.key});

  @override
  ConsumerState<IssueCardScreen> createState() => _IssueCardScreenState();
}

class _IssueCardScreenState extends ConsumerState<IssueCardScreen> {
  final _formKey = GlobalKey<FormState>();
  final _cardholderController = TextEditingController();
  final _dailyLimitController = TextEditingController();
  bool _onlineOnly = false;
  bool _submitting = false;

  @override
  void dispose() {
    _cardholderController.dispose();
    _dailyLimitController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _submitting = true);
    try {
      await ref.read(cardsControllerProvider.notifier).issueCard(
            cardholderName: _cardholderController.text.trim(),
            dailyLimit: _dailyLimitController.text.trim(),
            onlineOnly: _onlineOnly,
          );
      if (mounted) context.pop();
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
      appBar: AppBar(title: const Text('Issue a virtual card')),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Form(
            key: _formKey,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                TextFormField(
                  controller: _cardholderController,
                  decoration: const InputDecoration(labelText: 'Cardholder name'),
                  validator: (v) => (v == null || v.trim().isEmpty) ? 'Required' : null,
                ),
                const SizedBox(height: 16),
                TextFormField(
                  controller: _dailyLimitController,
                  keyboardType: const TextInputType.numberWithOptions(decimal: true),
                  decoration: const InputDecoration(labelText: 'Daily limit (optional)'),
                ),
                const SizedBox(height: 8),
                SwitchListTile(
                  contentPadding: EdgeInsets.zero,
                  title: const Text('Online purchases only'),
                  value: _onlineOnly,
                  onChanged: (v) => setState(() => _onlineOnly = v),
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
                      : const Text('Issue card'),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
