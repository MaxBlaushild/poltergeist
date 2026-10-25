import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'support/app_fonts.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/screens/travel_trip_screen.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/auth_service.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/widgets/travel_trip_hosts.dart';
import 'package:travel_angels/widgets/travel_trip_logistics.dart';
import 'package:travel_angels/widgets/travel_trip_settings.dart';

Map<String, dynamic> sampleTrip() => {
  'id': 'trip',
  'ownerId': 'owner',
  'calendarId': 'calendar',
  'revision': 7,
  'title': 'A slow week in Lisbon',
  'destination': 'Lisbon, Portugal',
  'description':
      'Long lunches, tiled streets, and a little Atlantic air.\nCome for the week or meet us for a few days. We’ll leave room for getting lost.',
  'startDate': '2099-06-01',
  'endDate': '2099-06-08',
  'datePrecision': 'fixed',
  'timeZone': 'Europe/Lisbon',
  'published': true,
  'joinOpen': true,
  'acceptingParticipation': true,
  'status': 'confirmed',
  'canManage': false,
  'canManagePermissions': false,
  'hostModel': true,
  'rosterVersion': 3,
  'canViewAttendees': false,
  'hosts': [
    {
      'userId': 'owner',
      'displayName': 'Max',
      'isOwner': true,
      'isAttending': true,
    },
    {
      'userId': 'host',
      'displayName': 'Jamie',
      'isOwner': false,
      'isAttending': false,
    },
  ],
  'attendees': [],
  'tripSettings': {
    'maxAttendees': 8,
    'showAttendeeCount': true,
    'rosterVisibility': 'attendees',
    'waitlistEnabled': true,
  },
  'capacity': {
    'full': false,
    'showCount': true,
    'confirmedCount': 4,
    'maximum': 8,
    'remainingPlaces': 4,
    'waitlistEnabled': true,
  },
  'lodging': [
    {
      'id': 'stay',
      'name': 'Casa da Colina',
      'type': 'rental',
      'generalLocation': 'Graça, Lisbon',
      'kind': 'group',
      'visibility': 'viewers',
      'status': 'selected',
      'datesTbd': false,
      'checkInDate': '2099-06-01',
      'checkOutDate': '2099-06-08',
      'timeZone': 'Europe/Lisbon',
      'description': 'A sunny base with a terrace above the rooftops.',
    },
  ],
  'flights': [
    {
      'id': 'flight',
      'label': 'Overnight to Lisbon',
      'kind': 'recommendation',
      'visibility': 'viewers',
      'status': 'suggested',
      'legs': [
        {
          'airline': 'TAP',
          'flightNumber': 'TP208',
          'departureAirport': 'JFK',
          'arrivalAirport': 'LIS',
          'departureDate': '2099-05-31',
          'departureTime': '22:00',
          'departureTimeZone': 'America/New_York',
          'arrivalDate': '2099-06-01',
          'arrivalTime': '10:00',
          'arrivalTimeZone': 'Europe/Lisbon',
        },
      ],
    },
  ],
};

class TripTestService extends TravelCalendarService {
  final requests = <Map<String, dynamic>>[];
  bool failWrite = false;
  Map<String, dynamic> preview = {};
  List<Map<String, dynamic>> participations = [];

  @override
  Future<Map<String, dynamic>> get(
    String path, {
    Map<String, dynamic>? query,
  }) async => path.endsWith('/participations')
      ? {'participations': participations}
      : {};

  Future<Map<String, dynamic>> _write(
    String method,
    String path,
    Map<String, dynamic>? data,
  ) async {
    requests.add({
      'method': method,
      'path': path,
      'data': jsonDecode(jsonEncode(data)),
    });
    if (failWrite) {
      throw const TravelCalendarException(
        'The trip changed. Refresh and review your draft.',
        statusCode: 409,
      );
    }
    if (path.endsWith('/preview')) return preview;
    return {};
  }

  @override
  Future<Map<String, dynamic>> post(
    String path, {
    Map<String, dynamic>? data,
  }) => _write('POST', path, data);
  @override
  Future<Map<String, dynamic>> put(String path, {Map<String, dynamic>? data}) =>
      _write('PUT', path, data);
}

