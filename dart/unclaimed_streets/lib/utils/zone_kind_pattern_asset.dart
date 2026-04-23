import 'dart:typed_data';

import 'package:http/http.dart' as http;
import 'package:image/image.dart' as img;

import '../constants/zone_kind_visuals.dart';

const _zoneKindPatternAssetVersion = 'v2';

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
      final processed = _processZoneKindPatternTile(bytes);
      if (processed == null || processed.isEmpty) {
        return null;
      }
      _zoneKindPatternAssetCache[trimmed] = processed;
      return processed;
    } catch (_) {
      return null;
    } finally {
      _zoneKindPatternAssetInFlight.remove(trimmed);
    }
  })();

  _zoneKindPatternAssetInFlight[trimmed] = future;
  return future;
}

Uint8List? _processZoneKindPatternTile(Uint8List bytes) {
  final decoded = img.decodeImage(bytes);
  if (decoded == null) {
    return bytes;
  }
  final contrasted = img.adjustColor(decoded, contrast: 1.16, saturation: 1.22);
  final output = img.Image.from(contrasted);
  for (var y = 0; y < output.height; y++) {
    for (var x = 0; x < output.width; x++) {
      final pixel = output.getPixel(x, y);
      final alpha = pixel.a.toInt();
      if (alpha == 0) {
        continue;
      }
      final luma =
          ((pixel.r.toInt() * 299) +
              (pixel.g.toInt() * 587) +
              (pixel.b.toInt() * 114)) ~/
          1000;
      final darkness = 255 - luma;
      if (darkness < 12) {
        output.setPixelRgba(
          x,
          y,
          pixel.r.toInt(),
          pixel.g.toInt(),
          pixel.b.toInt(),
          0,
        );
        continue;
      }
      final strengthenedAlpha = (darkness * 3).clamp(104, 244);
      output.setPixelRgba(
        x,
        y,
        pixel.r.toInt(),
        pixel.g.toInt(),
        pixel.b.toInt(),
        alpha < strengthenedAlpha ? alpha : strengthenedAlpha,
      );
    }
  }
  return Uint8List.fromList(img.encodePng(output));
}

String zoneKindPatternAssetImageId(String? rawKind, String? imageUrl) {
  final normalizedKind = normalizeZoneKindSlug(rawKind);
  final normalizedUrl = (imageUrl ?? '').trim();
  final hash = normalizedUrl.hashCode.abs();
  return 'zone_kind_pattern_${normalizedKind}_${_zoneKindPatternAssetVersion}_$hash';
}
