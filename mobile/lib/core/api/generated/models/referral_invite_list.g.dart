// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'referral_invite_list.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ReferralInviteList _$ReferralInviteListFromJson(Map<String, dynamic> json) =>
    ReferralInviteList(
      invites: (json['invites'] as List<dynamic>)
          .map((e) => ReferralInvite.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$ReferralInviteListToJson(ReferralInviteList instance) =>
    <String, dynamic>{'invites': instance.invites};
