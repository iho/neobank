// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

import 'payee.dart';

part 'payee_list.g.dart';

@JsonSerializable()
class PayeeList {
  const PayeeList({
    required this.payees,
  });
  
  factory PayeeList.fromJson(Map<String, Object?> json) => _$PayeeListFromJson(json);
  
  final List<Payee> payees;

  Map<String, Object?> toJson() => _$PayeeListToJson(this);
}
