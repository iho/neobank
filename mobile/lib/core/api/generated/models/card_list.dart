// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

import 'card.dart';

part 'card_list.g.dart';

@JsonSerializable()
class CardList {
  const CardList({
    required this.cards,
  });
  
  factory CardList.fromJson(Map<String, Object?> json) => _$CardListFromJson(json);
  
  final List<Card> cards;

  Map<String, Object?> toJson() => _$CardListToJson(this);
}
