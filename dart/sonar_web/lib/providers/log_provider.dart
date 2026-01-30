import 'package:flutter/foundation.dart';

import '../models/audit_item.dart';
import '../services/chat_service.dart';

class LogProvider with ChangeNotifier {
  final ChatService _service;

  LogProvider(this._service);

  List<AuditItem> _items = [];
  bool _loading = false;

  List<AuditItem> get items => _items;
  bool get loading => _loading;

  Future<void> refresh() async {
    _loading = true;
    notifyListeners();
    try {
      _items = await _service.getChat();
    } catch (_) {
      _items = [];
    }
    _loading = false;
    notifyListeners();
  }
}
