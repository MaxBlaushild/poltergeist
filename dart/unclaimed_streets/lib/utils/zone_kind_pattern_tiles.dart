import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:image/image.dart' as img;

import '../constants/zone_kind_visuals.dart';

const _zoneKindPatternVersion = 'v3';
const _zoneKindPatternTileSize = 32;

final Map<String, Uint8List> _zoneKindPatternTileCache = <String, Uint8List>{};

String zoneKindPatternImageId(String? rawKind) {
  final style = _zonePatternStyleForSlug(normalizeZoneKindSlug(rawKind));
  return 'zone_kind_pattern_${style.cacheKey}_$_zoneKindPatternVersion';
}

Uint8List zoneKindPatternTileBytes(String? rawKind) {
  final style = _zonePatternStyleForSlug(normalizeZoneKindSlug(rawKind));
  return _zoneKindPatternTileCache.putIfAbsent(style.cacheKey, () {
    final profile = zoneKindVisualProfileForSlug(style.canonicalSlug);
    final palette = _ZonePatternPalette.fromProfile(profile);
    final image = img.Image(
      width: _zoneKindPatternTileSize,
      height: _zoneKindPatternTileSize,
      numChannels: 4,
    );
    img.fill(image, color: img.ColorRgba8(0, 0, 0, 0));

    switch (style) {
      case _ZonePatternStyle.frontier:
        _drawFrontierPattern(image, palette);
        break;
      case _ZonePatternStyle.city:
        _drawCityPattern(image, palette);
        break;
      case _ZonePatternStyle.village:
        _drawVillagePattern(image, palette);
        break;
      case _ZonePatternStyle.academy:
        _drawAcademyPattern(image, palette);
        break;
      case _ZonePatternStyle.forest:
        _drawForestPattern(image, palette);
        break;
      case _ZonePatternStyle.swamp:
        _drawSwampPattern(image, palette);
        break;
      case _ZonePatternStyle.badlands:
        _drawBadlandsPattern(image, palette);
        break;
      case _ZonePatternStyle.farmland:
        _drawFarmlandPattern(image, palette);
        break;
      case _ZonePatternStyle.highlands:
        _drawHighlandsPattern(image, palette);
        break;
      case _ZonePatternStyle.mountain:
        _drawMountainPattern(image, palette);
        break;
      case _ZonePatternStyle.ruins:
        _drawRuinsPattern(image, palette);
        break;
      case _ZonePatternStyle.graveyard:
        _drawGraveyardPattern(image, palette);
        break;
      case _ZonePatternStyle.industrial:
        _drawIndustrialPattern(image, palette);
        break;
      case _ZonePatternStyle.desert:
        _drawDesertPattern(image, palette);
        break;
      case _ZonePatternStyle.templeGrounds:
        _drawTemplePattern(image, palette);
        break;
      case _ZonePatternStyle.volcanic:
        _drawVolcanicPattern(image, palette);
        break;
    }

    return Uint8List.fromList(img.encodePng(image));
  });
}

enum _ZonePatternStyle {
  frontier('frontier', null),
  city('city', 'city'),
  village('village', 'village'),
  academy('academy', 'academy'),
  forest('forest', 'forest'),
  swamp('swamp', 'swamp'),
  badlands('badlands', 'badlands'),
  farmland('farmland', 'farmland'),
  highlands('highlands', 'highlands'),
  mountain('mountain', 'mountain'),
  ruins('ruins', 'ruins'),
  graveyard('graveyard', 'graveyard'),
  industrial('industrial', 'industrial'),
  desert('desert', 'desert'),
  templeGrounds('temple_grounds', 'temple-grounds'),
  volcanic('volcanic', 'volcanic');

  const _ZonePatternStyle(this.cacheKey, this.canonicalSlug);

  final String cacheKey;
  final String? canonicalSlug;
}

