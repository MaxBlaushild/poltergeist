import 'dart:typed_data';

import 'package:file_picker/file_picker.dart';

import 'trip_document_download_stub.dart'
    if (dart.library.io) 'trip_document_download_io.dart'
    as platform;

/// Native desktop pickers choose a path; mobile and web pickers save the bytes.
/// Recheck the reader's session after any desktop picker before writing a file.
Future<void> saveTripDocumentBytes(
  Uint8List bytes,
  String filename, {
  required bool Function() isCurrent,
}) async {
  if (!isCurrent()) return;
  final safeName = filename
      .replaceAll(RegExp(r'[/\\\x00-\x1f<>:"|?*]'), '_')
      .replaceAll(RegExp(r'^[. ]+|[. ]+$'), '')
      .trim();
  final path = await FilePicker.platform.saveFile(
    dialogTitle: 'Save trip document',
    fileName: safeName.isEmpty ? 'trip-document' : safeName,
    bytes: platform.desktop ? null : bytes,
  );
  if (platform.desktop && path != null && isCurrent()) {
    await platform.writeBytes(path, bytes);
  }
}
