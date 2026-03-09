import 'dart:async';
import 'dart:typed_data';
import 'dart:html' as html;

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
  final input = html.FileUploadInputElement();
  input.accept = 'image/*';
  input.capture = useFrontCamera ? 'user' : 'environment';

  final completer = Completer<CapturedImage?>();

  input.onChange.listen((_) async {
    final files = input.files;
    if (files == null || files.isEmpty) {
      if (!completer.isCompleted) completer.complete(null);
      return;
    }
    final file = files.first;
    final reader = html.FileReader();
    reader.readAsArrayBuffer(file);
    await reader.onLoadEnd.first;
    final result = reader.result;
    if (result is ByteBuffer) {
      completer.complete(
        CapturedImage(
          bytes: Uint8List.view(result),
          mimeType: file.type,
          name: file.name,
        ),
      );
    } else {
      completer.complete(null);
    }
  });

  input.onError.listen((_) {
    if (!completer.isCompleted) completer.complete(null);
  });

  input.click();
  return completer.future;
}