_ZonePatternStyle _zonePatternStyleForSlug(String normalized) {
  switch (normalized) {
    case 'city':
      return _ZonePatternStyle.city;
    case 'village':
      return _ZonePatternStyle.village;
    case 'academy':
      return _ZonePatternStyle.academy;
    case 'forest':
      return _ZonePatternStyle.forest;
    case 'swamp':
      return _ZonePatternStyle.swamp;
    case 'badlands':
      return _ZonePatternStyle.badlands;
    case 'farmland':
      return _ZonePatternStyle.farmland;
    case 'highlands':
      return _ZonePatternStyle.highlands;
    case 'mountain':
      return _ZonePatternStyle.mountain;
    case 'ruins':
      return _ZonePatternStyle.ruins;
    case 'graveyard':
      return _ZonePatternStyle.graveyard;
    case 'industrial':
      return _ZonePatternStyle.industrial;
    case 'desert':
      return _ZonePatternStyle.desert;
    case 'temple-grounds':
      return _ZonePatternStyle.templeGrounds;
    case 'volcanic':
      return _ZonePatternStyle.volcanic;
    default:
      return _ZonePatternStyle.frontier;
  }
}

class _ZonePatternPalette {
  const _ZonePatternPalette({
    required this.ink,
    required this.highlight,
    required this.shadow,
  });

  factory _ZonePatternPalette.fromProfile(ZoneKindVisualProfile profile) {
    return _ZonePatternPalette(
      ink: _imgColor(
        Color.lerp(profile.panelAccent, Colors.white, 0.14) ??
            profile.panelAccent,
        alphaMultiplier: 1.0,
      ),
      highlight: _imgColor(
        Color.lerp(profile.panelChipText, Colors.white, 0.18) ??
            profile.panelChipText,
        alphaMultiplier: 1.0,
      ),
      shadow: _imgColor(
        Color.lerp(profile.panelAccent, Colors.black, 0.18) ??
            profile.panelAccent,
        alphaMultiplier: 0.86,
      ),
    );
  }

  final img.ColorRgba8 ink;
  final img.ColorRgba8 highlight;
  final img.ColorRgba8 shadow;
}

img.ColorRgba8 _imgColor(Color color, {double alphaMultiplier = 1.0}) {
  final argb = color.toARGB32();
  final alpha = (((argb >> 24) & 0xFF) * alphaMultiplier).round().clamp(0, 255);
  return img.ColorRgba8(
    (argb >> 16) & 0xFF,
    (argb >> 8) & 0xFF,
    argb & 0xFF,
    alpha,
  );
}

void _line(
  img.Image image,
  int x1,
  int y1,
  int x2,
  int y2,
  img.Color color, {
  num thickness = 1,
}) {
  img.drawLine(
    image,
    x1: x1,
    y1: y1,
    x2: x2,
    y2: y2,
    color: color,
    thickness: thickness,
  );
}

void _dot(img.Image image, int x, int y, int radius, img.Color color) {
  img.fillCircle(image, x: x, y: y, radius: radius, color: color);
}

void _rect(
  img.Image image,
  int x1,
  int y1,
  int x2,
  int y2,
  img.Color color, {
  num thickness = 1,
}) {
  img.drawRect(
    image,
    x1: x1,
    y1: y1,
    x2: x2,
    y2: y2,
    color: color,
    thickness: thickness,
  );
}

void _cross(img.Image image, int x, int y, _ZonePatternPalette palette) {
  _line(image, x, y - 2, x, y + 2, palette.ink);
  _line(image, x - 1, y, x + 1, y, palette.highlight);
}

void _polyline(
  img.Image image,
  List<(int, int)> points,
  img.Color color, {
  num thickness = 1,
}) {
  for (var i = 0; i < points.length - 1; i++) {
    _line(
      image,
      points[i].$1,
      points[i].$2,
      points[i + 1].$1,
      points[i + 1].$2,
      color,
      thickness: thickness,
    );
  }
}

void _drawFrontierPattern(img.Image image, _ZonePatternPalette palette) {
  for (var offset = -16; offset <= 32; offset += 12) {
    _line(image, offset, 0, offset + 16, 16, palette.shadow);
    _line(image, offset + 8, 16, offset + 24, 32, palette.ink);
  }
  for (var y = 8; y <= 24; y += 8) {
    for (var x = 8; x <= 24; x += 8) {
      _dot(image, x, y, 1, palette.highlight);
    }
  }
}

