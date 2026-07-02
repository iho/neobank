/// Mirrors `domain.KYCStatus` on the backend (services/user/internal/domain/user.go).
/// Before any KYC case exists the backend still reports `pending` — there is
/// no distinct "not started" wire value, so the form and the pending-review
/// screen share this one status.
enum KycStatus { pending, approved, rejected, manualReview }

KycStatus parseKycStatus(String raw) => switch (raw) {
      'approved' => KycStatus.approved,
      'rejected' => KycStatus.rejected,
      'manual_review' => KycStatus.manualReview,
      _ => KycStatus.pending,
    };

class KycStatusInfo {
  const KycStatusInfo({required this.status, this.rejectionReason});

  factory KycStatusInfo.fromJson(Map<String, dynamic> json) => KycStatusInfo(
        status: parseKycStatus(json['status'] as String),
        rejectionReason: json['rejection_reason'] as String?,
      );

  final KycStatus status;
  final String? rejectionReason;
}

class KycSubmitResult {
  const KycSubmitResult({
    required this.kycCaseId,
    required this.status,
    this.walletId,
    this.rejectionReason,
  });

  factory KycSubmitResult.fromJson(Map<String, dynamic> json) => KycSubmitResult(
        kycCaseId: json['kyc_case_id'] as String,
        status: parseKycStatus(json['status'] as String),
        walletId: json['wallet_id'] as String?,
        rejectionReason: json['rejection_reason'] as String?,
      );

  final String kycCaseId;
  final KycStatus status;
  final String? walletId;
  final String? rejectionReason;
}
