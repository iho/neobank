// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'update_card_controls_request.g.dart';

@JsonSerializable()
class UpdateCardControlsRequest {
  const UpdateCardControlsRequest({
    this.dailyLimit,
    this.onlineOnly,
  });
  
  factory UpdateCardControlsRequest.fromJson(Map<String, Object?> json) => _$UpdateCardControlsRequestFromJson(json);
  
  @JsonKey(name: 'daily_limit')
  final String? dailyLimit;
  @JsonKey(name: 'online_only')
  final bool? onlineOnly;

  Map<String, Object?> toJson() => _$UpdateCardControlsRequestToJson(this);
}
