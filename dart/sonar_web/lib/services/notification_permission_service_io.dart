import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';

import 'notification_permission_service.dart';

NotificationPermissionState _fromAuthorizationStatus(
  AuthorizationStatus status,
) {
  switch (status) {
    case AuthorizationStatus.authorized:
    case AuthorizationStatus.provisional:
      return NotificationPermissionState.granted;
    case AuthorizationStatus.denied:
      return NotificationPermissionState.denied;
    case AuthorizationStatus.notDetermined:
      return NotificationPermissionState.notDetermined;
  }
}

Future<NotificationPermissionState> getPermissionState() async {
  if (Firebase.apps.isEmpty) {
    await Firebase.initializeApp();
  }
  final settings = await FirebaseMessaging.instance.getNotificationSettings();
  return _fromAuthorizationStatus(settings.authorizationStatus);
}

Future<NotificationPermissionState> requestPermission() async {
  if (Firebase.apps.isEmpty) {
    await Firebase.initializeApp();
  }
  final settings = await FirebaseMessaging.instance.requestPermission(
    alert: true,
    badge: true,
    sound: true,
    provisional: false,
  );
  return _fromAuthorizationStatus(settings.authorizationStatus);
}
