import 'dart:async';
import 'dart:convert';

import '../models/equipment_item.dart';
import '../models/inventory_item.dart';
import '../models/outfit_generation.dart';
import 'api_client.dart';
import 'package:shared_preferences/shared_preferences.dart';

class InventoryCacheSnapshot {
  const InventoryCacheSnapshot({
    required this.items,
    required this.ownedItems,
    required this.equipment,
    this.cachedAt,
  });

  final List<InventoryItem> items;
  final List<OwnedInventoryItem> ownedItems;
  final List<EquippedItem> equipment;
  final DateTime? cachedAt;

  bool get hasData =>
      items.isNotEmpty || ownedItems.isNotEmpty || equipment.isNotEmpty;
}

class InventoryService {
  static const _inventoryCachePrefsKey = 'inventory_snapshot_v1';

  final ApiClient _api;
  List<Map<String, dynamic>>? _cachedItemsJson;
  List<Map<String, dynamic>>? _cachedOwnedItemsJson;
  List<Map<String, dynamic>>? _cachedEquipmentJson;
  DateTime? _cacheUpdatedAt;
  bool _cacheHydrated = false;
  Future<void> _cacheWriteChain = Future<void>.value();
  Future<void>? _cacheLoadFuture;

  InventoryService(this._api);

  Future<InventoryCacheSnapshot?> getCachedSnapshot() async {
    await _ensureCacheHydrated();
    return _buildSnapshotFromCache();
  }

  Future<List<InventoryItem>> getInventoryItems({
    bool preferCache = false,
  }) async {
    if (preferCache) {
      final cached = await getCachedSnapshot();
      if (cached != null && cached.items.isNotEmpty) {
        return cached.items;
      }
    }
    final fresh = await refreshInventoryItems();
    if (fresh != null) {
      return fresh;
    }
    final cached = await getCachedSnapshot();
    return cached?.items ?? const <InventoryItem>[];
  }

  Future<List<InventoryItem>?> refreshInventoryItems() async {
    try {
      final list = await _api.get<List<dynamic>>('/sonar/inventory-items');
      final normalized = _normalizeJsonList(list);
      final items = normalized
          .map((entry) => InventoryItem.fromJson(entry))
          .toList(growable: false);
      _cachedItemsJson = normalized;
      _cacheUpdatedAt = DateTime.now();
      _cacheHydrated = true;
      unawaited(_persistCache());
      return items;
    } catch (_) {
      return null;
    }
  }

  Future<List<OwnedInventoryItem>> getOwnedInventoryItems({
    bool preferCache = false,
  }) async {
    if (preferCache) {
      final cached = await getCachedSnapshot();
      if (cached != null && cached.ownedItems.isNotEmpty) {
        return cached.ownedItems;
      }
    }
    final fresh = await refreshOwnedInventoryItems();
    if (fresh != null) {
      return fresh;
    }
    final cached = await getCachedSnapshot();
    return cached?.ownedItems ?? const <OwnedInventoryItem>[];
  }

  Future<List<OwnedInventoryItem>?> refreshOwnedInventoryItems() async {
    try {
      final list = await _api.get<List<dynamic>>('/sonar/ownedInventoryItems');
      final normalized = _normalizeJsonList(list);
      final ownedItems = normalized
          .map((entry) => OwnedInventoryItem.fromJson(entry))
          .toList(growable: false);
      _cachedOwnedItemsJson = normalized;
      _cacheUpdatedAt = DateTime.now();
      _cacheHydrated = true;
      unawaited(_persistCache());
      return ownedItems;
    } catch (_) {
      return null;
    }
  }

  Future<List<EquippedItem>> getEquipment({bool preferCache = false}) async {
    if (preferCache) {
      final cached = await getCachedSnapshot();
      if (cached != null && cached.equipment.isNotEmpty) {
        return cached.equipment;
      }
    }
    final fresh = await refreshEquipment();
    if (fresh != null) {
      return fresh;
    }
    final cached = await getCachedSnapshot();
    return cached?.equipment ?? const <EquippedItem>[];
  }

  Future<List<EquippedItem>?> refreshEquipment() async {
    try {
      final list = await _api.get<List<dynamic>>('/sonar/equipment');
      final normalized = _normalizeJsonList(list);
      final equipment = normalized
          .map((entry) => EquippedItem.fromJson(entry))
          .toList(growable: false);
      _cachedEquipmentJson = normalized;
      _cacheUpdatedAt = DateTime.now();
      _cacheHydrated = true;
      unawaited(_persistCache());
      return equipment;
    } catch (_) {
      return null;
    }
  }

  Future<void> equipItem(
    String ownedInventoryItemId, {
    required String slot,
  }) async {
    await _api.post<dynamic>(
      '/sonar/equipment/equip',
      data: {'ownedInventoryItemId': ownedInventoryItemId, 'slot': slot},
    );
  }

  Future<void> unequipSlot(String slot) async {
    await _api.post<dynamic>('/sonar/equipment/unequip', data: {'slot': slot});
  }

  Future<void> unequipItem(String ownedInventoryItemId) async {
    await _api.post<dynamic>(
      '/sonar/equipment/unequip',
      data: {'ownedInventoryItemId': ownedInventoryItemId},
    );
  }

