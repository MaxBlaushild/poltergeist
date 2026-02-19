import 'dart:typed_data';

class CapturedImage {
  final Uint8List bytes;
  final String? mimeType;
  final String? name;

  const CapturedImage({
    required this.bytes,
    this.mimeType,
    this.name,
  });
}

Future<CapturedImage?> captureImageFromCamera({bool useFrontCamera = false}) async {
  return null;
}
