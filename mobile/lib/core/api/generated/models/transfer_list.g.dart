// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'transfer_list.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

TransferList _$TransferListFromJson(Map<String, dynamic> json) => TransferList(
  transfers: (json['transfers'] as List<dynamic>)
      .map((e) => Transfer.fromJson(e as Map<String, dynamic>))
      .toList(),
  nextCursor: json['next_cursor'] as String?,
);

Map<String, dynamic> _$TransferListToJson(TransferList instance) =>
    <String, dynamic>{
      'transfers': instance.transfers,
      'next_cursor': instance.nextCursor,
    };
