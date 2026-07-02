// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'provision_wallet_response.g.dart';

@JsonSerializable()
class ProvisionWalletResponse {
  const ProvisionWalletResponse({
    required this.walletId,
    required this.ledgerAccountId,
  });
  
  factory ProvisionWalletResponse.fromJson(Map<String, Object?> json) => _$ProvisionWalletResponseFromJson(json);
  
  @JsonKey(name: 'wallet_id')
  final String walletId;
  @JsonKey(name: 'ledger_account_id')
  final String ledgerAccountId;

  Map<String, Object?> toJson() => _$ProvisionWalletResponseToJson(this);
}
