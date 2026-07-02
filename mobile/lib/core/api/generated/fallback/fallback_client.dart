// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:dio/dio.dart';
import 'package:retrofit/retrofit.dart';

import '../models/authorization_list.dart';
import '../models/authorize_request.dart';
import '../models/card.dart';
import '../models/card_authorization.dart';
import '../models/card_list.dart';
import '../models/change_password_request.dart';
import '../models/create_payee_request.dart';
import '../models/create_transfer_request.dart';
import '../models/deposit_wallet_request.dart';
import '../models/deposit_wallet_response.dart';
import '../models/device_token.dart';
import '../models/health_response.dart';
import '../models/issue_card_request.dart';
import '../models/kyc_status_response.dart';
import '../models/limits_response.dart';
import '../models/login_request.dart';
import '../models/login_response.dart';
import '../models/mark_notifications_read_response.dart';
import '../models/notification.dart';
import '../models/notification_list.dart';
import '../models/notification_preferences.dart';
import '../models/payee.dart';
import '../models/payee_list.dart';
import '../models/profile.dart';
import '../models/provision_wallet_request.dart';
import '../models/provision_wallet_response.dart';
import '../models/referral_invite.dart';
import '../models/referral_invite_list.dart';
import '../models/refresh_token_request.dart';
import '../models/register_device_token_request.dart';
import '../models/register_request.dart';
import '../models/register_response.dart';
import '../models/submit_kyc_request.dart';
import '../models/submit_kyc_response.dart';
import '../models/transfer.dart';
import '../models/transfer_list.dart';
import '../models/update_card_controls_request.dart';
import '../models/update_notification_preferences_request.dart';
import '../models/wallet_balance.dart';
import '../models/wallet_balance_list.dart';
import '../models/wallet_transaction_list.dart';

part 'fallback_client.g.dart';

@RestApi()
abstract class FallbackClient {
  factory FallbackClient(Dio dio, {String? baseUrl}) = _FallbackClient;

  @GET('/health')
  Future<HealthResponse> getHealth();

  @POST('/v1/auth/login')
  Future<LoginResponse> login({
    @Body() required LoginRequest body,
  });

  @POST('/v1/auth/refresh')
  Future<LoginResponse> refreshToken({
    @Body() required RefreshTokenRequest body,
  });

  @POST('/v1/auth/register')
  Future<RegisterResponse> register({
    @Header('Idempotency-Key') required String idempotencyKey,
    @Body() required RegisterRequest body,
  });

