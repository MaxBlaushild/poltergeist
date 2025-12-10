import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/community_poll.dart';
import 'package:travel_angels/services/api_client.dart';

class CommunityPollService {
  final APIClient _apiClient;

  CommunityPollService(this._apiClient);

  /// Creates a community poll
  /// 
  /// [question] - The poll question
  /// [options] - List of options (3-10 items required)
  /// 
  /// Returns the created community poll
  Future<CommunityPoll> createCommunityPoll({
    required String question,
    required List<String> options,
  }) async {
    try {
      final data = <String, dynamic>{
        'question': question,
        'options': options,
      };

      final response = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.communityPollsEndpoint,
        data: data,
      );

      return CommunityPoll.fromJson(response);
    } catch (e) {
      rethrow;
    }
  }
}
