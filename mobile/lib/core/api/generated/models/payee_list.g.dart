// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'payee_list.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PayeeList _$PayeeListFromJson(Map<String, dynamic> json) => PayeeList(
  payees: (json['payees'] as List<dynamic>)
      .map((e) => Payee.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$PayeeListToJson(PayeeList instance) => <String, dynamic>{
  'payees': instance.payees,
};
