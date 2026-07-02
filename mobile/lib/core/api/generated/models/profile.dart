// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'profile.g.dart';

@JsonSerializable()
class Profile {
  const Profile({
    required this.userId,
    required this.email,
    required this.phone,
    required this.status,
    required this.kycStatus,
    required this.createdAt,
    this.fullName,
    this.dateOfBirth,
    this.countryCode,
  });
  
  factory Profile.fromJson(Map<String, Object?> json) => _$ProfileFromJson(json);
  
  @JsonKey(name: 'user_id')
  final String userId;
  final String email;
  final String phone;
  final String status;
  @JsonKey(name: 'full_name')
  final String? fullName;
  @JsonKey(name: 'date_of_birth')
  final DateTime? dateOfBirth;
  @JsonKey(name: 'country_code')
  final String? countryCode;
  @JsonKey(name: 'kyc_status')
  final String kycStatus;
  @JsonKey(name: 'created_at')
  final DateTime createdAt;

  Map<String, Object?> toJson() => _$ProfileToJson(this);
}
