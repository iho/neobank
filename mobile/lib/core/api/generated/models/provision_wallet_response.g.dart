// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'provision_wallet_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ProvisionWalletResponse _$ProvisionWalletResponseFromJson(
  Map<String, dynamic> json,
) => ProvisionWalletResponse(
  walletId: json['wallet_id'] as String,
  ledgerAccountId: json['ledger_account_id'] as String,
);

Map<String, dynamic> _$ProvisionWalletResponseToJson(
  ProvisionWalletResponse instance,
) => <String, dynamic>{
  'wallet_id': instance.walletId,
  'ledger_account_id': instance.ledgerAccountId,
};
