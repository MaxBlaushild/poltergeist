import 'dart:math' as math;
import 'dart:typed_data';

import 'package:http/http.dart' as http;
import 'package:image/image.dart' as img;

const _placeholderUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/poi-undiscovered.png';
const _thumbnailSize = 192;
const _cornerRadius = 14;

final Map<String, Uint8List> _thumbnailCache = {};
final Map<String, Future<Uint8List?>> _thumbnailInFlight = {};
final Map<String, Uint8List> _sourceImageCache = {};
final Map<String, Future<Uint8List?>> _sourceImageInFlight = {};

Future<Uint8List?> _loadThumbnailCached(
  String cacheKey,
  Future<Uint8List?> Function() loader,
) {
  final cached = _thumbnailCache[cacheKey];
  if (cached != null) return Future.value(cached);
  final inFlight = _thumbnailInFlight[cacheKey];
  if (inFlight != null) return inFlight;
  final future = loader()
      .then((bytes) {
        if (bytes != null) {
          _thumbnailCache[cacheKey] = bytes;
        }
        _thumbnailInFlight.remove(cacheKey);
        return bytes;
      })
      .catchError((_) {
        _thumbnailInFlight.remove(cacheKey);
        return null;
      });
  _thumbnailInFlight[cacheKey] = future;
  return future;
}

Future<Uint8List?> _loadSourceCached(String url) {
  final cached = _sourceImageCache[url];
  if (cached != null) return Future.value(cached);
  final inFlight = _sourceImageInFlight[url];
  if (inFlight != null) return inFlight;
  final future =
      () async {
            try {
              final response = await http.get(Uri.parse(url));
              if (response.statusCode != 200) return null;
              return response.bodyBytes;
            } catch (_) {
              return null;
            }
          }()
          .then((bytes) {
            if (bytes != null) {
              _sourceImageCache[url] = bytes;
            }
            _sourceImageInFlight.remove(url);
            return bytes;
          })
          .catchError((_) {
            _sourceImageInFlight.remove(url);
            return null;
          });
  _sourceImageInFlight[url] = future;
  return future;
}

Uint8List? peekPoiThumbnail(String? imageUrl) {
  final url = imageUrl != null && imageUrl.isNotEmpty
      ? imageUrl
      : _placeholderUrl;
  return _thumbnailCache['plain_v8|$url'];
}

Uint8List? peekPoiThumbnailWithQuestMarker(String? imageUrl) {
  final url = imageUrl != null && imageUrl.isNotEmpty
      ? imageUrl
      : _placeholderUrl;
  return _thumbnailCache['quest_v8|$url'];
}

Uint8List? peekPoiThumbnailWithMainStoryMarker(String? imageUrl) {
  final url = imageUrl != null && imageUrl.isNotEmpty
      ? imageUrl
      : _placeholderUrl;
  return _thumbnailCache['main_story_v8|$url'];
}

/// Fetches the POI image (or placeholder), resizes to a square, applies
/// rounded corners, and returns PNG bytes suitable for MapLibre addImage.
Future<Uint8List?> loadPoiThumbnail(String? imageUrl) {
  final url = imageUrl != null && imageUrl.isNotEmpty
      ? imageUrl
      : _placeholderUrl;
  final cacheKey = 'plain_v8|$url';
  return _loadThumbnailCached(cacheKey, () async {
    final bytes = await _loadSourceCached(url);
    if (bytes == null) return null;
    final decoded = img.decodeImage(bytes);
    if (decoded == null) return null;
    final square = _buildTransparentRoundedThumbnail(decoded);
    return Uint8List.fromList(img.encodePng(square));
  });
}

