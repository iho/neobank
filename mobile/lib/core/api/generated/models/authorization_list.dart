// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

import 'card_authorization.dart';

part 'authorization_list.g.dart';

@JsonSerializable()
class AuthorizationList {
  const AuthorizationList({
    required this.authorizations,
  });
  
  factory AuthorizationList.fromJson(Map<String, Object?> json) => _$AuthorizationListFromJson(json);
  
  final List<CardAuthorization> authorizations;

  Map<String, Object?> toJson() => _$AuthorizationListToJson(this);
}
