import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';
import 'dart:ui' as ui;

import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:travel_angels/screens/travel_trip_screen.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/travel_trip_hero.dart';

import 'support/app_fonts.dart';
import 'travel_trip_widgets_test.dart' show sampleTrip, TripTestService;

final _photo = base64Decode(
  'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRZkAAAAASUVORK5CYII=',
);

class _HeroService extends TripTestService {
  Completer<Uint8List>? pendingLoad;
  Completer<Map<String, dynamic>>? pendingUpload;
  Object? uploadFailure;
  final uploads = <Map<String, dynamic>>[];

  @override
  Future<Uint8List> loadHero(String url, {Map<String, dynamic>? query}) async =>
      pendingLoad?.future ?? _photo;

  @override
  Future<Map<String, dynamic>> uploadHero(
    String stopId,
    Uint8List bytes,
    String filename, {
    required String altText,
    required double focalX,
    required double focalY,
    required int revision,
    double cropX = 0,
    double cropY = 0,
    double cropWidth = 1,
    double cropHeight = 1,
  }) async {
    uploads.add({
      'stopId': stopId,
      'bytes': bytes,
      'filename': filename,
      'altText': altText,
      'revision': revision,
      'focalX': focalX,
      'focalY': focalY,
      'cropX': cropX,
      'cropY': cropY,
      'cropWidth': cropWidth,
      'cropHeight': cropHeight,
    });
    if (uploadFailure != null) throw uploadFailure!;
    return pendingUpload?.future ?? {};
  }
}

class _Picker extends FilePicker {
  Uint8List bytes = _photo;
  String filename = 'lisbon.png';
  bool requestedBytes = false;
  List<String>? extensions;
  Object? failure;

  @override
  Future<FilePickerResult?> pickFiles({
    String? dialogTitle,
    String? initialDirectory,
    FileType type = FileType.any,
    List<String>? allowedExtensions,
    Function(FilePickerStatus)? onFileLoading,
    bool allowCompression = true,
    int compressionQuality = 30,
    bool allowMultiple = false,
    bool withData = false,
    bool withReadStream = false,
    bool lockParentWindow = false,
    bool readSequential = false,
  }) async {
    requestedBytes = withData;
    extensions = allowedExtensions;
    if (failure != null) throw failure!;
    return FilePickerResult([
      PlatformFile(name: filename, size: bytes.length, bytes: bytes),
    ]);
  }
}

Map<String, dynamic> _existingCoverTrip() => sampleTrip()
  ..['hero'] = {
    'mediaId': 'old-photo',
    'url': '/travel-angels/calendar/stops/trip/hero/old-photo',
    'altText': 'The old cover',
    'focalX': .5,
    'focalY': .5,
    'cropX': 0.0,
    'cropY': 0.0,
    'cropWidth': 1.0,
    'cropHeight': 1.0,
  };

Future<void> _tap(WidgetTester tester, Finder finder) async {
  await tester.ensureVisible(finder);
  await tester.pumpAndSettle();
  await tester.tap(finder);
  await tester.pumpAndSettle();
}

Future<void> _pick(WidgetTester tester, String label) async {
  final finder = find.text(label);
  await tester.ensureVisible(finder);
  await tester.pump(const Duration(milliseconds: 300));
  await tester.runAsync(() async {
    await tester.tap(finder);
    // Image validation uses engine codecs outside the test's fake clock.
    await Future<void>.delayed(const Duration(milliseconds: 100));
  });
  await tester.pumpAndSettle();
}

Future<void> _openEditor(
  WidgetTester tester,
  _HeroService service, {
  Map<String, dynamic>? stop,
}) async {
  await tester.pumpWidget(
    MaterialApp(
      theme: AppTheme.lightTheme,
      home: Scaffold(
        body: Builder(
          builder: (context) => TextButton(
            onPressed: () => showDialog<bool>(
              context: context,
              builder: (_) => TravelTripHeroEditor(
                stop: stop ?? sampleTrip(),
                service: service,
              ),
            ),
            child: const Text('Open'),
          ),
        ),
      ),
    ),
  );
  await tester.tap(find.text('Open'));
  await tester.pump();
  await tester.pump(const Duration(milliseconds: 400));
}

