import 'dart:async';
import 'dart:io';
import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/models/friend_invite.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/auth_service.dart';
import 'package:travel_angels/services/friend_service.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/circle_people_dialog.dart';
import 'package:travel_angels/widgets/circle_person_row.dart';

import 'support/app_fonts.dart';

Finder _key(String key) => find.byKey(ValueKey(key));

class _Auth extends AuthProvider {
  _Auth() : super(AuthService(APIClient('https://example.test')));
  User? current = User(id: 'me', username: 'traveler');
  @override
  User? get user => current;
  @override
  Future<void> verifyToken() async {}
  void change(User? user) {
    current = user;
    notifyListeners();
  }
}

class _Friends extends FriendService {
  _Friends() : super(APIClient('https://example.test'));
  List<User> results = [
    User(id: 'maya', name: 'Maya Chen', username: 'mayatravels'),
    User(id: 'alex', name: 'Alex Rivera', username: 'alexrivera'),
  ];
  final searches = <String>[];
  final sent = <String>[];
  Object? searchError;
  Object? sendError;
  final pendingSearches = <String, Completer<List<User>>>{};
  Completer<FriendInvite>? pendingInvite;

  @override
  Future<List<User>> searchUsers(String query) async {
    searches.add(query);
    if (pendingSearches.containsKey(query)) {
      return pendingSearches[query]!.future;
    }
    if (searchError != null) throw searchError!;
    return results;
  }

  @override
  Future<FriendInvite> createFriendInvite(String inviteeId) async {
    sent.add(inviteeId);
    if (pendingInvite != null) return pendingInvite!.future;
    if (sendError != null) throw sendError!;
    return FriendInvite(
      id: 'sent-$inviteeId',
      inviterId: 'me',
      inviteeId: inviteeId,
    );
  }
}

Future<GlobalKey> _open(
  WidgetTester tester,
  _Friends service, {
  _Auth? auth,
  List<User> friends = const [],
  List<FriendInvite> invites = const [],
  ValueChanged<String?>? onClosed,
  double width = 1200,
  double scale = 1,
}) async {
  final identity = auth ?? _Auth();
  addTearDown(identity.dispose);
  final size = Size(width, 844);
  await tester.binding.setSurfaceSize(size);
  addTearDown(() => tester.binding.setSurfaceSize(null));
  final boundary = GlobalKey();
  await tester.pumpWidget(
    ChangeNotifierProvider<AuthProvider>.value(
      value: identity,
      child: RepaintBoundary(
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
          home: Builder(
            builder: (context) => Scaffold(
              body: TextButton(
                onPressed: () async {
                  final result = await showDialog<String>(
                    context: context,
                    builder: (_) => CirclePeopleDialog(
                      service: service,
                      friends: friends,
                      invites: invites,
                    ),
                  );
                  onClosed?.call(result);
                },
                child: const Text('Find people'),
              ),
            ),
          ),
        ),
      ),
    ),
  );
  await tester.tap(find.text('Find people'));
  await tester.pumpAndSettle();
  return boundary;
}

Future<void> _search(WidgetTester tester, [String query = 'ma']) async {
  await tester.enterText(_key('circle-people-search'), query);
  await tester.pump(const Duration(milliseconds: 350));
  await tester.pumpAndSettle();
}

Future<void> _tap(WidgetTester tester, Finder finder) async {
  await tester.ensureVisible(finder);
  await tester.pumpAndSettle();
  await tester.tap(finder);
  await tester.pumpAndSettle();
}

Future<void> _capture(
  WidgetTester tester,
  GlobalKey boundary,
  String label,
) async {
  if (!const bool.fromEnvironment('TRIP_SCREENSHOTS')) return;
  await tester.runAsync(() async {
    final image =
        await (boundary.currentContext!.findRenderObject()!
                as RenderRepaintBoundary)
            .toImage();
    final bytes = await image.toByteData(format: ui.ImageByteFormat.png);
    await File(
      '/private/tmp/travel-angels-circle-people-$label.png',
    ).writeAsBytes(bytes!.buffer.asUint8List());
    image.dispose();
  });
}

