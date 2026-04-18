import 'package:flutter/foundation.dart';

/// Shared presentation state for NewItemModal and UsedItemModal.
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
