import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:travel_angels/models/user_level.dart';
import 'package:travel_angels/services/user_level_service.dart';

class UserLevelProvider extends ChangeNotifier {
  final UserLevelService _userLevelService;
  UserLevel? _userLevel;
  bool _loading = false;
  String? _error;

  UserLevelProvider(this._userLevelService);

  UserLevel? get userLevel => _userLevel;
  bool get loading => _loading;
  String? get error => _error;

  /// Fetches the user level for the authenticated user
  Future<void> fetchUserLevel() async {
    _loading = true;
    _error = null;
    notifyListeners();

    try {
      _userLevel = await _userLevelService.getUserLevel();
    } catch (e) {
      _error = _extractErrorMessage(e);
      _userLevel = null;
    } finally {
      _loading = false;
      notifyListeners();
    }
  }

  /// Clears the user level data
  void clear() {
    _userLevel = null;
    _error = null;
    _loading = false;
    notifyListeners();
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