void main() {
  setUpAll(loadAppFonts);

  testWidgets(
    'debounces search and remembers newly sent invites across queries',
    (tester) async {
      final service = _Friends();
      await _open(tester, service);
      expect(
        find.text('Search by username to invite someone to your circle.'),
        findsOneWidget,
      );
      await tester.enterText(_key('circle-people-search'), 'm');
      await tester.pump(const Duration(milliseconds: 200));
      await tester.enterText(_key('circle-people-search'), '@maya');
      await tester.pump(const Duration(milliseconds: 350));
      await tester.pumpAndSettle();
      expect(service.searches, ['maya']);
      await _tap(tester, _key('circle-invite-maya'));
      expect(service.sent, ['maya']);
      expect(find.text('Invite sent'), findsOneWidget);
      expect(_key('circle-invite-maya'), findsNothing);
      await _search(tester, 'another');
      expect(find.text('Invite sent'), findsOneWidget);
      expect(service.sent, hasLength(1));
    },
  );

  testWidgets(
    'existing and pending people cannot be reinvited; incoming opens requests',
    (tester) async {
      final service = _Friends()
        ..results = [
          User(id: 'me', name: 'You'),
          User(id: '', name: 'Missing identity'),
          User(id: 'friend', name: 'Your friend'),
          User(id: 'sent', name: 'Already invited'),
          User(id: 'incoming', name: 'Incoming person'),
        ];
      String? closed;
      await _open(
        tester,
        service,
        friends: [User(id: 'friend')],
        invites: [
          FriendInvite(id: 'out', inviterId: 'me', inviteeId: 'sent'),
          FriendInvite(id: 'in', inviterId: 'incoming', inviteeId: 'me'),
        ],
        onClosed: (value) => closed = value,
      );
      await _search(tester);
      expect(find.text('You'), findsNothing);
      expect(find.text('Missing identity'), findsNothing);
      expect(find.text('In your circle'), findsOneWidget);
      expect(find.text('Invite sent'), findsOneWidget);
      expect(find.byType(OutlinedButton), findsNothing);
      await _tap(tester, _key('circle-review-incoming'));
      expect(closed, 'requests');
      expect(service.sent, isEmpty);
    },
  );

  testWidgets(
    'search and invite failures have inline retry without losing results',
    (tester) async {
      final service = _Friends()..searchError = Exception('offline');
      await _open(tester, service);
      await _search(tester);
      expect(find.text('Could not search people. Try again.'), findsOneWidget);
      service.searchError = null;
      await _tap(tester, _key('circle-people-retry'));
      expect(find.text('Maya Chen'), findsOneWidget);
      service.sendError = Exception('offline');
      await _tap(tester, _key('circle-invite-maya'));
      expect(find.text('Invite could not be sent. Try again.'), findsOneWidget);
      expect(find.text('Retry invite'), findsOneWidget);
      service.sendError = null;
      await _tap(tester, _key('circle-invite-maya'));
      expect(find.text('Invite sent'), findsOneWidget);
      expect(find.text('Invite could not be sent. Try again.'), findsNothing);
    },
  );

  testWidgets(
    'older search and disposed dialog completions never restore stale results',
    (tester) async {
      final service = _Friends();
      final first = Completer<List<User>>();
      final second = Completer<List<User>>();
      service.pendingSearches.addAll({'first': first, 'second': second});
      await _open(tester, service);
      await tester.enterText(_key('circle-people-search'), 'first');
      await tester.pump(const Duration(milliseconds: 350));
      await tester.enterText(_key('circle-people-search'), 'second');
      await tester.pump(const Duration(milliseconds: 350));
      second.complete([User(id: 'new', name: 'Current result')]);
      await tester.pumpAndSettle();
      first.complete([User(id: 'old', name: 'Stale result')]);
      await tester.pumpAndSettle();
      expect(find.text('Current result'), findsOneWidget);
      expect(find.text('Stale result'), findsNothing);
      final after = Completer<List<User>>();
      service.pendingSearches['after'] = after;
      await tester.enterText(_key('circle-people-search'), 'after');
      await tester.pump(const Duration(milliseconds: 350));
      await tester.tap(_key('circle-people-close'));
      await tester.pumpAndSettle();
      after.complete([User(id: 'late', name: 'Late result')]);
      await tester.pumpAndSettle();
      expect(find.byType(CirclePeopleDialog), findsNothing);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets('account switch dismisses dialog while an invite is pending', (
    tester,
  ) async {
    final auth = _Auth();
    final pending = Completer<FriendInvite>();
    final service = _Friends()..pendingInvite = pending;
    await _open(tester, service, auth: auth);
    await _search(tester);
    await _tap(tester, _key('circle-invite-maya'));
    expect(find.text('Sending…'), findsOneWidget);
    auth.change(User(id: 'other'));
    await tester.pumpAndSettle();
    expect(find.byType(CirclePeopleDialog), findsNothing);
    pending.complete(
      FriendInvite(id: 'sent', inviterId: 'me', inviteeId: 'maya'),
    );
    await tester.pumpAndSettle();
    expect(service.sent, ['maya']);
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'empty results describe eligibility and person rows omit private profile details',
    (tester) async {
      final service = _Friends()..results = [];
      await _open(tester, service);
      await _search(tester);
      expect(
        find.text('No new people found. Try another username.'),
        findsOneWidget,
      );
      service.results = [
        User(
          id: 'ada',
          name: '  Ada   Lovelace  ',
          username: 'ada',
          bio: 'Private biography',
          locationAddress: 'Private address',
          profilePictureUrl: 'https://example.test/broken-avatar.png',
        ),
      ];
      await _search(tester, 'ada');
      expect(circlePersonName(null), 'Traveler');
      expect(find.text('AL'), findsOneWidget);
      expect(find.text('Private biography'), findsNothing);
      expect(find.text('Private address'), findsNothing);
      expect(tester.takeException(), isNull);
    },
  );

  for (final (width, scale) in [(1200.0, 1.0), (390.0, 1.0), (320.0, 2.0)]) {
    testWidgets('people dialog stays usable at $width and ${scale}x', (
      tester,
    ) async {
      final service = _Friends()
        ..results = [
          User(id: 'maya', name: 'Maya Chen', username: 'mayatravels'),
          User(
            id: 'isabella',
            name: 'Isabella Montague',
            username: 'isabellatravels',
          ),
          User(id: 'alex', name: 'Alex Rivera', username: 'alexrivera'),
        ];
      final boundary = await _open(tester, service, width: width, scale: scale);
      await _search(tester);
      await _capture(tester, boundary, '${width.toInt()}-${scale}x');
      expect(_key('circle-people-close').hitTestable(), findsOneWidget);
      await _tap(tester, _key('circle-invite-alex'));
      expect(service.sent, ['alex']);
      expect(find.text('Invite sent'), findsOneWidget);
      expect(tester.takeException(), isNull);
      await _tap(tester, _key('circle-people-close'));
      expect(find.byType(CirclePeopleDialog), findsNothing);
    });
  }
}
