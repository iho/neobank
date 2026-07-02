import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/error/api_exception.dart';
import '../domain/card_models.dart';
import 'authorizations_controller.dart';

class AuthorizationDetailScreen extends ConsumerStatefulWidget {
  const AuthorizationDetailScreen({
    required this.authorizationId,
    required this.initialAuthorization,
    super.key,
  });

  final String authorizationId;
  final CardAuthorization? initialAuthorization;

  @override
  ConsumerState<AuthorizationDetailScreen> createState() => _AuthorizationDetailScreenState();
}

class _AuthorizationDetailScreenState extends ConsumerState<AuthorizationDetailScreen> {
  bool _capturing = false;

  Future<void> _capture() async {
    setState(() => _capturing = true);
    try {
      await ref.read(authorizationsControllerProvider.notifier).capture(widget.authorizationId);
    } on ApiException catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(e.message)));
    } finally {
      if (mounted) setState(() => _capturing = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final list = ref.watch(authorizationsControllerProvider).valueOrNull;
    CardAuthorization? auth = widget.initialAuthorization;
    if (list != null) {
      for (final a in list) {
        if (a.id == widget.authorizationId) {
          auth = a;
          break;
        }
      }
    }

    if (auth == null) {
      return const Scaffold(body: Center(child: Text('Authorization not found')));
    }
    final resolved = auth;

    return Scaffold(
      appBar: AppBar(title: Text(resolved.merchantName ?? 'Authorization')),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(20),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        '${resolved.amount} ${resolved.currency}',
                        style: Theme.of(context).textTheme.headlineSmall,
                      ),
                      const SizedBox(height: 12),
                      Text('Status: ${resolved.status}'),
                      if (resolved.merchantCategoryCode != null) ...[
                        const SizedBox(height: 4),
                        Text('Merchant category: ${resolved.merchantCategoryCode}'),
                      ],
                      if (resolved.failureReason != null) ...[
                        const SizedBox(height: 4),
                        Text('Reason: ${resolved.failureReason}'),
                      ],
                      if (resolved.createdAt != null) ...[
                        const SizedBox(height: 4),
                        Text('Authorized: ${resolved.createdAt!.toLocal()}'.split('.').first),
                      ],
                      if (resolved.capturedAt != null) ...[
                        const SizedBox(height: 4),
                        Text('Captured: ${resolved.capturedAt!.toLocal()}'.split('.').first),
                      ],
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 24),
              if (resolved.isCapturable)
                FilledButton(
                  onPressed: _capturing ? null : _capture,
                  child: _capturing
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Text('Capture'),
                ),
            ],
          ),
        ),
      ),
    );
  }
}
