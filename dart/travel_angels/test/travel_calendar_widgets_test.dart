import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/auth_service.dart';
import 'package:travel_angels/screens/travel_calendar_screen.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/widgets/travel_calendar_dialogs.dart';
import 'package:travel_angels/widgets/travel_calendar_delivery.dart';
import 'package:travel_angels/widgets/travel_calendar_views.dart';

Map<String, dynamic> stop(
  String id,
  String title, {
  String precision = 'fixed',
  String start = '2099-01-01',
  String end = '2099-01-10',
}) => {
  'id': id,
  'calendarId': 'calendar',
  'title': title,
  'destination': title,
  'startDate': start,
  'endDate': end,
  'datePrecision': precision,
  'status': 'confirmed',
  'published': true,
  'joinOpen': true,
  'canManage': false,
  'canManagePermissions': false,
};

class FakeTravelCalendarService extends TravelCalendarService {
  List<TravelStop> stops = [stop('a', 'Lisbon'), stop('b', 'Kyoto')];
  final List<Map<String, dynamic>> requests = [];
  int removals = 1;
  bool failSecondParticipation = false;
  bool previewVisible = false;
  Map<String, dynamic> notificationResponse = {};

  @override
  Future<Map<String, dynamic>> get(
    String path, {
    Map<String, dynamic>? query,
  }) async {
    if (path == '/calendar-notifications') return notificationResponse;
    if (path.endsWith('/permissions')) {
      return {
        'visibility': 'specific',
        'grants': [
          {'email': 'friend@example.com', 'recipientId': 'friend'},
        ],
        'collaborators': [],
      };
    }
    return {
      'calendar': {
        'id': 'calendar',
        'title': 'Our travel calendar',
        'canManage': false,
        'canManagePermissions': false,
      },
      'stops': stops,
    };
  }

  @override
  Future<Map<String, dynamic>> post(
    String path, {
    Map<String, dynamic>? data,
  }) async {
    requests.add({'method': 'POST', 'path': path, 'data': data});
    if (path == '/calendar-notifications/retry') return {'pending': 1};
    if (path.endsWith('/permissions/preview')) {
      return {'removedParticipationCount': removals};
    }
    if (path.endsWith('/preview')) {
      return {
        'canView': previewVisible,
        if (previewVisible) 'stop': stops.first,
      };
    }
    if (failSecondParticipation && path == '/calendar/stops/b/participation') {
      throw const TravelCalendarException('Dates have changed.');
    }
    return {};
  }

  @override
  Future<Map<String, dynamic>> put(
    String path, {
    Map<String, dynamic>? data,
  }) async {
    requests.add({'method': 'PUT', 'path': path, 'data': data});
    return {};
  }
}

class CalendarTestAuth extends AuthProvider {
  CalendarTestAuth(this._account)
    : super(AuthService(APIClient('https://example.test')));
  User? _account;

  @override
  User? get user => _account;

  @override
  bool get isAuthenticated => _account != null;

  @override
  bool get loading => false;

  @override
  Future<void> verifyToken() async {}

  void setAccount(User? account) {
    _account = account;
    notifyListeners();
  }
}