Future<Uint8List?> loadBaseDiamondMarker({bool isCurrentUserBase = false}) {
  final cacheKey = isCurrentUserBase
      ? 'base_diamond_marker_self_v6'
      : 'base_diamond_marker_v6';
  return _loadThumbnailCached(cacheKey, () async {
    final image = img.Image(
      width: _thumbnailSize,
      height: _thumbnailSize,
      numChannels: 4,
    );
    img.fill(image, color: img.ColorRgba8(0, 0, 0, 0));

    final outlineColor = isCurrentUserBase
        ? img.ColorRgba8(52, 73, 94, 255)
        : img.ColorRgba8(122, 78, 46, 255);
    final accentColor = isCurrentUserBase
        ? img.ColorRgba8(95, 171, 184, 255)
        : img.ColorRgba8(231, 195, 106, 255);
    final fillColor = isCurrentUserBase
        ? img.ColorRgba8(232, 241, 242, 255)
        : img.ColorRgba8(241, 226, 189, 255);
    final houseColor = outlineColor;

    final center = _thumbnailSize ~/ 2;
    _fillDiamond(image, center, center, 64, outlineColor);
    _fillDiamond(image, center, center, 58, accentColor);
    _fillDiamond(image, center, center, 52, fillColor);
    _drawBaseHouseGlyph(
      image,
      centerX: center,
      roofTop: 56,
      roofBaseY: 95,
      roofHalfWidth: 38,
      bodyLeft: 68,
      bodyTop: 95,
      bodyRight: 124,
      bodyBottom: 136,
      color: houseColor,
      cutoutColor: fillColor,
    );
    if (isCurrentUserBase) {
      _drawCurrentUserBaseBadge(
        image,
        centerX: center,
        centerY: 40,
        outlineColor: outlineColor,
        accentColor: accentColor,
      );
    }

    return Uint8List.fromList(img.encodePng(image));
  });
}

/// Same as [loadPoiThumbnail], but adds a gold border around the image.
/// Useful for quest highlights where we need a visible outline that isn't
/// dependent on map styling support.
Future<Uint8List?> loadPoiThumbnailWithBorder(
  String? imageUrl, {
  int borderWidth = 10,
}) {
  final url = imageUrl != null && imageUrl.isNotEmpty
      ? imageUrl
      : _placeholderUrl;
  final cacheKey = 'border_v7:$borderWidth|$url';
  return _loadThumbnailCached(cacheKey, () async {
    final bytes = await _loadSourceCached(url);
    if (bytes == null) return null;
    final decoded = img.decodeImage(bytes);
    if (decoded == null) return null;
    final square = _buildTransparentRoundedThumbnail(decoded);
    final borderedSize = _thumbnailSize + borderWidth * 2;
    final bordered = img.Image(
      width: borderedSize,
      height: borderedSize,
      numChannels: 4,
    );
    img.fill(bordered, color: img.ColorRgba8(0, 0, 0, 0));
    final gold = img.ColorRgba8(245, 197, 66, 255);
    final max = borderedSize - 1;
    for (var i = 0; i < borderWidth; i++) {
      img.drawRect(
        bordered,
        x1: i,
        y1: i,
        x2: max - i,
        y2: max - i,
        color: gold,
      );
    }
    img.compositeImage(bordered, square, dstX: borderWidth, dstY: borderWidth);
    return Uint8List.fromList(img.encodePng(bordered));
  });
}

/// Same as [loadPoiThumbnail], but keeps a separate cache namespace for
/// quest-available markers.
Future<Uint8List?> loadPoiThumbnailWithQuestMarker(String? imageUrl) {
  final url = imageUrl != null && imageUrl.isNotEmpty
      ? imageUrl
      : _placeholderUrl;
  final cacheKey = 'quest_v8|$url';
  return _loadThumbnailCached(cacheKey, () async {
    final bytes = await _loadSourceCached(url);
    if (bytes == null) return null;
    final decoded = img.decodeImage(bytes);
    if (decoded == null) return null;
    final square = _buildTransparentRoundedThumbnail(decoded);
    return Uint8List.fromList(img.encodePng(square));
  });
}

Future<Uint8List?> loadPoiThumbnailWithMainStoryMarker(String? imageUrl) {
  final url = imageUrl != null && imageUrl.isNotEmpty
      ? imageUrl
      : _placeholderUrl;
  final cacheKey = 'main_story_v8|$url';
  return _loadThumbnailCached(cacheKey, () async {
    final bytes = await _loadSourceCached(url);
    if (bytes == null) return null;
    final decoded = img.decodeImage(bytes);
    if (decoded == null) return null;
    final square = _buildTransparentRoundedThumbnail(decoded);
    _drawMainStoryFrame(square);
    return Uint8List.fromList(img.encodePng(square));
  });
}

img.Image _buildTransparentRoundedThumbnail(img.Image source) {
  final cropped = img.copyResizeCropSquare(
    source,
    size: _thumbnailSize,
    antialias: true,
  );
  final output = img.Image(
    width: _thumbnailSize,
    height: _thumbnailSize,
    numChannels: 4,
  );
  img.fill(output, color: img.ColorRgba8(0, 0, 0, 0));

  for (var y = 0; y < _thumbnailSize; y++) {
    for (var x = 0; x < _thumbnailSize; x++) {
      if (!_isInsideRoundedRect(x, y, _thumbnailSize, _thumbnailSize)) {
        continue;
      }
      output.setPixel(x, y, cropped.getPixel(x, y));
    }
  }

  return output;
}

