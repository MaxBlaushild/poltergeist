import 'dart:typed_data';

import 'package:http/http.dart' as http;

import '../constants/zone_kind_visuals.dart';

final Map<String, Uint8List> _zoneKindPatternAssetCache = <String, Uint8List>{};
final Map<String, Future<Uint8List?>> _zoneKindPatternAssetInFlight =
    <String, Future<Uint8List?>>{};

Future<Uint8List?> loadZoneKindPatternTile(String? imageUrl) {
  final trimmed = (imageUrl ?? '').trim();
  if (trimmed.isEmpty) {
    return Future<Uint8List?>.value(null);
  }
  final cached = _zoneKindPatternAssetCache[trimmed];
  if (cached != null) {
    return Future<Uint8List?>.value(cached);
  }
  final inFlight = _zoneKindPatternAssetInFlight[trimmed];
  if (inFlight != null) {
    return inFlight;
  }

  final future = (() async {
    try {
      final response = await http.get(Uri.parse(trimmed));
      if (response.statusCode < 200 || response.statusCode >= 300) {
        return null;
      }
      final bytes = response.bodyBytes;
      if (bytes.isEmpty) {
        return null;
      }
      _zoneKindPatternAssetCache[trimmed] = bytes;
      return bytes;
    } catch (_) {
      return null;
    } finally {
      _zoneKindPatternAssetInFlight.remove(trimmed);
    }
  })();

  _zoneKindPatternAssetInFlight[trimmed] = future;
  return future;
}

String zoneKindPatternAssetImageId(String? rawKind, String? imageUrl) {
  final normalizedKind = normalizeZoneKindSlug(rawKind);
  final normalizedUrl = (imageUrl ?? '').trim();
  final hash = normalizedUrl.hashCode.abs();
  return 'zone_kind_pattern_${normalizedKind}_$hash';
}
