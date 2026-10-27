import 'dart:async';
import 'dart:io';
import 'dart:typed_data';
import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/travel_lodging_card.dart';

import 'support/app_fonts.dart';
import 'travel_trip_widgets_test.dart' show TripTestAuth;

late Uint8List _image;

Future<void> _loadFixture() async {
  await loadAppFonts();
  final recorder = ui.PictureRecorder();
  final canvas = Canvas(recorder);
  const bounds = Rect.fromLTWH(0, 0, 640, 400);
  canvas.drawRect(
    bounds,
    Paint()
      ..shader = const LinearGradient(
        colors: [Color(0xffc9dbe3), Color(0xffffe6c1)],
        begin: Alignment.topCenter,
        end: Alignment.bottomCenter,
      ).createShader(bounds),
  );
  canvas.drawCircle(
    const Offset(480, 72),
    32,
    Paint()..color = const Color(0xffffeccb),
  );
  canvas.drawRect(
    const Rect.fromLTWH(0, 164, 640, 120),
    Paint()..color = const Color(0xffa5c5cd),
  );
  for (var i = 0; i < 8; i++) {
    final x = i * 88.0 - 20;
    final y = 208.0 + (i % 3) * 12;
    canvas.drawRect(
      Rect.fromLTWH(x, y, 75, 92),
      Paint()..color = const Color(0xfff1d7b2),
    );
    canvas.drawPath(
      Path()
        ..moveTo(x - 8, y)
        ..lineTo(x + 37, y - 36)
        ..lineTo(x + 83, y)
        ..close(),
      Paint()..color = const Color(0xffc47a58),
    );
    canvas.drawRect(
      Rect.fromLTWH(x + 30, y + 20, 14, 25),
      Paint()..color = const Color(0xff5d7274),
    );
  }
  canvas.drawRect(
    const Rect.fromLTWH(0, 300, 640, 100),
    Paint()..color = const Color(0xffd4b796),
  );
  canvas.drawRect(
    const Rect.fromLTWH(0, 285, 640, 18),
    Paint()..color = const Color(0xfff8e3c5),
  );
  canvas.drawOval(
    const Rect.fromLTWH(255, 325, 145, 35),
    Paint()..color = const Color(0xfff8e7cf),
  );
  canvas.drawLine(
    const Offset(327, 340),
    const Offset(327, 390),
    Paint()
      ..color = const Color(0xff79674f)
      ..strokeWidth = 8,
  );
  for (final x in [72.0, 560.0]) {
    canvas.drawRect(
      Rect.fromLTWH(x - 20, 330, 40, 48),
      Paint()..color = const Color(0xffb86b4b),
    );
    canvas.drawOval(
      Rect.fromLTWH(x - 25, 275, 34, 65),
      Paint()..color = const Color(0xff667c62),
    );
    canvas.drawOval(
      Rect.fromLTWH(x - 2, 285, 34, 56),
      Paint()..color = const Color(0xff7e906a),
    );
  }
  final picture = recorder.endRecording();
  final image = await picture.toImage(640, 400);
  final bytes = await image.toByteData(format: ui.ImageByteFormat.png);
  _image = bytes!.buffer.asUint8List();
  image.dispose();
  picture.dispose();
}

Map<String, dynamic> _stay() => {
  'id': 'stay',
  'name': 'Casa da Colina',
  'generalLocation': 'Graça, Lisbon',
  'description': 'A sunny base with a terrace above the rooftops.',
  'kind': 'group',
  'status': 'booked',
  'checkInDate': '2027-06-01',
  'checkOutDate': '2027-06-08',
  'timeZone': 'Europe/Lisbon',
  'url': 'https://example.test/hotel',
  'property': {'provider': 'google', 'placeId': 'lisbon-hotel'},
};

Map<String, dynamic> _property({String name = 'Casa da Colina Hotel'}) => {
  'property': {
    'provider': 'google',
    'placeId': 'lisbon-hotel',
    'name': name,
    'address': 'Rua da Graça 12, Lisbon',
    'mapsUrl': 'https://maps.google.com/?cid=123',
    'websiteUrl': 'https://example.test/property',
    'rating': 4.7,
    'ratingCount': 128,
    'photos': [
      {
        'url':
            '/travel-angels/calendar/stops/trip/lodging/stay/property/photos/0',
        'altText': 'The sunny terrace',
        'googleMapsUri': 'https://maps.google.com/photos/terrace',
        'authorAttributions': [
          {
            'displayName': 'Ana Photographer',
            'uri': 'https://maps.google.com/contrib/ana',
          },
        ],
      },
    ],
  },
};

