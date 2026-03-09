import 'package:flutter/foundation.dart';

import '../models/user_level.dart';
import '../providers/auth_provider.dart';
import '../services/user_level_service.dart';

class UserLevelProvider with ChangeNotifier {
  final UserLevelService _service;
  final AuthProvider _auth;

  UserLevel? _userLevel;
  bool _loading = false;

  UserLevelProvider(this._service, this._auth) {
    _auth.addListener(_onAuthChanged);
    _onAuthChanged();
  }

  UserLevel? get userLevel => _userLevel;
  bool get loading => _loading;

  void _onAuthChanged() {
    final u = _auth.user;
    if (u == null) {
      _userLevel = null;
      _loading = false;
      notifyListeners();
      return;
    }
    if (_userLevel == null && !_loading) {
      Future.microtask(() => refresh());
    }
  }

  Future<void> refresh() async {
    final uid = _auth.user?.id;
    if (uid == null || uid.isEmpty) {
      _userLevel = null;
      _loading = false;
      notifyListeners();
      return;
    }
    _loading = true;
    notifyListeners();
    try {
      _userLevel = await _service.getUserLevel();
    } catch (_) {
      _userLevel = null;
    }
    _loading = false;
    notifyListeners();
  }

  @override
  void dispose() {
    _auth.removeListener(_onAuthChanged);
    super.dispose();
  }
}
