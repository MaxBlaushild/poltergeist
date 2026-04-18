import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:crypto/crypto.dart';
import 'package:dio/dio.dart';
import 'package:path_provider/path_provider.dart';

class InventoryImageCache {
  InventoryImageCache._();

  static final InventoryImageCache instance = InventoryImageCache._();

  static const _cacheDirectoryName = 'inventory_item_image_cache';
  static const _maxCacheBytes = 64 * 1024 * 1024;
  static const _maxCacheEntries = 160;
  static const _refreshAfter = Duration(days: 21);

  final Dio _dio = Dio();
  final Map<String, Future<File?>> _inFlight = <String, Future<File?>>{};

  Directory? _cacheDirectory;
  Future<Directory>? _cacheDirectoryFuture;
  Future<void>? _cleanupFuture;

  Future<File?> getCachedFile(String imageUrl) async {
    final normalizedUrl = imageUrl.trim();
    if (normalizedUrl.isEmpty) {
      return null;
    }
    final file = await _cacheFileForUrl(normalizedUrl);
    if (!await file.exists()) {
      return null;
    }
    try {
      final stat = await file.stat();
      await _touch(file);
      if (DateTime.now().difference(stat.modified) > _refreshAfter) {
        unawaited(warmImage(normalizedUrl, forceRefresh: true));
      }
      return file;
    } catch (_) {
      return null;
    }
  }

  Future<File?> warmImage(String imageUrl, {bool forceRefresh = false}) {
    final normalizedUrl = imageUrl.trim();
    if (normalizedUrl.isEmpty) {
      return Future<File?>.value(null);
    }

    final existing = _inFlight[normalizedUrl];
    if (existing != null) {
      return existing;
    }

    final future = _downloadAndStore(normalizedUrl, forceRefresh: forceRefresh);
    _inFlight[normalizedUrl] = future;
    return future.whenComplete(() {
      _inFlight.remove(normalizedUrl);
    });
  }

  Future<File?> _downloadAndStore(
    String imageUrl, {
    required bool forceRefresh,
  }) async {
    try {
      final file = await _cacheFileForUrl(imageUrl);
      if (!forceRefresh && await file.exists()) {
        await _touch(file);
        return file;
      }

      final response = await _dio.get<List<int>>(
        imageUrl,
        options: Options(
          responseType: ResponseType.bytes,
          receiveTimeout: const Duration(seconds: 15),
          sendTimeout: const Duration(seconds: 15),
        ),
      );
      final bytes = response.data;
      if (bytes == null || bytes.isEmpty) {
        return await file.exists() ? file : null;
      }

      await file.parent.create(recursive: true);
      final tempFile = File('${file.path}.download');
      await tempFile.writeAsBytes(bytes, flush: true);
      if (await file.exists()) {
        await file.delete();
      }
      final storedFile = await tempFile.rename(file.path);
      await _touch(storedFile);
      unawaited(_cleanupIfNeeded());
      return storedFile;
    } catch (_) {
      return null;
    }
  }

  Future<File> _cacheFileForUrl(String imageUrl) async {
    final directory = await _ensureCacheDirectory();
    final filename =
        '${sha1.convert(utf8.encode(imageUrl)).toString()}${_extensionForUrl(imageUrl)}';
    return File('${directory.path}/$filename');
  }

  Future<Directory> _ensureCacheDirectory() async {
    if (_cacheDirectory != null) {
      return _cacheDirectory!;
    }
    final existingFuture = _cacheDirectoryFuture;
    if (existingFuture != null) {
      return existingFuture;
    }
    final future = () async {
      final root = await getTemporaryDirectory();
      final directory = Directory('${root.path}/$_cacheDirectoryName');
      if (!await directory.exists()) {
        await directory.create(recursive: true);
      }
      _cacheDirectory = directory;
      return directory;
    }();
    _cacheDirectoryFuture = future;
    final directory = await future;
    if (identical(_cacheDirectoryFuture, future)) {
      _cacheDirectoryFuture = null;
    }
    return directory;
  }

  Future<void> _cleanupIfNeeded() {
    final existingFuture = _cleanupFuture;
    if (existingFuture != null) {
      return existingFuture;
    }
    final future = _cleanupCache();
    _cleanupFuture = future;
    return future.whenComplete(() {
      if (identical(_cleanupFuture, future)) {
        _cleanupFuture = null;
      }
    });
  }

  Future<void> _cleanupCache() async {
    try {
      final directory = await _ensureCacheDirectory();
      final files = await directory
          .list()
          .where((entity) => entity is File)
          .cast<File>()
          .where((file) => !file.path.endsWith('.download'))
          .toList();
      if (files.isEmpty) {
        return;
      }

      final entries = <_CachedFileEntry>[];
      var totalBytes = 0;
      for (final file in files) {
        try {
          final stat = await file.stat();
          totalBytes += stat.size;
          entries.add(
            _CachedFileEntry(
              file: file,
              sizeBytes: stat.size,
              modifiedAt: stat.modified,
            ),
          );
        } catch (_) {}
      }

      entries.sort((a, b) => a.modifiedAt.compareTo(b.modifiedAt));
      while (entries.length > _maxCacheEntries || totalBytes > _maxCacheBytes) {
        final entry = entries.removeAt(0);
        totalBytes -= entry.sizeBytes;
        try {
          await entry.file.delete();
        } catch (_) {}
      }
    } catch (_) {
      // Cache cleanup is best-effort.
    }
  }

  Future<void> _touch(File file) async {
    try {
      await file.setLastModified(DateTime.now());
    } catch (_) {}
  }

  String _extensionForUrl(String imageUrl) {
    final path = Uri.tryParse(imageUrl)?.path ?? '';
    final slashIndex = path.lastIndexOf('/');
    final filename = slashIndex >= 0 ? path.substring(slashIndex + 1) : path;
    final dotIndex = filename.lastIndexOf('.');
    if (dotIndex < 0) {
      return '.img';
    }
    final extension = filename.substring(dotIndex).toLowerCase();
    switch (extension) {
      case '.png':
      case '.jpg':
      case '.jpeg':
      case '.webp':
      case '.gif':
      case '.bmp':
        return extension;
      default:
        return '.img';
    }
  }
}

class _CachedFileEntry {
  const _CachedFileEntry({
    required this.file,
    required this.sizeBytes,
    required this.modifiedAt,
  });

  final File file;
  final int sizeBytes;
  final DateTime modifiedAt;
}
