import 'api_client.dart';

class AdminService {
  final ApiClient _api;

  AdminService(this._api);

  Future<void> unlockPointOfInterestForTeam({
    required String teamId,
    required String pointOfInterestId,
  }) async {
    await _api.post<dynamic>(
      '/sonar/admin/pointOfInterest/unlock',
      data: {
        'teamId': teamId,
        'pointOfInterestId': pointOfInterestId,
      },
    );
  }

  /// Capture for team. Backend may not implement this endpoint.
  Future<void> capturePointOfInterestForTeam({
    required String teamId,
    required String pointOfInterestId,
    required int tier,
  }) async {
    await _api.post<dynamic>(
      '/sonar/admin/pointOfInterest/capture',
      data: {
        'teamId': teamId,
        'pointOfInterestId': pointOfInterestId,
        'tier': tier,
      },
    );
  }
}
