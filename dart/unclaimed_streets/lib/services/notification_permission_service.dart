import 'notification_permission_service_impl.dart' as impl;

enum NotificationPermissionState { unsupported, notDetermined, granted, denied }

class NotificationPermissionService {
  Future<NotificationPermissionState> getPermissionState() {
    return impl.getPermissionState();
  }

  Future<NotificationPermissionState> requestPermission() {
    return impl.requestPermission();
  }
}
