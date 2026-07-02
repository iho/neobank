// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'kyc_status_response.g.dart';

@JsonSerializable()
class KycStatusResponse {
  const KycStatusResponse({
    required this.status,
    this.rejectionReason,
  });
  
  factory KycStatusResponse.fromJson(Map<String, Object?> json) => _$KycStatusResponseFromJson(json);
  
  final String status;
  @JsonKey(name: 'rejection_reason')
  final String? rejectionReason;

  Map<String, Object?> toJson() => _$KycStatusResponseToJson(this);
}
