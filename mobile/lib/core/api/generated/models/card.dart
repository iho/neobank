// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'card.g.dart';

@JsonSerializable()
class Card {
  const Card({
    required this.id,
    required this.userId,
    required this.walletId,
    required this.lastFour,
    required this.status,
    required this.expiryMonth,
    required this.expiryYear,
    required this.onlineOnly,
    this.dailyLimit,
  });
  
  factory Card.fromJson(Map<String, Object?> json) => _$CardFromJson(json);
  
  final String id;
  @JsonKey(name: 'user_id')
  final String userId;
  @JsonKey(name: 'wallet_id')
  final String walletId;
  @JsonKey(name: 'last_four')
  final String lastFour;
  final String status;
  @JsonKey(name: 'expiry_month')
  final int expiryMonth;
  @JsonKey(name: 'expiry_year')
  final int expiryYear;
  @JsonKey(name: 'daily_limit')
  final String? dailyLimit;
  @JsonKey(name: 'online_only')
  final bool onlineOnly;

  Map<String, Object?> toJson() => _$CardToJson(this);
}
