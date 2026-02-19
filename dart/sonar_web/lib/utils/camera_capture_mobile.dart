import 'dart:typed_data';

import 'package:image_picker/image_picker.dart';

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
  final picker = ImagePicker();
  final file = await picker.pickImage(
    source: ImageSource.camera,
    preferredCameraDevice:
        useFrontCamera ? CameraDevice.front : CameraDevice.rear,
    imageQuality: 90,
  );
  if (file == null) return null;
  final bytes = await file.readAsBytes();
  if (bytes.isEmpty) return null;
  return CapturedImage(
    bytes: bytes,
    mimeType: file.mimeType,
    name: file.name,
  );
}
