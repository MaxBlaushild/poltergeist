import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';

class MapVisualSettingsProvider extends ChangeNotifier {
  static const String _zoneKindMapStylingPrefsKey =
      'zone_kind_map_styling_enabled';
  static const String _unselectedZoneKindTilingPrefsKey =
      'unselected_zone_kind_tiling_enabled';

  bool _zoneKindMapStylingEnabled = false;
  bool _unselectedZoneKindTilingEnabled = false;

  bool get zoneKindMapStylingEnabled => _zoneKindMapStylingEnabled;
  bool get unselectedZoneKindTilingEnabled => _unselectedZoneKindTilingEnabled;

  Future<void> load() async {
    final prefs = await SharedPreferences.getInstance();
    final enabled = prefs.getBool(_zoneKindMapStylingPrefsKey) ?? false;
    final tilingEnabled =
        prefs.getBool(_unselectedZoneKindTilingPrefsKey) ?? false;
    if (_zoneKindMapStylingEnabled == enabled &&
        _unselectedZoneKindTilingEnabled == tilingEnabled) {
      return;
    }
    _zoneKindMapStylingEnabled = enabled;
    _unselectedZoneKindTilingEnabled = tilingEnabled;
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
}
