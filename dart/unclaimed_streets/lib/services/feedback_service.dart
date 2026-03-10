import 'api_client.dart';

class FeedbackService {
  FeedbackService(this._api);

  final ApiClient _api;

  Future<void> submit({
    required String message,
    required String route,
    String? zoneId,
  }) async {
    await _api.post<dynamic>(
      '/sonar/feedback',
      data: {
        'message': message,
        'route': route,
        'zoneId': zoneId,
      },
    );
  }
}
