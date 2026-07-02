// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'kyc_status_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

KycStatusResponse _$KycStatusResponseFromJson(Map<String, dynamic> json) =>
    KycStatusResponse(
      status: json['status'] as String,
      rejectionReason: json['rejection_reason'] as String?,
    );

Map<String, dynamic> _$KycStatusResponseToJson(KycStatusResponse instance) =>
    <String, dynamic>{
      'status': instance.status,
      'rejection_reason': instance.rejectionReason,
    };
