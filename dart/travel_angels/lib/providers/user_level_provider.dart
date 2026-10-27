import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:travel_angels/models/user_level.dart';
import 'package:travel_angels/services/user_level_service.dart';

class UserLevelProvider extends ChangeNotifier {
  final UserLevelService _userLevelService;
  UserLevel? _userLevel;
  bool _loading = false;
  String? _error;
  int _requestRevision = 0;
  bool _disposed = false;

  UserLevelProvider(this._userLevelService);

  UserLevel? get userLevel => _userLevel;
  bool get loading => _loading;
  String? get error => _error;

  /// Fetches the user level for the authenticated user
  Future<void> fetchUserLevel() async {
    if (_disposed) return;
    final request = ++_requestRevision;
    bool isCurrent() => !_disposed && request == _requestRevision;
    _loading = true;
    _error = null;
    notifyListeners();

    try {
      final level = await _userLevelService.getUserLevel();
      if (!isCurrent()) return;
      _userLevel = level;
    } catch (e) {
      if (!isCurrent()) return;
      _error = _extractErrorMessage(e);
      _userLevel = null;
    } finally {
      if (isCurrent()) {
        _loading = false;
        notifyListeners();
      }
    }
  }

  /// Clears the user level data
  void clear() {
    ++_requestRevision;
    _userLevel = null;
    _error = null;
    _loading = false;
    notifyListeners();
  }

  @override
  void dispose() {
    _disposed = true;
    ++_requestRevision;
    super.dispose();
  }

  /// Extracts error message from DioException response body
  String _extractErrorMessage(dynamic error) {
    if (error is DioException) {
      // Try to get error message from response body
      if (error.response?.data != null) {
        final data = error.response!.data;
        if (data is Map<String, dynamic> && data.containsKey('error')) {
          return data['error'] as String;
        }
        if (data is String) {
          return data;
        }
      }
      // Fall back to DioException message
      return error.message ?? 'Failed to fetch user level';
    }
    // Fall back for non-Dio errors
    return error.toString();
  }
}
