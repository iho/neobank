// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'deposit_wallet_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

DepositWalletRequest _$DepositWalletRequestFromJson(
  Map<String, dynamic> json,
) => DepositWalletRequest(
  amount: json['amount'] as String,
  currency: json['currency'] as String? ?? 'USD',
);

Map<String, dynamic> _$DepositWalletRequestToJson(
  DepositWalletRequest instance,
) => <String, dynamic>{
  'amount': instance.amount,
  'currency': instance.currency,
};
