// ignore_for_file: avoid_web_libraries_in_flutter

import 'dart:html' as html;

import 'notification_permission_service.dart';

NotificationPermissionState _fromWebPermission(String? permission) {
  switch (permission) {
    case 'granted':
      return NotificationPermissionState.granted;
    case 'denied':
      return NotificationPermissionState.denied;
    case 'default':
      return NotificationPermissionState.notDetermined;
    default:
      return NotificationPermissionState.notDetermined;
  }
}

Future<NotificationPermissionState> getPermissionState() async {
  if (!html.Notification.supported) {
    return NotificationPermissionState.unsupported;
  }
  return _fromWebPermission(html.Notification.permission);
}

Future<NotificationPermissionState> requestPermission() async {
  if (!html.Notification.supported) {
    return NotificationPermissionState.unsupported;
  }
  final permission = await html.Notification.requestPermission();
  return _fromWebPermission(permission);
}
