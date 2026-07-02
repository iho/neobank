// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'transfer_limits.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

TransferLimits _$TransferLimitsFromJson(Map<String, dynamic> json) =>
    TransferLimits(
      hourlyTransferCount: LimitGauge.fromJson(
        json['hourly_transfer_count'] as Map<String, dynamic>,
      ),
      dailyTransferAmount: LimitGauge.fromJson(
        json['daily_transfer_amount'] as Map<String, dynamic>,
      ),
      singleTransferMax: json['single_transfer_max'] as String,
    );

Map<String, dynamic> _$TransferLimitsToJson(TransferLimits instance) =>
    <String, dynamic>{
      'hourly_transfer_count': instance.hourlyTransferCount,
      'daily_transfer_amount': instance.dailyTransferAmount,
      'single_transfer_max': instance.singleTransferMax,
    };