  /// POST /sonar/inventory/:ownedInventoryItemID/use
  /// Optional [targetTeamId] for items that require a team target.
  Future<Map<String, dynamic>> useItem(
    String ownedInventoryItemId, {
    String? targetTeamId,
    String? targetUserId,
    double? baseLatitude,
    double? baseLongitude,
  }) async {
    final raw = await _api.post<dynamic>(
      '/sonar/inventory/$ownedInventoryItemId/use',
      data: {
        if (targetTeamId != null && targetTeamId.isNotEmpty)
          'targetTeamId': targetTeamId,
        if (targetUserId != null && targetUserId.isNotEmpty)
          'targetUserId': targetUserId,
        if (baseLatitude != null) 'baseLatitude': baseLatitude,
        if (baseLongitude != null) 'baseLongitude': baseLongitude,
      },
    );
    return raw is Map ? Map<String, dynamic>.from(raw) : <String, dynamic>{};
  }

  Future<OutfitGeneration?> getOutfitGenerationStatus(
    String ownedInventoryItemId,
  ) async {
    try {
      final data = await _api.get<Map<String, dynamic>>(
        '/sonar/inventory/$ownedInventoryItemId/outfit-generation',
      );
      return OutfitGeneration.fromJson(data);
    } catch (_) {
      return null;
    }
  }

  Future<OutfitGeneration> useOutfitItem(
    String ownedInventoryItemId, {
    required String selfieUrl,
  }) async {
    final data = await _api.post<Map<String, dynamic>>(
      '/sonar/inventory/$ownedInventoryItemId/use-outfit',
      data: {'selfieUrl': selfieUrl},
    );
    return OutfitGeneration.fromJson(data);
  }

  Future<void> _ensureCacheHydrated() async {
    if (_cacheHydrated) {
      return;
    }
    final loadFuture = _cacheLoadFuture ??= _loadCacheFromPrefs();
    await loadFuture;
    if (identical(_cacheLoadFuture, loadFuture)) {
      _cacheLoadFuture = null;
    }
  }

  Future<void> _loadCacheFromPrefs() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final raw = prefs.getString(_inventoryCachePrefsKey);
      if (raw == null || raw.isEmpty) {
        _cacheHydrated = true;
        return;
      }
      final decoded = jsonDecode(raw);
      if (decoded is! Map) {
        _cacheHydrated = true;
        return;
      }
      final normalized = _normalizeJsonMap(decoded);
      _cachedItemsJson = _readCachedList(normalized['items']);
      _cachedOwnedItemsJson = _readCachedList(normalized['ownedItems']);
      _cachedEquipmentJson = _readCachedList(normalized['equipment']);
      final cachedAtRaw = normalized['cachedAt']?.toString();
      _cacheUpdatedAt = cachedAtRaw == null
          ? null
          : DateTime.tryParse(cachedAtRaw);
    } catch (_) {
      // Ignore cache hydration failures and fall back to live network data.
    } finally {
      _cacheHydrated = true;
    }
  }

  InventoryCacheSnapshot? _buildSnapshotFromCache() {
    if (!_cacheHydrated &&
        _cachedItemsJson == null &&
        _cachedOwnedItemsJson == null &&
        _cachedEquipmentJson == null) {
      return null;
    }
    return InventoryCacheSnapshot(
      items: (_cachedItemsJson ?? const <Map<String, dynamic>>[])
          .map((entry) => InventoryItem.fromJson(entry))
          .toList(growable: false),
      ownedItems: (_cachedOwnedItemsJson ?? const <Map<String, dynamic>>[])
          .map((entry) => OwnedInventoryItem.fromJson(entry))
          .toList(growable: false),
      equipment: (_cachedEquipmentJson ?? const <Map<String, dynamic>>[])
          .map((entry) => EquippedItem.fromJson(entry))
          .toList(growable: false),
      cachedAt: _cacheUpdatedAt,
    );
  }

  Future<void> _persistCache() {
    _cacheWriteChain = _cacheWriteChain.then((_) async {
      final prefs = await SharedPreferences.getInstance();
      final payload = <String, dynamic>{
        'cachedAt': (_cacheUpdatedAt ?? DateTime.now()).toIso8601String(),
        'items': _cachedItemsJson ?? const <Map<String, dynamic>>[],
        'ownedItems': _cachedOwnedItemsJson ?? const <Map<String, dynamic>>[],
        'equipment': _cachedEquipmentJson ?? const <Map<String, dynamic>>[],
      };
      await prefs.setString(_inventoryCachePrefsKey, jsonEncode(payload));
    });
    return _cacheWriteChain;
  }

  List<Map<String, dynamic>>? _readCachedList(dynamic raw) {
    if (raw is! List) {
      return null;
    }
    return raw.whereType<Map>().map(_normalizeJsonMap).toList(growable: false);
  }

  List<Map<String, dynamic>> _normalizeJsonList(List<dynamic> rawList) {
    return rawList
        .whereType<Map>()
        .map(_normalizeJsonMap)
        .toList(growable: false);
  }

  Map<String, dynamic> _normalizeJsonMap(Map raw) {
    final normalized = <String, dynamic>{};
    raw.forEach((key, value) {
      normalized[key.toString()] = _normalizeJsonValue(value);
    });
    return normalized;
  }

  dynamic _normalizeJsonValue(dynamic value) {
    if (value is Map) {
      return _normalizeJsonMap(value);
    }
    if (value is List) {
      return value.map(_normalizeJsonValue).toList(growable: false);
    }
    return value;
  }
}
