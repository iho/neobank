// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

import 'wallet_balance.dart';

part 'wallet_balance_list.g.dart';

@JsonSerializable()
class WalletBalanceList {
  const WalletBalanceList({
    required this.wallets,
  });
  
  factory WalletBalanceList.fromJson(Map<String, Object?> json) => _$WalletBalanceListFromJson(json);
  
  final List<WalletBalance> wallets;

  Map<String, Object?> toJson() => _$WalletBalanceListToJson(this);
}
