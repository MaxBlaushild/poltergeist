import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'support/app_fonts.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/auth_service.dart';
import 'package:travel_angels/screens/travel_calendar_screen.dart';
import 'package:travel_angels/screens/travel_trip_screen.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/widgets/travel_calendar_dialogs.dart';
import 'package:travel_angels/widgets/travel_calendar_delivery.dart';
import 'package:travel_angels/widgets/travel_calendar_views.dart';
import 'package:travel_angels/widgets/travel_trip_list.dart';

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
  'acceptingParticipation': true,
  'canManage': false,
  'canManagePermissions': false,
};

class FakeTravelCalendarService extends TravelCalendarService {
  List<TravelStop> stops = [stop('a', 'Lisbon'), stop('b', 'Kyoto')];
  Map<String, dynamic> calendar = {
    'id': 'calendar',
    'title': 'Our travel calendar',
    'canManage': false,
    'canManagePermissions': false,
  };
  final List<String> reads = [];
  final List<Map<String, dynamic>> requests = [];
  int removals = 1;
  bool failSecondParticipation = false;
  bool previewVisible = false;
  Map<String, dynamic> notificationResponse = {};
  Map<String, dynamic> flightTravelers = {};
  List<Map<String, dynamic>> cohostInvitations = [];

  @override
  Future<Map<String, dynamic>> get(
    String path, {
    Map<String, dynamic>? query,
  }) async {
    reads.add(path);
    if (path.contains('/flights/') && path.endsWith('/travelers')) {
      return flightTravelers;
    }
    if (path == '/calendar-notifications') return notificationResponse;
    if (path.startsWith('/shared/stops/')) return {'stop': stops.first};
    if (RegExp(r'^/calendar/stops/[^/]+$').hasMatch(path)) {
      return {
        'stop': stops.firstWhere((stop) => stop['id'] == path.split('/').last),
      };
    }
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
      'calendar': calendar,
      'stops': stops,
      'cohostInvitations': cohostInvitations,
    };
  }

  @override
  Future<Map<String, dynamic>> post(
    String path, {
    Map<String, dynamic>? data,
    Map<String, dynamic>? query,
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
    Map<String, dynamic>? query,
  }) async {
    requests.add({'method': 'PUT', 'path': path, 'data': data});
    return {};
  }
}

Future<void> switchToList(WidgetTester tester) async {
  await tester.scrollUntilVisible(
    find.byTooltip('List view'),
    -200,
    scrollable: find.byType(Scrollable).first,
  );
  await tester.tap(find.byTooltip('List view'));
  await tester.pumpAndSettle();
}

Future<void> openCalendarFilters(WidgetTester tester) async {
  await tester.scrollUntilVisible(
    find.byTooltip('Filter trips'),
    -200,
    scrollable: find.byType(Scrollable).first,
  );
  await tester.tap(find.byTooltip('Filter trips'));
  await tester.pumpAndSettle();
}

