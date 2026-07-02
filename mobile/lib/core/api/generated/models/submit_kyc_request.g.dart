// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'submit_kyc_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

SubmitKycRequest _$SubmitKycRequestFromJson(Map<String, dynamic> json) =>
    SubmitKycRequest(
      fullName: json['full_name'] as String,
      dateOfBirth: DateTime.parse(json['date_of_birth'] as String),
      countryCode: json['country_code'] as String,
      documentType: json['document_type'] as String?,
      documentNumber: json['document_number'] as String?,
    );

Map<String, dynamic> _$SubmitKycRequestToJson(SubmitKycRequest instance) =>
    <String, dynamic>{
      'full_name': instance.fullName,
      'date_of_birth': instance.dateOfBirth.toIso8601String(),
      'country_code': instance.countryCode,
      'document_type': instance.documentType,
      'document_number': instance.documentNumber,
    };
