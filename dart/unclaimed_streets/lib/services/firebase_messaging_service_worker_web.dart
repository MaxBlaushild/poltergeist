// ignore_for_file: avoid_web_libraries_in_flutter, deprecated_member_use

import 'dart:html' as html;

import '../config/firebase_config.dart';

const _firebaseMessagingScope = 'firebase-cloud-messaging-push-scope';

Future<void> ensureRegistered() async {
  if (!FirebaseConfig.isConfigured) {
    return;
  }

  final container = html.window.navigator.serviceWorker;
  if (container == null) {
    return;
  }

  final scriptUrl = Uri(
    path: 'firebase-messaging-sw.js',
    queryParameters: {
      'apiKey': FirebaseConfig.apiKey,
      'appId': FirebaseConfig.appId,
      'messagingSenderId': FirebaseConfig.messagingSenderId,
      'projectId': FirebaseConfig.projectId,
      if (FirebaseConfig.authDomain.isNotEmpty)
        'authDomain': FirebaseConfig.authDomain,
      if (FirebaseConfig.storageBucket.isNotEmpty)
        'storageBucket': FirebaseConfig.storageBucket,
      if (FirebaseConfig.measurementId.isNotEmpty)
        'measurementId': FirebaseConfig.measurementId,
    },
  ).toString();

  await container.register(scriptUrl, <String, dynamic>{
    'scope': _firebaseMessagingScope,
  });
}