Finder calendarSearch() => find.byWidgetPredicate(
  (widget) =>
      widget is TextField &&
      widget.decoration?.labelText == 'Filter by destination or title',
);

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
  setUpAll(loadAppFonts);
  setUp(() => SharedPreferences.setMockInitialValues({}));

  testWidgets(
    'empty calendars default to a navigable month with discoverable filters',
    (tester) async {
      final service = FakeTravelCalendarService()..stops = [];
      await tester.pumpWidget(
        MaterialApp(
          theme: AppTheme.lightTheme,
          home: TravelCalendarScreen(calendarToken: 'share', service: service),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.byType(TravelCalendarMonth), findsOneWidget);
      expect(find.byType(TravelStopCard), findsNothing);
      expect(calendarSearch(), findsNothing);
      expect(find.byTooltip('Filter trips'), findsOneWidget);
      final before = tester
          .widget<TravelCalendarMonth>(find.byType(TravelCalendarMonth))
          .month;
      await tester.ensureVisible(find.byTooltip('Next month'));
      await tester.tap(find.byTooltip('Next month'));
      await tester.pumpAndSettle();
      expect(
        tester
            .widget<TravelCalendarMonth>(find.byType(TravelCalendarMonth))
            .month,
        DateTime(before.year, before.month + 1),
      );
      await openCalendarFilters(tester);
      expect(calendarSearch(), findsOneWidget);
      expect(find.text('All dates'), findsOneWidget);
      expect(find.text('Date range'), findsOneWidget);
      await tester.enterText(calendarSearch(), 'No matching destination');
      await tester.pumpAndSettle();
      expect(find.byType(TravelCalendarMonth), findsOneWidget);
      await tester.tap(find.byTooltip('Filter trips'));
      await tester.pumpAndSettle();
      expect(calendarSearch(), findsNothing);
      expect(find.byTooltip('Clear destination filter'), findsOneWidget);
      await openCalendarFilters(tester);
      expect(
        tester.widget<TextField>(calendarSearch()).controller!.text,
        'No matching destination',
      );
      await tester.tap(find.byTooltip('Clear destination filter'));
      await tester.pumpAndSettle();
      expect(
        tester.widget<TextField>(calendarSearch()).controller!.text,
        isEmpty,
      );
      expect(service.requests, isEmpty);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'compact manager controls and expanded filters fit at 320px with enlarged text',
    (tester) async {
      await tester.binding.setSurfaceSize(const Size(320, 800));
      addTearDown(() => tester.binding.setSurfaceSize(null));
      final service = FakeTravelCalendarService()
        ..stops = []
        ..calendar = {
          'id': 'calendar',
          'title': 'Our year of travel together',
          'canManage': true,
          'canManagePermissions': true,
          'canCreateTrips': true,
        };
      await tester.pumpWidget(
        MaterialApp(
          theme: AppTheme.lightTheme,
          builder: (_, child) => MediaQuery(
            data: const MediaQueryData(
              size: Size(320, 800),
              textScaler: TextScaler.linear(2),
            ),
            child: child!,
          ),
          home: TravelCalendarScreen(service: service),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.byTooltip('Create trip'), findsOneWidget);
      expect(tester.takeException(), isNull);
      await openCalendarFilters(tester);
      await tester.enterText(
        calendarSearch(),
        'A destination with a deliberately long name',
      );
      await tester.pumpAndSettle();
      final periods = find.byType(DropdownButton<String>);
      await tester.ensureVisible(periods);
      await tester.tap(periods);
      await tester.pumpAndSettle();
      await tester.ensureVisible(find.text('Approximate / TBD').last);
      await tester.tap(find.text('Approximate / TBD').last);
      await tester.pumpAndSettle();
      expect(
        tester.widget<DropdownButton<String>>(periods).value,
        'unscheduled',
      );
      expect(find.byTooltip('Clear destination filter'), findsOneWidget);
      expect(find.byTooltip('Clear period filter'), findsOneWidget);
      expect(tester.takeException(), isNull);
      await tester.tap(find.byTooltip('Calendar actions'));
      await tester.pumpAndSettle();
      expect(find.text('Default audience'), findsOneWidget);
      expect(tester.takeException(), isNull);
      expect(service.requests, isEmpty);
    },
  );

  testWidgets(
    'compact calendar actions retain owner controls and restrict editor and guest roles',
    (tester) async {
      final owner = FakeTravelCalendarService()
        ..stops = []
        ..calendar = {
          'id': 'calendar',
          'title': 'Owned calendar',
          'canManage': true,
          'canManagePermissions': true,
          'canCreateTrips': true,
        };
      await tester.pumpWidget(
        MaterialApp(
          theme: AppTheme.lightTheme,
          home: TravelCalendarScreen(
            key: const ValueKey('owner'),
            service: owner,
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.byTooltip('Create trip'), findsOneWidget);
      expect(find.text('Invite friends'), findsNothing);
      await tester.tap(find.byTooltip('Calendar actions'));
      await tester.pumpAndSettle();
      for (final label in [
        'Share calendar',
        'Invite friends',
        'Broadcast update',
        'Rename calendar',
        'Default audience',
        'Calendar editors',
        'Manage share link',
      ]) {
        expect(find.text(label), findsOneWidget);
      }
      expect(find.text('Verify invited email'), findsNothing);
      await tester.tap(find.text('Default audience'));
      await tester.pumpAndSettle();
      expect(find.byType(TravelPermissionsDialog), findsOneWidget);
      expect(owner.reads, contains('/calendar/permissions'));
      expect(owner.requests, isEmpty);
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();

      final editor = FakeTravelCalendarService()
        ..stops = []
        ..calendar = {
          'id': 'calendar',
          'title': 'Edited calendar',
          'canManage': true,
          'canManagePermissions': false,
          'canCreateTrips': false,
        };
      await tester.pumpWidget(
        MaterialApp(
          theme: AppTheme.lightTheme,
          home: TravelCalendarScreen(
            key: const ValueKey('editor'),
            service: editor,
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.byTooltip('Create trip'), findsNothing);
      await tester.tap(find.byTooltip('Calendar actions'));
      await tester.pumpAndSettle();
      expect(find.text('Invite friends'), findsOneWidget);
      expect(find.text('Broadcast update'), findsOneWidget);
      for (final label in [
        'Share calendar',
        'Rename calendar',
        'Default audience',
        'Calendar editors',
        'Manage share link',
      ]) {
        expect(find.text(label), findsNothing);
      }
      expect(editor.requests, isEmpty);
      await tester.pumpWidget(const SizedBox.shrink());

      final guest = FakeTravelCalendarService()..stops = [];
      await tester.pumpWidget(
        MaterialApp(
          theme: AppTheme.lightTheme,
          home: TravelCalendarScreen(calendarToken: 'share', service: guest),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text('Follow updates'), findsOneWidget);
      expect(find.byTooltip('Create trip'), findsNothing);
      await tester.tap(find.byTooltip('Calendar actions'));
      await tester.pumpAndSettle();
      expect(find.text('Follow / manage updates'), findsOneWidget);
      expect(find.text('Verify invited email'), findsOneWidget);
      for (final label in [
        'Share calendar',
        'Invite friends',
        'Broadcast update',
        'Rename calendar',
        'Default audience',
        'Calendar editors',
        'Manage share link',
      ]) {
        expect(find.text(label), findsNothing);
      }
      await tester.tap(find.text('Verify invited email'));
      await tester.pumpAndSettle();
      expect(find.byType(TravelGuestVerificationDialog), findsOneWidget);
      expect(guest.requests, isEmpty);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'full host acceptance separates organizing from attendance and requires consent',
    (tester) async {
      final service = FakeTravelCalendarService()
        ..cohostInvitations = [
          {
            'id': 'host-invite',
            'stopId': 'a',
            'isHost': true,
            'tripTitle': 'A week in Lisbon',
            'ownerDisplayName': 'Max',
            'rosterVersion': 3,
            'rosterVisibility': 'attendees',
          },
        ];
      await tester.pumpWidget(
        MaterialApp(
          theme: AppTheme.lightTheme,
          home: TravelCalendarScreen(service: service),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text('A week in Lisbon'), findsOneWidget);
      await tester.tap(find.text('Accept'));
      await tester.pumpAndSettle();
      final attendance = find.widgetWithText(
        CheckboxListTile,
        'I am also attending',
      );
      expect(tester.widget<CheckboxListTile>(attendance).value, isFalse);
      await tester.tap(attendance);
      await tester.pumpAndSettle();
      final accept = find.widgetWithText(FilledButton, 'Accept');
      expect(tester.widget<FilledButton>(accept).onPressed, isNull);
      await tester.tap(
        find.text('Hosts and confirmed attendees can see my name and avatar.'),
      );
      await tester.pumpAndSettle();
      await tester.tap(accept);
      await tester.pumpAndSettle();
      expect(
        service.requests.single['path'],
        '/calendar/collaborators/host-invite/accept',
      );
      expect(service.requests.single['data'], {
        'attending': true,
        'rosterConsentVersion': 3,
        'shareDates': false,
      });
    },
  );

  testWidgets(
    'an existing stop share link opens the trip page without a login gate',
    (tester) async {
      final service = FakeTravelCalendarService();
      await tester.pumpWidget(
        MaterialApp(
          theme: AppTheme.lightTheme,
          home: TravelCalendarScreen(
            stopToken: 'shared-trip-token',
            service: service,
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.byType(TravelTripScreen), findsOneWidget);
      expect(find.text('Lisbon'), findsWidgets);
      expect(find.text('Our travel calendar'), findsNothing);
      expect(service.requests, isEmpty);
      await tester.pumpWidget(const SizedBox.shrink());
    },
  );

  testWidgets(
    'calendar trip navigation preserves the destination filter on return',
    (tester) async {
      final service = FakeTravelCalendarService();
      await tester.pumpWidget(
        MaterialApp(
          theme: AppTheme.lightTheme,
          home: TravelCalendarScreen(calendarToken: 'share', service: service),
        ),
      );
      await tester.pumpAndSettle();
      await switchToList(tester);
      await openCalendarFilters(tester);
      await tester.enterText(find.byType(TextField).first, 'Kyoto');
      await tester.pumpAndSettle();
      final card = find.byWidgetPredicate(
        (widget) => widget is TravelTripListRow && widget.stop['id'] == 'b',
      );
      await tester.ensureVisible(card);
      await tester.tap(card);
      await tester.pumpAndSettle();
      expect(find.byType(TravelTripScreen), findsOneWidget);
      Navigator.of(tester.element(find.byType(TravelTripScreen))).pop();
      await tester.pumpAndSettle();
      expect(
        tester.widget<TextField>(find.byType(TextField).first).controller!.text,
        'Kyoto',
      );
      expect(
        find.byWidgetPredicate(
          (widget) => widget is TravelTripListRow && widget.stop['id'] == 'a',
        ),
        findsNothing,
      );
      expect(card, findsOneWidget);
    },
  );

  testWidgets(
    'attendee preview renders the filtered lodging and flight response',
    (tester) async {
      final service = FakeTravelCalendarService()
        ..previewVisible = true
        ..stops = [
          {
            ...stop('a', 'Lisbon'),
            'lodging': [
              {
                'id': 'stay',
                'name': 'Harbor House',
                'datesTbd': true,
                'privateDetails': {'address': '10 Harbor Road'},
              },
            ],
            'flights': [
              {'id': 'flight', 'label': 'Outbound via Madrid', 'legs': []},
            ],
          },
        ];
      await tester.pumpWidget(
        MaterialApp(
          theme: AppTheme.lightTheme,
          home: Scaffold(
            body: TravelStopPreviewDialog(
              service: service,
              stop: service.stops.first,
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      await tester.tap(find.byType(DropdownButtonFormField<String>));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Confirmed attendee').last);
      await tester.pumpAndSettle();
      expect(
        service.requests.last['data'],
        containsPair('audience', 'attendee'),
      );
      expect(find.text('Harbor House'), findsOneWidget);
      expect(find.text('Outbound via Madrid'), findsOneWidget);
      expect(find.textContaining('Named traveler assignments'), findsOneWidget);
      await tester.ensureVisible(find.text('Stay details'));
      await tester.tap(find.text('Stay details'));
      await tester.pumpAndSettle();
      expect(find.text('10 Harbor Road'), findsOneWidget);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'draft editor saves destination time zone without converting local dates',
    (tester) async {
      final service = FakeTravelCalendarService();
      await tester.pumpWidget(
        MaterialApp(
          theme: AppTheme.lightTheme,
          home: Builder(
            builder: (context) => Scaffold(
              body: TextButton(
                onPressed: () => showDialog<void>(
                  context: context,
                  builder: (_) => TravelStopEditor(
                    service: service,
                    stop: {
                      ...stop('a', 'Auckland'),
                      'revision': 9,
                      'draft': {
                        ...stop('a', 'Auckland'),
                        'hero': {'mediaId': 'protected-cover'},
                        'lodging': [
                          {'id': 'stay', 'name': 'Harbor House'},
                        ],
                        'flights': [
                          {'id': 'flight', 'label': 'Outbound'},
                        ],
                      },
                    },
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
      await tester.ensureVisible(find.text('More options'));
      await tester.tap(find.text('More options'));
      await tester.pumpAndSettle();
      final timeZoneField = find.byWidgetPredicate(
        (widget) =>
            widget is TextField &&
            widget.decoration?.labelText == 'Destination time zone',
      );
      await tester.ensureVisible(timeZoneField);
      await tester.enterText(timeZoneField, 'Pacific/Auckland');
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
      final data = service.requests.single['data'] as Map;
      expect(data['revision'], 9);
      expect((data['hero'] as Map)['mediaId'], 'protected-cover');
      expect((data['lodging'] as List).single['name'], 'Harbor House');
      expect((data['flights'] as List).single['label'], 'Outbound');
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
      Widget calendarPage(String instance) => MaterialApp(
        theme: AppTheme.lightTheme,
        home: TravelCalendarScreen(
          key: ValueKey(instance),
          calendarToken: 'share',
          service: service,
        ),
      );
      await tester.pumpWidget(calendarPage('expired'));
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
      await tester.pumpWidget(calendarPage('verified'));
      await tester.pumpAndSettle();
      expect(find.text('1 stop selected · nothing submitted'), findsOneWidget);
      expect(find.text('Restore saved selections'), findsNothing);
      expect(
        prefs.getString('travel_calendar_selections:share:guest'),
        contains('2099-01-03'),
      );
      // A current verified session can now establish actual loss of access.
      service.stops = [];
      await tester.pumpWidget(calendarPage('revoked'));
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
            theme: AppTheme.lightTheme,
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
            theme: AppTheme.lightTheme,
            home: TravelCalendarScreen(
              calendarToken: 'share',
              service: service,
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      await switchToList(tester);
      expect(find.text('1 stop selected · nothing submitted'), findsOneWidget);
      expect(find.text('Private host plan'), findsWidgets);
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
            theme: AppTheme.lightTheme,
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
      await tester.ensureVisible(
        find.text('Claim account & review selections'),
      );
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
          theme: AppTheme.lightTheme,
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

  test('temporal filters ignore saved dates when a trip is TBD', () {
    final plans = [
      stop(
        'undecided',
        'Japan',
        precision: 'tbd',
        start: '2027-01-01',
        end: '2027-01-10',
      ),
      stop(
        'approximate',
        'Lisbon',
        precision: 'approximate',
        start: '2027-01-01',
        end: '2027-01-10',
      ),
    ];
    for (final entry in {
      'upcoming': DateTime(2026, 12, 31),
      'current': DateTime(2027, 1, 5),
      'past': DateTime(2027, 1, 11),
    }.entries) {
      expect(
        filterTravelStops(
          plans,
          period: entry.key,
          today: entry.value,
        ).map((trip) => trip['id']),
        ['approximate'],
      );
    }
    expect(
      filterTravelStops(
        plans,
        range: DateTimeRange(
          start: DateTime(2027, 1, 3),
          end: DateTime(2027, 1, 5),
        ),
      ).map((trip) => trip['id']),
      ['approximate'],
    );
    expect(
      filterTravelStops(plans, period: 'unscheduled').map((trip) => trip['id']),
      containsAll(['undecided', 'approximate']),
    );
  });

  test(
    'destination filters match all selected places, aliases and countries',
    () {
      final plans = [
        {
          ...stop('japan', 'Autumn adventure'),
          'destination': 'Tokyo → Kyoto',
          'destinations': [
            {
              'id': 'tokyo',
              'name': 'Tokyo',
              'kind': 'city',
              'countryName': 'Japan',
              'countryCode': 'JP',
            },
            {
              'id': 'kyoto',
              'name': 'Kyoto',
              'kind': 'city',
              'countryName': 'Japan',
              'aliases': ['京都'],
            },
          ],
        },
        stop('legacy', 'Lisbon'),
      ];
      for (final query in ['Tokyo', 'Kyoto', 'Japan', 'JP', '京都']) {
        expect(
          filterTravelStops(plans, destination: query).map((s) => s['id']),
          ['japan'],
        );
      }
      expect(
        filterTravelStops(plans, destination: 'Lisbon').map((s) => s['id']),
        ['legacy'],
      );
    },
  );

  testWidgets(
    'shared plans browse without an authentication provider and filter destinations',
    (tester) async {
      final service = FakeTravelCalendarService();
      await tester.pumpWidget(
        MaterialApp(
          theme: AppTheme.lightTheme,
          home: TravelCalendarScreen(calendarToken: 'share', service: service),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text('Our travel calendar'), findsOneWidget);
      expect(find.byTooltip('Create trip'), findsNothing);
      await switchToList(tester);
      await openCalendarFilters(tester);
      await tester.enterText(find.byType(TextField).first, 'Kyoto');
      await tester.pump();
      expect(
        find.byWidgetPredicate(
          (widget) =>
              widget is TravelTripListRow && widget.stop['title'] == 'Kyoto',
        ),
        findsOneWidget,
      );
      expect(
        find.byWidgetPredicate(
          (widget) =>
              widget is TravelTripListRow && widget.stop['title'] == 'Lisbon',
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
          theme: AppTheme.lightTheme,
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
          theme: AppTheme.lightTheme,
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
    'flight claim preserves pending plans and requires explicit flight selection',
    (tester) async {
      SharedPreferences.setMockInitialValues({
        'travel_calendar_selections:flight-share:guest': jsonEncode({
          'a': {'stopId': 'a', 'status': 'interested'},
        }),
      });
      final auth = CalendarTestAuth(null);
      final service = FakeTravelCalendarService()
        ..flightTravelers = {
          'requiresAccount': true,
          'canSelect': false,
          'travelers': [],
        }
        ..stops = [
          {
            ...stop('a', 'Lisbon'),
            'flights': [
              {
                'id': 'outbound',
                'label': 'Fly together',
                'visibility': 'viewers',
                'legs': [
                  {
                    'id': 'leg',
                    'airline': 'TAP',
                    'flightNumber': 'TP208',
                    'departureAirport': 'JFK',
                    'arrivalAirport': 'LIS',
                    'departureDate': '2099-01-01',
                    'arrivalDate': '2099-01-02',
                  },
                ],
              },
            ],
          },
        ];
      var claims = 0;
      await tester.pumpWidget(
        ChangeNotifierProvider<AuthProvider>.value(
          value: auth,
          child: MaterialApp(
            theme: AppTheme.lightTheme,
            home: TravelCalendarScreen(
              stopToken: 'flight-share',
              service: service,
              claimAccount: (_, _) async {
                claims++;
                service.flightTravelers = {
                  'requiresAccount': false,
                  'canSelect': true,
                  'flightVersion': 'current',
                  'rosterVersion': 1,
                  'travelers': [],
                };
                auth.setAccount(User(id: 'flight-traveler'));
                return true;
              },
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      final select = find.byKey(const ValueKey('flight-select'));
      await tester.ensureVisible(select);
      await tester.tap(select);
      await tester.pumpAndSettle();
      expect(claims, 1);
      expect(service.requests, isEmpty);
      final prefs = await SharedPreferences.getInstance();
      expect(
        prefs.getString('travel_calendar_selections:flight-share:guest'),
        isNull,
      );
      expect(
        prefs.getString(
          'travel_calendar_selections:flight-share:flight-traveler',
        ),
        contains('interested'),
      );
      await tester.ensureVisible(select);
      await tester.tap(select);
      await tester.pumpAndSettle();
      expect(
        find.byKey(const ValueKey('flight-selection-status')),
        findsOneWidget,
      );
      expect(service.requests, isEmpty);
      expect(tester.takeException(), isNull);
      await tester.pumpWidget(const SizedBox.shrink());
    },
  );

  testWidgets(
    'guests can choose a stop before any authentication or RSVP request',
    (tester) async {
      var claims = 0;
      final service = FakeTravelCalendarService()
        ..stops = [stop('a', 'Lisbon')];
      await tester.pumpWidget(
        MaterialApp(
          theme: AppTheme.lightTheme,
          home: TravelCalendarScreen(
            calendarToken: 'share',
            service: service,
            claimAccount: (_, _) async {
              claims++;
              return false;
            },
          ),
        ),
      );
      await tester.pumpAndSettle();
      await switchToList(tester);
      await tester.ensureVisible(find.byType(TravelTripListRow));
      await tester.pumpAndSettle();
      await tester.tap(find.byType(TravelTripListRow));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Request to join').last);
      await tester.pumpAndSettle();
      expect(claims, 0);
      expect(find.text('Continue'), findsOneWidget);
      expect(service.requests, isEmpty);
      await tester.tap(find.text('Continue'));
      await tester.pumpAndSettle();
      expect(claims, 1);
      expect(service.requests, isEmpty);
      final prefs = await SharedPreferences.getInstance();
      final saved =
          jsonDecode(prefs.getString('travel_calendar_selections:share:guest')!)
              as Map;
      expect(saved['a'], containsPair('status', 'requested'));
      expect(saved['a'], containsPair('startDate', '2099-01-01'));
      await tester.pumpWidget(const SizedBox.shrink());
    },
  );

  testWidgets(
    'trip page requires roster consent then claims and reviews partial attendance',
    (tester) async {
      SharedPreferences.setMockInitialValues({
        'travel_calendar_selections:trip-link:guest': jsonEncode({
          'a': {
            'stopId': 'a',
            'status': 'requested',
            'startDate': '2099-01-03',
            'endDate': '2099-01-07',
          },
        }),
      });
      var claims = 0;
      final auth = CalendarTestAuth(null);
      final service = FakeTravelCalendarService()
        ..stops = [
          {
            ...stop('a', 'Lisbon'),
            'hostModel': true,
            'rosterVersion': 2,
            'tripSettings': {'rosterVisibility': 'attendees'},
          },
        ];
      await tester.pumpWidget(
        ChangeNotifierProvider<AuthProvider>.value(
          value: auth,
          child: MaterialApp(
            theme: AppTheme.lightTheme,
            home: TravelCalendarScreen(
              stopToken: 'trip-link',
              service: service,
              claimAccount: (_, _) async {
                claims++;
                auth.setAccount(User(id: 'claimed-friend'));
                return true;
              },
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      await tester.tap(find.text('Request to join').last);
      await tester.pumpAndSettle();
      final continueButton = find.widgetWithText(FilledButton, 'Continue');
      expect(tester.widget<FilledButton>(continueButton).onPressed, isNull);
      final consent = find.text(
        'Trip hosts and confirmed attendees can see my name and avatar when I attend.',
      );
      await tester.ensureVisible(consent);
      await tester.tap(consent);
      await tester.pumpAndSettle();
      expect(claims, 0);
      await tester.tap(continueButton);
      await tester.pumpAndSettle();
      expect(claims, 1);
      expect(find.text('Review your selections'), findsOneWidget);
      expect(service.requests, isEmpty);
      await tester.tap(find.text('Submit selections'));
      await tester.pumpAndSettle();
      expect(
        service.requests.single['path'],
        '/calendar/stops/a/participation',
      );
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
        containsPair('rosterConsentVersion', 2),
      );
      expect(
        service.requests.single['data'],
        containsPair('stopToken', 'trip-link'),
      );
      expect(
        service.requests.single['data'],
        containsPair('status', 'requested'),
      );
      await tester.pumpWidget(const SizedBox.shrink());
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
          theme: AppTheme.lightTheme,
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
        theme: AppTheme.lightTheme,
        home: TravelCalendarScreen(
          calendarToken: 'share',
          service: service,
          claimAccount: (_, _) async => false,
        ),
      ),
    );
    await tester.pumpAndSettle();
    await tester.ensureVisible(find.text('Claim account & review selections'));
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
          theme: AppTheme.lightTheme,
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
        theme: AppTheme.lightTheme,
        home: TravelCalendarScreen(
          calendarToken: 'share',
          service: service,
          claimAccount: (_, _) async => true,
        ),
      ),
    );
    await tester.pumpAndSettle();
    await tester.ensureVisible(find.text('Claim account & review selections'));
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
    await tester.scrollUntilVisible(
      find.textContaining('Dates have changed.'),
      -200,
      scrollable: find.byType(Scrollable).first,
    );
    expect(find.textContaining('Dates have changed.'), findsOneWidget);
  });
}
