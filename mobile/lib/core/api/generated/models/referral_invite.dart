// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'referral_invite.g.dart';

@JsonSerializable()
class ReferralInvite {
  const ReferralInvite({
    required this.id,
    required this.inviterUserId,
    required this.inviteCode,
    required this.status,
    required this.createdAt,
    this.inviteeUserId,
    this.acceptedAt,
  });
  
  factory ReferralInvite.fromJson(Map<String, Object?> json) => _$ReferralInviteFromJson(json);
  
  final String id;
  @JsonKey(name: 'inviter_user_id')
  final String inviterUserId;
  @JsonKey(name: 'invite_code')
  final String inviteCode;
  @JsonKey(name: 'invitee_user_id')
  final String? inviteeUserId;
  final String status;
  @JsonKey(name: 'created_at')
  final DateTime createdAt;
  @JsonKey(name: 'accepted_at')
  final DateTime? acceptedAt;

  Map<String, Object?> toJson() => _$ReferralInviteToJson(this);
}
