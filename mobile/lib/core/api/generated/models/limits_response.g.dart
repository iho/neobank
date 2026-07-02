// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'limits_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

LimitsResponse _$LimitsResponseFromJson(Map<String, dynamic> json) =>
    LimitsResponse(
      p2p: TransferLimits.fromJson(json['p2p'] as Map<String, dynamic>),
    );

Map<String, dynamic> _$LimitsResponseToJson(LimitsResponse instance) =>
    <String, dynamic>{'p2p': instance.p2p};
