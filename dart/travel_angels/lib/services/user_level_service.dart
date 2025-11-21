import 'package:travel_angels/models/user_level.dart';
import 'package:travel_angels/services/api_client.dart';

class UserLevelService {
  final APIClient _apiClient;

  UserLevelService(this._apiClient);

  /// Gets the user level for the authenticated user
  /// Returns the UserLevel object
  Future<UserLevel> getUserLevel() async {
    try {
      final response = await _apiClient.get<Map<String, dynamic>>(
        '/travel-angels/level',
      );

      return UserLevel.fromJson(response);
    } catch (e) {
      rethrow;
    }
  }
}
