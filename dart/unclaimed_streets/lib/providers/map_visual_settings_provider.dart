import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';

class MapVisualSettingsProvider extends ChangeNotifier {
  static const int minContentLevelOffset = -10;
  static const int maxContentLevelOffset = 10;
  static const String _zoneKindMapStylingPrefsKey =
      'zone_kind_map_styling_enabled';
  static const String _unselectedZoneKindTilingPrefsKey =
      'unselected_zone_kind_tiling_enabled';
  static const String _proximityBypassPrefsKey =
      'debug_proximity_bypass_enabled';
  static const String _contentLevelOffsetPrefsKey = 'content_level_offset';

  bool _zoneKindMapStylingEnabled = true;
  bool _unselectedZoneKindTilingEnabled = true;
  bool _proximityBypassEnabled = false;
  int _contentLevelOffset = 0;

  bool get zoneKindMapStylingEnabled => _zoneKindMapStylingEnabled;
  bool get unselectedZoneKindTilingEnabled => _unselectedZoneKindTilingEnabled;
  bool get proximityBypassEnabled => _proximityBypassEnabled;
  int get contentLevelOffset => _contentLevelOffset;

  Future<void> load() async {
    final prefs = await SharedPreferences.getInstance();
    final enabled = prefs.getBool(_zoneKindMapStylingPrefsKey) ?? true;
    final tilingEnabled =
        prefs.getBool(_unselectedZoneKindTilingPrefsKey) ?? true;
    final proximityBypassEnabled =
        prefs.getBool(_proximityBypassPrefsKey) ?? false;
    final contentLevelOffset = _clampContentLevelOffset(
      prefs.getInt(_contentLevelOffsetPrefsKey) ?? 0,
    );
    if (_zoneKindMapStylingEnabled == enabled &&
        _unselectedZoneKindTilingEnabled == tilingEnabled &&
        _proximityBypassEnabled == proximityBypassEnabled &&
        _contentLevelOffset == contentLevelOffset) {
      return;
    }
    _zoneKindMapStylingEnabled = enabled;
    _unselectedZoneKindTilingEnabled = tilingEnabled;
    _proximityBypassEnabled = proximityBypassEnabled;
    _contentLevelOffset = contentLevelOffset;
    notifyListeners();
  }

  Future<void> setZoneKindMapStylingEnabled(bool enabled) async {
    if (_zoneKindMapStylingEnabled == enabled) {
      return;
    }
    _zoneKindMapStylingEnabled = enabled;
    notifyListeners();

    final prefs = await SharedPreferences.getInstance();
    await prefs.setBool(_zoneKindMapStylingPrefsKey, enabled);
  }

  Future<void> setUnselectedZoneKindTilingEnabled(bool enabled) async {
    if (_unselectedZoneKindTilingEnabled == enabled) {
      return;
    }
    _unselectedZoneKindTilingEnabled = enabled;
    notifyListeners();

    final prefs = await SharedPreferences.getInstance();
    await prefs.setBool(_unselectedZoneKindTilingPrefsKey, enabled);
  }

  Future<void> setProximityBypassEnabled(bool enabled) async {
    if (_proximityBypassEnabled == enabled) {
      return;
    }
    _proximityBypassEnabled = enabled;
    notifyListeners();

    final prefs = await SharedPreferences.getInstance();
    await prefs.setBool(_proximityBypassPrefsKey, enabled);
  }

  Future<void> setContentLevelOffset(int offset) async {
    final clamped = _clampContentLevelOffset(offset);
    if (_contentLevelOffset == clamped) {
      return;
    }
    _contentLevelOffset = clamped;
    notifyListeners();

    final prefs = await SharedPreferences.getInstance();
    await prefs.setInt(_contentLevelOffsetPrefsKey, clamped);
  }

  static int _clampContentLevelOffset(int offset) {
    if (offset < minContentLevelOffset) {
      return minContentLevelOffset;
    }
    if (offset > maxContentLevelOffset) {
      return maxContentLevelOffset;
    }
    return offset;
  }
}
