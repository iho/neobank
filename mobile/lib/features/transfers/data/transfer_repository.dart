import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/dio_client.dart';
import '../../../core/network/error_mapper.dart';
import '../domain/transfer_models.dart';

final transferRepositoryProvider = Provider<TransferRepository>((ref) {
  return TransferRepository(ref.watch(dioProvider));
});

class TransferRepository {
  TransferRepository(this._dio);

  final Dio _dio;

  /// [idempotencyKey] must be supplied by the caller (and reused verbatim on
  /// retry) rather than left to the interceptor's default — a client-side
  /// timeout must not spawn a second transfer for the same user intent.
  Future<Transfer> createTransfer({
    required String idempotencyKey,
    required String amount,
    String? recipientPhone,
    String? recipientEmail,
    String? recipientUserId,
    String? currency,
    String? memo,
  }) async {
    try {
      final response = await _dio.post<Map<String, dynamic>>(
        '/v1/transfers',
        options: Options(headers: {'Idempotency-Key': idempotencyKey}),
        data: {
          'amount': amount,
          if (recipientPhone != null && recipientPhone.isNotEmpty)
            'recipient_phone': recipientPhone,
          if (recipientEmail != null && recipientEmail.isNotEmpty)
            'recipient_email': recipientEmail,
          if (recipientUserId != null && recipientUserId.isNotEmpty)
            'recipient_user_id': recipientUserId,
          if (currency != null && currency.isNotEmpty) 'currency': currency,
          if (memo != null && memo.isNotEmpty) 'memo': memo,
        },
      );
      return Transfer.fromJson(response.data!);
    } on DioException catch (e) {
      // 422 carries a Transfer body describing the decline (e.g.
      // fraud-blocked) — that's a domain result, not a transport failure.
      final data = e.response?.data;
      if (e.response?.statusCode == 422 && data is Map<String, dynamic>) {
        return Transfer.fromJson(data);
      }
      throw mapDioException(e);
    }
  }

  Future<TransferPage> listTransfers({String? cursor, int limit = 20}) async {
    try {
      final response = await _dio.get<Map<String, dynamic>>(
        '/v1/transfers',
        queryParameters: {
          'limit': limit,
          'cursor': ?cursor,
        },
      );
      return TransferPage.fromJson(response.data!);
    } on DioException catch (e) {
      throw mapDioException(e);
    }
  }

  Future<Transfer> getTransfer(String id) async {
    try {
      final response = await _dio.get<Map<String, dynamic>>('/v1/transfers/$id');
      return Transfer.fromJson(response.data!);
    } on DioException catch (e) {
      throw mapDioException(e);
    }
  }
}
