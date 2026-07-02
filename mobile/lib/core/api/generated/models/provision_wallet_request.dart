// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:json_annotation/json_annotation.dart';

part 'provision_wallet_request.g.dart';

@JsonSerializable()
class ProvisionWalletRequest {
  const ProvisionWalletRequest({
    this.currency,
  });
  
  factory ProvisionWalletRequest.fromJson(Map<String, Object?> json) => _$ProvisionWalletRequestFromJson(json);
  
  final String? currency;

  Map<String, Object?> toJson() => _$ProvisionWalletRequestToJson(this);
}
