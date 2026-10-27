import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';
import 'dart:ui' as ui;
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/travel_lodging_editor.dart';
import 'package:travel_angels/widgets/travel_trip_design.dart';
import 'package:travel_angels/widgets/travel_trip_logistics.dart';
import 'support/app_fonts.dart';
import 'travel_trip_widgets_test.dart' show sampleTrip, showTestDialog;

final _photo = base64Decode(
  'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRZkAAAAASUVORK5CYII=',
);
const _property = <String, dynamic>{
  'provider': 'google',
  'placeId': 'place-1',
  'name': 'Hotel from Google',
  'address': 'Google-only street address',
  'rating': 4.6,
  'photos': <dynamic>[],
  'attributions': [
    {
      'displayName': 'Property data partner',
      'uri': 'https://example.test/source',
    },
  ],
};

class _Service extends TravelCalendarService {
  final writes = <Map<String, dynamic>>[];
  final uploads = <Map<String, dynamic>>[];
  final searches = <String, Completer<Map<String, dynamic>>>{};
  Map<String, dynamic> stop = sampleTrip()
    ..['draft'] = {'lodging': <dynamic>[]};
  bool failSecondPhoto = false;
  bool failPropertyDetails = false;
  @override
  Future<Map<String, dynamic>> get(
    String path, {
    Map<String, dynamic>? query,
  }) async {
    if (path.endsWith('/participations')) return {};
    if (path.endsWith('/lodging-properties')) {
      if (searches.containsKey(query?['q'])) {
        return searches[query!['q']]!.future;
      }
      return {
        'properties': [_property],
      };
    }
    if (failPropertyDetails) {
      throw const TravelCalendarException('Property lookup unavailable.');
    }
    return {'property': _property};
  }

  @override
  Future<Uint8List> loadLodgingImage(
    String url, {
    Map<String, dynamic>? query,
  }) async => _photo;
  @override
  Future<Map<String, dynamic>> put(
    String path, {
    Map<String, dynamic>? data,
  }) async {
    writes.add(tripCopy(data!));
    final draft = tripCopy(data)..remove('revision');
    for (final item in (draft['lodging'] as List? ?? [])) {
      item['id'] ??= 'stay-1';
    }
    stop = {...stop, 'revision': (data['revision'] as int) + 1, 'draft': draft};
    return {'stop': tripCopy(stop)};
  }

  @override
  Future<Map<String, dynamic>> uploadLodgingPhoto(
    String stopId,
    String entryId,
    Uint8List bytes,
    String filename, {
    required String altText,
    required int revision,
  }) async {
    uploads.add({
      'id': entryId,
      'filename': filename,
      'revision': revision,
      'altText': altText,
    });
    if (failSecondPhoto && filename == 'second.png') {
      throw const TravelCalendarException('Image upload interrupted.');
    }
    final draft = tripDraft(stop);
    final lodging = tripMaps(draft['lodging']);
    final item = lodging.firstWhere((item) => item['id'] == entryId);
    item['photos'] = [
      ...tripMaps(item['photos']),
      {
        'mediaId': filename,
        'altText': altText,
        'url':
            '/travel-angels/calendar/stops/trip/lodging/stay-1/photos/photo-1',
      },
    ];
    draft['lodging'] = lodging;
    stop = {...stop, 'revision': revision + 1, 'draft': draft};
    return {'stop': tripCopy(stop)};
  }
}

Finder _field(String label) => find.widgetWithText(TextFormField, label);
Future<void> _tap(WidgetTester tester, Finder finder) async {
  await tester.ensureVisible(finder);
  await tester.pump();
  await tester.tap(finder);
  // Engine image decoding needs real async time after newly visible photos build.
  for (var i = 0; i < 2; i++) {
    await tester.pump(const Duration(milliseconds: 400));
    await tester.runAsync(() async {
      await Future<void>.delayed(const Duration(milliseconds: 50));
    });
  }
  await tester.pumpAndSettle();
}

Future<void> _open(WidgetTester tester, _Service service) => showTestDialog(
  tester,
  TravelTripLogisticsEditor(
    stop: sampleTrip(),
    kind: 'lodging',
    service: service,
  ),
);

class _Adapter implements HttpClientAdapter {
  final requests = <RequestOptions>[];
  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? stream,
    Future<void>? cancelFuture,
  ) async {
    requests.add(options);
    if (options.method == 'POST') {
      return ResponseBody.fromString(
        '{"stop":{"revision":8}}',
        200,
        headers: {
          Headers.contentTypeHeader: ['application/json'],
        },
      );
    }
    return ResponseBody.fromBytes(_photo, 200);
  }

  @override
  void close({bool force = false}) {}
}

