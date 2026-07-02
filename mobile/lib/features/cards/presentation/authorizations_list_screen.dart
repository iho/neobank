import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../domain/card_models.dart';
import 'authorizations_controller.dart';

class AuthorizationsListScreen extends ConsumerWidget {
  const AuthorizationsListScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final authorizations = ref.watch(authorizationsControllerProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Authorizations')),
      body: authorizations.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (error, _) => Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text('$error'),
              const SizedBox(height: 12),
              OutlinedButton(
                onPressed: () => ref.read(authorizationsControllerProvider.notifier).refresh(),
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
        data: (list) => list.isEmpty
            ? const Center(child: Text('No authorizations yet'))
            : RefreshIndicator(
                onRefresh: () => ref.read(authorizationsControllerProvider.notifier).refresh(),
                child: ListView.builder(
                  itemCount: list.length,
                  itemBuilder: (context, index) => _AuthorizationTile(auth: list[index]),
                ),
              ),
      ),
    );
  }
}

class _AuthorizationTile extends StatelessWidget {
  const _AuthorizationTile({required this.auth});

  final CardAuthorization auth;

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: Icon(
        switch (auth.status) {
          'captured' => Icons.check_circle,
          'declined' => Icons.cancel,
          _ => Icons.hourglass_top,
        },
      ),
      title: Text(auth.merchantName ?? 'Unknown merchant'),
      subtitle: Text(auth.status),
      trailing: Text('${auth.amount} ${auth.currency}'),
      onTap: () => context.push('/authorizations/${auth.id}', extra: auth),
    );
  }
}
