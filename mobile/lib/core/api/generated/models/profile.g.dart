// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'profile.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

Profile _$ProfileFromJson(Map<String, dynamic> json) => Profile(
  userId: json['user_id'] as String,
  email: json['email'] as String,
  phone: json['phone'] as String,
  status: json['status'] as String,
  kycStatus: json['kyc_status'] as String,
  createdAt: DateTime.parse(json['created_at'] as String),
  fullName: json['full_name'] as String?,
  dateOfBirth: json['date_of_birth'] == null
      ? null
      : DateTime.parse(json['date_of_birth'] as String),
  countryCode: json['country_code'] as String?,
);

Map<String, dynamic> _$ProfileToJson(Profile instance) => <String, dynamic>{
  'user_id': instance.userId,
  'email': instance.email,
  'phone': instance.phone,
  'status': instance.status,
  'full_name': instance.fullName,
  'date_of_birth': instance.dateOfBirth?.toIso8601String(),
  'country_code': instance.countryCode,
  'kyc_status': instance.kycStatus,
  'created_at': instance.createdAt.toIso8601String(),
};
