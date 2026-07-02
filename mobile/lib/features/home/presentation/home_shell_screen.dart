import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../cards/presentation/cards_list_screen.dart';
import '../../notifications/presentation/notifications_screen.dart';
import '../../onboarding_kyc/domain/kyc_models.dart';
import '../../onboarding_kyc/presentation/kyc_controller.dart';
import '../../onboarding_kyc/presentation/kyc_form_screen.dart';
import '../../wallet/presentation/wallet_home_screen.dart';

/// Root screen after login. Gates the whole app on KYC status before
/// showing the tabbed shell — see mobile/TODO.md, "Onboarding/KYC ...
/// gate wallet features on approved".
class HomeShellScreen extends ConsumerWidget {
  const HomeShellScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final kyc = ref.watch(kycControllerProvider);

    return kyc.when(
      loading: () => const Scaffold(body: Center(child: CircularProgressIndicator())),
      error: (error, _) => Scaffold(
        body: Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text('$error'),
              const SizedBox(height: 12),
              OutlinedButton(
                onPressed: () => ref.read(kycControllerProvider.notifier).refresh(),
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
      ),
      data: (info) {
        switch (info.status) {
          case KycStatus.approved:
            return const _MainTabsScreen();
          case KycStatus.rejected:
            return KycFormScreen(rejectionReason: info.rejectionReason);
          case KycStatus.pending:
          case KycStatus.manualReview:
            final hasSubmitted =
                ref.read(kycControllerProvider.notifier).hasSubmittedThisSession;
            return hasSubmitted ? const KycPendingScreen() : const KycFormScreen(rejectionReason: null);
        }
      },
    );
  }
}

class _MainTabsScreen extends StatefulWidget {
  const _MainTabsScreen();

  @override
  State<_MainTabsScreen> createState() => _MainTabsScreenState();
}

class _MainTabsScreenState extends State<_MainTabsScreen> {
  int _index = 0;

  static const _tabs = [
    WalletHomeScreen(),
    CardsListScreen(),
    NotificationsScreen(),
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: IndexedStack(index: _index, children: _tabs),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _index,
        onDestinationSelected: (i) => setState(() => _index = i),
        destinations: const [
          NavigationDestination(icon: Icon(Icons.account_balance_wallet), label: 'Wallet'),
          NavigationDestination(icon: Icon(Icons.credit_card), label: 'Cards'),
          NavigationDestination(icon: Icon(Icons.notifications), label: 'Alerts'),
        ],
      ),
    );
  }
}
