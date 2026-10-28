import 'dart:async';
import 'dart:io';
import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/travel_flight_card.dart';
import 'package:travel_angels/widgets/travel_flight_lookup.dart';
import 'package:travel_angels/widgets/travel_trip_design.dart';
import 'package:travel_angels/widgets/travel_trip_logistics.dart';
import 'support/app_fonts.dart';
import 'travel_trip_widgets_test.dart'
    show TripTestAuth, sampleTrip, showTestDialog;

Map<String, dynamic> _leg({String id = 'leg-1', String number = 'BA178'}) => {
  'id': id,
  'airline': 'British Airways',
  'flightNumber': number,
  'departureAirport': 'JFK',
  'arrivalAirport': 'LHR',
  'departureDate': '2099-06-01',
  'arrivalDate': '2099-06-02',
  'departureTime': '22:00',
  'arrivalTime': '10:00',
  'departureTimeZone': 'America/New_York',
  'arrivalTimeZone': 'Europe/London',
  'provider': 'aeroapi',
  'providerId': 'flight-instance',
};
Map<String, dynamic> _flight({bool connection = false}) => {
  'id': 'flight-1',
  'label': 'Across the Atlantic',
  'kind': 'group',
  'status': 'planned',
  'legs': [
    _leg(),
    if (connection)
      {
        ..._leg(id: 'leg-2', number: 'BA500'),
        'departureAirport': 'LHR',
        'arrivalAirport': 'LIS',
        'departureDate': '2099-06-02',
        'arrivalDate': '2099-06-02',
        'departureTime': '12:00',
        'arrivalTime': '14:45',
        'departureTimeZone': 'Europe/London',
        'arrivalTimeZone': 'Europe/Lisbon',
      },
  ],
};

class _Service extends TravelCalendarService {
  final gets = <Map<String, dynamic>>[];
  final posts = <Map<String, dynamic>>[];
  final puts = <Map<String, dynamic>>[];
  Map<String, dynamic> state = {
    'flightVersion': 'v1',
    'rosterVersion': 3,
    'canSelect': true,
    'canViewTravelers': true,
    'canShareWithTravelers': true,
    'sharingAudience': 'attendees',
    'requiresAccount': false,
    'requiresAttendance': false,
    'ownSelection': null,
    'travelers': [
      {
        'userId': 'host',
        'displayName': 'Max',
        'status': 'booked',
        'legIds': ['leg-1'],
        'isHost': true,
      },
    ],
  };
  Map<String, dynamic>? response;
  Map<String, dynamic> search = {
    'configured': true,
    'flights': [_leg()],
  };
  final searches = <String, Completer<Map<String, dynamic>>>{};
  Completer<Map<String, dynamic>>? travelerGate;
  Object? failure;
  Object? postFailure;
  @override
  Future<Map<String, dynamic>> get(
    String path, {
    Map<String, dynamic>? query,
  }) async {
    gets.add({'path': path, 'query': query});
    if (path.endsWith('/participations')) return {};
    if (path.endsWith('/flight-search')) {
      if (searches.containsKey(query?['flightNumber'])) {
        return searches[query!['flightNumber']]!.future;
      }
      if (failure != null) throw failure!;
      return search;
    }
    if (travelerGate != null) return travelerGate!.future;
    if (failure != null) throw failure!;
    return response ?? tripCopy(state);
  }

  @override
  Future<Map<String, dynamic>> post(
    String path, {
    Map<String, dynamic>? data,
  }) async {
    posts.add({'path': path, 'data': tripCopy(data!)});
    if (postFailure != null) throw postFailure!;
    state['ownSelection'] = data['status'] == 'removed'
        ? null
        : {...data, 'needsReconfirmation': false};
    return {};
  }

  @override
  Future<Map<String, dynamic>> put(
    String path, {
    Map<String, dynamic>? data,
  }) async {
    puts.add(tripCopy(data!));
    return {};
  }
}

Finder _key(String key) => find.byKey(ValueKey(key));
Future<void> _tap(WidgetTester tester, Finder finder) async {
  await tester.ensureVisible(finder);
  await tester.pumpAndSettle();
  await tester.tap(finder);
  await tester.pumpAndSettle();
}