bool _isInsideRoundedRect(int x, int y, int width, int height) {
  final innerLeft = _cornerRadius;
  final innerRight = width - _cornerRadius - 1;
  final innerTop = _cornerRadius;
  final innerBottom = height - _cornerRadius - 1;

  if ((x >= innerLeft && x <= innerRight) ||
      (y >= innerTop && y <= innerBottom)) {
    return true;
  }

  final cornerCenterX = x < innerLeft ? innerLeft : innerRight;
  final cornerCenterY = y < innerTop ? innerTop : innerBottom;
  final dx = x - cornerCenterX;
  final dy = y - cornerCenterY;
  return dx * dx + dy * dy <= _cornerRadius * _cornerRadius;
}

void _drawMainStoryFrame(img.Image image) {
  final ruby = img.ColorRgba8(130, 16, 28, 255);
  final gold = img.ColorRgba8(255, 219, 125, 255);
  final outer = math.max(4, (_thumbnailSize * 0.028).round());
  final inner = math.max(2, (_thumbnailSize * 0.014).round());
  final max = _thumbnailSize - 1;

  for (var i = 0; i < outer; i++) {
    img.drawRect(image, x1: i, y1: i, x2: max - i, y2: max - i, color: ruby);
  }
  for (var i = 0; i < inner; i++) {
    img.drawRect(
      image,
      x1: outer + i,
      y1: outer + i,
      x2: max - (outer + i),
      y2: max - (outer + i),
      color: gold,
    );
  }
}

void _fillDiamond(
  img.Image image,
  int centerX,
  int centerY,
  int radius,
  img.Color color,
) {
  final minY = math.max(0, centerY - radius);
  final maxY = math.min(image.height - 1, centerY + radius);
  for (var y = minY; y <= maxY; y++) {
    final verticalDistance = (y - centerY).abs();
    final halfWidth = radius - verticalDistance;
    final minX = math.max(0, centerX - halfWidth);
    final maxX = math.min(image.width - 1, centerX + halfWidth);
    for (var x = minX; x <= maxX; x++) {
      image.setPixel(x, y, color);
    }
  }
}

void _drawBaseHouseGlyph(
  img.Image image, {
  required int centerX,
  required int roofTop,
  required int roofBaseY,
  required int roofHalfWidth,
  required int bodyLeft,
  required int bodyTop,
  required int bodyRight,
  required int bodyBottom,
  required img.Color color,
  required img.Color cutoutColor,
}) {
  img.fillRect(
    image,
    x1: bodyLeft,
    y1: bodyTop,
    x2: bodyRight,
    y2: bodyBottom,
    color: color,
  );
  img.fillRect(
    image,
    x1: bodyRight - 11,
    y1: roofTop + 10,
    x2: bodyRight - 3,
    y2: roofTop + 28,
    color: color,
  );
  for (var y = roofTop; y <= roofBaseY; y++) {
    final progress = (y - roofTop) / (roofBaseY - roofTop);
    final halfWidth = (roofHalfWidth * progress).round();
    final minX = centerX - halfWidth;
    final maxX = centerX + halfWidth;
    img.fillRect(image, x1: minX, y1: y, x2: maxX, y2: y, color: color);
  }

  img.fillRect(
    image,
    x1: centerX - 10,
    y1: bodyBottom - 20,
    x2: centerX + 10,
    y2: bodyBottom,
    color: cutoutColor,
  );
  img.fillRect(
    image,
    x1: bodyLeft + 9,
    y1: bodyTop + 9,
    x2: bodyLeft + 20,
    y2: bodyTop + 20,
    color: cutoutColor,
  );
  img.fillRect(
    image,
    x1: bodyRight - 20,
    y1: bodyTop + 9,
    x2: bodyRight - 9,
    y2: bodyTop + 20,
    color: cutoutColor,
  );
}

void _drawCurrentUserBaseBadge(
  img.Image image, {
  required int centerX,
  required int centerY,
  required img.Color outlineColor,
  required img.Color accentColor,
}) {
  final glowColor = img.ColorRgba8(247, 214, 114, 255);
  final centerFillColor = img.ColorRgba8(255, 245, 216, 255);
  _fillDiamond(image, centerX, centerY, 14, glowColor);
  _fillDiamond(image, centerX, centerY, 10, outlineColor);
  _fillDiamond(image, centerX, centerY, 6, accentColor);
  _fillDiamond(image, centerX, centerY, 3, centerFillColor);
}
