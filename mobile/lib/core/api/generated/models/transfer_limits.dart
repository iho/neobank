// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

import 'limit_gauge.dart';

part 'transfer_limits.g.dart';

@JsonSerializable()
class TransferLimits {
  const TransferLimits({
    required this.hourlyTransferCount,
    required this.dailyTransferAmount,
    required this.singleTransferMax,
  });
  
  factory TransferLimits.fromJson(Map<String, Object?> json) => _$TransferLimitsFromJson(json);
  
  @JsonKey(name: 'hourly_transfer_count')
  final LimitGauge hourlyTransferCount;
  @JsonKey(name: 'daily_transfer_amount')
  final LimitGauge dailyTransferAmount;
  @JsonKey(name: 'single_transfer_max')
  final String singleTransferMax;

  Map<String, Object?> toJson() => _$TransferLimitsToJson(this);
}
