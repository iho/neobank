import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/dio_client.dart';
import '../../../core/network/error_mapper.dart';
import '../domain/wallet_models.dart';

final walletRepositoryProvider = Provider<WalletRepository>((ref) {
  return WalletRepository(ref.watch(dioProvider));
});

class WalletRepository {
  WalletRepository(this._dio);

  final Dio _dio;

  Future<WalletBalance> getBalance({String currency = 'USD'}) async {
    try {
      final response = await _dio.get<Map<String, dynamic>>(
        '/v1/wallet',
        queryParameters: {'currency': currency},
      );
      return WalletBalance.fromJson(response.data!);
    } on DioException catch (e) {
      throw mapDioException(e);
    }
  }

  Future<WalletTransactionPage> listTransactions({String? cursor, int limit = 20}) async {
    try {
      final response = await _dio.get<Map<String, dynamic>>(
        '/v1/wallet/transactions',
        queryParameters: {
          'limit': limit,
          'cursor': ?cursor,
        },
      );
      return WalletTransactionPage.fromJson(response.data!);
    } on DioException catch (e) {
      throw mapDioException(e);
    }
  }
}