void main() {
  setUpAll(loadAppFonts);
  testWidgets(
    'property selection saves only its reference and keeps host facts',
    (tester) async {
      final service = _Service();
      await _open(tester, service);
      await tester.enterText(_field('Property name'), 'Our seaside base');
      await tester.enterText(
        find.byKey(const ValueKey('lodging-property-search')),
        'Hotel',
      );
      await tester.pump(const Duration(milliseconds: 400));
      await tester.pumpAndSettle();
      expect(find.text('Property data partner'), findsOneWidget);
      await _tap(tester, find.text('Hotel from Google'));
      expect(find.text('Google-only street address'), findsOneWidget);
      await _tap(tester, find.text('Use this property'));
      await _tap(tester, find.text('Save logistics draft'));
      final entry = (service.writes.single['lodging'] as List).last as Map;
      expect(entry['name'], 'Our seaside base');
      expect(entry['property'], {'provider': 'google', 'placeId': 'place-1'});
      expect(entry['status'], 'considering');
      expect(entry['datesTbd'], true);
      expect(entry['visibility'], 'attendees');
      expect(jsonEncode(service.writes), isNot(contains('Google-only')));
      expect(jsonEncode(service.writes), isNot(contains('Hotel from Google')));
      expect(tester.takeException(), isNull);
    },
  );
  testWidgets('late property searches cannot replace newer matches', (
    tester,
  ) async {
    final service = _Service();
    service.searches['Lis'] = Completer();
    service.searches['Tokyo'] = Completer();
    await _open(tester, service);
    final search = find.byKey(const ValueKey('lodging-property-search'));
    await tester.enterText(search, 'Lis');
    await tester.pump(const Duration(milliseconds: 400));
    await tester.enterText(search, 'Tokyo');
    await tester.pump(const Duration(milliseconds: 400));
    service.searches['Tokyo']!.complete({
      'properties': [
        {..._property, 'name': 'Tokyo stay'},
      ],
    });
    await tester.pumpAndSettle();
    service.searches['Lis']!.complete({
      'properties': [
        {..._property, 'name': 'Lisbon stay'},
      ],
    });
    await tester.pumpAndSettle();
    expect(find.text('Tokyo stay'), findsOneWidget);
    expect(find.text('Lisbon stay'), findsNothing);
    await _tap(tester, find.text('Cancel'));
  });
  testWidgets(
    'partial photo upload retries preserve saved photos and revisions',
    (tester) async {
      final service = _Service()..failSecondPhoto = true;
      await _open(tester, service);
      await tester.enterText(_field('Property name'), 'Our rental');
      final first = PendingLodgingPhoto(bytes: _photo, filename: 'first.png')
        ..altText = 'Terrace';
      final second = PendingLodgingPhoto(bytes: _photo, filename: 'second.png')
        ..altText = 'Bedroom';
      tester
          .widget<TravelLodgingSourceEditor>(
            find.byType(TravelLodgingSourceEditor),
          )
          .onPendingChanged([first, second]);
      await tester.pumpAndSettle();
      await _tap(tester, find.text('Save logistics draft'));
      expect(service.uploads.map((item) => item['revision']), [8, 9]);
      expect(
        find.textContaining('Your stay is saved as a draft.'),
        findsOneWidget,
      );
      expect(
        tester
            .widget<TravelLodgingSourceEditor>(
              find.byType(TravelLodgingSourceEditor),
            )
            .pendingPhotos
            .single,
        second,
      );
      service.failSecondPhoto = false;
      await _tap(tester, find.text('Save logistics draft'));
      expect(service.writes.last['revision'], 9);
      expect(service.uploads.map((item) => item['filename']), [
        'first.png',
        'second.png',
        'second.png',
      ]);
      expect(service.uploads.last['revision'], 10);
      expect(
        (tripMaps(
              tripDraft(service.stop)['lodging'],
            ).firstWhere((stay) => stay['id'] == 'stay-1')['photos']
            as List),
        hasLength(2),
      );
      expect(tripMaps(tripDraft(service.stop)['lodging']), hasLength(2));
      expect(find.byType(TravelTripLogisticsEditor), findsNothing);
    },
  );
  testWidgets(
    'native image selection prevents saving an incomplete selection',
    (tester) async {
      final service = _Service();
      await _open(tester, service);
      final editor = tester.widget<TravelLodgingSourceEditor>(
        find.byType(TravelLodgingSourceEditor),
      );
      editor.onPickingChanged(true);
      await tester.pump();
      expect(
        tester
            .widget<FilledButton>(
              find.widgetWithText(FilledButton, 'Save logistics draft'),
            )
            .onPressed,
        isNull,
      );
      editor.onPickingChanged(false);
      await tester.pump();
      expect(
        tester
            .widget<FilledButton>(
              find.widgetWithText(FilledButton, 'Save logistics draft'),
            )
            .onPressed,
        isNotNull,
      );
      expect(service.writes, isEmpty);
    },
  );
  testWidgets('an unavailable linked property can still be unlinked', (
    tester,
  ) async {
    final service = _Service()..failPropertyDetails = true;
    await showTestDialog(
      tester,
      TravelTripLogisticsEditor(
        stop: sampleTrip(),
        kind: 'lodging',
        entry: {
          ...tripMaps(sampleTrip()['lodging']).single,
          'property': {'provider': 'google', 'placeId': 'place-1'},
        },
        service: service,
      ),
    );
    expect(
      find.textContaining('Property details are unavailable'),
      findsOneWidget,
    );
    await _tap(tester, find.text('Unlink'));
    expect(
      find.textContaining('Property details are unavailable'),
      findsNothing,
    );
    expect(_field('Property name'), findsOneWidget);
    await _tap(tester, find.text('Save logistics draft'));
    expect(
      tripMaps(service.writes.single['lodging']).single['property'],
      isNull,
    );
  });
  for (final (width, scale) in [(390.0, 1.0), (320.0, 2.0)]) {
    testWidgets('lodging editor fits ${width.toInt()}px at ${scale}x text', (
      tester,
    ) async {
      await tester.binding.setSurfaceSize(Size(width, 844));
      addTearDown(() => tester.binding.setSurfaceSize(null));
      tester.view.physicalSize = Size(width, 844);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final boundary = GlobalKey();
      final service = _Service();
      await tester.pumpWidget(
        RepaintBoundary(
          key: boundary,
          child: MaterialApp(
            theme: AppTheme.lightTheme,
            builder: (context, child) => MediaQuery(
              data: MediaQuery.of(
                context,
              ).copyWith(textScaler: TextScaler.linear(scale)),
              child: child!,
            ),
            home: Scaffold(
              body: Builder(
                builder: (context) => TextButton(
                  onPressed: () => showDialog<void>(
                    context: context,
                    builder: (_) => TravelTripLogisticsEditor(
                      stop: sampleTrip(),
                      kind: 'lodging',
                      service: service,
                    ),
                  ),
                  child: const Text('Open'),
                ),
              ),
            ),
          ),
        ),
      );
      await _tap(tester, find.text('Open'));
      final source = tester.widget<TravelLodgingSourceEditor>(
        find.byType(TravelLodgingSourceEditor),
      );
      source.onPendingChanged([
        PendingLodgingPhoto(bytes: _photo, filename: 'terrace.png')
          ..altText = 'Our terrace',
      ]);
      await tester.pumpAndSettle();
      await tester.ensureVisible(_field('Photo description'));
      await tester.pumpAndSettle();
      expect(
        tester.getSize(_field('Photo description')).width,
        greaterThan(140),
      );
      if (const bool.fromEnvironment('TRIP_SCREENSHOTS')) {
        await tester.runAsync(() async {
          final image =
              await (boundary.currentContext!.findRenderObject()!
                      as RenderRepaintBoundary)
                  .toImage();
          final bytes = await image.toByteData(format: ui.ImageByteFormat.png);
          await File(
            '/private/tmp/travel-angels-lodging-editor-${width.toInt()}-${scale}x.png',
          ).writeAsBytes(bytes!.buffer.asUint8List());
          image.dispose();
        });
      }
    });
  }
  test(
    'lodging images enforce endpoint and origin while forwarding viewer context',
    () async {
      SharedPreferences.setMockInitialValues({'token': 'token'});
      final adapter = _Adapter();
      final service = TravelCalendarService(
        client: Dio(BaseOptions(baseUrl: 'https://example.test'))
          ..httpClientAdapter = adapter,
      );
      await service.loadLodgingImage(
        '/travel-angels/calendar/stops/trip/lodging/stay/property/photos/0?photoVersion=selected-photo',
        query: {'draft': true, 'stopToken': 'link'},
      );
      expect(adapter.requests.single.uri.queryParameters, {
        'photoVersion': 'selected-photo',
        'draft': 'true',
        'stopToken': 'link',
      });
      expect(adapter.requests.single.headers['Authorization'], 'Bearer token');
      expect(adapter.requests.single.followRedirects, false);
      for (final path in [
        'https://foreign.test/travel-angels/calendar/stops/trip/lodging/stay/photos/photo',
        '/unrelated',
        '/travel-angels/calendar/stops/trip/hero/photo',
      ]) {
        await expectLater(
          service.loadLodgingImage(path),
          throwsA(isA<TravelCalendarException>()),
        );
      }
      await service.uploadLodgingPhoto(
        'trip',
        'stay',
        _photo,
        'photo.png',
        altText: 'The terrace',
        revision: 7,
      );
      final fields = Map.fromEntries(
        (adapter.requests.last.data as FormData).fields,
      );
      expect(fields, containsPair('revision', '7'));
      expect(fields, containsPair('altText', 'The terrace'));
      expect(adapter.requests, hasLength(2));
    },
  );
}