void _drawCityPattern(img.Image image, _ZonePatternPalette palette) {
  for (var x = 0; x <= 32; x += 8) {
    _line(image, x, 0, x, 32, x % 16 == 0 ? palette.ink : palette.shadow);
  }
  for (var y = 0; y <= 32; y += 8) {
    _line(image, 0, y, 32, y, y % 16 == 0 ? palette.ink : palette.shadow);
  }
  for (var x = 8; x <= 24; x += 16) {
    for (var y = 8; y <= 24; y += 16) {
      _dot(image, x, y, 1, palette.highlight);
    }
  }
}

void _drawVillagePattern(img.Image image, _ZonePatternPalette palette) {
  _rect(image, 2, 2, 12, 12, palette.shadow);
  _rect(image, 18, 2, 28, 12, palette.shadow);
  _rect(image, 2, 18, 12, 28, palette.shadow);
  _rect(image, 18, 18, 28, 28, palette.shadow);
  _line(image, 15, 0, 15, 32, palette.ink);
  _line(image, 0, 15, 32, 15, palette.ink);
  _dot(image, 15, 15, 1, palette.highlight);
}

void _drawAcademyPattern(img.Image image, _ZonePatternPalette palette) {
  for (var offset = -16; offset <= 32; offset += 16) {
    _line(image, offset, 0, offset + 16, 16, palette.shadow);
    _line(image, offset + 16, 16, offset, 32, palette.shadow);
  }
  _rect(image, 10, 10, 22, 22, palette.ink);
  _dot(image, 16, 16, 2, palette.highlight);
  _dot(image, 0, 0, 1, palette.highlight);
  _dot(image, 32, 0, 1, palette.highlight);
  _dot(image, 0, 32, 1, palette.highlight);
  _dot(image, 32, 32, 1, palette.highlight);
}

void _drawForestPattern(img.Image image, _ZonePatternPalette palette) {
  for (var x = 4; x <= 28; x += 8) {
    _line(image, x, 2, x - 2, 13, palette.shadow, thickness: 2);
    _line(image, x - 1, 6, x + 4, 13, palette.ink, thickness: 2);
    _line(image, x - 1, 6, x - 6, 13, palette.ink, thickness: 2);
  }
  for (var x = 8; x <= 24; x += 8) {
    _dot(image, x, 19, 2, palette.highlight);
    _dot(image, x + 2, 25, 1, palette.highlight);
    _dot(image, x - 3, 24, 1, palette.shadow);
  }
}

void _drawSwampPattern(img.Image image, _ZonePatternPalette palette) {
  for (var y = 6; y <= 26; y += 10) {
    _polyline(image, [
      (0, y),
      (6, y - 1),
      (12, y + 1),
      (18, y),
      (24, y + 2),
      (32, y + 1),
    ], palette.shadow);
  }
  _dot(image, 7, 20, 2, palette.highlight);
  _dot(image, 22, 11, 1, palette.highlight);
  _dot(image, 26, 24, 1, palette.ink);
}

void _drawBadlandsPattern(img.Image image, _ZonePatternPalette palette) {
  _polyline(image, [
    (0, 6),
    (7, 9),
    (12, 4),
    (19, 8),
    (26, 3),
    (32, 6),
  ], palette.shadow);
  _polyline(image, [
    (4, 18),
    (10, 14),
    (16, 19),
    (21, 12),
    (28, 18),
  ], palette.ink);
  _polyline(image, [
    (0, 28),
    (8, 24),
    (14, 28),
    (20, 22),
    (28, 26),
    (32, 24),
  ], palette.shadow);
}

void _drawFarmlandPattern(img.Image image, _ZonePatternPalette palette) {
  for (var x = -8; x <= 32; x += 8) {
    _line(image, x, 0, x + 8, 32, palette.shadow);
  }
  for (var x = -4; x <= 32; x += 16) {
    _line(image, x, 0, x + 8, 32, palette.ink, thickness: 1.5);
  }
  _line(image, 0, 16, 32, 16, palette.highlight);
}