class _Service extends TravelCalendarService {
  final requests = <Map<String, dynamic>>[];
  final images = <Map<String, dynamic>>[];
  final pending = <Completer<Map<String, dynamic>>>[];
  bool defer = false;
  bool fail = false;
  @override
  Future<Map<String, dynamic>> get(
    String path, {
    Map<String, dynamic>? query,
  }) async {
    requests.add({'path': path, 'query': query});
    if (fail) {
      throw const TravelCalendarException('Unavailable', statusCode: 503);
    }
    if (defer) {
      final response = Completer<Map<String, dynamic>>();
      pending.add(response);
      return response.future;
    }
    return _property();
  }

  @override
  Future<Uint8List> loadLodgingImage(
    String url, {
    Map<String, dynamic>? query,
  }) async {
    images.add({'url': url, 'query': query});
    return _image;
  }
}

Future<void> _settle(WidgetTester tester) async {
  // Build newly visible photos before allowing their engine codecs to finish.
  for (var i = 0; i < 2; i++) {
    await tester.pump(const Duration(milliseconds: 400));
    await tester.runAsync(() async {
      await Future<void>.delayed(const Duration(milliseconds: 50));
    });
  }
  await tester.pumpAndSettle();
}

Future<void> _tap(WidgetTester tester, Finder finder) async {
  await tester.ensureVisible(finder);
  await tester.pump();
  await tester.tap(finder);
  await _settle(tester);
}

Future<void> _show(
  WidgetTester tester,
  _Service service,
  Map<String, dynamic> entry, {
  bool draft = false,
  Map<String, dynamic> accessQuery = const {},
  double scale = 1,
  GlobalKey? boundary,
  TripTestAuth? auth,
}) async {
  Widget app = MaterialApp(
    theme: AppTheme.lightTheme,
    builder: (context, child) => MediaQuery(
      data: MediaQuery.of(
        context,
      ).copyWith(textScaler: TextScaler.linear(scale)),
      child: child!,
    ),
    home: RepaintBoundary(
      key: boundary,
      child: Scaffold(
        body: SingleChildScrollView(
          padding: const EdgeInsets.all(16),
          child: TravelTripLodgingCard(
            entry: entry,
            service: service,
            stopId: 'trip',
            accessQuery: accessQuery,
            draft: draft,
          ),
        ),
      ),
    ),
  );
  if (auth != null) {
    app = ChangeNotifierProvider<AuthProvider>.value(value: auth, child: app);
  }
  await tester.pumpWidget(app);
  await _settle(tester);
}

