import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:travel_angels/main.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';

class _SharedCalendarService extends TravelCalendarService {
  @override
  Future<Map<String, dynamic>> get(
    String path, {
    Map<String, dynamic>? query,
  }) async => {
    'calendar': {
      'id': 'calendar-1',
      'title': 'Our travel calendar',
      'canManage': false,
    },
    'stops': [
      {
        'id': 'stop-1',
        'title': 'Meet us in Lisbon',
        'destination': 'Lisbon, Portugal',
        'startDate': '2032-12-20',
        'endDate': '2033-01-10',
        'datePrecision': 'fixed',
        'status': 'confirmed',
        'published': true,
        'joinOpen': true,
      },
    ],
  };
}

void main() {
  testWidgets('shared links open permitted plans without the app login gate', (
    tester,
  ) async {
    SharedPreferences.setMockInitialValues({});
    await tester.pumpWidget(
      MyApp(
        initialRoute: '/calendar/example_shared_token_123456789',
        calendarService: _SharedCalendarService(),
        listenForDeepLinks: false,
      ),
    );
    await tester.pumpAndSettle();
    expect(find.text('Meet us in Lisbon'), findsOneWidget);
    expect(find.text('Our travel calendar'), findsOneWidget);
    expect(find.text('Phone number'), findsNothing);
    expect(tester.takeException(), isNull);
  });
}
