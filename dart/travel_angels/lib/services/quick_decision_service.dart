import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/quick_decision_request.dart';
import 'package:travel_angels/services/api_client.dart';

class QuickDecisionService {
  final APIClient _apiClient;

  QuickDecisionService(this._apiClient);

  /// Creates a quick decision request
  /// 
  /// [question] - The user's question
  /// [option1] - First option (required)
  /// [option2] - Second option (required)
  /// [option3] - Third option (optional)
  /// 
  /// Returns the created quick decision request
  Future<QuickDecisionRequest> createQuickDecisionRequest({
    required String question,
    required String option1,
    required String option2,
    String? option3,
  }) async {
    try {
      final data = <String, dynamic>{
        'question': question,
        'option1': option1,
        'option2': option2,
      };

      if (option3 != null && option3.isNotEmpty) {
        data['option3'] = option3;
      }

      final response = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.quickDecisionRequestsEndpoint,
        data: data,
      );

      return QuickDecisionRequest.fromJson(response);
    } catch (e) {
      rethrow;
    }
  }
}
