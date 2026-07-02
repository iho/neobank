import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';

import '../data/transfer_repository.dart';
import '../domain/transfer_models.dart';

final transferSubmitControllerProvider =
    NotifierProvider.autoDispose<TransferSubmitController, AsyncValue<Transfer?>>(
  TransferSubmitController.new,
);

/// One instance per transfer flow (`.autoDispose` — torn down when the flow
/// screen is popped). Generates the `Idempotency-Key` once and reuses it on
/// every retry, so a client timeout followed by a retry can't double-send
/// the transfer.
class TransferSubmitController extends AutoDisposeNotifier<AsyncValue<Transfer?>> {
  final _idempotencyKey = const Uuid().v4();

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
