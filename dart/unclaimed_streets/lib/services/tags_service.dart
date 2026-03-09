import '../models/tag.dart';
import 'api_client.dart';

class TagsService {
  final ApiClient _api;

  TagsService(this._api);

  Future<List<Tag>> getTags() async {
    try {
      final list = await _api.get<List<dynamic>>('/sonar/tags');
      return list.map((e) => Tag.fromJson(e as Map<String, dynamic>)).toList();
    } catch (_) {
      return [];
    }
  }

  Future<List<TagGroup>> getTagGroups() async {
    try {
      final list = await _api.get<List<dynamic>>('/sonar/tagGroups');
      return list
          .map((e) => TagGroup.fromJson(e as Map<String, dynamic>))
          .toList();
    } catch (_) {
      return [];
    }
  }
}
