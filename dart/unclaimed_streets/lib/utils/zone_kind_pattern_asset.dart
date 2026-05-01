import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:crypto/crypto.dart';
import 'package:http/http.dart' as http;
import 'package:image/image.dart' as img;
import 'package:path_provider/path_provider.dart';

import '../constants/zone_kind_visuals.dart';

const _zoneKindPatternAssetVersion = 'v2';
const _zonePatternAssetDiskCacheDirectoryName = 'zone_pattern_tile_cache';
const _zonePatternAssetDiskMaxCacheBytes = 24 * 1024 * 1024;
const _zonePatternAssetDiskMaxCacheEntries = 128;

final Map<String, Uint8List> _zoneKindPatternAssetCache = <String, Uint8List>{};
final Map<String, Future<Uint8List?>> _zoneKindPatternAssetInFlight =
    <String, Future<Uint8List?>>{};
Directory? _zonePatternAssetCacheDirectory;
Future<Directory>? _zonePatternAssetCacheDirectoryFuture;
Future<void>? _zonePatternAssetCacheCleanupFuture;

Future<Uint8List?> loadZoneKindPatternTile(String? imageUrl) {
  return _loadZonePatternTileAsset(
    imageUrl,
    cacheNamespace: 'zone-kind',
    processor: _processZoneKindPatternTile,
  );
}

Future<Uint8List?> loadZonePatternTileAsset(String? imageUrl) {
  return loadZoneKindPatternTile(imageUrl);
}

Future<Uint8List?> loadZoneShroudPatternTile(String? imageUrl) {
  return _loadZonePatternTileAsset(
    imageUrl,
    cacheNamespace: 'zone-shroud',
    processor: _processZoneShroudPatternTile,
  );
}

Future<Uint8List?> _loadZonePatternTileAsset(
  String? imageUrl, {
  required String cacheNamespace,
  required Uint8List? Function(Uint8List bytes) processor,
}) {
  final trimmed = (imageUrl ?? '').trim();
  if (trimmed.isEmpty) {
    return Future<Uint8List?>.value(null);
  }
  final cacheKey = '$cacheNamespace::$trimmed';
  final cached = _zoneKindPatternAssetCache[cacheKey];
  if (cached != null) {
    return Future<Uint8List?>.value(cached);
  }
  final inFlight = _zoneKindPatternAssetInFlight[cacheKey];
  if (inFlight != null) {
    return inFlight;
  }

  final future = (() async {
    try {
      final diskBytes = await _readZonePatternTileFromDisk(cacheKey);
      if (diskBytes != null) {
        _zoneKindPatternAssetCache[cacheKey] = diskBytes;
        return diskBytes;
      }

      final response = await http.get(Uri.parse(trimmed));
      if (response.statusCode < 200 || response.statusCode >= 300) {
        return null;
      }
      final bytes = response.bodyBytes;
      if (bytes.isEmpty) {
        return null;
      }
      final processed = processor(bytes);
      if (processed == null || processed.isEmpty) {
        return null;
      }
      _zoneKindPatternAssetCache[cacheKey] = processed;
      unawaited(_writeZonePatternTileToDisk(cacheKey, processed));
      return processed;
    } catch (_) {
      return null;
    } finally {
      _zoneKindPatternAssetInFlight.remove(cacheKey);
    }
  })();

  _zoneKindPatternAssetInFlight[cacheKey] = future;
  return future;
}

Future<File> _zonePatternTileCacheFileForKey(String cacheKey) async {
  final directory = await _ensureZonePatternTileCacheDirectory();
  final filename = '${sha1.convert(utf8.encode(cacheKey)).toString()}.png';
  return File('${directory.path}/$filename');
}

Future<Directory> _ensureZonePatternTileCacheDirectory() async {
  if (_zonePatternAssetCacheDirectory != null) {
    return _zonePatternAssetCacheDirectory!;
  }
  final existingFuture = _zonePatternAssetCacheDirectoryFuture;
  if (existingFuture != null) {
    return existingFuture;
  }
  final future = () async {
    final root = await getTemporaryDirectory();
    final directory = Directory(
      '${root.path}/$_zonePatternAssetDiskCacheDirectoryName',
    );
    if (!await directory.exists()) {
      await directory.create(recursive: true);
    }
    _zonePatternAssetCacheDirectory = directory;
    return directory;
  }();
  _zonePatternAssetCacheDirectoryFuture = future;
  final directory = await future;
  if (identical(_zonePatternAssetCacheDirectoryFuture, future)) {
    _zonePatternAssetCacheDirectoryFuture = null;
  }
  return directory;
}

