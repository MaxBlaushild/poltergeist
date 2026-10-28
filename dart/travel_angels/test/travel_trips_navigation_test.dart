import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/screens/travel_calendar_screen.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/travel_calendar_views.dart';
import 'package:travel_angels/widgets/travel_trip_list.dart';

import 'support/app_fonts.dart';
import 'travel_calendar_widgets_test.dart'
    show
        CalendarTestAuth,
        FakeTravelCalendarService,
        calendarSearch,
        openCalendarFilters,
        stop,
        switchToList;

class _DelayedCalendarService extends FakeTravelCalendarService {
  Completer<Map<String, dynamic>>? nextCalendar;

  @override
  Future<Map<String, dynamic>> get(String path, {Map<String, dynamic>? query}) {
    if (path == '/calendar' && nextCalendar != null) {
      final pending = nextCalendar!;
      nextCalendar = null;
      return pending.future;
    }
    return super.get(path, query: query);
  }
}

Widget _screen({
  required FakeTravelCalendarService service,
  CalendarTestAuth? auth,
  String? calendarToken,
  String instance = 'first',
}) {
  final app = MaterialApp(
    theme: AppTheme.lightTheme,
    home: TravelCalendarScreen(
      key: ValueKey(instance),
      calendarToken: calendarToken,
      service: service,
    ),
  );
  return auth == null
      ? app
      : ChangeNotifierProvider<AuthProvider>.value(value: auth, child: app);
}

Future<void> _switchToCalendar(WidgetTester tester) async {
  await tester.ensureVisible(find.byKey(const ValueKey('trips-calendar-view')));
  await tester.tap(find.byKey(const ValueKey('trips-calendar-view')));
  await tester.pumpAndSettle();
}

void main() {
  setUpAll(loadAppFonts);
  setUp(() => SharedPreferences.setMockInitialValues({}));

  testWidgets(
    'calendar is the default and the chosen view survives recreation',
    (tester) async {
      final service = FakeTravelCalendarService();
      final auth = CalendarTestAuth(User(id: 'traveler-a'));
      await tester.pumpWidget(_screen(service: service, auth: auth));
      await tester.pumpAndSettle();
      expect(find.byType(TravelCalendarMonth), findsOneWidget);
      expect(find.byKey(const ValueKey('trips-calendar-view')), findsOneWidget);
      expect(find.byKey(const ValueKey('trips-list-view')), findsOneWidget);

      await switchToList(tester);
      expect(find.byType(TravelCalendarMonth), findsNothing);
      expect(find.byType(TravelTripList), findsOneWidget);
      final prefs = await SharedPreferences.getInstance();
      expect(prefs.getString('travel_trips_view:traveler-a'), 'list');
      expect(prefs.getString('travel_trips_view:guest'), isNull);

      await tester.pumpWidget(
        _screen(service: service, auth: auth, instance: 'reopened'),
      );
      await tester.pumpAndSettle();
      expect(find.byType(TravelTripList), findsOneWidget);
      expect(find.byType(TravelCalendarMonth), findsNothing);
      await _switchToCalendar(tester);
      expect(prefs.getString('travel_trips_view:traveler-a'), 'calendar');
      expect(service.requests, isEmpty);
    },
  );

  testWidgets(
    'toggling views retains the chosen month and destination filter',
    (tester) async {
      final service = FakeTravelCalendarService();
      await tester.pumpWidget(
        _screen(service: service, calendarToken: 'share'),
      );
      await tester.pumpAndSettle();
      await tester.ensureVisible(find.byTooltip('Next month'));
      await tester.tap(find.byTooltip('Next month'));
      await tester.pumpAndSettle();
      final chosenMonth = tester
          .widget<TravelCalendarMonth>(find.byType(TravelCalendarMonth))
          .month;

      await openCalendarFilters(tester);
      await tester.enterText(calendarSearch(), 'Kyoto');
      await tester.pumpAndSettle();
      await switchToList(tester);
      expect(find.byType(TravelTripListRow), findsOneWidget);
      expect(
        tester
            .widget<TravelTripListRow>(find.byType(TravelTripListRow))
            .stop['id'],
        'b',
      );
      await _switchToCalendar(tester);
      final month = tester.widget<TravelCalendarMonth>(
        find.byType(TravelCalendarMonth),
      );
      expect(month.month, chosenMonth);
      expect(month.stops.map((trip) => trip['id']), ['b']);
      expect(
        tester.widget<TextField>(calendarSearch()).controller!.text,
        'Kyoto',
      );
      expect(service.requests, isEmpty);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets('personal trips use one Trips heading and shared titles remain', (
    tester,
  ) async {
    final service = FakeTravelCalendarService();
    await tester.pumpWidget(_screen(service: service));
    await tester.pumpAndSettle();
    expect(
      find.descendant(of: find.byType(AppBar), matching: find.text('Trips')),
      findsOneWidget,
    );
    expect(find.text('Trips'), findsOneWidget);
    expect(find.text('Our travel calendar'), findsNothing);

    service.calendar = {
      ...service.calendar,
      'title': 'A year around the world',
    };
    await tester.pumpWidget(
      _screen(service: service, calendarToken: 'share', instance: 'shared'),
    );
    await tester.pumpAndSettle();
    expect(
      find.descendant(
        of: find.byType(AppBar),
        matching: find.text('A year around the world'),
      ),
      findsNothing,
    );
    expect(find.text('A year around the world'), findsOneWidget);
    expect(find.text('Trips'), findsNothing);
  });

  testWidgets('account changes clear old trips and restore each account view', (
    tester,
  ) async {
    SharedPreferences.setMockInitialValues({
      'travel_trips_view:traveler-a': 'list',
      'travel_trips_view:traveler-b': 'calendar',
    });
    final service = _DelayedCalendarService()
      ..stops = [stop('private', 'Private host trip')];
    final auth = CalendarTestAuth(User(id: 'traveler-a'));
    await tester.pumpWidget(_screen(service: service, auth: auth));
    await tester.pumpAndSettle();
    expect(find.byType(TravelTripList), findsOneWidget);
    expect(find.text('Private host trip'), findsWidgets);

    final nextResponse = Completer<Map<String, dynamic>>();
    service.nextCalendar = nextResponse;
    auth.setAccount(User(id: 'traveler-b'));
    await tester.pump();
    expect(find.text('Private host trip'), findsNothing);
    expect(find.byType(TravelTripListRow), findsNothing);
    nextResponse.complete({
      'calendar': service.calendar,
      'stops': [stop('public', 'New account trip')],
    });
    await tester.pumpAndSettle();
    expect(find.byType(TravelCalendarMonth), findsOneWidget);
    expect(find.byType(TravelTripList), findsNothing);
    expect(
      tester
          .widget<TravelCalendarMonth>(find.byType(TravelCalendarMonth))
          .stops
          .map((trip) => trip['id']),
      ['public'],
    );
    final prefs = await SharedPreferences.getInstance();
    expect(prefs.getString('travel_trips_view:traveler-a'), 'list');
    expect(prefs.getString('travel_trips_view:traveler-b'), 'calendar');
    expect(service.requests, isEmpty);
  });
}
