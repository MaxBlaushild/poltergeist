import 'dart:io';
import 'dart:typed_data';

bool get desktop => Platform.isMacOS || Platform.isLinux || Platform.isWindows;

Future<void> writeBytes(String path, Uint8List bytes) =>
    File(path).writeAsBytes(bytes, flush: true);
