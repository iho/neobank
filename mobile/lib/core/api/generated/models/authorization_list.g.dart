// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'authorization_list.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

AuthorizationList _$AuthorizationListFromJson(Map<String, dynamic> json) =>
    AuthorizationList(
      authorizations: (json['authorizations'] as List<dynamic>)
          .map((e) => CardAuthorization.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$AuthorizationListToJson(AuthorizationList instance) =>
    <String, dynamic>{'authorizations': instance.authorizations};
