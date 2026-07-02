import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/dio_client.dart';
import '../../../core/network/error_mapper.dart';
import '../domain/card_models.dart';

final cardRepositoryProvider = Provider<CardRepository>((ref) {
  return CardRepository(ref.watch(dioProvider));
});

class CardRepository {
  CardRepository(this._dio);

  final Dio _dio;

  Future<List<BankCard>> listCards() async {
    try {
      final response = await _dio.get<Map<String, dynamic>>('/v1/cards');
      return (response.data!['cards'] as List<dynamic>)
          .map((e) => BankCard.fromJson(e as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      throw mapDioException(e);
    }
  }

  Future<BankCard> issueCard({
    required String cardholderName,
    String? walletId,
    String? dailyLimit,
    bool? onlineOnly,
  }) async {
    try {
      final response = await _dio.post<Map<String, dynamic>>(
        '/v1/cards',
        data: {
          'cardholder_name': cardholderName,
          'wallet_id': ?walletId,
          if (dailyLimit != null && dailyLimit.isNotEmpty) 'daily_limit': dailyLimit,
          'online_only': ?onlineOnly,
        },
      );
      return BankCard.fromJson(response.data!);
    } on DioException catch (e) {
      throw mapDioException(e);
    }
  }

  Future<BankCard> getCard(String id) async {
    try {
      final response = await _dio.get<Map<String, dynamic>>('/v1/cards/$id');
      return BankCard.fromJson(response.data!);
    } on DioException catch (e) {
      throw mapDioException(e);
    }
  }

  Future<BankCard> freeze(String id) async {
    try {
      final response = await _dio.post<Map<String, dynamic>>('/v1/cards/$id/freeze');
      return BankCard.fromJson(response.data!);
    } on DioException catch (e) {
      throw mapDioException(e);
    }
  }

  Future<BankCard> unfreeze(String id) async {
    try {
      final response = await _dio.post<Map<String, dynamic>>('/v1/cards/$id/unfreeze');
      return BankCard.fromJson(response.data!);
    } on DioException catch (e) {
      throw mapDioException(e);
    }
  }

  Future<List<CardAuthorization>> listAuthorizations() async {
    try {
      final response = await _dio.get<Map<String, dynamic>>('/v1/authorizations');
      return (response.data!['authorizations'] as List<dynamic>)
          .map((e) => CardAuthorization.fromJson(e as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      throw mapDioException(e);
    }
  }

  Future<CardAuthorization> getAuthorization(String id) async {
    try {
      final response = await _dio.get<Map<String, dynamic>>('/v1/authorizations/$id');
      return CardAuthorization.fromJson(response.data!);
    } on DioException catch (e) {
      throw mapDioException(e);
    }
  }

  Future<CardAuthorization> captureAuthorization(String id) async {
    try {
      final response = await _dio.post<Map<String, dynamic>>('/v1/authorizations/$id/capture');
      return CardAuthorization.fromJson(response.data!);
    } on DioException catch (e) {
      throw mapDioException(e);
    }
  }
}
