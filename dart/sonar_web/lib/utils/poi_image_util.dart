import 'dart:typed_data';

import 'package:http/http.dart' as http;
import 'package:image/image.dart' as img;

const _placeholderUrl =
    'https://crew-points-of-interest.s3.amazonaws.com/question-mark.webp';
const _thumbnailSize = 192;
const _cornerRadius = 12;
const _questMarkerSize = 44;
const _questMarkerPadding = 6;

/// Fetches the POI image (or placeholder), resizes to a square, applies
/// rounded corners, and returns PNG bytes suitable for MapLibre addImage.
Future<Uint8List?> loadPoiThumbnail(String? imageUrl) async {
  final url = imageUrl != null && imageUrl.isNotEmpty
      ? imageUrl
      : _placeholderUrl;
  try {
    final response = await http.get(Uri.parse(url));
    if (response.statusCode != 200) return null;
    final bytes = response.bodyBytes;
    final decoded = img.decodeImage(bytes);
    if (decoded == null) return null;
    final square = img.copyResizeCropSquare(
      decoded,
      size: _thumbnailSize,
      radius: _cornerRadius,
      antialias: true,
    );
    return Uint8List.fromList(img.encodePng(square));
  } catch (_) {
    return null;
  }
}

/// Same as [loadPoiThumbnail], but adds a gold border around the image.
/// Useful for quest highlights where we need a visible outline that isn't
/// dependent on map styling support.
Future<Uint8List?> loadPoiThumbnailWithBorder(
  String? imageUrl, {
  int borderWidth = 10,
}) async {
  final url = imageUrl != null && imageUrl.isNotEmpty
      ? imageUrl
      : _placeholderUrl;
  try {
    final response = await http.get(Uri.parse(url));
    if (response.statusCode != 200) return null;
    final bytes = response.bodyBytes;
    final decoded = img.decodeImage(bytes);
    if (decoded == null) return null;
    final square = img.copyResizeCropSquare(
      decoded,
      size: _thumbnailSize,
      radius: _cornerRadius,
      antialias: true,
    );
    final borderedSize = _thumbnailSize + borderWidth * 2;
    final bordered = img.Image(width: borderedSize, height: borderedSize);
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
  } catch (_) {
    return null;
  }
}

/// Same as [loadPoiThumbnail], but adds a golden quest marker in the corner.
Future<Uint8List?> loadPoiThumbnailWithQuestMarker(String? imageUrl) async {
  final url = imageUrl != null && imageUrl.isNotEmpty
      ? imageUrl
      : _placeholderUrl;
  try {
    final response = await http.get(Uri.parse(url));
    if (response.statusCode != 200) return null;
    final bytes = response.bodyBytes;
    final decoded = img.decodeImage(bytes);
    if (decoded == null) return null;
    final square = img.copyResizeCropSquare(
      decoded,
      size: _thumbnailSize,
      radius: _cornerRadius,
      antialias: true,
    );
    _drawQuestMarker(square);
    return Uint8List.fromList(img.encodePng(square));
  } catch (_) {
    return null;
  }
}

void _drawQuestMarker(img.Image image) {
  final radius = (_questMarkerSize / 2).round();
  final centerX = _thumbnailSize - _questMarkerPadding - radius;
  final centerY = _questMarkerPadding + radius;
  final gold = img.ColorRgba8(245, 197, 66, 255);
  final goldEdge = img.ColorRgba8(255, 233, 168, 255);
  final dark = img.ColorRgba8(54, 35, 0, 255);

  img.fillCircle(
    image,
    x: centerX,
    y: centerY,
    radius: radius,
    color: gold,
    antialias: true,
  );
  img.drawCircle(
    image,
    x: centerX,
    y: centerY,
    radius: radius,
    color: goldEdge,
    antialias: true,
  );

  final barWidth = 6;
  final barHeight = 18;
  img.fillRect(
    image,
    x1: centerX - (barWidth ~/ 2),
    y1: centerY - (barHeight ~/ 2) - 3,
    x2: centerX + (barWidth ~/ 2),
    y2: centerY + (barHeight ~/ 2) - 3,
    color: dark,
  );
  img.fillCircle(
    image,
    x: centerX,
    y: centerY + (barHeight ~/ 2) + 4,
    radius: 3,
    color: dark,
  );
}
