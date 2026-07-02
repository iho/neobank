// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'health_response.g.dart';

@JsonSerializable()
class HealthResponse {
  const HealthResponse({
    required this.status,
    required this.service,
  });
  
  factory HealthResponse.fromJson(Map<String, Object?> json) => _$HealthResponseFromJson(json);
  
  final String status;
  final String service;

  Map<String, Object?> toJson() => _$HealthResponseToJson(this);
}