Future<GlobalKey> _mount(
  WidgetTester tester,
  Widget child, {
  double width = 390,
  double scale = 1,
}) async {
  final size = Size(width, 844);
  await tester.binding.setSurfaceSize(size);
  addTearDown(() => tester.binding.setSurfaceSize(null));
  final boundary = GlobalKey();
  await tester.pumpWidget(
    RepaintBoundary(
      key: boundary,
      child: MaterialApp(
        debugShowCheckedModeBanner: false,
        theme: AppTheme.lightTheme,
        builder: (_, child) => MediaQuery(
          data: MediaQueryData(
            size: size,
            textScaler: TextScaler.linear(scale),
          ),
          child: child!,
        ),
        home: Scaffold(
          body: SingleChildScrollView(
            padding: const EdgeInsets.all(16),
            child: child,
          ),
        ),
      ),
    ),
  );
  await tester.pumpAndSettle();
  return boundary;
}

TravelTripFlightCard _card(
  _Service service, {
  Map<String, dynamic>? entry,
  bool draft = false,
  Future<void> Function()? claim,
  Future<void> Function()? join,
}) => TravelTripFlightCard(
  entry: entry ?? _flight(),
  service: service,
  stopId: 'trip',
  draft: draft,
  onClaim: claim,
  onJoinTrip: join,
);
Future<void> _capture(
  WidgetTester tester,
  GlobalKey boundary,
  String name,
) async {
  if (!const bool.fromEnvironment('TRIP_SCREENSHOTS')) return;
  await tester.runAsync(() async {
    final image =
        await (boundary.currentContext!.findRenderObject()!
                as RenderRepaintBoundary)
            .toImage();
    final bytes = await image.toByteData(format: ui.ImageByteFormat.png);
    await File(
      '/private/tmp/travel-angels-flight-$name.png',
    ).writeAsBytes(bytes!.buffer.asUint8List());
    image.dispose();
  });
}

