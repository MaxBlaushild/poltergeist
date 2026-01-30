import 'dart:typed_data';

import 'package:http/http.dart' as http;
import 'package:image/image.dart' as img;

const _placeholderUrl =
    'https://crew-points-of-interest.s3.amazonaws.com/question-mark.webp';
const _thumbnailSize = 192;
const _cornerRadius = 12;

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
