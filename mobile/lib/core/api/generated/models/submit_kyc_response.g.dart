// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'submit_kyc_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

SubmitKycResponse _$SubmitKycResponseFromJson(Map<String, dynamic> json) =>
    SubmitKycResponse(
      kycCaseId: json['kyc_case_id'] as String,
      status: json['status'] as String,
      walletId: json['wallet_id'] as String?,
      rejectionReason: json['rejection_reason'] as String?,
    );

Map<String, dynamic> _$SubmitKycResponseToJson(SubmitKycResponse instance) =>
    <String, dynamic>{
      'kyc_case_id': instance.kycCaseId,
      'status': instance.status,
      'wallet_id': instance.walletId,
      'rejection_reason': instance.rejectionReason,
    };