Future<Uint8List?> _readZonePatternTileFromDisk(String cacheKey) async {
  try {
    final file = await _zonePatternTileCacheFileForKey(cacheKey);
    if (!await file.exists()) {
      return null;
    }
    final bytes = await file.readAsBytes();
    if (bytes.isEmpty) {
      return null;
    }
    await _touchZonePatternTileCacheFile(file);
    return bytes;
  } catch (_) {
    return null;
  }
}

Future<void> _writeZonePatternTileToDisk(
  String cacheKey,
  Uint8List bytes,
) async {
  try {
    final file = await _zonePatternTileCacheFileForKey(cacheKey);
    await file.parent.create(recursive: true);
    final tempFile = File('${file.path}.download');
    await tempFile.writeAsBytes(bytes, flush: true);
    if (await file.exists()) {
      await file.delete();
    }
    final storedFile = await tempFile.rename(file.path);
    await _touchZonePatternTileCacheFile(storedFile);
    unawaited(_cleanupZonePatternTileCacheIfNeeded());
  } catch (_) {
    // Disk cache writes are best-effort.
  }
}

Future<void> _touchZonePatternTileCacheFile(File file) async {
  try {
    await file.setLastModified(DateTime.now());
  } catch (_) {}
}

Future<void> _cleanupZonePatternTileCacheIfNeeded() {
  final existingFuture = _zonePatternAssetCacheCleanupFuture;
  if (existingFuture != null) {
    return existingFuture;
  }
  final future = _cleanupZonePatternTileCache();
  _zonePatternAssetCacheCleanupFuture = future;
  return future.whenComplete(() {
    if (identical(_zonePatternAssetCacheCleanupFuture, future)) {
      _zonePatternAssetCacheCleanupFuture = null;
    }
  });
}

Future<void> _cleanupZonePatternTileCache() async {
  try {
    final directory = await _ensureZonePatternTileCacheDirectory();
    final files = await directory
        .list()
        .where((entity) => entity is File)
        .cast<File>()
        .where((file) => !file.path.endsWith('.download'))
        .toList();
    if (files.isEmpty) {
      return;
    }

    final entries = <_ZonePatternTileCacheEntry>[];
    var totalBytes = 0;
    for (final file in files) {
      try {
        final stat = await file.stat();
        totalBytes += stat.size;
        entries.add(
          _ZonePatternTileCacheEntry(
            file: file,
            sizeBytes: stat.size,
            modifiedAt: stat.modified,
          ),
        );
      } catch (_) {}
    }

    entries.sort((a, b) => a.modifiedAt.compareTo(b.modifiedAt));
    while (entries.length > _zonePatternAssetDiskMaxCacheEntries ||
        totalBytes > _zonePatternAssetDiskMaxCacheBytes) {
      final entry = entries.removeAt(0);
      totalBytes -= entry.sizeBytes;
      try {
        await entry.file.delete();
      } catch (_) {}
    }
  } catch (_) {
    // Disk cache cleanup is best-effort.
  }
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

Uint8List? _processZoneShroudPatternTile(Uint8List bytes) {
  final decoded = img.decodeImage(bytes);
  if (decoded == null) {
    return bytes;
  }
  final contrasted = img.adjustColor(decoded, contrast: 1.22, saturation: 0.92);
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
      if (darkness < 8) {
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
      final strengthenedAlpha = (darkness * 4.6).clamp(136, 255);
      final tinted = img.ColorRgb8(
        ((pixel.r.toInt() * 0.36) + 82).round().clamp(0, 255),
        ((pixel.g.toInt() * 0.42) + 94).round().clamp(0, 255),
        ((pixel.b.toInt() * 0.5) + 104).round().clamp(0, 255),
      );
      output.setPixelRgba(
        x,
        y,
        tinted.r.toInt(),
        tinted.g.toInt(),
        tinted.b.toInt(),
        alpha < strengthenedAlpha ? alpha : strengthenedAlpha,
      );
    }
  }
  return Uint8List.fromList(img.encodePng(output));
}

String zoneKindPatternAssetImageId(String? rawKind, String? imageUrl) {
  final normalizedKind = normalizeZoneKindSlug(rawKind);
  return zonePatternAssetImageId('zone_kind_pattern_$normalizedKind', imageUrl);
}

String zoneShroudPatternAssetImageId(String? imageUrl) {
  return zonePatternAssetImageId('zone_shroud_pattern', imageUrl);
}

class _ZonePatternTileCacheEntry {
  const _ZonePatternTileCacheEntry({
    required this.file,
    required this.sizeBytes,
    required this.modifiedAt,
  });

  final File file;
  final int sizeBytes;
  final DateTime modifiedAt;
}

String zonePatternAssetImageId(String prefix, String? imageUrl) {
  final normalizedUrl = (imageUrl ?? '').trim();
  final hash = normalizedUrl.hashCode.abs();
  return '${prefix}_${_zoneKindPatternAssetVersion}_$hash';
}
