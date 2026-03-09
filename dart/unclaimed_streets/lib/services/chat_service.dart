import '../models/audit_item.dart';
import 'api_client.dart';

class ChatService {
  final ApiClient _api;

  ChatService(this._api);

  Future<List<AuditItem>> getChat() async {
    try {
      final list = await _api.get<List<dynamic>>('/sonar/chat');
      return list
          .map((e) => AuditItem.fromJson(e as Map<String, dynamic>))
          .toList();
    } catch (_) {
      return [];
    }
  }
}
