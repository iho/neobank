// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'deposit_wallet_request.g.dart';

@JsonSerializable()
class DepositWalletRequest {
  const DepositWalletRequest({
    required this.amount,
    this.currency = 'USD',
  });
  
  factory DepositWalletRequest.fromJson(Map<String, Object?> json) => _$DepositWalletRequestFromJson(json);
  
  final String amount;
  final String currency;

  Map<String, Object?> toJson() => _$DepositWalletRequestToJson(this);
}
