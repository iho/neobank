import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';

import '../data/transfer_repository.dart';
import '../domain/transfer_models.dart';

final transferSubmitControllerProvider =
    NotifierProvider.autoDispose<TransferSubmitController, AsyncValue<Transfer?>>(
  TransferSubmitController.new,
);

/// One instance per transfer flow (`.autoDispose` — torn down when the flow
/// screen is popped).
///
/// The `Idempotency-Key` is kept stable across a retry *only* when the
/// previous attempt ended in a transport-level error (timeout, dropped
/// connection, etc.) — in that case we genuinely don't know whether the
/// server processed the request, and resending the same key lets the
/// idempotency middleware return the original outcome instead of double-
/// sending. But once the server has given a definitive answer — completed
/// *or* failed (e.g. declined by a fraud/risk rule) — that key is
/// permanently associated with that outcome: retrying with the same key
/// would just replay the exact same cached response forever, so a declined
/// transfer could never be retried successfully. A fresh key is minted for
/// any retry after a definitive response, since that's a deliberate new
/// attempt.
class TransferSubmitController extends AutoDisposeNotifier<AsyncValue<Transfer?>> {
  String _idempotencyKey = const Uuid().v4();

  @override
  AsyncValue<Transfer?> build() => const AsyncData(null);

  Future<void> submit({
    required String amount,
    String? recipientPhone,
    String? recipientEmail,
    String? recipientUserId,
    String? currency,
    String? memo,
  }) async {
    if (state.valueOrNull != null) {
      _idempotencyKey = const Uuid().v4();
    }
    state = const AsyncLoading();
    state = await AsyncValue.guard(
      () => ref.read(transferRepositoryProvider).createTransfer(
            idempotencyKey: _idempotencyKey,
            amount: amount,
            recipientPhone: recipientPhone,
            recipientEmail: recipientEmail,
            recipientUserId: recipientUserId,
            currency: currency,
            memo: memo,
          ),
    );
  }
}
