import '../constants/api_constants.dart';
import '../models/user_character_profile.dart';
import 'api_client.dart';

class UserCharacterService {
  final ApiClient _api;

  UserCharacterService(this._api);

  Future<UserCharacterProfile?> getProfile(String userId) async {
    try {
      final data = await _api
          .get<Map<String, dynamic>>(ApiConstants.userCharacterEndpoint(userId));
      return UserCharacterProfile.fromJson(data);
    } catch (_) {
      return null;
    }
  }
}