void main() {
  setUp(() => SharedPreferences.setMockInitialValues({}));

  testWidgets(
    'draft editor saves destination time zone without converting local dates',
    (tester) async {
      final service = FakeTravelCalendarService();
      await tester.pumpWidget(
        MaterialApp(
          home: Builder(
            builder: (context) => Scaffold(
              body: TextButton(
                onPressed: () => showDialog<void>(
                  context: context,
                  builder: (_) => TravelStopEditor(
                    service: service,
                    stop: stop('a', 'Auckland'),
                  ),
                ),
                child: const Text('Edit stop'),
              ),
            ),
          ),
        ),
      );
      await tester.tap(find.text('Edit stop'));
      await tester.pumpAndSettle();
      await tester.enterText(
        find.byWidgetPredicate(
          (widget) =>
              widget is TextField &&
              widget.decoration?.labelText == 'Destination time zone',
        ),
        'Pacific/Auckland',
      );
      await tester.tap(find.text('Save draft'));
      await tester.pumpAndSettle();
      expect(
        service.requests.single['data'],
        containsPair('timeZone', 'Pacific/Auckland'),
      );
      expect(
        service.requests.single['data'],
        containsPair('startDate', '2099-01-01'),
      );
      expect(
        service.requests.single['data'],
        containsPair('endDate', '2099-01-10'),
      );
    },
  );

  testWidgets(
    'expired guest viewing preserves selections without rendering restricted details',
    (tester) async {
      SharedPreferences.setMockInitialValues({
        TravelCalendarService.guestEmailKey: 'friend@example.com',
        TravelCalendarService.guestTokenKey: 'expired-token',
        TravelCalendarService.guestExpiryKey: '2000-01-01T00:00:00Z',
        'travel_calendar_selections:share:guest': jsonEncode({
          'a': {
            'stopId': 'a',
            'status': 'requested',
            'startDate': '2099-01-03',
            'endDate': '2099-01-07',
          },
        }),
      });
      final service = FakeTravelCalendarService()..stops = [];
      await tester.pumpWidget(
        MaterialApp(
          home: TravelCalendarScreen(calendarToken: 'share', service: service),
        ),
      );
      await tester.pumpAndSettle();
      final prefs = await SharedPreferences.getInstance();
      expect(
        prefs.getString('travel_calendar_selections:share:guest'),
        contains('2099-01-03'),
      );
      expect(
        find.text(
          'Verify your invited email to restore saved selections for restricted plans.',
        ),
        findsOneWidget,
      );
      expect(find.textContaining('2099-01-03'), findsNothing);
      expect(find.text('1 stop selected · nothing submitted'), findsNothing);
      expect(find.text('Lisbon'), findsNothing);
      // A fresh email verification restores the same viewing identity.
      await prefs.setString(TravelCalendarService.guestTokenKey, 'fresh-token');
      await prefs.setString(
        TravelCalendarService.guestExpiryKey,
        '2099-12-31T00:00:00Z',
      );
      service.stops = [stop('a', 'Lisbon')];
      await tester.tap(find.byTooltip('Refresh calendar'));
      await tester.pumpAndSettle();
      expect(find.text('1 stop selected · nothing submitted'), findsOneWidget);
      expect(find.text('Restore saved selections'), findsNothing);
      expect(
        prefs.getString('travel_calendar_selections:share:guest'),
        contains('2099-01-03'),
      );
      // A current verified session can now establish actual loss of access.
      service.stops = [];
      await tester.tap(find.byTooltip('Refresh calendar'));
      await tester.pumpAndSettle();
      expect(prefs.getString('travel_calendar_selections:share:guest'), isNull);
      expect(
        find.textContaining('Some selected plans are no longer available.'),
        findsOneWidget,
      );
    },
  );

  testWidgets(
    'expired historical guest session does not preserve revoked claimed-account selections',
    (tester) async {
      SharedPreferences.setMockInitialValues({
        TravelCalendarService.guestEmailKey: 'friend@example.com',
        TravelCalendarService.guestClaimedAccountKey: 'account-a',
        TravelCalendarService.guestTokenKey: 'expired-token',
        TravelCalendarService.guestExpiryKey: '2000-01-01T00:00:00Z',
        'travel_calendar_selections:share:account-a': jsonEncode({
          'a': {
            'stopId': 'a',
            'status': 'requested',
            'startDate': '2099-01-03',
            'endDate': '2099-01-07',
          },
        }),
      });
      final auth = CalendarTestAuth(User(id: 'account-a'));
      final service = FakeTravelCalendarService()..stops = [];
      await tester.pumpWidget(
        ChangeNotifierProvider<AuthProvider>.value(
          value: auth,
          child: MaterialApp(
            home: TravelCalendarScreen(
              calendarToken: 'share',
              service: service,
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      final prefs = await SharedPreferences.getInstance();
      expect(
        prefs.getString('travel_calendar_selections:share:account-a'),
        isNull,
      );
      expect(find.text('Restore saved selections'), findsNothing);
    },
  );

  testWidgets(
    'logging out removes manager data and does not transfer account selections',
    (tester) async {
      SharedPreferences.setMockInitialValues({
        'travel_calendar_selections:share:account-a': jsonEncode({
          'a': {
            'stopId': 'a',
            'status': 'interested',
            'startDate': '',
            'endDate': '',
          },
        }),
      });
      final auth = CalendarTestAuth(User(id: 'account-a'));
      final service = FakeTravelCalendarService()
        ..stops = [
          {
            ...stop('a', 'Private host plan'),
            'canManage': true,
            'visibility': 'private',
          },
        ];
      await tester.pumpWidget(
        ChangeNotifierProvider<AuthProvider>.value(
          value: auth,
          child: MaterialApp(
            home: TravelCalendarScreen(
              calendarToken: 'share',
              service: service,
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text('1 stop selected · nothing submitted'), findsOneWidget);
      service.stops = [stop('public', 'Public plans')];
      auth.setAccount(null);
      await tester.pumpAndSettle();
      expect(find.text('Private host plan'), findsNothing);
      expect(find.text('1 stop selected · nothing submitted'), findsNothing);
      final prefs = await SharedPreferences.getInstance();
      expect(
        prefs.getString('travel_calendar_selections:share:account-a'),
        isNotNull,
      );
      expect(prefs.getString('travel_calendar_selections:share:guest'), isNull);
    },
  );

  testWidgets(
    'cancelled claim retains selections until account association succeeds',
    (tester) async {
      SharedPreferences.setMockInitialValues({
        'travel_calendar_selections:share:guest': jsonEncode({
          'a': {
            'stopId': 'a',
            'status': 'interested',
            'startDate': '',
            'endDate': '',
          },
        }),
      });
      final auth = CalendarTestAuth(null);
      final service = FakeTravelCalendarService();
      var success = false;
      await tester.pumpWidget(
        ChangeNotifierProvider<AuthProvider>.value(
          value: auth,
          child: MaterialApp(
            home: TravelCalendarScreen(
              calendarToken: 'share',
              service: service,
              claimAccount: (_, _) async {
                auth.setAccount(User(id: 'account-a'));
                return success;
              },
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      await tester.drag(find.byType(ListView).first, const Offset(0, -350));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Claim account & review selections'));
      await tester.pumpAndSettle();
      expect(find.text('1 stop selected · nothing submitted'), findsOneWidget);
      final prefs = await SharedPreferences.getInstance();
      expect(
        prefs.getString('travel_calendar_selections:share:guest'),
        isNotNull,
      );
      expect(
        prefs.getString('travel_calendar_selections:share:account-a'),
        isNull,
      );
      success = true;
      await tester.ensureVisible(find.text('Review selections'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Review selections'));
      await tester.pumpAndSettle();
      expect(find.text('Review your selections'), findsOneWidget);
      expect(service.requests, isEmpty);
      expect(prefs.getString('travel_calendar_selections:share:guest'), isNull);
      expect(
        prefs.getString('travel_calendar_selections:share:account-a'),
        isNotNull,
      );
    },
  );

  testWidgets(
    'notification panel is read-only until own failed notices are retried',
    (tester) async {
      final service = FakeTravelCalendarService()
        ..notificationResponse = {
          'notifications': [
            {
              'id': 'notice',
              'event': 'access_revoked',
              'message': 'Access to a travel plan ended.',
              'createdAt': '2099-01-01T12:00:00Z',
            },
          ],
          'deliveries': [
            {
              'id': 'delivery',
              'status': 'failed',
              'event': 'participation_requested',
              'createdAt': '2099-01-01T12:00:00Z',
            },
          ],
        };
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: TravelCalendarNotificationsDialog(service: service),
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text('Access to a travel plan ended.'), findsOneWidget);
      expect(service.requests, isEmpty);
      expect(
        tester.widget<ListTile>(find.byType(ListTile).first).onTap,
        isNull,
      );
      await tester.tap(find.text('Retry my failed email notices'));
      await tester.pumpAndSettle();
      expect(service.requests.single['path'], '/calendar-notifications/retry');
      expect(
        find.text('1 failed email notice(s) queued to retry.'),
        findsOneWidget,
      );
    },
  );

  test('date and destination filters include overlaps across years', () {
    final plans = [
      stop('a', 'Lisbon', start: '2026-12-25', end: '2027-01-10'),
      stop('b', 'Kyoto', start: '2027-02-01', end: '2027-02-05'),
      stop('c', 'Maybe Seoul', precision: 'tbd', start: '', end: ''),
    ];
    expect(
      filterTravelStops(
        plans,
        destination: 'lis',
        range: DateTimeRange(
          start: DateTime(2027, 1, 1),
          end: DateTime(2027, 1, 5),
        ),
      ).map((s) => s['id']),
      ['a'],
    );
    expect(
      filterTravelStops(
        plans,
        period: 'current',
        today: DateTime(2027, 1, 1),
      ).map((s) => s['id']),
      ['a'],
    );
    expect(
      filterTravelStops(plans, period: 'unscheduled').map((s) => s['id']),
      ['c'],
    );
  });

  testWidgets(
    'shared plans browse without an authentication provider and filter destinations',
    (tester) async {
      final service = FakeTravelCalendarService();
      await tester.pumpWidget(
        MaterialApp(
          home: TravelCalendarScreen(calendarToken: 'share', service: service),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text('Our travel calendar'), findsOneWidget);
      expect(find.text('Add stop'), findsNothing);
      await tester.enterText(find.byType(TextField).first, 'Kyoto');
      await tester.pump();
      expect(
        find.byWidgetPredicate(
          (widget) =>
              widget is TravelStopCard && widget.stop['title'] == 'Kyoto',
        ),
        findsOneWidget,
      );
      expect(
        find.byWidgetPredicate(
          (widget) =>
              widget is TravelStopCard && widget.stop['title'] == 'Lisbon',
        ),
        findsNothing,
      );
      expect(service.requests, isEmpty);
    },
  );

  testWidgets(
    'month navigation crosses years and approximate plans never occupy exact days',
    (tester) async {
      tester.view.resetPhysicalSize();
      tester.view.physicalSize = const Size(320, 800);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      DateTime? next;
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SingleChildScrollView(
              child: TravelCalendarMonth(
                month: DateTime(2026, 12),
                stops: [
                  stop(
                    'a',
                    'Tentative Paris',
                    precision: 'approximate',
                    start: '2026-12-01',
                    end: '2026-12-10',
                  ),
                ],
                onMonthChanged: (month) => next = month,
                onStopTap: (_) {},
              ),
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text('December 2026'), findsOneWidget);
      expect(find.text('1 stop'), findsNothing);
      expect(find.text('Approximate & unscheduled plans'), findsOneWidget);
      await tester.tap(find.byTooltip('Next month'));
      expect(next, DateTime(2027, 1));
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'visitor preview does not render details when the audience lacks access',
    (tester) async {
      final service = FakeTravelCalendarService();
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: TravelStopPreviewDialog(
              service: service,
              stop: service.stops.first,
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(
        find.textContaining('This visitor cannot view the stop'),
        findsOneWidget,
      );
      expect(find.text('Lisbon'), findsNothing);
      expect(service.requests.single['data'], {'audience': 'anonymous'});
    },
  );

  testWidgets(
    'guests can choose a stop before any authentication or RSVP request',
    (tester) async {
      final service = FakeTravelCalendarService()
        ..stops = [stop('a', 'Lisbon')];
      await tester.pumpWidget(
        MaterialApp(
          home: TravelCalendarScreen(calendarToken: 'share', service: service),
        ),
      );
      await tester.pumpAndSettle();
      await tester.drag(find.byType(ListView).first, const Offset(0, -350));
      await tester.pumpAndSettle();
      await tester.tap(find.byType(TravelStopCard));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Choose dates / interest'));
      await tester.pumpAndSettle();
      expect(find.text('Add to selections'), findsOneWidget);
      await tester.tap(find.text('Add to selections'));
      await tester.pumpAndSettle();
      expect(service.requests, isEmpty);
      final prefs = await SharedPreferences.getInstance();
      final saved =
          jsonDecode(prefs.getString('travel_calendar_selections:share:guest')!)
              as Map;
      expect(saved['a'], containsPair('status', 'interested'));
      expect(saved['a'], containsPair('startDate', '2099-01-01'));
    },
  );

  testWidgets(
    'claiming preserves partial dates and requires explicit submission',
    (tester) async {
      SharedPreferences.setMockInitialValues({
        'travel_calendar_selections:share:guest': jsonEncode({
          'a': {
            'stopId': 'a',
            'status': 'requested',
            'startDate': '2099-01-03',
            'endDate': '2099-01-07',
          },
        }),
      });
      final service = FakeTravelCalendarService();
      var claims = 0;
      await tester.pumpWidget(
        MaterialApp(
          home: TravelCalendarScreen(
            calendarToken: 'share',
            service: service,
            claimAccount: (_, _) async {
              claims++;
              return true;
            },
          ),
        ),
      );
      await tester.pumpAndSettle();
      await tester.ensureVisible(
        find.text('Claim account & review selections'),
      );
      await tester.pumpAndSettle();
      await tester.tap(find.text('Claim account & review selections'));
      await tester.pumpAndSettle();
      expect(claims, 1);
      expect(find.text('Review your selections'), findsOneWidget);
      expect(service.requests, isEmpty);
      await tester.tap(find.text('Submit selections'));
      await tester.pumpAndSettle();
      expect(service.requests, hasLength(1));
      expect(
        service.requests.single['data'],
        containsPair('startDate', '2099-01-03'),
      );
      expect(
        service.requests.single['data'],
        containsPair('endDate', '2099-01-07'),
      );
      expect(
        service.requests.single['data'],
        containsPair('status', 'requested'),
      );
      expect(
        service.requests.single['data'],
        containsPair('calendarToken', 'share'),
      );
      expect(find.text('1 stop selected · nothing submitted'), findsNothing);
    },
  );

  testWidgets('failed claim keeps unsent selections across rebuilds', (
    tester,
  ) async {
    SharedPreferences.setMockInitialValues({
      'travel_calendar_selections:share:guest': jsonEncode({
        'a': {
          'stopId': 'a',
          'status': 'interested',
          'startDate': '',
          'endDate': '',
        },
      }),
    });
    final service = FakeTravelCalendarService();
    await tester.pumpWidget(
      MaterialApp(
        home: TravelCalendarScreen(
          calendarToken: 'share',
          service: service,
          claimAccount: (_, _) async => false,
        ),
      ),
    );
    await tester.pumpAndSettle();
    await tester.drag(find.byType(ListView).first, const Offset(0, -350));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Claim account & review selections'));
    await tester.pumpAndSettle();
    expect(service.requests, isEmpty);
    expect(find.text('1 stop selected · nothing submitted'), findsOneWidget);
    final prefs = await SharedPreferences.getInstance();
    expect(
      prefs.getString('travel_calendar_selections:share:guest'),
      contains('interested'),
    );
  });

  testWidgets(
    'permission removals preview and require confirmation before saving',
    (tester) async {
      final service = FakeTravelCalendarService();
      await tester.pumpWidget(
        MaterialApp(
          home: Builder(
            builder: (context) => Scaffold(
              body: TextButton(
                onPressed: () => showDialog<void>(
                  context: context,
                  builder: (_) => TravelPermissionsDialog(service: service),
                ),
                child: const Text('Open permissions'),
              ),
            ),
          ),
        ),
      );
      await tester.tap(find.text('Open permissions'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Remove viewer'));
      await tester.tap(find.text('Review changes'));
      await tester.pumpAndSettle();
      expect(service.requests.single['path'], '/calendar/permissions/preview');
      expect(
        find.textContaining('1 participation record(s) will be removed'),
        findsOneWidget,
      );
      expect(service.requests.where((r) => r['method'] == 'PUT'), isEmpty);
      await tester.tap(find.text('Apply permissions'));
      await tester.pumpAndSettle();
      expect(service.requests.last['method'], 'PUT');
      expect(
        service.requests.last['data'],
        containsPair('confirmRemoval', true),
      );
      expect(service.requests.last['data'], containsPair('grants', isEmpty));
    },
  );

  testWidgets('one failed participation keeps only that selection for review', (
    tester,
  ) async {
    SharedPreferences.setMockInitialValues({
      'travel_calendar_selections:share:guest': jsonEncode({
        for (final id in ['a', 'b'])
          id: {
            'stopId': id,
            'status': 'requested',
            'startDate': '2099-01-03',
            'endDate': '2099-01-07',
          },
      }),
    });
    final service = FakeTravelCalendarService()..failSecondParticipation = true;
    await tester.pumpWidget(
      MaterialApp(
        home: TravelCalendarScreen(
          calendarToken: 'share',
          service: service,
          claimAccount: (_, _) async => true,
        ),
      ),
    );
    await tester.pumpAndSettle();
    await tester.drag(find.byType(ListView).first, const Offset(0, -350));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Claim account & review selections'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Submit selections'));
    await tester.pumpAndSettle();
    final prefs = await SharedPreferences.getInstance();
    final saved =
        jsonDecode(prefs.getString('travel_calendar_selections:share:guest')!)
            as Map;
    expect(saved.keys, ['b']);
    expect(service.requests, hasLength(2));
    expect(find.textContaining('Dates have changed.'), findsOneWidget);
  });
}
