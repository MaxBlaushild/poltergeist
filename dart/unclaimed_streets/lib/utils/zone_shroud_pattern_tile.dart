import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:image/image.dart' as img;

const _zoneShroudPatternVersion = 'v1';
const _zoneShroudPatternTileSize = 32;

Uint8List? _zoneShroudPatternTileCache;

String zoneShroudPatternImageId() =>
    'zone_shroud_pattern_$_zoneShroudPatternVersion';

Uint8List zoneShroudPatternTileBytes() {
  return _zoneShroudPatternTileCache ??= () {
    final image = img.Image(
      width: _zoneShroudPatternTileSize,
      height: _zoneShroudPatternTileSize,
      numChannels: 4,
    );
    img.fill(image, color: img.ColorRgba8(0, 0, 0, 0));

    const mist = Color(0xFF758293);
    const haze = Color(0xFF97A4B2);
    const shadow = Color(0xFF232B34);

    final mistInk = _imgColor(mist, alpha: 56);
    final hazeInk = _imgColor(haze, alpha: 34);
    final shadowInk = _imgColor(shadow, alpha: 28);

    for (var offset = -8; offset <= 32; offset += 12) {
      _line(image, offset + 2, 7, offset + 10, 4, mistInk, thickness: 1.2);
      _line(image, offset + 6, 17, offset + 14, 14, hazeInk, thickness: 1.0);
      _line(image, offset + 1, 27, offset + 9, 24, mistInk, thickness: 1.0);
    }

    for (final dot in const <Offset>[
      Offset(4, 4),
      Offset(14, 6),
      Offset(24, 4),
      Offset(8, 13),
      Offset(20, 12),
      Offset(28, 16),
      Offset(5, 22),
      Offset(15, 24),
      Offset(25, 22),
      Offset(10, 30),
      Offset(22, 29),
    ]) {
      _dot(image, dot.dx.round(), dot.dy.round(), 1, hazeInk);
    }

    for (final puff in const <Offset>[
      Offset(6, 10),
      Offset(18, 18),
      Offset(29, 9),
      Offset(27, 27),
    ]) {
      _dot(image, puff.dx.round(), puff.dy.round(), 2, shadowInk);
      _dot(image, puff.dx.round() + 1, puff.dy.round() + 1, 1, hazeInk);
    }

    return Uint8List.fromList(img.encodePng(image));
  }();
}

img.ColorRgba8 _imgColor(Color color, {required int alpha}) {
  return img.ColorRgba8(
    (color.r * 255.0).round().clamp(0, 255),
    (color.g * 255.0).round().clamp(0, 255),
    (color.b * 255.0).round().clamp(0, 255),
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
