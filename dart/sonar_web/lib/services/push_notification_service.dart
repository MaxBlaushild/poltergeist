import 'dart:async';

import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../config/firebase_config.dart';
import '../constants/api_constants.dart';
import 'api_client.dart';
import 'notification_permission_service.dart';

class InAppPushEvent {
  InAppPushEvent({
    required this.type,
    required this.title,
    required this.body,
    required this.data,
    this.openedFromNotification = false,
  });

  final String type;
  final String title;
  final String body;
  final Map<String, String> data;
  final bool openedFromNotification;
}

class PartySubmissionResultEnvelope {
  PartySubmissionResultEnvelope({
    required this.id,
    required this.type,
    required this.data,
    required this.createdAt,
  });

  final String id;
  final String type;
  final Map<String, dynamic> data;
  final DateTime? createdAt;

  factory PartySubmissionResultEnvelope.fromJson(Map<String, dynamic> json) {
    return PartySubmissionResultEnvelope(
      id: json['id']?.toString() ?? '',
      type: json['type']?.toString() ?? '',
      data: json['data'] is Map
          ? Map<String, dynamic>.from(json['data'] as Map)
          : <String, dynamic>{},
      createdAt: DateTime.tryParse(json['createdAt']?.toString() ?? ''),
    );
  }
}

class PushNotificationService {
  PushNotificationService(this._apiClient);

  final ApiClient _apiClient;
  final NotificationPermissionService _permissionService =
      NotificationPermissionService();

  static const String _registeredTokenKey = 'push_registered_token';
  static const String _registeredUserIdKey = 'push_registered_user_id';

  bool _firebaseInitAttempted = false;
  bool _firebaseReady = false;
  bool _tokenRefreshListenerAttached = false;
  bool _foregroundListenerAttached = false;
  bool _openedListenerAttached = false;
  bool _initialMessageHandled = false;
  String? _activeUserId;
  StreamSubscription<String>? _tokenRefreshSubscription;
  StreamSubscription<RemoteMessage>? _foregroundMessageSubscription;
  StreamSubscription<RemoteMessage>? _openedMessageSubscription;
  final StreamController<InAppPushEvent> _inAppEventController =
      StreamController<InAppPushEvent>.broadcast();

  Stream<InAppPushEvent> get inAppEvents => _inAppEventController.stream;

  Future<bool> registerDeviceTokenForUser(
    String? userId, {
    bool force = false,
  }) async {
    final trimmedUserID = userId?.trim() ?? '';
    if (trimmedUserID.isEmpty) return false;
    _activeUserId = trimmedUserID;

    final permission = await _permissionService.getPermissionState();
    if (permission != NotificationPermissionState.granted) {
      return false;
    }

    final firebaseReady = await _ensureFirebaseReady();
    if (!firebaseReady) {
      return false;
    }

    final token = await _getFcmToken();
    if (token == null || token.trim().isEmpty) {
      return false;
    }

    final prefs = await SharedPreferences.getInstance();
    final previousToken = prefs.getString(_registeredTokenKey);
    final previousUserId = prefs.getString(_registeredUserIdKey);
    if (!force && previousToken == token && previousUserId == trimmedUserID) {
      _attachTokenRefreshListener();
      _attachForegroundMessageListener();
      return true;
    }

    await _registerTokenWithApi(
      token: token,
      userId: trimmedUserID,
      updateLocalCache: true,
    );
    _attachTokenRefreshListener();
    _attachForegroundMessageListener();
    return true;
  }

  Future<void> initializeForegroundMessageHandling() async {
    final permission = await _permissionService.getPermissionState();
    if (permission != NotificationPermissionState.granted) {
      return;
    }
    final firebaseReady = await _ensureFirebaseReady();
    if (!firebaseReady) {
      return;
    }
    _attachForegroundMessageListener();
    _attachOpenedMessageListener();
    await _emitInitialMessageIfPresent();
  }

  Future<bool> _ensureFirebaseReady() async {
    if (_firebaseReady) return true;
    if (_firebaseInitAttempted) {
      return _firebaseReady;
    }
    _firebaseInitAttempted = true;

    try {
      if (Firebase.apps.isEmpty) {
        if (kIsWeb) {
          final options = FirebaseConfig.options();
          if (options == null) {
            if (kDebugMode) {
              debugPrint(
                '[push] Firebase config missing; set FIREBASE_* --dart-define values.',
              );
            }
            return false;
          }
          await Firebase.initializeApp(options: options);
        } else {
          await Firebase.initializeApp();
        }
      }
      _firebaseReady = true;
      return true;
    } catch (err, stack) {
      if (kDebugMode) {
        debugPrint('[push] Firebase init failed: $err');
        debugPrintStack(stackTrace: stack);
      }
      _firebaseReady = Firebase.apps.isNotEmpty;
      return _firebaseReady;
    }
  }

