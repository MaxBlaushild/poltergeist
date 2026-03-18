import 'package:flutter/foundation.dart';

import '../models/inventory_item.dart';

class BasePlacementRequest {
  const BasePlacementRequest({
    required this.ownedInventoryItem,
    required this.inventoryItem,
  });

  final OwnedInventoryItem ownedInventoryItem;
  final InventoryItem inventoryItem;
}

class BasePlacementProvider with ChangeNotifier {
  BasePlacementRequest? _pendingRequest;

  BasePlacementRequest? get pendingRequest => _pendingRequest;

  void requestPlacement(
    OwnedInventoryItem ownedInventoryItem,
    InventoryItem inventoryItem,
  ) {
    _pendingRequest = BasePlacementRequest(
      ownedInventoryItem: ownedInventoryItem,
      inventoryItem: inventoryItem,
    );
    notifyListeners();
  }

  void clearRequest() {
    if (_pendingRequest == null) return;
    _pendingRequest = null;
    notifyListeners();
  }
}
