import 'package:flutter/material.dart';
import 'package:skunkworks/models/notification.dart';
import 'package:skunkworks/services/notification_service.dart';

class NotificationProvider extends ChangeNotifier {
  final NotificationService _notificationService;
  List<Notification> _notifications = [];
  int _unreadCount = 0;
  bool _loading = false;
  String? _error;

  NotificationProvider(this._notificationService);

  List<Notification> get notifications => _notifications;
  int get unreadCount => _unreadCount;
  bool get loading => _loading;
  String? get error => _error;

  Future<void> loadNotifications() async {
    _loading = true;
    _error = null;
    notifyListeners();

    try {
      final result = await _notificationService.getNotifications();
      _notifications = List<Notification>.from(result['notifications'] as List);
      _unreadCount = result['unreadCount'] as int? ?? 0;
    } catch (e) {
      _error = e.toString();
    } finally {
      _loading = false;
      notifyListeners();
    }
  }

  Future<void> markAsRead(String notificationId) async {
    try {
      await _notificationService.markAsRead(notificationId);
      final idx = _notifications.indexWhere((n) => n.id == notificationId);
      if (idx >= 0 && _unreadCount > 0) {
        _unreadCount--;
        notifyListeners();
      }
      await loadNotifications();
    } catch (_) {}
  }

  Future<void> markAllAsRead() async {
    try {
      await _notificationService.markAllAsRead();
      _unreadCount = 0;
      await loadNotifications();
    } catch (_) {}
  }

  void setUnreadCount(int count) {
    _unreadCount = count;
    notifyListeners();
  }
}