class TripTestAuth extends AuthProvider {
  TripTestAuth(this.account)
    : super(AuthService(APIClient('https://example.test')));
  User? account;
  @override
  User? get user => account;
  @override
  bool get loading => false;
  @override
  Future<void> verifyToken() async {}
  void change(User? user) {
    account = user;
    notifyListeners();
  }
}

Future<void> showTestDialog(WidgetTester tester, Widget dialog) async {
  await tester.pumpWidget(
    MaterialApp(
      theme: AppTheme.lightTheme,
      home: Scaffold(
        body: Builder(
          builder: (context) => TextButton(
            onPressed: () =>
                showDialog<void>(context: context, builder: (_) => dialog),
            child: const Text('Open'),
          ),
        ),
      ),
    ),
  );
  await tester.tap(find.text('Open'));
  await tester.pumpAndSettle();
}

void main() {
  setUpAll(loadAppFonts);
  testWidgets(
    'guest page hides identities and exact counts while showing permitted planning content',
    (tester) async {
      await tester.binding.setSurfaceSize(const Size(1200, 1000));
      addTearDown(() => tester.binding.setSurfaceSize(null));
      final stop = sampleTrip()
        ..['attendees'] = [
          {'displayName': 'Private traveler'},
        ]
        ..['capacity'] = {
          'showCount': false,
          'full': false,
          'confirmedCount': 981,
          'maximum': 999,
          'remainingPlaces': 18,
        };
      await tester.pumpWidget(
        MaterialApp(
          theme: AppTheme.lightTheme,
          home: TravelTripScreen(
            initialStop: stop,
            service: TripTestService(),
            loadStop: () async => stop,
            onAction: (_, _) async {},
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text('A slow week in Lisbon'), findsOneWidget);
      expect(find.text('Private traveler'), findsNothing);
      expect(find.textContaining('981'), findsNothing);
      expect(find.textContaining('999'), findsNothing);
      expect(find.text('18 places available'), findsNothing);
      expect(find.text('Manage trip'), findsNothing);
      await tester.tap(find.text('Lodging'));
      await tester.pumpAndSettle();
      expect(find.text('Casa da Colina'), findsOneWidget);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'full trips offer waitlist and pending travelers can edit their date selection',
    (tester) async {
      await tester.binding.setSurfaceSize(const Size(1200, 1000));
      addTearDown(() => tester.binding.setSurfaceSize(null));
      for (final status in ['new', 'requested', 'waitlisted']) {
        final stop = sampleTrip()
          ..['capacity'] = {
            'full': true,
            'waitlistEnabled': true,
            'showCount': false,
          };
        if (status != 'new') stop['ownParticipation'] = {'status': status};
        final calls = <String>[];
        await tester.pumpWidget(
          MaterialApp(
            theme: AppTheme.lightTheme,
            home: TravelTripScreen(
              key: ValueKey(status),
              initialStop: stop,
              service: TripTestService(),
              loadStop: () async => stop,
              onAction: (action, _) async => calls.add(action),
            ),
          ),
        );
        await tester.pumpAndSettle();
        final label = status == 'requested'
            ? 'Update request'
            : status == 'waitlisted'
            ? 'Update waitlist dates'
            : 'Join the waitlist';
        await tester.tap(find.text(label));
        await tester.pumpAndSettle();
        expect(calls, [status == 'requested' ? 'request' : 'waitlist']);
      }
    },
  );

  testWidgets(
    'host attendance requires acknowledgement and preserves partial dates',
    (tester) async {
      await tester.binding.setSurfaceSize(const Size(900, 1100));
      addTearDown(() => tester.binding.setSurfaceSize(null));
      final service = TripTestService();
      final stop = sampleTrip()
        ..['ownParticipation'] = {
          'status': 'joined',
          'startDate': '2099-06-03',
          'endDate': '2099-06-05',
          'shareDates': false,
        };
      await showTestDialog(
        tester,
        TravelTripConsentDialog(
          stop: stop,
          service: service,
          hostAttendance: true,
        ),
      );
      await tester.tap(find.text('Save changes'));
      await tester.pumpAndSettle();
      expect(service.requests, isEmpty);
      await tester.tap(
        find.text('I understand who can see my name and avatar'),
      );
      await tester.enterText(
        find.widgetWithText(TextField, 'My departure date'),
        '2099-06-06',
      );
      await tester.tap(find.text('Save changes'));
      await tester.pumpAndSettle();
      expect(
        service.requests.single['data'],
        containsPair('startDate', '2099-06-03'),
      );
      expect(
        service.requests.single['data'],
        containsPair('endDate', '2099-06-06'),
      );
      expect(
        service.requests.single['data'],
        containsPair('rosterConsentVersion', 3),
      );
      expect(
        service.requests.single['data'],
        containsPair('shareDates', false),
      );
    },
  );

  testWidgets(
    'settings apply immediately and failed revision keeps the edited maximum',
    (tester) async {
      await tester.binding.setSurfaceSize(const Size(900, 1100));
      addTearDown(() => tester.binding.setSurfaceSize(null));
      final service = TripTestService()..failWrite = true;
      await showTestDialog(
        tester,
        TravelTripSettingsDialog(stop: sampleTrip(), service: service),
      );
      await tester.enterText(find.byType(TextField), '12');
      await tester.tap(find.text('Accept new attendance requests'));
      await tester.tap(find.text('Apply settings now'));
      await tester.pumpAndSettle();
      expect(service.requests.single['path'], '/calendar/stops/trip/settings');
      expect(service.requests.single['data'], containsPair('maxAttendees', 12));
      expect(service.requests.single['data'], containsPair('joinOpen', false));
      expect(find.text('12'), findsOneWidget);
      expect(
        find.text('The trip changed. Refresh and review your draft.'),
        findsOneWidget,
      );
    },
  );

  testWidgets(
    'lodging edits retain unrelated draft fields and reserved traveler assignments',
    (tester) async {
      await tester.binding.setSurfaceSize(const Size(1000, 1500));
      addTearDown(() => tester.binding.setSurfaceSize(null));
      final stop = sampleTrip();
      stop['draft'] = {
        ...stop,
        'title': 'Saved unpublished title',
        'hero': {'mediaId': 'photo'},
        'description': 'Private unfinished draft',
      };
      final service = TripTestService()
        ..participations = [
          {
            'userId': 'reserved',
            'displayName': 'Reserved traveler',
            'status': 'needs_reconfirmation',
            'reserved': true,
          },
        ];
      await showTestDialog(
        tester,
        TravelTripLogisticsEditor(
          stop: stop,
          service: service,
          kind: 'lodging',
          entry: (stop['lodging'] as List).first,
        ),
      );
      final nameField = find.widgetWithText(TextFormField, 'Property name');
      await tester.enterText(nameField, 'Casa with a terrace');
      await tester.ensureVisible(find.text('Reserved traveler'));
      await tester.tap(find.text('Reserved traveler'));
      await tester.tap(find.text('Save logistics draft'));
      await tester.pumpAndSettle();
      final data = service.requests.single['data'] as Map;
      expect(data['title'], 'Saved unpublished title');
      expect(data['hero'], {'mediaId': 'photo'});
      expect(data['flights'], stop['flights']);
      expect((data['lodging'] as List).first['assignedUserIds'], ['reserved']);
      expect(data['revision'], 7);
    },
  );

  testWidgets(
    'ownership transfer confirms the exact previewed audience fingerprint',
    (tester) async {
      final service = TripTestService()
        ..preview = {
          'audienceFingerprint': 'reviewed-audience',
          'audienceChange': {
            'fromVisibility': 'private',
            'toVisibility': 'specific',
            'addedViewerCount': 1,
            'removedViewerCount': 0,
            'addedViewers': [
              {'displayName': 'New viewer'},
            ],
          },
        };
      await showTestDialog(
        tester,
        TravelTripTransferDialog(stop: sampleTrip(), service: service),
      );
      await tester.tap(find.byType(DropdownButtonFormField<String>));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Jamie').last);
      await tester.pumpAndSettle();
      await tester.tap(find.text('Review audience impact'));
      await tester.pumpAndSettle();
      expect(find.text('Gains access: New viewer'), findsOneWidget);
      await tester.tap(find.widgetWithText(FilledButton, 'Transfer ownership'));
      await tester.pumpAndSettle();
      expect(
        service.requests.last['data'],
        containsPair('audienceFingerprint', 'reviewed-audience'),
      );
    },
  );

  testWidgets(
    'revocation while a host dialog is open dismisses protected content',
    (tester) async {
      final stop = sampleTrip()
        ..['canManage'] = true
        ..['canManagePermissions'] = true;
      var denied = false;
      await tester.pumpWidget(
        MaterialApp(
          theme: AppTheme.lightTheme,
          home: TravelTripScreen(
            initialStop: stop,
            service: TripTestService(),
            loadStop: () async {
              if (denied) {
                throw const TravelCalendarException('Denied', statusCode: 403);
              }
              return stop;
            },
            onAction: (_, _) async {},
          ),
        ),
      );
      await tester.pumpAndSettle();
      await tester.tap(find.text('Manage trip'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Attendance & privacy'));
      await tester.pumpAndSettle();
      expect(find.byType(TravelTripSettingsDialog), findsOneWidget);
      denied = true;
      await tester.pump(const Duration(seconds: 31));
      await tester.pumpAndSettle();
      expect(find.byType(TravelTripSettingsDialog), findsNothing);
      expect(find.text('A slow week in Lisbon'), findsNothing);
      expect(
        find.textContaining('not available to your current account'),
        findsOneWidget,
      );
    },
  );

  testWidgets(
    'account changes clear protected page before refreshed access resolves',
    (tester) async {
      final auth = TripTestAuth(User(id: 'owner'));
      final stop = sampleTrip()
        ..['canViewAttendees'] = true
        ..['attendees'] = [
          {'displayName': 'Private traveler'},
        ];
      final pending = Completer<Map<String, dynamic>>();
      var switched = false;
      await tester.pumpWidget(
        ChangeNotifierProvider<AuthProvider>.value(
          value: auth,
          child: MaterialApp(
            theme: AppTheme.lightTheme,
            home: TravelTripScreen(
              initialStop: stop,
              service: TripTestService(),
              loadStop: () => switched ? pending.future : Future.value(stop),
              onAction: (_, _) async {},
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      switched = true;
      auth.change(null);
      await tester.pump();
      expect(find.text('Private traveler'), findsNothing);
      expect(find.text('A slow week in Lisbon'), findsNothing);
      pending.complete(sampleTrip());
      await tester.pumpAndSettle();
    },
  );

  testWidgets('mobile and desktop trip page render without overflow', (
    tester,
  ) async {
    const screenshots = bool.fromEnvironment('TRIP_SCREENSHOTS');
    addTearDown(() => tester.binding.setSurfaceSize(null));
    for (final size in [const Size(1440, 1000), const Size(390, 844)]) {
      await tester.binding.setSurfaceSize(size);
      final boundary = GlobalKey();
      final stop = sampleTrip();
      await tester.pumpWidget(
        MaterialApp(
          theme: AppTheme.lightTheme,
          home: RepaintBoundary(
            key: boundary,
            child: TravelTripScreen(
              key: ValueKey(size),
              initialStop: stop,
              service: TripTestService(),
              loadStop: () async => stop,
              onAction: (_, _) async {},
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
      if (screenshots) {
        await tester.runAsync(() async {
          final image =
              await (boundary.currentContext!.findRenderObject()!
                      as RenderRepaintBoundary)
                  .toImage();
          final bytes = await image.toByteData(format: ui.ImageByteFormat.png);
          await File(
            '/private/tmp/travel-trip-${size.width.toInt()}.png',
          ).writeAsBytes(bytes!.buffer.asUint8List());
          image.dispose();
        });
      }
      await tester.tap(find.text('Flights'));
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
      if (size.width < 500) {
        await tester.pumpWidget(
          MaterialApp(
            theme: AppTheme.lightTheme,
            builder: (_, child) => MediaQuery(
              data: const MediaQueryData(textScaler: TextScaler.linear(2)),
              child: child!,
            ),
            home: TravelTripScreen(
              key: UniqueKey(),
              initialStop: stop,
              service: TripTestService(),
              loadStop: () async => stop,
              onAction: (_, _) async {},
            ),
          ),
        );
        await tester.pumpAndSettle();
        expect(tester.takeException(), isNull);
      }
    }
  });
}
