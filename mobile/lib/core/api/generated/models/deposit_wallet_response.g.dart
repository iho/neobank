// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'deposit_wallet_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

DepositWalletResponse _$DepositWalletResponseFromJson(
  Map<String, dynamic> json,
) => DepositWalletResponse(
  id: json['id'] as String,
  walletId: json['wallet_id'] as String,
  amount: json['amount'] as String,
  currency: json['currency'] as String,
  status: json['status'] as String,
  ledgerTransferId: json['ledger_transfer_id'] as String?,
  createdAt: json['created_at'] == null
      ? null
      : DateTime.parse(json['created_at'] as String),
);

Map<String, dynamic> _$DepositWalletResponseToJson(
  DepositWalletResponse instance,
) => <String, dynamic>{
  'id': instance.id,
  'wallet_id': instance.walletId,
  'amount': instance.amount,
  'currency': instance.currency,
  'ledger_transfer_id': instance.ledgerTransferId,
  'status': instance.status,
  'created_at': instance.createdAt?.toIso8601String(),
};
