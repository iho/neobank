// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'submit_kyc_request.g.dart';

@JsonSerializable()
class SubmitKycRequest {
  const SubmitKycRequest({
    required this.fullName,
    required this.dateOfBirth,
    required this.countryCode,
    this.documentType,
    this.documentNumber,
  });
  
  factory SubmitKycRequest.fromJson(Map<String, Object?> json) => _$SubmitKycRequestFromJson(json);
  
  @JsonKey(name: 'full_name')
  final String fullName;
  @JsonKey(name: 'date_of_birth')
  final DateTime dateOfBirth;
  @JsonKey(name: 'country_code')
  final String countryCode;
  @JsonKey(name: 'document_type')
  final String? documentType;
  @JsonKey(name: 'document_number')
  final String? documentNumber;

  Map<String, Object?> toJson() => _$SubmitKycRequestToJson(this);
}
