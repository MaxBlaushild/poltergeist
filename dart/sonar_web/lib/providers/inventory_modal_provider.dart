import 'package:flutter/foundation.dart';

/// Minimal provider for NewItemModal and UsedItemModal.
/// Items are set when inventory actions occur (not wired in this polish pass).
class InventoryModalProvider with ChangeNotifier {
  Map<String, dynamic>? _presentedItem;
  Map<String, dynamic>? _usedItem;

  Map<String, dynamic>? get presentedItem => _presentedItem;
  Map<String, dynamic>? get usedItem => _usedItem;

  void setPresentedItem(Map<String, dynamic>? item) {
    _presentedItem = item;
    notifyListeners();
  }

  void setUsedItem(Map<String, dynamic>? item) {
    _usedItem = item;
    notifyListeners();
  }
}
