// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'card_list.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

CardList _$CardListFromJson(Map<String, dynamic> json) => CardList(
  cards: (json['cards'] as List<dynamic>)
      .map((e) => Card.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$CardListToJson(CardList instance) => <String, dynamic>{
  'cards': instance.cards,
};
