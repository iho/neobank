// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'limit_gauge.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

LimitGauge _$LimitGaugeFromJson(Map<String, dynamic> json) => LimitGauge(
  limit: json['limit'] as String,
  used: json['used'] as String,
  remaining: json['remaining'] as String,
);

Map<String, dynamic> _$LimitGaugeToJson(LimitGauge instance) =>
    <String, dynamic>{
      'limit': instance.limit,
      'used': instance.used,
      'remaining': instance.remaining,
    };