void main() {
  setUpAll(loadAppFonts);
  late _Picker picker;
  setUp(() {
    picker = _Picker();
    FilePicker.platform = picker;
  });

  testWidgets('only hosts can open the cover editor directly from the image', (
    tester,
  ) async {
    await tester.binding.setSurfaceSize(const Size(1200, 1000));
    addTearDown(() => tester.binding.setSurfaceSize(null));
    for (final (canManage, width) in [
      (false, 1200.0),
      (true, 1200.0),
      (true, 390.0),
    ]) {
      await tester.binding.setSurfaceSize(Size(width, 1000));
      final boundary = GlobalKey();
      final stop = sampleTrip()..['canManage'] = canManage;
      await tester.pumpWidget(
        MaterialApp(
          theme: AppTheme.lightTheme,
          home: RepaintBoundary(
            key: boundary,
            child: TravelTripScreen(
              key: ValueKey('$canManage:$width'),
              initialStop: stop,
              service: _HeroService(),
              loadStop: () async => stop,
              onAction: (_, _) async {},
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      final edit = find.byKey(const ValueKey('trip-edit-cover'));
      expect(edit, canManage ? findsOneWidget : findsNothing);
      if (canManage) {
        if (const bool.fromEnvironment('TRIP_SCREENSHOTS')) {
          await tester.runAsync(() async {
            final image =
                await (boundary.currentContext!.findRenderObject()!
                        as RenderRepaintBoundary)
                    .toImage();
            final bytes = await image.toByteData(
              format: ui.ImageByteFormat.png,
            );
            await File(
              '/private/tmp/travel-angels-trip-cover-host-${width.toInt()}.png',
            ).writeAsBytes(bytes!.buffer.asUint8List());
            image.dispose();
          });
        }
        await _tap(tester, edit);
        expect(find.byType(TravelTripHeroEditor), findsOneWidget);
        expect(find.text('Edit cover image'), findsOneWidget);
        await _tap(tester, find.text('Cancel'));
      }
    }
    expect(tester.takeException(), isNull);
  });

  testWidgets('draft cover removal does not inherit the published cover', (
    tester,
  ) async {
    await tester.binding.setSurfaceSize(const Size(1200, 1000));
    addTearDown(() => tester.binding.setSurfaceSize(null));
    final stop = _existingCoverTrip()..['canManage'] = true;
    stop['draft'] = {...sampleTrip(), 'title': 'Unpublished title'};
    await tester.pumpWidget(
      MaterialApp(
        theme: AppTheme.lightTheme,
        home: TravelTripScreen(
          initialStop: stop,
          service: _HeroService(),
          loadStop: () async => stop,
          onAction: (_, _) async {},
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(
      tester
          .widget<TripProtectedHero>(find.byType(TripProtectedHero))
          .hero['mediaId'],
      'old-photo',
    );
    await _tap(tester, find.text('View draft'));
    expect(
      tester.widget<TripProtectedHero>(find.byType(TripProtectedHero)).hero,
      isEmpty,
    );
    await _tap(tester, find.text('View published'));
    expect(
      tester
          .widget<TripProtectedHero>(find.byType(TripProtectedHero))
          .hero['mediaId'],
      'old-photo',
    );
  });

  testWidgets(
    'uploads bytes-only image with description and current revision',
    (tester) async {
      final service = _HeroService();
      await _openEditor(tester, service);
      await _pick(tester, 'Choose image');
      expect(picker.requestedBytes, isTrue);
      expect(picker.extensions, containsAll(['jpg', 'jpeg', 'png']));
      expect(find.text('Replace image'), findsOneWidget);
      await tester.enterText(find.byType(TextField), ' Lisbon in the evening ');
      await _tap(tester, find.text('Save cover draft'));
      expect(service.uploads.single, containsPair('stopId', 'trip'));
      expect(service.uploads.single['bytes'], _photo);
      expect(service.uploads.single['filename'], 'lisbon.png');
      expect(service.uploads.single['altText'], 'Lisbon in the evening');
      expect(service.uploads.single['revision'], 7);
      expect(find.byType(TravelTripHeroEditor), findsNothing);
      expect(service.requests, isEmpty);
    },
  );

  testWidgets('removing cover preserves the rest of the unpublished draft', (
    tester,
  ) async {
    final stop = _existingCoverTrip();
    stop['draft'] = {...stop, 'title': 'Unpublished title'};
    final service = _HeroService();
    await _openEditor(tester, service, stop: stop);
    await _tap(tester, find.text('Remove image'));
    await _tap(tester, find.text('Save cover draft'));
    final request = service.requests.single;
    expect(request['method'], 'PUT');
    expect(request['path'], '/calendar/stops/trip');
    final data = request['data'] as Map;
    expect(data, containsPair('hero', null));
    expect(data['title'], 'Unpublished title');
    expect(data['lodging'], stop['lodging']);
    expect(data['flights'], stop['flights']);
    expect(data['revision'], 7);
    expect(service.uploads, isEmpty);
    expect(find.byType(TravelTripHeroEditor), findsNothing);
  });

  testWidgets(
    'late original cover load cannot overwrite a picked replacement',
    (tester) async {
      final oldBytes = Uint8List.fromList([..._photo, 0]);
      final service = _HeroService()..pendingLoad = Completer<Uint8List>();
      await _openEditor(tester, service, stop: _existingCoverTrip());
      final pickLabel = find.text('Replace image').evaluate().isEmpty
          ? 'Choose image'
          : 'Replace image';
      await _pick(tester, pickLabel);
      service.pendingLoad!.complete(oldBytes);
      await tester.pumpAndSettle();
      final preview = tester.widget<TripHeroCanvas>(
        find.byType(TripHeroCanvas),
      );
      expect(preview.bytes, _photo);
      await _tap(tester, find.text('Save cover draft'));
      expect(service.uploads.single['bytes'], _photo);
      expect(service.uploads.single['filename'], 'lisbon.png');
    },
  );

  testWidgets(
    'upload failure preserves selected image and description for retry',
    (tester) async {
      final service = _HeroService()
        ..uploadFailure = const TravelCalendarException(
          'Upload did not finish. Try again.',
          statusCode: 503,
        );
      await _openEditor(tester, service);
      await _pick(tester, 'Choose image');
      await tester.enterText(find.byType(TextField), 'Rooftops at sunset');
      await _tap(tester, find.text('Save cover draft'));
      expect(find.text('Upload did not finish. Try again.'), findsOneWidget);
      expect(find.text('Replace image'), findsOneWidget);
      expect(
        tester.widget<TextField>(find.byType(TextField)).controller!.text,
        'Rooftops at sunset',
      );
      service.uploadFailure = null;
      await _tap(tester, find.text('Save cover draft'));
      expect(service.uploads, hasLength(2));
      expect(service.uploads.last['bytes'], _photo);
      expect(find.byType(TravelTripHeroEditor), findsNothing);
    },
  );

  testWidgets('malformed image is rejected before it can replace the cover', (
    tester,
  ) async {
    picker.bytes = Uint8List.fromList(utf8.encode('not an image'));
    final service = _HeroService();
    await _openEditor(tester, service);
    await _pick(tester, 'Choose image');
    expect(find.text('Choose image'), findsOneWidget);
    expect(find.text('Replace image'), findsNothing);
    expect(service.uploads, isEmpty);
    expect(service.requests, isEmpty);
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'saving disables image mutations and prevents duplicate uploads',
    (tester) async {
      final service = _HeroService()
        ..pendingUpload = Completer<Map<String, dynamic>>();
      await _openEditor(tester, service);
      await _pick(tester, 'Choose image');
      await tester.enterText(find.byType(TextField), 'Rooftops at sunset');
      await _tap(tester, find.text('Save cover draft'));
      expect(service.uploads, hasLength(1));
      expect(
        tester
            .widget<TextButton>(
              find.ancestor(
                of: find.text('Replace image'),
                matching: find.byWidgetPredicate(
                  (widget) => widget is TextButton,
                ),
              ),
            )
            .onPressed,
        isNull,
      );
      expect(
        tester
            .widget<TextButton>(find.widgetWithText(TextButton, 'Remove image'))
            .onPressed,
        isNull,
      );
      expect(
        tester
            .widget<FilledButton>(find.widgetWithText(FilledButton, 'Saving…'))
            .onPressed,
        isNull,
      );
      service.pendingUpload!.complete({});
      await tester.pumpAndSettle();
      expect(find.byType(TravelTripHeroEditor), findsNothing);
      expect(service.uploads, hasLength(1));
    },
  );

  testWidgets('mobile cover editor keeps framing optional and save reachable', (
    tester,
  ) async {
    await tester.binding.setSurfaceSize(const Size(390, 844));
    addTearDown(() => tester.binding.setSurfaceSize(null));
    final boundary = GlobalKey();
    final service = _HeroService();
    await tester.pumpWidget(
      MaterialApp(
        theme: AppTheme.lightTheme,
        home: RepaintBoundary(
          key: boundary,
          child: Scaffold(
            body: TravelTripHeroEditor(stop: sampleTrip(), service: service),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.text('Save cover draft').hitTestable(), findsOneWidget);
    expect(find.byType(Slider).hitTestable(), findsNothing);
    expect(tester.takeException(), isNull);
    if (const bool.fromEnvironment('TRIP_SCREENSHOTS')) {
      await tester.runAsync(() async {
        final image =
            await (boundary.currentContext!.findRenderObject()!
                    as RenderRepaintBoundary)
                .toImage();
        final bytes = await image.toByteData(format: ui.ImageByteFormat.png);
        await File(
          '/private/tmp/travel-angels-trip-cover-editor-mobile.png',
        ).writeAsBytes(bytes!.buffer.asUint8List());
        image.dispose();
      });
    }
    await _pick(tester, 'Choose image');
    await _tap(tester, find.text('Adjust framing'));
    expect(find.byType(Slider), findsNWidgets(3));
    expect(tester.takeException(), isNull);
  });
}