void main() {
  setUpAll(loadAppFonts);
  test(
    'booking links validate scheme and credentials and search encodes flight details',
    () {
      final entry = _flight()..['url'] = 'https://airline.example/book';
      expect(flightSearchLink(entry).host, 'airline.example');
      entry['url'] = 'https://user:password@airline.example';
      final fallback = flightSearchLink(entry);
      expect(fallback.host, 'www.google.com');
      expect(
        fallback.queryParameters['q'],
        contains('BA178 2099-06-01 JFK to LHR'),
      );
      entry['url'] = 'javascript:alert(1)';
      expect(flightSearchLink(entry).scheme, 'https');
    },
  );
  testWidgets('draft cards never load or expose traveler actions', (
    tester,
  ) async {
    final service = _Service();
    await _mount(tester, _card(service, draft: true));
    expect(service.gets, isEmpty);
    expect(_key('flight-select'), findsNothing);
    expect(find.textContaining('Max'), findsNothing);
    expect(_key('flight-find'), findsOneWidget);
  });
  testWidgets('only authorized shared names render, never host assignments', (
    tester,
  ) async {
    final service = _Service()
      ..state['canViewTravelers'] = false
      ..state['sharingAudience'] = 'hosts';
    final entry = _flight()..['assignedUserIds'] = ['secret-person'];
    await _mount(tester, _card(service, entry: entry));
    expect(find.textContaining('Max'), findsNothing);
    expect(find.textContaining('secret-person'), findsNothing);
    await _tap(tester, _key('flight-select'));
    expect(_key('flight-selection-share'), findsOneWidget);
    expect(find.text('Share with trip hosts'), findsOneWidget);
    await _tap(tester, find.text('Save my plans'));
    expect(service.posts.single['data']['shareWithTravelers'], false);
    expect(service.puts, isEmpty);
  });
  testWidgets(
    'traveler can choose one connection, self report booking and share',
    (tester) async {
      final service = _Service();
      await _mount(tester, _card(service, entry: _flight(connection: true)));
      await _tap(tester, _key('flight-select'));
      await _tap(tester, _key('flight-selection-status'));
      await _tap(tester, find.text('Booked').last);
      await _tap(tester, _key('flight-selection-leg-leg-2'));
      await _tap(tester, _key('flight-selection-share'));
      await _tap(tester, find.text('Save my plans'));
      final data = service.posts.single['data'];
      expect(data, {
        'status': 'booked',
        'legIds': ['leg-1'],
        'flightVersion': 'v1',
        'rosterVersion': 3,
        'shareWithTravelers': true,
      });
      expect(
        find.text('You’re booked · JFK to LHR · self-reported'),
        findsOneWidget,
      );
      await _tap(tester, _key('flight-select'));
      await _tap(tester, _key('flight-selection-remove'));
      expect(service.posts.last['data']['status'], 'removed');
      expect(service.puts, isEmpty);
    },
  );
  testWidgets('claim and attendance gates never record a selection', (
    tester,
  ) async {
    final service = _Service()
      ..state['canSelect'] = false
      ..state['requiresAccount'] = true;
    var claims = 0;
    var joins = 0;
    await _mount(
      tester,
      _card(
        service,
        claim: () async {
          claims++;
          service.state['requiresAccount'] = false;
          service.state['requiresAttendance'] = true;
        },
        join: () async {
          joins++;
          service.state['requiresAttendance'] = false;
          service.state['canSelect'] = true;
        },
      ),
    );
    await _tap(tester, _key('flight-select'));
    expect(claims, 1);
    expect(service.posts, isEmpty);
    expect(find.byType(TravelFlightSelectionDialog), findsNothing);
    await _tap(tester, _key('flight-select'));
    expect(joins, 1);
    expect(service.posts, isEmpty);
    await _tap(tester, _key('flight-select'));
    expect(find.byType(TravelFlightSelectionDialog), findsOneWidget);
    expect(service.posts, isEmpty);
  });
  testWidgets(
    'reconfirmation uses server flight snapshot and requires review',
    (tester) async {
      final service = _Service()
        ..state.addAll({
          'flight': _flight()
            ..['legs'] = [
              {..._leg(), 'flightNumber': 'BA180', 'departureTime': '23:00'},
            ],
          'flightVersion': 'v2',
          'ownSelection': {
            'status': 'booked',
            'legIds': ['leg-1'],
            'needsReconfirmation': true,
            'shareWithTravelers': false,
          },
        });
      await _mount(tester, _card(service));
      await _tap(tester, _key('flight-select'));
      expect(find.textContaining('BA180'), findsWidgets);
      await _tap(tester, find.text('Confirm my plans'));
      expect(service.posts, isEmpty);
      await _tap(tester, _key('flight-selection-reconfirm'));
      await _tap(tester, find.text('Confirm my plans'));
      expect(service.posts.single['data']['flightVersion'], 'v2');
    },
  );
  testWidgets('selection conflict keeps changes and requires a new review', (
    tester,
  ) async {
    final service = _Service()
      ..postFailure = const TravelCalendarException('Changed', statusCode: 409);
    await _mount(tester, _card(service));
    await _tap(tester, _key('flight-select'));
    await _tap(tester, find.text('Save my plans'));
    expect(
      find.textContaining('The flight or sharing settings changed'),
      findsOneWidget,
    );
    expect(find.byType(TravelFlightSelectionDialog), findsOneWidget);
  });
  testWidgets(
    'identity change closes the selection and clears protected names',
    (tester) async {
      final service = _Service();
      final auth = TripTestAuth(User(id: 'owner', name: 'Owner'));
      await _mount(
        tester,
        ChangeNotifierProvider<AuthProvider>.value(
          value: auth,
          child: _card(service),
        ),
      );
      await _tap(tester, _key('flight-select'));
      auth.account = null;
      auth.notifyListeners();
      await tester.pumpAndSettle();
      expect(find.byType(TravelFlightSelectionDialog), findsNothing);
      expect(find.textContaining('Max'), findsNothing);
      expect(_key('flight-card-flight-1'), findsNothing);
      await tester.pumpWidget(const SizedBox());
      auth.dispose();
    },
  );
  testWidgets(
    'cancelled flight keeps existing plans removable without rejoining',
    (tester) async {
      final service = _Service()
        ..state.addAll({
          'canSelect': false,
          'canRemoveSelection': true,
          'requiresAttendance': true,
          'ownSelection': {
            'status': 'booked',
            'legIds': ['leg-1'],
            'shareWithTravelers': false,
          },
        });
      var joins = 0;
      await _mount(
        tester,
        _card(
          service,
          entry: _flight()..['status'] = 'cancelled',
          join: () async {
            joins++;
          },
        ),
      );
      expect(find.textContaining('Cancelled'), findsOneWidget);
      await _tap(tester, _key('flight-select'));
      expect(joins, 0);
      expect(
        tester
            .widget<FilledButton>(
              find.widgetWithText(FilledButton, 'Save my plans'),
            )
            .onPressed,
        isNull,
      );
      await _tap(tester, _key('flight-selection-remove'));
      expect(service.posts.single['data']['status'], 'removed');
    },
  );
  testWidgets(
    'stale roster consent disables sharing but preserves private selection',
    (tester) async {
      final service = _Service()..state['canShareWithTravelers'] = false;
      await _mount(tester, _card(service));
      await _tap(tester, _key('flight-select'));
      expect(
        tester
            .widget<CheckboxListTile>(_key('flight-selection-share'))
            .onChanged,
        isNull,
      );
      expect(
        find.textContaining('Review the trip’s attendee-list privacy settings'),
        findsOneWidget,
      );
      await _tap(tester, find.text('Save my plans'));
      expect(service.posts.single['data']['shareWithTravelers'], false);
    },
  );
  testWidgets('access denied while saving dismisses protected flight details', (
    tester,
  ) async {
    final service = _Service();
    await _mount(tester, _card(service));
    await _tap(tester, _key('flight-select'));
    service.postFailure = const TravelCalendarException(
      'Access removed',
      statusCode: 403,
    );
    service.failure = const TravelCalendarException(
      'Access removed',
      statusCode: 403,
    );
    await _tap(tester, find.text('Save my plans'));
    expect(find.byType(TravelFlightSelectionDialog), findsNothing);
    expect(_key('flight-card-flight-1'), findsNothing);
  });
  testWidgets(
    'automatic roster reload keeps unchanged dialog but closes it when permissions change',
    (tester) async {
      final service = _Service();
      var token = 0;
      late VoidCallback update;
      await _mount(
        tester,
        StatefulBuilder(
          builder: (context, setState) {
            update = () => setState(() => token++);
            return TravelTripFlightCard(
              entry: _flight(),
              service: service,
              stopId: 'trip',
              refreshToken: token,
            );
          },
        ),
      );
      await _tap(tester, _key('flight-select'));
      update();
      await tester.pumpAndSettle();
      expect(find.byType(TravelFlightSelectionDialog), findsOneWidget);
      service.state['travelers'] = [
        {
          'userId': 'nora',
          'displayName': 'Nora',
          'status': 'booked',
          'legIds': ['leg-1'],
        },
      ];
      update();
      await tester.pumpAndSettle();
      expect(find.textContaining('Nora'), findsOneWidget);
      expect(find.textContaining('Max'), findsNothing);
      expect(find.byType(TravelFlightSelectionDialog), findsOneWidget);
      service.failure = const TravelCalendarException('Network unavailable');
      update();
      await tester.pumpAndSettle();
      expect(find.byType(TravelFlightSelectionDialog), findsOneWidget);
      service.failure = null;
      service.state['canShareWithTravelers'] = false;
      update();
      await tester.pumpAndSettle();
      expect(find.byType(TravelFlightSelectionDialog), findsNothing);
    },
  );
  testWidgets(
    'lookup supports route, prefilled date and stale result suppression',
    (tester) async {
      final service = _Service();
      Map<String, dynamic>? selected;
      await _mount(
        tester,
        TravelFlightLookup(
          stopId: 'trip',
          service: service,
          departureDate: '2099-06-01',
          onSelected: (leg) => selected = leg,
        ),
      );
      expect(
        tester.widget<TextField>(_key('flight-search-date')).controller!.text,
        '2099-06-01',
      );
      await _tap(tester, find.text('Route'));
      await tester.enterText(_key('flight-search-origin'), 'jfk');
      await tester.enterText(_key('flight-search-destination'), 'lhr');
      await _tap(tester, _key('flight-search-submit'));
      expect(service.gets.last['query'], {
        'origin': 'JFK',
        'destination': 'LHR',
        'departureDate': '2099-06-01',
      });
      await _tap(tester, _key('flight-search-use-0'));
      expect(selected!['providerId'], 'flight-instance');
      expect(selected!.containsKey('id'), false);
      await _tap(tester, find.text('Flight number'));
      final late = Completer<Map<String, dynamic>>();
      service.searches['BA178'] = late;
      await tester.enterText(_key('flight-search-number'), 'BA178');
      await tester.tap(_key('flight-search-submit'));
      await tester.pump();
      await tester.enterText(_key('flight-search-number'), 'BA180');
      late.complete({
        'configured': true,
        'flights': [_leg()],
      });
      await tester.pumpAndSettle();
      expect(_key('flight-search-result-0'), findsNothing);
    },
  );
  testWidgets(
    'lookup unconfigured and failure states keep manual fallback clear',
    (tester) async {
      final service = _Service()..search = {'configured': false, 'flights': []};
      await _mount(
        tester,
        TravelFlightLookup(
          stopId: 'trip',
          service: service,
          departureDate: '2099-06-01',
          onSelected: (_) {},
        ),
      );
      await tester.enterText(_key('flight-search-number'), 'BA178');
      await _tap(tester, _key('flight-search-submit'));
      expect(
        find.textContaining('Flight lookup is not available yet'),
        findsOneWidget,
      );
      service.failure = const TravelCalendarException('Provider unavailable');
      await _tap(tester, _key('flight-search-submit'));
      expect(
        find.textContaining('You can still add the flight manually below'),
        findsOneWidget,
      );
    },
  );
  testWidgets(
    'host lookup fills first leg and draft only, preserving lodging',
    (tester) async {
      final service = _Service();
      final stop = sampleTrip();
      await showTestDialog(
        tester,
        TravelTripLogisticsEditor(stop: stop, kind: 'flight', service: service),
      );
      await tester.enterText(_key('flight-search-number'), 'BA178');
      await _tap(tester, _key('flight-search-submit'));
      await _tap(tester, _key('flight-search-use-0'));
      await _tap(tester, find.text('Save logistics draft'));
      final write = service.puts.single;
      expect(write['flights'].last['legs'], hasLength(1));
      expect(write['flights'].last['legs'].first['flightNumber'], 'BA178');
      expect(write['flights'].last['label'], 'JFK to LHR');
      expect(write['lodging'], stop['lodging']);
      expect(service.posts, isEmpty);
    },
  );
  for (final (width, scale) in [(390.0, 1.0), (320.0, 2.0)]) {
    testWidgets('flight card and timeline fit $width at ${scale}x', (
      tester,
    ) async {
      final service = _Service();
      final boundary = await _mount(
        tester,
        _card(service, entry: _flight(connection: true)),
        width: width,
        scale: scale,
      );
      expect(tester.takeException(), isNull);
      await _capture(tester, boundary, '${width.toInt()}-${scale}x');
      await _tap(tester, _key('flight-details'));
      expect(find.text('+1 day'), findsWidgets);
      expect(find.text('Connection in LHR'), findsOneWidget);
      if (scale > 1) {
        expect(tester.getSize(find.text('LHR').last).height, lessThan(80));
        expect(tester.getSize(find.text('22:00').last).height, lessThan(70));
      }
      expect(tester.takeException(), isNull);
      await _capture(tester, boundary, 'details-${width.toInt()}-${scale}x');
      await _tap(tester, find.text('Close'));
      await _tap(tester, _key('flight-select'));
      expect(tester.takeException(), isNull);
      await _capture(tester, boundary, 'selection-${width.toInt()}-${scale}x');
    });
  }
}
