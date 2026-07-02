// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'referral_invite.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ReferralInvite _$ReferralInviteFromJson(Map<String, dynamic> json) =>
    ReferralInvite(
      id: json['id'] as String,
      inviterUserId: json['inviter_user_id'] as String,
      inviteCode: json['invite_code'] as String,
      status: json['status'] as String,
      createdAt: DateTime.parse(json['created_at'] as String),
      inviteeUserId: json['invitee_user_id'] as String?,
      acceptedAt: json['accepted_at'] == null
          ? null
          : DateTime.parse(json['accepted_at'] as String),
    );

Map<String, dynamic> _$ReferralInviteToJson(ReferralInvite instance) =>
    <String, dynamic>{
      'id': instance.id,
      'inviter_user_id': instance.inviterUserId,
      'invite_code': instance.inviteCode,
      'invitee_user_id': instance.inviteeUserId,
      'status': instance.status,
      'created_at': instance.createdAt.toIso8601String(),
      'accepted_at': instance.acceptedAt?.toIso8601String(),
    };
