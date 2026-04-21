import 'dart:math' as math;
import 'dart:typed_data';

import 'package:http/http.dart' as http;
import 'package:image/image.dart' as img;

import '../models/point_of_interest.dart';

const _placeholderUrl =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/poi-undiscovered.png';
const _poiCategoryPlaceholderUrlPrefix =
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/poi-marker-category-';
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
  return _thumbnailCache['quest_v9|$url'];
}

Uint8List? peekPoiThumbnailWithMainStoryMarker(String? imageUrl) {
  final url = imageUrl != null && imageUrl.isNotEmpty
      ? imageUrl
      : _placeholderUrl;
  return _thumbnailCache['main_story_v11|$url'];
}

Uint8List? peekPoiCategoryThumbnail(PoiMarkerCategory category) {
  return _thumbnailCache['poi_category_plain_v2|${category.wireValue}'];
}

Uint8List? peekPoiCategoryThumbnailWithQuestMarker(PoiMarkerCategory category) {
  return _thumbnailCache['poi_category_quest_v3|${category.wireValue}'];
}

Uint8List? peekPoiCategoryThumbnailWithMainStoryMarker(
  PoiMarkerCategory category,
) {
  return _thumbnailCache['poi_category_main_story_v5|${category.wireValue}'];
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

Future<Uint8List?> loadPoiCategoryThumbnail(PoiMarkerCategory category) {
  final cacheKey = 'poi_category_plain_v2|${category.wireValue}';
  return _loadThumbnailCached(cacheKey, () async {
    final generated = await _loadGeneratedPoiCategoryThumbnail(category);
    if (generated != null) return generated;
    final image = _buildPoiCategoryThumbnail(category);
    return Uint8List.fromList(img.encodePng(image));
  });
}

Future<Uint8List?> loadPoiCategoryThumbnailWithQuestMarker(
  PoiMarkerCategory category,
) {
  final cacheKey = 'poi_category_quest_v3|${category.wireValue}';
  return _loadThumbnailCached(cacheKey, () async {
    final generated = await _loadGeneratedPoiCategoryThumbnail(category);
    if (generated != null) {
      final decoded = img.decodeImage(generated);
      if (decoded == null) return generated;
      drawQuestAvailabilityBadge(decoded);
      return Uint8List.fromList(img.encodePng(decoded));
    }
    final image = _buildPoiCategoryThumbnail(category);
    drawQuestAvailabilityBadge(image);
    return Uint8List.fromList(img.encodePng(image));
  });
}

Future<Uint8List?> loadPoiCategoryThumbnailWithMainStoryMarker(
  PoiMarkerCategory category,
) {
  final cacheKey = 'poi_category_main_story_v5|${category.wireValue}';
  return _loadThumbnailCached(cacheKey, () async {
    final generated = await _loadGeneratedPoiCategoryThumbnail(category);
    if (generated != null) {
      final decoded = img.decodeImage(generated);
      if (decoded == null) return generated;
      drawMainStoryCrest(decoded);
      return Uint8List.fromList(img.encodePng(decoded));
    }
    final image = _buildPoiCategoryThumbnail(category);
    drawMainStoryCrest(image);
    return Uint8List.fromList(img.encodePng(image));
  });
}

String _poiCategoryPlaceholderUrl(PoiMarkerCategory category) {
  return '$_poiCategoryPlaceholderUrlPrefix${category.wireValue}.png';
}

Future<Uint8List?> _loadGeneratedPoiCategoryThumbnail(
  PoiMarkerCategory category,
) async {
  final bytes = await _loadSourceCached(_poiCategoryPlaceholderUrl(category));
  if (bytes == null) return null;
  final decoded = img.decodeImage(bytes);
  if (decoded == null) return null;
  final square = img.copyResizeCropSquare(
    decoded,
    size: _thumbnailSize,
    antialias: true,
  );
  return Uint8List.fromList(img.encodePng(square));
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

Uint8List? peekPlayerPresenceMarker(
  String? imageUrl, {
  bool usePortrait = true,
}) {
  final normalizedUrl = usePortrait && imageUrl != null && imageUrl.isNotEmpty
      ? imageUrl
      : '__fallback__';
  return _thumbnailCache['player_presence_v2|$normalizedUrl'];
}

Future<Uint8List?> loadPlayerPresenceMarker(
  String? imageUrl, {
  bool usePortrait = true,
}) {
  final normalizedUrl = usePortrait && imageUrl != null && imageUrl.isNotEmpty
      ? imageUrl
      : '__fallback__';
  final cacheKey = 'player_presence_v2|$normalizedUrl';
  return _loadThumbnailCached(cacheKey, () async {
    img.Image? portrait;
    if (usePortrait && imageUrl != null && imageUrl.isNotEmpty) {
      final bytes = await _loadSourceCached(imageUrl);
      if (bytes != null) {
        final decoded = img.decodeImage(bytes);
        if (decoded != null) {
          portrait = img.copyResizeCropSquare(
            decoded,
            size: 136,
            antialias: true,
          );
        }
      }
    }

    final marker = _buildPlayerPresenceMarker(portrait: portrait);
    return Uint8List.fromList(img.encodePng(marker));
  });
}

img.Image _buildPlayerPresenceMarker({img.Image? portrait}) {
  final image = img.Image(
    width: _thumbnailSize,
    height: _thumbnailSize,
    numChannels: 4,
  );
  img.fill(image, color: img.ColorRgba8(0, 0, 0, 0));

  final shadowColor = img.ColorRgba8(9, 16, 24, 72);
  final outlineColor = img.ColorRgba8(33, 46, 61, 255);
  final accentColor = img.ColorRgba8(223, 191, 113, 255);
  final portraitRingColor = img.ColorRgba8(98, 160, 176, 255);
  final highlightColor = img.ColorRgba8(255, 249, 237, 255);
  final panelColor = img.ColorRgba8(242, 232, 208, 255);
  final ribbonColor = img.ColorRgba8(70, 123, 139, 255);

  _fillCircle(image, 96, 98, 72, shadowColor);
  _fillTriangle(
    image,
    x1: 72,
    y1: 120,
    x2: 120,
    y2: 120,
    x3: 96,
    y3: 178,
    color: shadowColor,
  );

  _fillTriangle(
    image,
    x1: 68,
    y1: 116,
    x2: 124,
    y2: 116,
    x3: 96,
    y3: 176,
    color: outlineColor,
  );
  _fillTriangle(
    image,
    x1: 74,
    y1: 120,
    x2: 118,
    y2: 120,
    x3: 96,
    y3: 168,
    color: accentColor,
  );
  _fillTriangle(
    image,
    x1: 79,
    y1: 124,
    x2: 113,
    y2: 124,
    x3: 96,
    y3: 160,
    color: ribbonColor,
  );

  _fillCircle(image, 96, 80, 64, outlineColor);
  _fillCircle(image, 96, 80, 58, accentColor);
  _fillCircle(image, 96, 80, 52, portraitRingColor);
  _fillCircle(image, 96, 80, 47, highlightColor);

  if (portrait != null) {
    _stampCircularImage(image, portrait, centerX: 96, centerY: 80, radius: 44);
  } else {
    _fillCircle(image, 96, 80, 44, panelColor);
    _drawFallbackPlayerGlyph(
      image,
      centerX: 96,
      centerY: 82,
      outlineColor: outlineColor,
      accentColor: accentColor,
      ribbonColor: ribbonColor,
      highlightColor: highlightColor,
    );
  }

  _fillRect(image, x1: 68, y1: 122, x2: 124, y2: 138, color: outlineColor);
  _fillRect(image, x1: 72, y1: 126, x2: 120, y2: 134, color: accentColor);
  _fillRect(image, x1: 80, y1: 128, x2: 112, y2: 132, color: highlightColor);

  return image;
}

img.Image _buildPoiCategoryThumbnail(PoiMarkerCategory category) {
  final palette = _poiMarkerPalette(category);
  final image = img.Image(
    width: _thumbnailSize,
    height: _thumbnailSize,
    numChannels: 4,
  );
  img.fill(image, color: img.ColorRgba8(0, 0, 0, 0));

  _fillRoundedRect(image, palette.background);
  _fillRoundedInset(image, inset: 10, color: palette.frame);
  _fillRoundedInset(image, inset: 18, color: palette.panel);
  _fillRect(image, x1: 28, y1: 26, x2: 164, y2: 46, color: palette.ribbon);
  _drawPoiCategoryGlyph(image, category, palette);
  return image;
}

void _drawPoiCategoryGlyph(
  img.Image image,
  PoiMarkerCategory category,
  _PoiMarkerPalette palette,
) {
  switch (category) {
    case PoiMarkerCategory.generic:
      _drawGenericPoiGlyph(image, palette);
      return;
    case PoiMarkerCategory.coffeehouse:
      _drawCoffeehousePoiGlyph(image, palette);
      return;
    case PoiMarkerCategory.tavern:
      _drawTavernPoiGlyph(image, palette);
      return;
    case PoiMarkerCategory.eatery:
      _drawEateryPoiGlyph(image, palette);
      return;
    case PoiMarkerCategory.market:
      _drawMarketPoiGlyph(image, palette);
      return;
    case PoiMarkerCategory.archive:
      _drawArchivePoiGlyph(image, palette);
      return;
    case PoiMarkerCategory.park:
      _drawParkPoiGlyph(image, palette);
      return;
    case PoiMarkerCategory.waterfront:
      _drawWaterfrontPoiGlyph(image, palette);
      return;
    case PoiMarkerCategory.museum:
      _drawMuseumPoiGlyph(image, palette);
      return;
    case PoiMarkerCategory.theater:
      _drawTheaterPoiGlyph(image, palette);
      return;
    case PoiMarkerCategory.landmark:
      _drawLandmarkPoiGlyph(image, palette);
      return;
    case PoiMarkerCategory.civic:
      _drawCivicPoiGlyph(image, palette);
      return;
    case PoiMarkerCategory.arena:
      _drawArenaPoiGlyph(image, palette);
      return;
  }
}

class _PoiMarkerPalette {
  const _PoiMarkerPalette({
    required this.background,
    required this.frame,
    required this.panel,
    required this.ribbon,
    required this.ink,
    required this.highlight,
  });

  final img.Color background;
  final img.Color frame;
  final img.Color panel;
  final img.Color ribbon;
  final img.Color ink;
  final img.Color highlight;
}

_PoiMarkerPalette _poiMarkerPalette(PoiMarkerCategory category) {
  switch (category) {
    case PoiMarkerCategory.coffeehouse:
      return _PoiMarkerPalette(
        background: img.ColorRgba8(92, 61, 46, 255),
        frame: img.ColorRgba8(214, 170, 111, 255),
        panel: img.ColorRgba8(242, 225, 198, 255),
        ribbon: img.ColorRgba8(166, 98, 59, 255),
        ink: img.ColorRgba8(82, 47, 34, 255),
        highlight: img.ColorRgba8(255, 247, 228, 255),
      );
    case PoiMarkerCategory.tavern:
      return _PoiMarkerPalette(
        background: img.ColorRgba8(102, 35, 30, 255),
        frame: img.ColorRgba8(221, 175, 83, 255),
        panel: img.ColorRgba8(245, 222, 192, 255),
        ribbon: img.ColorRgba8(133, 59, 40, 255),
        ink: img.ColorRgba8(83, 34, 28, 255),
        highlight: img.ColorRgba8(255, 242, 207, 255),
      );
    case PoiMarkerCategory.eatery:
      return _PoiMarkerPalette(
        background: img.ColorRgba8(123, 73, 40, 255),
        frame: img.ColorRgba8(231, 188, 103, 255),
        panel: img.ColorRgba8(247, 232, 203, 255),
        ribbon: img.ColorRgba8(187, 104, 56, 255),
        ink: img.ColorRgba8(86, 48, 27, 255),
        highlight: img.ColorRgba8(255, 247, 226, 255),
      );
    case PoiMarkerCategory.market:
      return _PoiMarkerPalette(
        background: img.ColorRgba8(51, 96, 92, 255),
        frame: img.ColorRgba8(227, 194, 116, 255),
        panel: img.ColorRgba8(232, 240, 229, 255),
        ribbon: img.ColorRgba8(66, 128, 121, 255),
        ink: img.ColorRgba8(36, 72, 69, 255),
        highlight: img.ColorRgba8(250, 252, 236, 255),
      );
    case PoiMarkerCategory.archive:
      return _PoiMarkerPalette(
        background: img.ColorRgba8(52, 66, 112, 255),
        frame: img.ColorRgba8(223, 195, 129, 255),
        panel: img.ColorRgba8(232, 228, 214, 255),
        ribbon: img.ColorRgba8(78, 93, 146, 255),
        ink: img.ColorRgba8(40, 49, 84, 255),
        highlight: img.ColorRgba8(251, 247, 230, 255),
      );
    case PoiMarkerCategory.park:
      return _PoiMarkerPalette(
        background: img.ColorRgba8(56, 113, 63, 255),
        frame: img.ColorRgba8(218, 189, 118, 255),
        panel: img.ColorRgba8(230, 241, 214, 255),
        ribbon: img.ColorRgba8(84, 143, 74, 255),
        ink: img.ColorRgba8(39, 76, 43, 255),
        highlight: img.ColorRgba8(245, 252, 231, 255),
      );
    case PoiMarkerCategory.waterfront:
      return _PoiMarkerPalette(
        background: img.ColorRgba8(39, 97, 128, 255),
        frame: img.ColorRgba8(216, 195, 135, 255),
        panel: img.ColorRgba8(217, 238, 245, 255),
        ribbon: img.ColorRgba8(77, 148, 189, 255),
        ink: img.ColorRgba8(29, 69, 93, 255),
        highlight: img.ColorRgba8(246, 252, 255, 255),
      );
    case PoiMarkerCategory.museum:
      return _PoiMarkerPalette(
        background: img.ColorRgba8(93, 76, 119, 255),
        frame: img.ColorRgba8(224, 194, 136, 255),
        panel: img.ColorRgba8(235, 226, 236, 255),
        ribbon: img.ColorRgba8(126, 104, 150, 255),
        ink: img.ColorRgba8(70, 54, 92, 255),
        highlight: img.ColorRgba8(251, 245, 252, 255),
      );
    case PoiMarkerCategory.theater:
      return _PoiMarkerPalette(
        background: img.ColorRgba8(111, 41, 63, 255),
        frame: img.ColorRgba8(227, 193, 109, 255),
        panel: img.ColorRgba8(241, 222, 223, 255),
        ribbon: img.ColorRgba8(150, 61, 84, 255),
        ink: img.ColorRgba8(84, 31, 46, 255),
        highlight: img.ColorRgba8(255, 241, 231, 255),
      );
    case PoiMarkerCategory.landmark:
      return _PoiMarkerPalette(
        background: img.ColorRgba8(112, 84, 45, 255),
        frame: img.ColorRgba8(232, 201, 127, 255),
        panel: img.ColorRgba8(241, 234, 209, 255),
        ribbon: img.ColorRgba8(165, 122, 61, 255),
        ink: img.ColorRgba8(84, 60, 32, 255),
        highlight: img.ColorRgba8(255, 248, 227, 255),
      );
    case PoiMarkerCategory.civic:
      return _PoiMarkerPalette(
        background: img.ColorRgba8(70, 82, 111, 255),
        frame: img.ColorRgba8(220, 194, 127, 255),
        panel: img.ColorRgba8(232, 236, 244, 255),
        ribbon: img.ColorRgba8(98, 112, 147, 255),
        ink: img.ColorRgba8(51, 60, 85, 255),
        highlight: img.ColorRgba8(251, 246, 229, 255),
      );
    case PoiMarkerCategory.arena:
      return _PoiMarkerPalette(
        background: img.ColorRgba8(96, 63, 72, 255),
        frame: img.ColorRgba8(226, 189, 108, 255),
        panel: img.ColorRgba8(236, 226, 221, 255),
        ribbon: img.ColorRgba8(140, 87, 91, 255),
        ink: img.ColorRgba8(73, 47, 54, 255),
        highlight: img.ColorRgba8(255, 243, 225, 255),
      );
    case PoiMarkerCategory.generic:
      return _PoiMarkerPalette(
        background: img.ColorRgba8(87, 83, 73, 255),
        frame: img.ColorRgba8(220, 192, 125, 255),
        panel: img.ColorRgba8(233, 228, 214, 255),
        ribbon: img.ColorRgba8(128, 120, 104, 255),
        ink: img.ColorRgba8(64, 59, 50, 255),
        highlight: img.ColorRgba8(250, 245, 229, 255),
      );
  }
}

void _drawGenericPoiGlyph(img.Image image, _PoiMarkerPalette palette) {
  _fillDiamond(image, 96, 102, 34, palette.ink);
  _fillDiamond(image, 96, 102, 22, palette.highlight);
  _fillCircle(image, 96, 102, 8, palette.ink);
}

void _drawCoffeehousePoiGlyph(img.Image image, _PoiMarkerPalette palette) {
  _fillRect(image, x1: 63, y1: 84, x2: 123, y2: 127, color: palette.ink);
  _fillRect(image, x1: 71, y1: 91, x2: 115, y2: 120, color: palette.highlight);
  _fillRect(image, x1: 55, y1: 80, x2: 131, y2: 88, color: palette.ink);
  _fillRect(image, x1: 78, y1: 129, x2: 108, y2: 136, color: palette.ink);
  _fillRect(image, x1: 106, y1: 136, x2: 134, y2: 142, color: palette.ink);
  _fillRect(image, x1: 126, y1: 93, x2: 133, y2: 118, color: palette.ink);
  _fillRect(image, x1: 134, y1: 98, x2: 139, y2: 113, color: palette.ink);
  _fillRect(image, x1: 74, y1: 62, x2: 79, y2: 81, color: palette.ink);
  _fillRect(image, x1: 92, y1: 56, x2: 97, y2: 81, color: palette.ink);
  _fillRect(image, x1: 110, y1: 62, x2: 115, y2: 81, color: palette.ink);
}

void _drawTavernPoiGlyph(img.Image image, _PoiMarkerPalette palette) {
  _fillRect(image, x1: 62, y1: 86, x2: 122, y2: 131, color: palette.ink);
  _fillRect(image, x1: 70, y1: 95, x2: 114, y2: 124, color: palette.highlight);
  _fillRect(image, x1: 78, y1: 132, x2: 106, y2: 140, color: palette.ink);
  _fillRect(image, x1: 60, y1: 78, x2: 124, y2: 85, color: palette.ink);
  _fillCircle(image, 73, 78, 10, palette.highlight);
  _fillCircle(image, 93, 74, 12, palette.highlight);
  _fillCircle(image, 112, 78, 10, palette.highlight);
  _fillRect(image, x1: 123, y1: 95, x2: 130, y2: 122, color: palette.ink);
  _fillRect(image, x1: 130, y1: 101, x2: 136, y2: 116, color: palette.ink);
}

void _drawEateryPoiGlyph(img.Image image, _PoiMarkerPalette palette) {
  _fillCircle(image, 96, 106, 34, palette.ink);
  _fillCircle(image, 96, 106, 25, palette.highlight);
  _fillRect(image, x1: 54, y1: 82, x2: 59, y2: 134, color: palette.ink);
  _fillRect(image, x1: 48, y1: 82, x2: 65, y2: 88, color: palette.ink);
  _fillRect(image, x1: 48, y1: 100, x2: 65, y2: 106, color: palette.ink);
  _fillRect(image, x1: 48, y1: 118, x2: 65, y2: 124, color: palette.ink);
  _fillRect(image, x1: 131, y1: 80, x2: 136, y2: 134, color: palette.ink);
  _fillRect(image, x1: 124, y1: 80, x2: 144, y2: 89, color: palette.ink);
}

void _drawMarketPoiGlyph(img.Image image, _PoiMarkerPalette palette) {
  _fillRect(image, x1: 58, y1: 95, x2: 134, y2: 136, color: palette.ink);
  _fillRect(image, x1: 65, y1: 102, x2: 127, y2: 129, color: palette.highlight);
  _fillRect(image, x1: 52, y1: 76, x2: 140, y2: 94, color: palette.ink);
  for (var i = 0; i < 4; i++) {
    final start = 58 + (i * 20);
    _fillRect(
      image,
      x1: start,
      y1: 80,
      x2: start + 11,
      y2: 94,
      color: i.isEven ? palette.highlight : palette.ribbon,
    );
  }
  _fillRect(image, x1: 70, y1: 112, x2: 84, y2: 129, color: palette.ink);
  _fillRect(image, x1: 90, y1: 112, x2: 104, y2: 129, color: palette.ink);
  _fillRect(image, x1: 110, y1: 112, x2: 124, y2: 129, color: palette.ink);
}

void _drawArchivePoiGlyph(img.Image image, _PoiMarkerPalette palette) {
  _fillRect(image, x1: 60, y1: 77, x2: 131, y2: 136, color: palette.ink);
  _fillRect(image, x1: 66, y1: 83, x2: 90, y2: 130, color: palette.highlight);
  _fillRect(image, x1: 101, y1: 83, x2: 125, y2: 130, color: palette.highlight);
  _fillRect(image, x1: 93, y1: 83, x2: 98, y2: 130, color: palette.ribbon);
  _fillRect(image, x1: 78, y1: 90, x2: 82, y2: 124, color: palette.ink);
  _fillRect(image, x1: 109, y1: 90, x2: 113, y2: 124, color: palette.ink);
}

void _drawParkPoiGlyph(img.Image image, _PoiMarkerPalette palette) {
  _fillRect(image, x1: 90, y1: 103, x2: 101, y2: 140, color: palette.ink);
  _fillCircle(image, 96, 80, 26, palette.ink);
  _fillCircle(image, 76, 95, 20, palette.ink);
  _fillCircle(image, 116, 95, 20, palette.ink);
  _fillCircle(image, 96, 87, 18, palette.highlight);
  _fillRect(image, x1: 56, y1: 141, x2: 136, y2: 147, color: palette.ink);
}

void _drawWaterfrontPoiGlyph(img.Image image, _PoiMarkerPalette palette) {
  _fillRect(image, x1: 60, y1: 78, x2: 65, y2: 134, color: palette.ink);
  _fillRect(image, x1: 84, y1: 86, x2: 89, y2: 134, color: palette.ink);
  _fillRect(image, x1: 108, y1: 94, x2: 113, y2: 134, color: palette.ink);
  _fillRect(image, x1: 60, y1: 74, x2: 126, y2: 80, color: palette.ink);
  _fillRect(image, x1: 56, y1: 112, x2: 136, y2: 118, color: palette.ribbon);
  _fillRect(image, x1: 66, y1: 122, x2: 146, y2: 128, color: palette.ink);
  _fillRect(image, x1: 56, y1: 132, x2: 136, y2: 138, color: palette.ribbon);
  _fillCircle(image, 132, 74, 11, palette.highlight);
}

void _drawMuseumPoiGlyph(img.Image image, _PoiMarkerPalette palette) {
  _fillTriangle(
    image,
    x1: 54,
    y1: 88,
    x2: 138,
    y2: 88,
    x3: 96,
    y3: 56,
    color: palette.ink,
  );
  _fillRect(image, x1: 56, y1: 88, x2: 136, y2: 96, color: palette.ink);
  for (var i = 0; i < 4; i++) {
    final left = 62 + (i * 18);
    _fillRect(
      image,
      x1: left,
      y1: 96,
      x2: left + 10,
      y2: 134,
      color: palette.ink,
    );
    _fillRect(
      image,
      x1: left + 2,
      y1: 102,
      x2: left + 8,
      y2: 134,
      color: palette.highlight,
    );
  }
  _fillRect(image, x1: 52, y1: 134, x2: 140, y2: 142, color: palette.ink);
}

void _drawTheaterPoiGlyph(img.Image image, _PoiMarkerPalette palette) {
  _fillRect(image, x1: 58, y1: 66, x2: 133, y2: 139, color: palette.ink);
  _fillRect(image, x1: 66, y1: 74, x2: 125, y2: 131, color: palette.highlight);
  _fillRect(image, x1: 58, y1: 66, x2: 70, y2: 139, color: palette.ribbon);
  _fillRect(image, x1: 121, y1: 66, x2: 133, y2: 139, color: palette.ribbon);
  _fillCircle(image, 96, 96, 16, palette.ink);
  _fillCircle(image, 90, 94, 3, palette.highlight);
  _fillCircle(image, 102, 94, 3, palette.highlight);
  _fillRect(image, x1: 84, y1: 108, x2: 108, y2: 114, color: palette.highlight);
}

void _drawLandmarkPoiGlyph(img.Image image, _PoiMarkerPalette palette) {
  _fillRect(image, x1: 84, y1: 83, x2: 107, y2: 134, color: palette.ink);
  _fillTriangle(
    image,
    x1: 80,
    y1: 83,
    x2: 111,
    y2: 83,
    x3: 96,
    y3: 50,
    color: palette.ink,
  );
  _fillRect(image, x1: 74, y1: 134, x2: 117, y2: 144, color: palette.ink);
  _fillCircle(image, 126, 66, 10, palette.highlight);
}

void _drawCivicPoiGlyph(img.Image image, _PoiMarkerPalette palette) {
  _fillRect(image, x1: 58, y1: 82, x2: 133, y2: 132, color: palette.ink);
  _fillRect(image, x1: 64, y1: 88, x2: 127, y2: 126, color: palette.highlight);
  _fillTriangle(
    image,
    x1: 64,
    y1: 88,
    x2: 96,
    y2: 112,
    x3: 127,
    y3: 88,
    color: palette.ink,
  );
  _fillTriangle(
    image,
    x1: 64,
    y1: 126,
    x2: 96,
    y2: 102,
    x3: 127,
    y3: 126,
    color: palette.ink,
  );
  _fillCircle(image, 125, 128, 11, palette.ribbon);
}

void _drawArenaPoiGlyph(img.Image image, _PoiMarkerPalette palette) {
  _fillRect(image, x1: 54, y1: 101, x2: 138, y2: 138, color: palette.ink);
  _fillRect(image, x1: 62, y1: 109, x2: 130, y2: 130, color: palette.highlight);
  for (var i = 0; i < 3; i++) {
    final left = 70 + (i * 20);
    _fillRect(
      image,
      x1: left,
      y1: 114,
      x2: left + 8,
      y2: 130,
      color: palette.ink,
    );
  }
  _fillTriangle(
    image,
    x1: 90,
    y1: 52,
    x2: 114,
    y2: 52,
    x3: 102,
    y3: 80,
    color: palette.ribbon,
  );
  _fillRect(image, x1: 99, y1: 80, x2: 104, y2: 101, color: palette.ink);
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
  final cacheKey = 'quest_v9|$url';
  return _loadThumbnailCached(cacheKey, () async {
    final bytes = await _loadSourceCached(url);
    if (bytes == null) return null;
    final decoded = img.decodeImage(bytes);
    if (decoded == null) return null;
    final square = _buildTransparentRoundedThumbnail(decoded);
    drawQuestAvailabilityBadge(square);
    return Uint8List.fromList(img.encodePng(square));
  });
}

Future<Uint8List?> loadPoiThumbnailWithMainStoryMarker(String? imageUrl) {
  final url = imageUrl != null && imageUrl.isNotEmpty
      ? imageUrl
      : _placeholderUrl;
  final cacheKey = 'main_story_v11|$url';
  return _loadThumbnailCached(cacheKey, () async {
    final bytes = await _loadSourceCached(url);
    if (bytes == null) return null;
    final decoded = img.decodeImage(bytes);
    if (decoded == null) return null;
    final square = _buildTransparentRoundedThumbnail(decoded);
    drawMainStoryCrest(square);
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

void drawQuestAvailabilityBadge(img.Image image) {
  final bounds = _findOpaqueBounds(image) ?? _fallbackOpaqueBounds(image);
  final outerRadius = math.max(13, math.min(18, (bounds.width * 0.11).round()));
  final centerX = (bounds.right + outerRadius - 2)
      .clamp(outerRadius + 4, image.width - outerRadius - 5)
      .toInt();
  final centerY = (bounds.top + outerRadius - 2)
      .clamp(outerRadius + 4, image.height - outerRadius - 5)
      .toInt();

  final shadow = img.ColorRgba8(28, 18, 8, 110);
  final ring = img.ColorRgba8(255, 250, 240, 255);
  final gold = img.ColorRgba8(245, 197, 66, 255);
  final highlight = img.ColorRgba8(255, 232, 156, 255);
  final ink = img.ColorRgba8(58, 36, 0, 255);
  final stemHalfWidth = math.max(2, outerRadius ~/ 4).toInt();
  final stemTop = (centerY - outerRadius + math.max(4, outerRadius ~/ 3))
      .toInt();
  final stemBottom = (centerY + outerRadius - math.max(8, outerRadius ~/ 2))
      .toInt();
  final dotRadius = math.max(2, outerRadius ~/ 5).toInt();

  _fillCircle(image, centerX + 2, centerY + 3, outerRadius + 2, shadow);
  _fillCircle(image, centerX, centerY, outerRadius + 2, ring);
  _fillCircle(image, centerX, centerY, outerRadius, gold);
  _fillCircle(
    image,
    centerX - math.max(3, outerRadius ~/ 3),
    centerY - math.max(3, outerRadius ~/ 3),
    math.max(3, outerRadius ~/ 3),
    highlight,
  );
  _fillRect(
    image,
    x1: centerX - stemHalfWidth,
    y1: stemTop,
    x2: centerX + stemHalfWidth,
    y2: stemBottom,
    color: ink,
  );
  _fillCircle(
    image,
    centerX,
    centerY + outerRadius - math.max(4, outerRadius ~/ 3),
    dotRadius,
    ink,
  );
}

void drawMainStoryCrest(img.Image image) {
  final bounds = _findOpaqueBounds(image) ?? _fallbackOpaqueBounds(image);
  final outline = img.ColorRgba8(87, 24, 35, 255);
  final ruby = img.ColorRgba8(130, 16, 28, 255);
  final gold = img.ColorRgba8(255, 219, 125, 255);
  final outerRadius = math.max(14, math.min(21, (bounds.width * 0.12).round()));
  final ringInset = math.max(2, outerRadius ~/ 4);
  final sealInset = math.max(4, outerRadius ~/ 3);
  final overlap = math.max(4, outerRadius ~/ 3);
  final centerX = bounds.centerX.round().clamp(
    outerRadius + 4,
    image.width - outerRadius - 5,
  );
  final centerY = (bounds.top - outerRadius + overlap)
      .clamp(outerRadius + 4, image.height - outerRadius - 5)
      .toInt();
  final sealBottom = centerY + outerRadius;
  final anchorBottom = math.min(
    image.height - 1,
    math.max(bounds.top + (overlap ~/ 2), sealBottom),
  );
  final connectorHalfWidth = math.max(4, outerRadius ~/ 3);

  _fillTriangle(
    image,
    x1: centerX - connectorHalfWidth,
    y1: sealBottom - 1,
    x2: centerX + connectorHalfWidth,
    y2: sealBottom - 1,
    x3: centerX,
    y3: anchorBottom,
    color: outline,
  );
  _fillTriangle(
    image,
    x1: centerX - (connectorHalfWidth - 1),
    y1: sealBottom,
    x2: centerX + (connectorHalfWidth - 1),
    y2: sealBottom,
    x3: centerX,
    y3: math.max(sealBottom + 1, anchorBottom - 1),
    color: gold,
  );
  _fillTriangle(
    image,
    x1: centerX - math.max(2, connectorHalfWidth - 3),
    y1: sealBottom + 1,
    x2: centerX + math.max(2, connectorHalfWidth - 3),
    y2: sealBottom + 1,
    x3: centerX,
    y3: math.max(sealBottom + 2, anchorBottom - 2),
    color: ruby,
  );

  _fillCircle(image, centerX, centerY, outerRadius, outline);
  _fillCircle(image, centerX, centerY, outerRadius - 2, gold);
  _fillCircle(image, centerX, centerY, outerRadius - ringInset, ruby);
  _fillDiamond(
    image,
    centerX,
    centerY + 1,
    math.max(4, outerRadius - sealInset),
    gold,
  );
  _fillDiamond(
    image,
    centerX,
    centerY + 1,
    math.max(2, outerRadius - sealInset - 4),
    ruby,
  );
  _fillCircle(image, centerX, centerY + 1, math.max(2, outerRadius ~/ 5), gold);
}

class _OpaqueBounds {
  const _OpaqueBounds({
    required this.left,
    required this.top,
    required this.right,
    required this.bottom,
  });

  final int left;
  final int top;
  final int right;
  final int bottom;

  int get width => right - left + 1;
  double get centerX => (left + right) / 2;
}

_OpaqueBounds? _findOpaqueBounds(img.Image image, {int alphaThreshold = 24}) {
  var left = image.width;
  var top = image.height;
  var right = -1;
  var bottom = -1;

  for (var y = 0; y < image.height; y++) {
    for (var x = 0; x < image.width; x++) {
      if (_pixelAlpha(image, x, y) <= alphaThreshold) {
        continue;
      }
      if (x < left) left = x;
      if (x > right) right = x;
      if (y < top) top = y;
      if (y > bottom) bottom = y;
    }
  }

  if (right < left || bottom < top) {
    return null;
  }
  return _OpaqueBounds(left: left, top: top, right: right, bottom: bottom);
}

_OpaqueBounds _fallbackOpaqueBounds(img.Image image) {
  final horizontalInset = math.max(20, (image.width * 0.2).round());
  final topInset = math.max(18, (image.height * 0.18).round());
  final bottomInset = math.max(24, (image.height * 0.16).round());
  return _OpaqueBounds(
    left: horizontalInset,
    top: topInset,
    right: image.width - horizontalInset - 1,
    bottom: image.height - bottomInset - 1,
  );
}

int _pixelAlpha(img.Image image, int x, int y) {
  final pixel = image.getPixel(x, y);
  if (image.numChannels < 4) {
    return pixel.maxChannelValue.toInt();
  }
  return pixel.a.round();
}

void _stampCircularImage(
  img.Image destination,
  img.Image source, {
  required int centerX,
  required int centerY,
  required int radius,
}) {
  final diameter = radius * 2;
  final portrait = source.width == diameter && source.height == diameter
      ? source
      : img.copyResizeCropSquare(source, size: diameter, antialias: true);
  final top = centerY - radius;
  final left = centerX - radius;
  final radiusSquared = radius * radius;

  for (var y = 0; y < diameter; y++) {
    for (var x = 0; x < diameter; x++) {
      final dx = x - radius;
      final dy = y - radius;
      if ((dx * dx) + (dy * dy) > radiusSquared) continue;
      final dstX = left + x;
      final dstY = top + y;
      if (dstX < 0 ||
          dstX >= destination.width ||
          dstY < 0 ||
          dstY >= destination.height) {
        continue;
      }
      if (_pixelAlpha(portrait, x, y) <= 0) continue;
      destination.setPixel(dstX, dstY, portrait.getPixel(x, y));
    }
  }
}

void _drawFallbackPlayerGlyph(
  img.Image image, {
  required int centerX,
  required int centerY,
  required img.Color outlineColor,
  required img.Color accentColor,
  required img.Color ribbonColor,
  required img.Color highlightColor,
}) {
  _fillDiamond(image, centerX, centerY - 3, 24, accentColor);
  _fillDiamond(image, centerX, centerY - 3, 16, highlightColor);
  _fillCircle(image, centerX, centerY - 18, 10, outlineColor);
  _fillCircle(image, centerX, centerY - 18, 6, highlightColor);
  _fillTriangle(
    image,
    x1: centerX - 18,
    y1: centerY + 20,
    x2: centerX + 18,
    y2: centerY + 20,
    x3: centerX,
    y3: centerY - 2,
    color: ribbonColor,
  );
  _fillTriangle(
    image,
    x1: centerX - 12,
    y1: centerY + 16,
    x2: centerX + 12,
    y2: centerY + 16,
    x3: centerX,
    y3: centerY + 2,
    color: outlineColor,
  );
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

void _fillRoundedRect(img.Image image, img.Color color) {
  for (var y = 0; y < image.height; y++) {
    for (var x = 0; x < image.width; x++) {
      if (_isInsideRoundedRect(x, y, image.width, image.height)) {
        image.setPixel(x, y, color);
      }
    }
  }
}

void _fillRoundedInset(
  img.Image image, {
  required int inset,
  required img.Color color,
}) {
  final width = image.width - (inset * 2);
  final height = image.height - (inset * 2);
  for (var y = inset; y < image.height - inset; y++) {
    for (var x = inset; x < image.width - inset; x++) {
      if (_isInsideRoundedRect(x - inset, y - inset, width, height)) {
        image.setPixel(x, y, color);
      }
    }
  }
}

void _fillRect(
  img.Image image, {
  required int x1,
  required int y1,
  required int x2,
  required int y2,
  required img.Color color,
}) {
  img.fillRect(image, x1: x1, y1: y1, x2: x2, y2: y2, color: color);
}

void _fillCircle(
  img.Image image,
  int centerX,
  int centerY,
  int radius,
  img.Color color,
) {
  final minY = math.max(0, centerY - radius);
  final maxY = math.min(image.height - 1, centerY + radius);
  final radiusSquared = radius * radius;
  for (var y = minY; y <= maxY; y++) {
    final dy = y - centerY;
    final maxDx = math
        .sqrt((radiusSquared - (dy * dy)).clamp(0, radiusSquared))
        .floor();
    final minX = math.max(0, centerX - maxDx);
    final maxX = math.min(image.width - 1, centerX + maxDx);
    for (var x = minX; x <= maxX; x++) {
      image.setPixel(x, y, color);
    }
  }
}

void _fillTriangle(
  img.Image image, {
  required int x1,
  required int y1,
  required int x2,
  required int y2,
  required int x3,
  required int y3,
  required img.Color color,
}) {
  final minX = math.max(0, math.min(x1, math.min(x2, x3)));
  final maxX = math.min(image.width - 1, math.max(x1, math.max(x2, x3)));
  final minY = math.max(0, math.min(y1, math.min(y2, y3)));
  final maxY = math.min(image.height - 1, math.max(y1, math.max(y2, y3)));

  final area = ((x2 - x1) * (y3 - y1)) - ((x3 - x1) * (y2 - y1));
  if (area == 0) return;

  for (var y = minY; y <= maxY; y++) {
    for (var x = minX; x <= maxX; x++) {
      final w1 = ((x2 - x1) * (y - y1)) - ((y2 - y1) * (x - x1));
      final w2 = ((x3 - x2) * (y - y2)) - ((y3 - y2) * (x - x2));
      final w3 = ((x1 - x3) * (y - y3)) - ((y1 - y3) * (x - x3));
      final hasNegative = w1 < 0 || w2 < 0 || w3 < 0;
      final hasPositive = w1 > 0 || w2 > 0 || w3 > 0;
      if (!(hasNegative && hasPositive)) {
        image.setPixel(x, y, color);
      }
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
