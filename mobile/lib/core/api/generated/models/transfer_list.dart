// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

import 'transfer.dart';

part 'transfer_list.g.dart';

@JsonSerializable()
class TransferList {
  const TransferList({
    required this.transfers,
    this.nextCursor,
  });
  
  factory TransferList.fromJson(Map<String, Object?> json) => _$TransferListFromJson(json);
  
  final List<Transfer> transfers;
  @JsonKey(name: 'next_cursor')
  final String? nextCursor;

  Map<String, Object?> toJson() => _$TransferListToJson(this);
}
