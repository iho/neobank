// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

import 'referral_invite.dart';

part 'referral_invite_list.g.dart';

@JsonSerializable()
class ReferralInviteList {
  const ReferralInviteList({
    required this.invites,
  });
  
  factory ReferralInviteList.fromJson(Map<String, Object?> json) => _$ReferralInviteListFromJson(json);
  
  final List<ReferralInvite> invites;

  Map<String, Object?> toJson() => _$ReferralInviteListToJson(this);
}