void _drawHighlandsPattern(img.Image image, _ZonePatternPalette palette) {
  for (var y = 6; y <= 26; y += 8) {
    _polyline(image, [
      (0, y),
      (6, y - 1),
      (13, y),
      (20, y - 2),
      (26, y - 1),
      (32, y - 2),
    ], y == 14 ? palette.ink : palette.shadow);
  }
  _dot(image, 10, 10, 1, palette.highlight);
  _dot(image, 22, 22, 1, palette.highlight);
}

void _drawMountainPattern(img.Image image, _ZonePatternPalette palette) {
  for (var x = -2; x <= 30; x += 12) {
    _polyline(
      image,
      [(x, 24), (x + 6, 10), (x + 12, 24)],
      palette.shadow,
      thickness: 1.5,
    );
    _line(image, x + 6, 10, x + 6, 16, palette.highlight);
  }
  _line(image, 0, 24, 32, 24, palette.ink);
}

void _drawRuinsPattern(img.Image image, _ZonePatternPalette palette) {
  _rect(image, 2, 3, 12, 11, palette.shadow);
  _rect(image, 18, 5, 29, 13, palette.shadow);
  _rect(image, 5, 18, 14, 28, palette.shadow);
  _line(image, 7, 3, 9, 3, palette.highlight);
  _line(image, 22, 13, 25, 13, palette.highlight);
  _line(image, 5, 22, 5, 25, palette.highlight);
  _line(image, 20, 20, 30, 30, palette.ink);
}

void _drawGraveyardPattern(img.Image image, _ZonePatternPalette palette) {
  for (final point in const [(8, 9), (23, 9), (8, 24), (23, 24)]) {
    _cross(image, point.$1, point.$2, palette);
  }
  _line(image, 16, 4, 16, 28, palette.shadow);
  _dot(image, 16, 16, 1, palette.highlight);
}

void _drawIndustrialPattern(img.Image image, _ZonePatternPalette palette) {
  for (var x = 0; x <= 32; x += 10) {
    _line(image, x, 0, x, 32, palette.shadow, thickness: 1.5);
  }
  for (var y = 0; y <= 32; y += 10) {
    _line(image, 0, y, 32, y, palette.shadow, thickness: 1.5);
  }
  _rect(image, 5, 5, 27, 27, palette.ink);
  for (final point in const [(5, 5), (27, 5), (5, 27), (27, 27), (16, 16)]) {
    _dot(image, point.$1, point.$2, 1, palette.highlight);
  }
}

void _drawDesertPattern(img.Image image, _ZonePatternPalette palette) {
  for (var y = 7; y <= 25; y += 9) {
    _polyline(image, [
      (0, y),
      (8, y - 2),
      (16, y + 1),
      (24, y - 1),
      (32, y + 2),
    ], palette.shadow);
  }
  _polyline(image, [(4, 29), (12, 24), (20, 27), (28, 23)], palette.ink);
  _dot(image, 26, 8, 1, palette.highlight);
}

void _drawTemplePattern(img.Image image, _ZonePatternPalette palette) {
  _rect(image, 8, 8, 24, 24, palette.shadow);
  _rect(image, 12, 12, 20, 20, palette.ink);
  _line(image, 16, 0, 16, 32, palette.shadow);
  _line(image, 0, 16, 32, 16, palette.shadow);
  _dot(image, 16, 16, 2, palette.highlight);
}

void _drawVolcanicPattern(img.Image image, _ZonePatternPalette palette) {
  _polyline(
    image,
    [(0, 8), (6, 6), (11, 12), (17, 9), (23, 15), (32, 12)],
    palette.shadow,
    thickness: 1.5,
  );
  _polyline(
    image,
    [(5, 32), (11, 24), (16, 28), (22, 20), (30, 24)],
    palette.ink,
    thickness: 1.5,
  );
  for (final point in const [(9, 17), (19, 6), (25, 27), (28, 14)]) {
    _dot(image, point.$1, point.$2, 1, palette.highlight);
  }
}
