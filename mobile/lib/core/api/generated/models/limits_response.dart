// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

import 'transfer_limits.dart';

part 'limits_response.g.dart';

@JsonSerializable()
class LimitsResponse {
  const LimitsResponse({
    required this.p2p,
  });
  
  factory LimitsResponse.fromJson(Map<String, Object?> json) => _$LimitsResponseFromJson(json);
  
  final TransferLimits p2p;

  Map<String, Object?> toJson() => _$LimitsResponseToJson(this);
}