  @POST('/v1/auth/change-password')
  Future<void> changePassword({
    @Body() required ChangePasswordRequest body,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @GET('/v1/me')
  Future<Profile> getProfile({
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @POST('/v1/kyc')
  Future<SubmitKycResponse> submitKyc({
    @Header('Idempotency-Key') required String idempotencyKey,
    @Body() required SubmitKycRequest body,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @GET('/v1/kyc/status')
  Future<KycStatusResponse> getKycStatus({
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @GET('/v1/wallet')
  Future<WalletBalance> getWalletBalance({
    @Query('currency') String? currency = 'USD',
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @GET('/v1/payees')
  Future<PayeeList> listPayees({
    @Query('limit') int? limit = 20,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @POST('/v1/payees')
  Future<Payee> createPayee({
    @Body() required CreatePayeeRequest body,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @DELETE('/v1/payees/{id}')
  Future<void> deletePayee({
    @Path('id') required String id,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @GET('/v1/wallet/transactions/export')
  Future<String> exportWalletTransactions({
    @Query('from') required DateTime from,
    @Query('to') required DateTime to,
    @Query('format') String? format = 'csv',
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @GET('/v1/wallet/transactions')
  Future<WalletTransactionList> listWalletTransactions({
    @Query('limit') int? limit = 20,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
    @Query('cursor') String? cursor,
  });

  @POST('/v1/wallet/deposit')
  Future<DepositWalletResponse> depositWallet({
    @Header('Idempotency-Key') required String idempotencyKey,
    @Body() required DepositWalletRequest body,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @GET('/v1/invites')
  Future<ReferralInviteList> listReferralInvites({
    @Query('limit') int? limit = 20,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @POST('/v1/invites')
  Future<ReferralInvite> createReferralInvite({
    @Header('Idempotency-Key') required String idempotencyKey,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @GET('/v1/wallets')
  Future<WalletBalanceList> listWallets({
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @POST('/v1/wallets')
  Future<ProvisionWalletResponse> provisionWallet({
    @Header('Idempotency-Key') required String idempotencyKey,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
    @Body() ProvisionWalletRequest? body,
  });

  @GET('/v1/limits')
  Future<LimitsResponse> getLimits({
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @GET('/v1/transfers')
  Future<TransferList> listTransfers({
    @Query('limit') int? limit = 20,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
    @Query('cursor') String? cursor,
  });

  @POST('/v1/transfers')
  Future<Transfer> createTransfer({
    @Header('Idempotency-Key') required String idempotencyKey,
    @Body() required CreateTransferRequest body,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @GET('/v1/transfers/{id}')
  Future<Transfer> getTransfer({
    @Path('id') required String id,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @GET('/v1/cards')
  Future<CardList> listCards({
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @POST('/v1/cards')
  Future<Card> issueCard({
    @Header('Idempotency-Key') required String idempotencyKey,
    @Body() required IssueCardRequest body,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @GET('/v1/cards/{id}')
  Future<Card> getCard({
    @Path('id') required String id,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @PATCH('/v1/cards/{id}/controls')
  Future<Card> updateCardControls({
    @Path('id') required String id,
    @Body() required UpdateCardControlsRequest body,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @POST('/v1/cards/{id}/freeze')
  Future<Card> freezeCard({
    @Path('id') required String id,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @POST('/v1/cards/{id}/unfreeze')
  Future<Card> unfreezeCard({
    @Path('id') required String id,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @POST('/v1/cards/{id}/authorize')
  Future<CardAuthorization> authorizeTransaction({
    @Header('Idempotency-Key') required String idempotencyKey,
    @Path('id') required String id,
    @Body() required AuthorizeRequest body,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @GET('/v1/authorizations')
  Future<AuthorizationList> listAuthorizations({
    @Query('limit') int? limit = 20,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @GET('/v1/authorizations/{id}')
  Future<CardAuthorization> getAuthorization({
    @Path('id') required String id,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @POST('/v1/authorizations/{id}/capture')
  Future<CardAuthorization> captureAuthorization({
    @Path('id') required String id,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @POST('/v1/notifications/read-all')
  Future<MarkNotificationsReadResponse> markAllNotificationsRead({
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @POST('/v1/notifications/{id}/read')
  Future<Notification> markNotificationRead({
    @Path('id') required String id,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @POST('/v1/devices')
  Future<DeviceToken> registerDeviceToken({
    @Header('Idempotency-Key') required String idempotencyKey,
    @Body() required RegisterDeviceTokenRequest body,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @DELETE('/v1/devices/{id}')
  Future<void> deleteDeviceToken({
    @Header('Idempotency-Key') required String idempotencyKey,
    @Path('id') required String id,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @POST('/v1/account/close')
  Future<void> closeAccount({
    @Header('Idempotency-Key') required String idempotencyKey,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @GET('/v1/notification-preferences')
  Future<NotificationPreferences> getNotificationPreferences({
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @PATCH('/v1/notification-preferences')
  Future<NotificationPreferences> updateNotificationPreferences({
    @Body() required UpdateNotificationPreferencesRequest body,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
  });

  @GET('/v1/notifications')
  Future<NotificationList> listNotifications({
    @Query('limit') int? limit = 20,
    @Header('Authorization') String? authorization,
    @Header('X-User-Id') String? xUserId,
    @Query('cursor') String? cursor,
  });
}
