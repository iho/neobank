// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'limit_gauge.g.dart';

@JsonSerializable()
class LimitGauge {
  const LimitGauge({
    required this.limit,
    required this.used,
    required this.remaining,
  });
  
  factory LimitGauge.fromJson(Map<String, Object?> json) => _$LimitGaugeFromJson(json);
  
  final String limit;
  final String used;
  final String remaining;

  Map<String, Object?> toJson() => _$LimitGaugeToJson(this);
}
