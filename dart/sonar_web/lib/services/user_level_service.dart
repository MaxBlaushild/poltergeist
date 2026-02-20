import '../models/user_level.dart';
import 'api_client.dart';

class UserLevelService {
  final ApiClient _api;

  UserLevelService(this._api);

  Future<UserLevel?> getUserLevel() async {
    try {
      final data = await _api.get<Map<String, dynamic>>('/sonar/level');
      return UserLevel.fromJson(data);
    } catch (_) {
      return null;
    }
  }
}
