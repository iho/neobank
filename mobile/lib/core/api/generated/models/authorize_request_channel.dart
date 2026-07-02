// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

@JsonEnum()
enum AuthorizeRequestChannel {
  @JsonValue('pos')
  pos('pos'),
  @JsonValue('online')
  online('online'),
  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const AuthorizeRequestChannel(this.json);

  factory AuthorizeRequestChannel.fromJson(String json) => values.firstWhere(
        (e) => e.json == json,
        orElse: () => $unknown,
      );

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();
  /// Returns all defined enum values excluding the $unknown value.
  static List<AuthorizeRequestChannel> get $valuesDefined => values.where((value) => value != $unknown).toList();
}
