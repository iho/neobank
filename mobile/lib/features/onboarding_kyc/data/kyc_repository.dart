import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/dio_client.dart';
import '../../../core/network/error_mapper.dart';
import '../domain/kyc_models.dart';

final kycRepositoryProvider = Provider<KycRepository>((ref) {
  return KycRepository(ref.watch(dioProvider));
});

class KycRepository {
  KycRepository(this._dio);

  final Dio _dio;

  Future<KycStatusInfo> getStatus() async {
    try {
      final response = await _dio.get<Map<String, dynamic>>('/v1/kyc/status');
      return KycStatusInfo.fromJson(response.data!);
    } on DioException catch (e) {
      throw mapDioException(e);
    }
  }

  Future<KycSubmitResult> submit({
    required String fullName,
    required String dateOfBirth,
    required String countryCode,
    String? documentType,
    String? documentNumber,
  }) async {
    try {
      final response = await _dio.post<Map<String, dynamic>>(
        '/v1/kyc',
        data: {
          'full_name': fullName,
          'date_of_birth': dateOfBirth,
          'country_code': countryCode,
          if (documentType != null && documentType.isNotEmpty) 'document_type': documentType,
          if (documentNumber != null && documentNumber.isNotEmpty)
            'document_number': documentNumber,
        },
      );
      return KycSubmitResult.fromJson(response.data!);
    } on DioException catch (e) {
      throw mapDioException(e);
    }
  }
}
