import 'dart:async';
import 'dart:io';
import 'dart:typed_data';

import 'package:file_picker/file_picker.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:travel_angels/utils/trip_document_download.dart';

class _SavePicker extends FilePicker {
  final result = Completer<String?>();
  String? filename;
  Uint8List? bytes;
  int calls = 0;

  @override
  Future<String?> saveFile({
    String? dialogTitle,
    String? fileName,
    String? initialDirectory,
    FileType type = FileType.any,
    List<String>? allowedExtensions,
    Uint8List? bytes,
    bool lockParentWindow = false,
  }) {
    calls++;
    filename = fileName;
    this.bytes = bytes;
    return result.future;
  }
}

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  setUp(() => FilePicker.platform = _SavePicker());

  test(
    'a session change while selecting a destination does not save the file',
    () async {
      final old = FilePicker.platform;
      final picker = _SavePicker();
      FilePicker.platform = picker;
      addTearDown(() => FilePicker.platform = old);
      final directory = await Directory.systemTemp.createTemp(
        'trip-document-test-',
      );
      addTearDown(() => directory.delete(recursive: true));
      final target = File('${directory.path}/private.pdf');
      var current = true;
      final saving = saveTripDocumentBytes(
        Uint8List.fromList([37, 80, 68, 70]),
        'private.pdf',
        isCurrent: () => current,
      );
      current = false;
      picker.result.complete(target.path);
      await saving;
      expect(await target.exists(), isFalse);
      expect(picker.bytes, isNull);
    },
  );

  test('desktop saves original bytes at the selected destination', () async {
    final old = FilePicker.platform;
    final picker = _SavePicker();
    FilePicker.platform = picker;
    addTearDown(() => FilePicker.platform = old);
    final directory = await Directory.systemTemp.createTemp(
      'trip-document-test-',
    );
    addTearDown(() => directory.delete(recursive: true));
    final target = File('${directory.path}/itinerary.pdf');
    picker.result.complete(target.path);
    final bytes = Uint8List.fromList([37, 80, 68, 70]);
    await saveTripDocumentBytes(
      bytes,
      '../itinerary.pdf',
      isCurrent: () => true,
    );
    expect(await target.readAsBytes(), bytes);
    expect(picker.filename, '_itinerary.pdf');
    expect(picker.bytes, isNull);
  });

  test('closed readers never open a save picker', () async {
    final old = FilePicker.platform;
    final picker = _SavePicker();
    FilePicker.platform = picker;
    addTearDown(() => FilePicker.platform = old);
    await saveTripDocumentBytes(
      Uint8List(4),
      'file.pdf',
      isCurrent: () => false,
    );
    expect(picker.calls, 0);
  });
}
