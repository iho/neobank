/// The OpenAPI `Transfer` schema declares no `required` fields at all, so
/// every property here is nullable to match the contract exactly (a 422
/// "failed" response, for instance, may omit `completed_at`/`ledger_transfer_id`).
class Transfer {
  const Transfer({
    this.id,
    this.status,
    this.senderUserId,
    this.recipientUserId,
    this.amount,
    this.currency,
    this.failureReason,
    this.memo,
    this.createdAt,
    this.completedAt,
  });

  factory Transfer.fromJson(Map<String, dynamic> json) => Transfer(
        id: json['id'] as String?,
        status: json['status'] as String?,
        senderUserId: json['sender_user_id'] as String?,
        recipientUserId: json['recipient_user_id'] as String?,
        amount: json['amount'] as String?,
        currency: json['currency'] as String?,
        failureReason: json['failure_reason'] as String?,
        memo: json['memo'] as String?,
        createdAt: json['created_at'] != null ? DateTime.parse(json['created_at'] as String) : null,
        completedAt:
            json['completed_at'] != null ? DateTime.parse(json['completed_at'] as String) : null,
      );

  final String? id;
  final String? status;
  final String? senderUserId;
  final String? recipientUserId;
  final String? amount;
  final String? currency;
  final String? failureReason;
  final String? memo;
  final DateTime? createdAt;
  final DateTime? completedAt;

  bool get isCompleted => status == 'completed';
  bool get isFailed => status == 'failed' || status == 'declined';
}

class TransferPage {
  const TransferPage({required this.transfers, this.nextCursor});

  factory TransferPage.fromJson(Map<String, dynamic> json) => TransferPage(
        transfers: (json['transfers'] as List<dynamic>)
            .map((e) => Transfer.fromJson(e as Map<String, dynamic>))
            .toList(),
        nextCursor: json['next_cursor'] as String?,
      );

  final List<Transfer> transfers;
  final String? nextCursor;
}
