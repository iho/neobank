// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'mark_notifications_read_response.g.dart';

@JsonSerializable()
class MarkNotificationsReadResponse {
  const MarkNotificationsReadResponse({
    required this.markedCount,
  });
  
  factory MarkNotificationsReadResponse.fromJson(Map<String, Object?> json) => _$MarkNotificationsReadResponseFromJson(json);
  
  @JsonKey(name: 'marked_count')
  final int markedCount;

  Map<String, Object?> toJson() => _$MarkNotificationsReadResponseToJson(this);
}
