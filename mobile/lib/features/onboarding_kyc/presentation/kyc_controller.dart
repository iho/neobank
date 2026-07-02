import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/kyc_repository.dart';
import '../domain/kyc_models.dart';

final kycControllerProvider =
    AsyncNotifierProvider<KycController, KycStatusInfo>(KycController.new);

class KycController extends AsyncNotifier<KycStatusInfo> {
  /// The backend can't distinguish "never submitted" from "submitted, still
  /// pending" — both report status `pending`. Track submission locally so the
  /// home shell can tell the form apart from the pending-review screen.
  bool hasSubmittedThisSession = false;

  @override
  Future<KycStatusInfo> build() => ref.read(kycRepositoryProvider).getStatus();

  Future<void> refresh() async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(() => ref.read(kycRepositoryProvider).getStatus());
  }

  /// Throws on failure — the form screen shows the error inline rather than
  /// this notifier going into an error state (which would blank the form).
  Future<void> submit({
    required String fullName,
    required String dateOfBirth,
    required String countryCode,
    String? documentType,
    String? documentNumber,
  }) async {
    final result = await ref.read(kycRepositoryProvider).submit(
          fullName: fullName,
          dateOfBirth: dateOfBirth,
          countryCode: countryCode,
          documentType: documentType,
          documentNumber: documentNumber,
        );
    hasSubmittedThisSession = true;
    state = AsyncData(
      KycStatusInfo(status: result.status, rejectionReason: result.rejectionReason),
    );
  }
}