  Future<String?> _getFcmToken() async {
    try {
      final vapidKey = FirebaseConfig.webVapidKey.trim();
      if (vapidKey.isNotEmpty) {
        return FirebaseMessaging.instance.getToken(vapidKey: vapidKey);
      }
      String? token = await FirebaseMessaging.instance.getToken();
      if (token != null && token.trim().isNotEmpty) {
        return token;
      }

      // APNs/FCM token bridging on iOS can be delayed right after permission.
      for (var i = 0; i < 4; i++) {
        await Future<void>.delayed(const Duration(milliseconds: 600));
        token = await FirebaseMessaging.instance.getToken();
        if (token != null && token.trim().isNotEmpty) {
          return token;
        }
      }
      return null;
    } catch (err, stack) {
      if (kDebugMode) {
        debugPrint('[push] Failed to get FCM token: $err');
        debugPrintStack(stackTrace: stack);
      }
      return null;
    }
  }

  Future<void> _registerTokenWithApi({
    required String token,
    required String userId,
    required bool updateLocalCache,
  }) async {
    await _apiClient.post<dynamic>(
      ApiConstants.deviceTokensEndpoint,
      data: {'token': token, 'platform': _platformLabel()},
    );

    if (!updateLocalCache) {
      return;
    }
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_registeredTokenKey, token);
    await prefs.setString(_registeredUserIdKey, userId);
  }

  Future<Map<String, dynamic>> sendTestPush({int delaySeconds = 0}) async {
    final response = await _apiClient.post<Map<String, dynamic>>(
      ApiConstants.pushTestEndpoint,
      data: {'delaySeconds': delaySeconds},
    );
    return response;
  }

  Future<List<PartySubmissionResultEnvelope>>
  getPendingPartySubmissionResults() async {
    final raw = await _apiClient.get<List<dynamic>>(
      ApiConstants.partySubmissionPendingResultsEndpoint,
    );
    final out = <PartySubmissionResultEnvelope>[];
    for (final entry in raw) {
      if (entry is Map<String, dynamic>) {
        out.add(PartySubmissionResultEnvelope.fromJson(entry));
      } else if (entry is Map) {
        out.add(
          PartySubmissionResultEnvelope.fromJson(
            Map<String, dynamic>.from(entry),
          ),
        );
      }
    }
    return out;
  }

  void _attachTokenRefreshListener() {
    if (_tokenRefreshListenerAttached) return;
    _tokenRefreshListenerAttached = true;
    _tokenRefreshSubscription = FirebaseMessaging.instance.onTokenRefresh
        .listen((token) async {
          final userId = _activeUserId;
          if (userId == null || token.trim().isEmpty) return;
          try {
            await _registerTokenWithApi(
              token: token,
              userId: userId,
              updateLocalCache: true,
            );
          } catch (err, stack) {
            if (kDebugMode) {
              debugPrint('[push] Failed to refresh token registration: $err');
              debugPrintStack(stackTrace: stack);
            }
          }
        });
  }

  void _attachForegroundMessageListener() {
    if (_foregroundListenerAttached) return;
    _foregroundListenerAttached = true;
    _foregroundMessageSubscription = FirebaseMessaging.onMessage.listen(
      (message) => _handleMessage(message, openedFromNotification: false),
    );
  }

  void _attachOpenedMessageListener() {
    if (_openedListenerAttached) return;
    _openedListenerAttached = true;
    _openedMessageSubscription = FirebaseMessaging.onMessageOpenedApp.listen(
      (message) => _handleMessage(message, openedFromNotification: true),
    );
  }

  Future<void> _emitInitialMessageIfPresent() async {
    if (_initialMessageHandled) {
      return;
    }
    _initialMessageHandled = true;
    try {
      final initialMessage = await FirebaseMessaging.instance
          .getInitialMessage();
      if (initialMessage == null) return;
      _handleMessage(initialMessage, openedFromNotification: true);
    } catch (_) {}
  }

  void _handleMessage(
    RemoteMessage message, {
    required bool openedFromNotification,
  }) {
    final data = message.data.map(
      (key, value) => MapEntry(key, value.toString()),
    );
    final type = (data['type'] ?? '').trim();
    final title = (message.notification?.title ?? '').trim();
    final body = (message.notification?.body ?? '').trim();

    _inAppEventController.add(
      InAppPushEvent(
        type: type,
        title: title,
        body: body,
        data: data,
        openedFromNotification: openedFromNotification,
      ),
    );
  }

  String _platformLabel() {
    if (kIsWeb) return 'web';
    switch (defaultTargetPlatform) {
      case TargetPlatform.iOS:
      case TargetPlatform.macOS:
        return 'ios';
      case TargetPlatform.android:
        return 'android';
      case TargetPlatform.linux:
      case TargetPlatform.windows:
      case TargetPlatform.fuchsia:
        return 'web';
    }
  }

  void dispose() {
    _tokenRefreshSubscription?.cancel();
    _tokenRefreshSubscription = null;
    _tokenRefreshListenerAttached = false;
    _foregroundMessageSubscription?.cancel();
    _foregroundMessageSubscription = null;
    _foregroundListenerAttached = false;
    _openedMessageSubscription?.cancel();
    _openedMessageSubscription = null;
    _openedListenerAttached = false;
    _inAppEventController.close();
  }
}
