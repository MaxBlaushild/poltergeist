import 'package:skunkworks/constants/api_constants.dart';
import 'package:skunkworks/models/notification.dart';
import 'package:skunkworks/services/api_client.dart';

class NotificationService {
  final APIClient _api;

  NotificationService(this._api);

  Future<Map<String, dynamic>> getNotifications({int limit = 50, int offset = 0}) async {
    final response = await _api.get<Map<String, dynamic>>(
      ApiConstants.notificationsEndpoint,
      params: {'limit': limit, 'offset': offset},
    );
    final list = response['notifications'] as List<dynamic>? ?? [];
    final notifications = list.map((e) => Notification.fromJson(e as Map<String, dynamic>)).toList();
    final unreadCount = response['unreadCount'] as int? ?? 0;
    return {'notifications': notifications, 'unreadCount': unreadCount};
  }

  Future<void> markAsRead(String notificationId) async {
    await _api.patch(ApiConstants.notificationReadEndpoint(notificationId));
  }

  Future<void> markAllAsRead() async {
    await _api.patch(ApiConstants.notificationsReadAllEndpoint);
  }

  Future<void> registerDeviceToken(String token, String platform) async {
    await _api.post(
      ApiConstants.deviceTokensEndpoint,
      data: {'token': token, 'platform': platform},
    );
  }
}
