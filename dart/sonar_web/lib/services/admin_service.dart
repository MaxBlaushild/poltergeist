import 'api_client.dart';
import '../models/zone_seed_job.dart';

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

  Future<ZoneSeedJob> seedZoneDraft({
    required String zoneId,
    int? placeCount,
    int? characterCount,
    int? questCount,
  }) async {
    final payload = <String, dynamic>{};
    if (placeCount != null) payload['placeCount'] = placeCount;
    if (characterCount != null) payload['characterCount'] = characterCount;
    if (questCount != null) payload['questCount'] = questCount;

    final data = await _api.post<Map<String, dynamic>>(
      '/sonar/admin/zones/$zoneId/seed-draft',
      data: payload,
    );
    return ZoneSeedJob.fromJson(data);
  }

  Future<List<ZoneSeedJob>> getZoneSeedJobs({String? zoneId, int? limit}) async {
    final params = <String, dynamic>{};
    if (zoneId != null && zoneId.isNotEmpty) params['zoneId'] = zoneId;
    if (limit != null) params['limit'] = limit.toString();
    final list = await _api.get<List<dynamic>>(
      '/sonar/admin/zone-seed-jobs',
      params: params.isEmpty ? null : params,
    );
    return list
        .whereType<Map<String, dynamic>>()
        .map(ZoneSeedJob.fromJson)
        .toList();
  }

  Future<ZoneSeedJob> getZoneSeedJob(String jobId) async {
    final data = await _api.get<Map<String, dynamic>>(
      '/sonar/admin/zone-seed-jobs/$jobId',
    );
    return ZoneSeedJob.fromJson(data);
  }

  Future<ZoneSeedJob> approveZoneSeedJob(String jobId) async {
    final data = await _api.post<Map<String, dynamic>>(
      '/sonar/admin/zone-seed-jobs/$jobId/approve',
    );
    return ZoneSeedJob.fromJson(data);
  }
}
