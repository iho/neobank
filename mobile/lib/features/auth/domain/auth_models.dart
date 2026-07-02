/// Response shape shared by `/v1/auth/login` and `/v1/auth/register`.
class AuthTokens {
  const AuthTokens({
    required this.userId,
    required this.accessToken,
    required this.refreshToken,
  });

  factory AuthTokens.fromJson(Map<String, dynamic> json) => AuthTokens(
        userId: json['user_id'] as String,
        accessToken: json['access_token'] as String,
        refreshToken: json['refresh_token'] as String,
      );

  final String userId;
  final String accessToken;
  final String refreshToken;
}

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

  factory Profile.fromJson(Map<String, dynamic> json) => Profile(
        userId: json['user_id'] as String,
        email: json['email'] as String,
        phone: json['phone'] as String,
        status: json['status'] as String,
        kycStatus: json['kyc_status'] as String,
        createdAt: DateTime.parse(json['created_at'] as String),
        fullName: json['full_name'] as String?,
        dateOfBirth: json['date_of_birth'] as String?,
        countryCode: json['country_code'] as String?,
      );

  final String userId;
  final String email;
  final String phone;
  final String status;
  final String kycStatus;
  final DateTime createdAt;
  final String? fullName;
  final String? dateOfBirth;
  final String? countryCode;
}
