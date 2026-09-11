import 'package:flutter_test/flutter_test.dart';
import 'package:travel_angels/models/travel_calendar_route.dart';

void main() {
  const token = 'example_shared_token_123456789';

  test(
    'shared calendar links work with static hosting and native deep links',
    () {
      for (final location in [
        '/calendar/$token',
        'https://example.test/#/calendar/$token',
        'https://example.test/calendar/$token',
        'travelangels://calendar/$token',
      ]) {
        final route = TravelCalendarRoute.parse(location);
        expect(route?.kind, 'calendar');
        expect(route?.token, token);
      }
    },
  );

  test(
    'preferences and stop links retain purpose and reject unrelated routes',
    () {
      expect(
        TravelCalendarRoute.parse('/preferences/$token')?.kind,
        'preferences',
      );
      expect(TravelCalendarRoute.parse('/stops/$token')?.kind, 'stops');
      for (final location in [
        '/',
        '/calendar/short',
        '/calendar/$token/extra',
        '/calendar/${Uri.encodeComponent('a/b' * 10)}',
        'travelangels://credits/purchase/success',
      ]) {
        expect(TravelCalendarRoute.parse(location), isNull, reason: location);
      }
    },
  );
}
