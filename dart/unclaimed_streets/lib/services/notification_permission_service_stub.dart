import 'notification_permission_service.dart';

Future<NotificationPermissionState> getPermissionState() async {
  return NotificationPermissionState.unsupported;
}

Future<NotificationPermissionState> requestPermission() async {
  return NotificationPermissionState.unsupported;
}
