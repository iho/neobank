// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'submit_kyc_response.g.dart';

@JsonSerializable()
class SubmitKycResponse {
  const SubmitKycResponse({
    required this.kycCaseId,
    required this.status,
    this.walletId,
    this.rejectionReason,
  });
  
  factory SubmitKycResponse.fromJson(Map<String, Object?> json) => _$SubmitKycResponseFromJson(json);
  
  @JsonKey(name: 'kyc_case_id')
  final String kycCaseId;
  final String status;
  @JsonKey(name: 'wallet_id')
  final String? walletId;
  @JsonKey(name: 'rejection_reason')
  final String? rejectionReason;

  Map<String, Object?> toJson() => _$SubmitKycResponseToJson(this);
}