void main() {
  setUpAll(_loadFixture);

  testWidgets(
    'property stay prioritizes uploaded photos and opens attributed gallery',
    (tester) async {
      final service = _Service();
      final entry = _stay()
        ..['photos'] = [
          {
            'mediaId': 'upload',
            'url':
                '/travel-angels/calendar/stops/trip/lodging/stay/photos/upload',
            'altText': 'Our terrace',
          },
        ];
      await _show(tester, service, entry, accessQuery: {'share': 'guest-link'});
      expect(
        service.requests.single['path'],
        '/calendar/stops/trip/lodging/stay/property',
      );
      expect(service.requests.single['query'], {'share': 'guest-link'});
      expect(service.images.first['url'], contains('/photos/upload'));
      expect(tester.widget<RawImage>(find.byType(RawImage)).image, isNotNull);
      expect(find.text('Casa da Colina'), findsOneWidget);
      expect(find.textContaining('7 nights'), findsOneWidget);
      expect(find.text('Directions'), findsOneWidget);
      expect(find.text('2 photos'), findsOneWidget);
      expect(find.text('Google Maps'), findsOneWidget);
      await _tap(tester, find.text('2 photos'));
      expect(find.text('Your stay'), findsOneWidget);
      expect(find.text('1 / 2'), findsOneWidget);
      await _tap(tester, find.byTooltip('Next photo'));
      expect(find.text('2 / 2'), findsOneWidget);
      expect(find.text('Ana Photographer'), findsOneWidget);
      expect(find.text('Photo source'), findsOneWidget);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'manual stay avoids lookup and retains permitted arrival and host details',
    (tester) async {
      final service = _Service();
      final entry = _stay()..remove('property');
      entry['privateDetails'] = {
        'address': 'Private flat, floor 3',
        'arrivalInstructions': 'Collect keys from the host.',
      };
      entry['hostNotes'] = 'Paid the first night.';
      entry['roomsOrBeds'] = '3 bedrooms';
      entry['cost'] = {'currency': 'EUR', 'amount': 1200, 'basis': 'total'};
      await _show(tester, service, entry);
      expect(service.requests, isEmpty);
      expect(service.images, isEmpty);
      expect(find.text('Google Maps'), findsNothing);
      expect(find.text('Directions'), findsNothing);
      await _tap(tester, find.text('Stay details'));
      expect(find.text('Private flat, floor 3'), findsOneWidget);
      expect(find.text('Collect keys from the host.'), findsOneWidget);
      expect(find.text('Host notes: Paid the first night.'), findsOneWidget);
      expect(find.text('Estimate: EUR 1200 total'), findsOneWidget);
    },
  );

  testWidgets('provider failure leaves host details and original link usable', (
    tester,
  ) async {
    final service = _Service()..fail = true;
    await _show(tester, service, _stay());
    expect(find.text('Casa da Colina'), findsOneWidget);
    expect(find.text('View property'), findsOneWidget);
    expect(find.textContaining('7 nights'), findsOneWidget);
    expect(find.byType(CircularProgressIndicator), findsNothing);
    await _tap(tester, find.text('Stay details'));
    expect(find.text('Your stay'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('late lookup cannot decorate a different lodging entry', (
    tester,
  ) async {
    final service = _Service()..defer = true;
    await _show(tester, service, _stay());
    expect(service.pending, hasLength(1));
    final changed = _stay()
      ..['id'] = 'other-stay'
      ..['name'] = 'Kyoto Guesthouse'
      ..['generalLocation'] = 'Kyoto, Japan'
      ..['property'] = {'provider': 'google', 'placeId': 'kyoto-house'};
    await _show(tester, service, changed);
    expect(service.pending, hasLength(2));
    service.pending.first.complete(_property());
    await _settle(tester);
    expect(find.text('Kyoto Guesthouse'), findsOneWidget);
    expect(find.text('Directions'), findsNothing);
    expect(service.images, isEmpty);
    service.pending.last.complete({
      'property': {
        'name': 'Kyoto Guesthouse',
        'address': 'Kyoto',
        'rating': 4.2,
        'ratingCount': 9,
      },
    });
    await _settle(tester);
    expect(find.textContaining('4.2'), findsOneWidget);
    expect(find.textContaining('4.7'), findsNothing);
  });

  testWidgets(
    'draft and access changes reset live property and close stale details',
    (tester) async {
      final service = _Service();
      final boundary = GlobalKey();
      await _show(
        tester,
        service,
        _stay(),
        boundary: boundary,
        accessQuery: {'share': 'first'},
      );
      await _tap(tester, find.text('Stay details'));
      expect(find.text('Your stay'), findsOneWidget);
      await _show(
        tester,
        service,
        _stay(),
        boundary: boundary,
        draft: true,
        accessQuery: {'share': 'second'},
      );
      expect(find.text('Your stay'), findsNothing);
      expect(service.requests.last['query'], {
        'share': 'second',
        'draft': true,
      });
      expect(service.images.last['query'], {'share': 'second', 'draft': true});
    },
  );

  testWidgets(
    'account change closes detail gallery and discards live property',
    (tester) async {
      final service = _Service();
      final auth = TripTestAuth(User(id: 'host'));
      addTearDown(auth.dispose);
      await _show(tester, service, _stay(), auth: auth);
      await _tap(tester, find.text('Stay details'));
      expect(find.text('Your stay'), findsOneWidget);
      auth.change(User(id: 'other'));
      await _settle(tester);
      expect(find.text('Your stay'), findsNothing);
      expect(find.text('Google Maps'), findsNothing);
      expect(find.byType(LodgingPhoto), findsNothing);
    },
  );

  for (final (width, scale) in [(390.0, 1.0), (320.0, 2.0)]) {
    testWidgets('card and details fit ${width.toInt()}px at ${scale}x text', (
      tester,
    ) async {
      await tester.binding.setSurfaceSize(Size(width, 844));
      addTearDown(() => tester.binding.setSurfaceSize(null));
      final boundary = GlobalKey();
      final service = _Service();
      await _show(tester, service, _stay(), boundary: boundary, scale: scale);
      expect(tester.takeException(), isNull);
      if (const bool.fromEnvironment('TRIP_SCREENSHOTS')) {
        await tester.runAsync(() async {
          final image =
              await (boundary.currentContext!.findRenderObject()!
                      as RenderRepaintBoundary)
                  .toImage();
          final bytes = await image.toByteData(format: ui.ImageByteFormat.png);
          await File(
            '/private/tmp/travel-angels-lodging-${width.toInt()}-${scale}x.png',
          ).writeAsBytes(bytes!.buffer.asUint8List());
          image.dispose();
        });
      }
      await _tap(tester, find.text('Stay details'));
      expect(
        find.byTooltip('Close stay details').hitTestable(),
        findsOneWidget,
      );
      expect(tester.takeException(), isNull);
    });
  }
}
