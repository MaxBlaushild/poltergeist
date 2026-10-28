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
import 'package:travel_angels/screens/my_network_screen.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/auth_service.dart';
import 'package:travel_angels/services/friend_service.dart';
import 'package:travel_angels/theme/app_theme.dart';

import 'support/app_fonts.dart';

Finder _key(String value) => find.byKey(ValueKey(value));

User _person(String id, String name, [String? username]) =>
    User(id: id, name: name, username: username);

final _me = _person('me', 'Isabella Montague', 'isabellatravels');

FriendInvite _invite(String id, User from, User to) => FriendInvite(
  id: id,
  inviterId: from.id!,
  inviteeId: to.id!,
  inviter: from,
  invitee: to,
  createdAt: DateTime(2026, 10, 24),
);

class _CircleAuth extends AuthProvider {
  _CircleAuth() : super(AuthService(APIClient('https://example.test')));
  User current = _me;
  @override
  User get user => current;
  @override
  bool get isAuthenticated => true;
  @override
  bool get loading => false;
  @override
  Future<void> verifyToken() async {}

  void change(User user) {
    current = user;
    notifyListeners();
  }
}

class _CircleFriends extends FriendService {
  _CircleFriends() : super(APIClient('https://example.test'));
  List<User> friends = [];
  List<FriendInvite> invites = [];
  Future<List<User>> Function()? loadFriends;
  Future<List<FriendInvite>> Function()? loadInvites;
  Future<void> Function(String)? accept;
  Future<void> Function(String)? remove;
  int friendLoads = 0;
  int inviteLoads = 0;
  final accepted = <String>[];
  final removed = <String>[];

  @override
  Future<List<User>> getFriends() async {
    friendLoads++;
    return loadFriends == null ? List.of(friends) : await loadFriends!();
  }

  @override
  Future<List<FriendInvite>> getFriendInvites() async {
    inviteLoads++;
    return loadInvites == null ? List.of(invites) : await loadInvites!();
  }

  @override
  Future<void> acceptFriendInvite(String inviteId) async {
    accepted.add(inviteId);
    await accept?.call(inviteId);
    final invite = invites.firstWhere((item) => item.id == inviteId);
    if (invite.inviter != null) friends.add(invite.inviter!);
    invites.removeWhere((item) => item.id == inviteId);
  }

  @override
  Future<void> deleteFriendInvite(String inviteId) async {
    removed.add(inviteId);
    await remove?.call(inviteId);
    invites.removeWhere((item) => item.id == inviteId);
  }

  @override
  Future<List<User>> searchUsers(String query) async => [];
}

_CircleFriends _sample() => _CircleFriends()
  ..friends = [
    _person('maya', 'Maya Chen', 'mayainmotion'),
    _person('leo', 'Leo Martinez', 'leoinlisbon'),
    _person('ava', 'Ava Thompson', 'avasatlas'),
    _person('oliver', 'Oliver Alexander Montgomery', 'oliveraroundtheworld'),
    _person('sofia', 'Sofia Rossi', 'sofiawanders'),
  ]
  ..invites = [
    _invite('incoming', _person('nina', 'Nina Patel', 'ninatravels'), _me),
    _invite(
      'outgoing',
      _me,
      _person('james', 'James Wilson', 'jamesontheroad'),
    ),
    _invite(
      'unrelated',
      _person('outsider1', 'Private unrelated sender'),
      _person('outsider2', 'Private unrelated recipient'),
    ),
  ];

Future<GlobalKey> _mount(
  WidgetTester tester,
  _CircleFriends friends, {
  _CircleAuth? auth,
  double width = 390,
  double scale = 1,
}) async {
  final size = Size(width, width >= 600 ? 1000 : 844);
  tester.view.physicalSize = size;
  tester.view.devicePixelRatio = 1;
  addTearDown(tester.view.resetPhysicalSize);
  addTearDown(tester.view.resetDevicePixelRatio);
  final currentAuth = auth ?? _CircleAuth();
  addTearDown(currentAuth.dispose);
  final boundary = GlobalKey();
  await tester.pumpWidget(
    ChangeNotifierProvider<AuthProvider>.value(
      value: currentAuth,
      child: MaterialApp(
        debugShowCheckedModeBanner: false,
        theme: AppTheme.lightTheme,
        builder: (context, child) => MediaQuery(
          data: MediaQuery.of(
            context,
          ).copyWith(textScaler: TextScaler.linear(scale)),
          child: child!,
        ),
        home: RepaintBoundary(
          key: boundary,
          child: MyNetworkScreen(service: friends),
        ),
      ),
    ),
  );
  await tester.pump();
  return boundary;
}

Future<void> _tap(WidgetTester tester, String key) async {
  final finder = _key(key);
  await tester.ensureVisible(finder);
  await tester.pumpAndSettle();
  await tester.tap(finder);
  await tester.pumpAndSettle();
}

Future<void> _capture(
  GlobalKey boundary,
  WidgetTester tester,
  String name,
) async {
  if (!const bool.fromEnvironment('TRIP_SCREENSHOTS')) return;
  await tester.runAsync(() async {
    final image =
        await (boundary.currentContext!.findRenderObject()!
                as RenderRepaintBoundary)
            .toImage();
    final data = await image.toByteData(format: ui.ImageByteFormat.png);
    await File(
      '/private/tmp/travel-angels-circle-$name.png',
    ).writeAsBytes(data!.buffer.asUint8List());
    image.dispose();
  });
}

void main() {
  setUpAll(loadAppFonts);

  testWidgets(
    'Circle routes friends, received requests and sent invites separately',
    (tester) async {
      final service = _sample();
      await _mount(tester, service);
      await tester.pumpAndSettle();
      expect(_key('circle-friend-maya'), findsOneWidget);
      expect(_key('circle-page-next'), findsNothing);
      expect(_key('circle-accept-incoming'), findsNothing);
      expect(_key('circle-cancel-outgoing'), findsNothing);

      await _tap(tester, 'circle-tab-requests');
      expect(_key('circle-accept-incoming'), findsOneWidget);
      expect(_key('circle-decline-incoming'), findsOneWidget);
      expect(_key('circle-cancel-outgoing'), findsNothing);
      expect(find.text('Private unrelated sender'), findsNothing);
      expect(find.text('Private unrelated recipient'), findsNothing);

      await _tap(tester, 'circle-tab-sent');
      expect(_key('circle-cancel-outgoing'), findsOneWidget);
      expect(_key('circle-accept-incoming'), findsNothing);
      expect(find.text('Private unrelated sender'), findsNothing);
      await _tap(tester, 'circle-tab-friends');
      await _tap(tester, 'circle-find-people');
      expect(_key('circle-people-close'), findsOneWidget);
      await _tap(tester, 'circle-people-close');
      expect(_key('circle-friend-maya'), findsOneWidget);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'friend filters and pagination survive returning from Find people',
    (tester) async {
      final service = _CircleFriends()
        ..friends = List.generate(12, (index) {
          final number = (index + 1).toString().padLeft(2, '0');
          return _person('friend$number', 'Traveler $number', 'atlas$number');
        });
      await _mount(tester, service);
      await tester.pumpAndSettle();
      expect(_key('circle-friend-friend01'), findsOneWidget);
      expect(_key('circle-friend-friend11'), findsNothing);
      await _tap(tester, 'circle-page-next');
      expect(_key('circle-friend-friend11'), findsOneWidget);
      expect(_key('circle-friend-friend01'), findsNothing);
      await _tap(tester, 'circle-find-people');
      await _tap(tester, 'circle-people-close');
      expect(_key('circle-friend-friend11'), findsOneWidget);
      expect(service.friendLoads, 2);

      await _tap(tester, 'circle-search-toggle');
      await tester.enterText(_key('circle-search'), 'ATLAS03');
      await tester.pumpAndSettle();
      expect(_key('circle-friend-friend03'), findsOneWidget);
      expect(_key('circle-friend-friend11'), findsNothing);
      expect(_key('circle-page-next'), findsNothing);
      await _tap(tester, 'circle-find-people');
      await _tap(tester, 'circle-people-close');
      expect(
        tester.widget<TextField>(_key('circle-search')).controller!.text,
        'ATLAS03',
      );
      expect(_key('circle-friend-friend03'), findsOneWidget);
      await tester.enterText(_key('circle-search'), 'traveler 12');
      await tester.pumpAndSettle();
      expect(_key('circle-friend-friend12'), findsOneWidget);
      await tester.enterText(_key('circle-search'), 'nobody matches this');
      await tester.pumpAndSettle();
      expect(_key('circle-friend-friend12'), findsNothing);
      await tester.enterText(_key('circle-search'), '');
      await tester.pumpAndSettle();
      expect(_key('circle-friend-friend01'), findsOneWidget);
      await _tap(tester, 'circle-tab-sent');
      await _tap(tester, 'circle-find-people');
      await _tap(tester, 'circle-people-close');
      expect(_key('circle-friend-friend01'), findsNothing);
      expect(tester.takeException(), isNull);
    },
  );

  for (final failFriends in [true, false]) {
    testWidgets(
      'independent ${failFriends ? 'friends' : 'invites'} load error leaves other data usable and retries',
      (tester) async {
        final service = _sample();
        if (failFriends) {
          service.loadFriends = () async => throw StateError('offline');
        } else {
          service.loadInvites = () async => throw StateError('offline');
        }
        await _mount(tester, service);
        await tester.pumpAndSettle();
        if (failFriends) {
          expect(_key('circle-friends-retry'), findsOneWidget);
          await _tap(tester, 'circle-tab-requests');
          expect(_key('circle-accept-incoming'), findsOneWidget);
          await _tap(tester, 'circle-tab-friends');
          service.loadFriends = null;
          await _tap(tester, 'circle-friends-retry');
          expect(_key('circle-friend-maya'), findsOneWidget);
          expect(service.inviteLoads, 1);
        } else {
          expect(_key('circle-friend-maya'), findsOneWidget);
          await _tap(tester, 'circle-tab-requests');
          expect(_key('circle-invites-retry'), findsOneWidget);
          service.loadInvites = null;
          await _tap(tester, 'circle-invites-retry');
          expect(_key('circle-accept-incoming'), findsOneWidget);
          expect(service.friendLoads, 1);
        }
        expect(tester.takeException(), isNull);
      },
    );
  }

  for (final action in ['accept', 'decline', 'cancel']) {
    testWidgets(
      '$action prevents duplicate requests and supports retry after failure',
      (tester) async {
        final service = _sample();
        final pending = Completer<void>();
        final isAccept = action == 'accept';
        final id = action == 'cancel' ? 'outgoing' : 'incoming';
        final key = 'circle-$action-$id';
        service.invites.add(
          action == 'cancel'
              ? _invite('other', _me, _person('other', 'Other pending friend'))
              : _invite('other', _person('other', 'Other pending friend'), _me),
        );
        if (isAccept) {
          service.accept = (_) => pending.future;
        } else {
          service.remove = (_) => pending.future;
        }
        await _mount(tester, service);
        await tester.pumpAndSettle();
        await _tap(
          tester,
          action == 'cancel' ? 'circle-tab-sent' : 'circle-tab-requests',
        );
        await tester.ensureVisible(_key(key));
        await tester.tap(_key(key));
        await tester.pump();
        expect(isAccept ? service.accepted : service.removed, [id]);
        await tester.tap(_key(key), warnIfMissed: false);
        await tester.pump();
        expect(isAccept ? service.accepted : service.removed, [id]);
        if (action != 'cancel') {
          await tester.tap(
            _key('circle-${isAccept ? 'decline' : 'accept'}-$id'),
            warnIfMissed: false,
          );
          await tester.pump();
          expect(isAccept ? service.removed : service.accepted, isEmpty);
        }
        expect(service.invites.any((invite) => invite.id == id), isTrue);
        pending.completeError(StateError('Temporary connection failure'));
        await tester.pumpAndSettle();
        expect(_key(key), findsOneWidget);
        expect(find.byType(SnackBar), findsOneWidget);
        service.accept = null;
        service.remove = null;
        await _tap(tester, key);
        expect(isAccept ? service.accepted : service.removed, [id, id]);
        expect(_key(key), findsNothing);
        expect(
          _key('circle-${action == 'cancel' ? 'cancel' : 'accept'}-other'),
          findsOneWidget,
        );
        if (isAccept) {
          await _tap(tester, 'circle-tab-friends');
          expect(_key('circle-friend-nina'), findsOneWidget);
        }
        expect(tester.takeException(), isNull);
      },
    );
  }

  testWidgets(
    'disposing Circle during acceptance prevents a late reload or UI update',
    (tester) async {
      final service = _sample();
      final pending = Completer<void>();
      service.accept = (_) => pending.future;
      await _mount(tester, service);
      await tester.pumpAndSettle();
      await _tap(tester, 'circle-tab-requests');
      await tester.tap(_key('circle-accept-incoming'));
      await tester.pump();
      await tester.pumpWidget(const SizedBox());
      pending.complete();
      await tester.pumpAndSettle();
      expect(service.friendLoads, 1);
      expect(service.inviteLoads, 1);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets('account change clears prior people and rejects late loads', (
    tester,
  ) async {
    final service = _CircleFriends();
    final auth = _CircleAuth();
    final oldFriends = Completer<List<User>>();
    final oldInvites = Completer<List<FriendInvite>>();
    service.loadFriends = () => oldFriends.future;
    service.loadInvites = () => oldInvites.future;
    await _mount(tester, service, auth: auth);
    service.loadFriends = () async => [_person('new', 'New account friend')];
    service.loadInvites = () async => [];
    auth.change(_person('other-account', 'Another traveler'));
    await tester.pump();
    await tester.pump();
    oldFriends.complete([_person('old', 'Old account private friend')]);
    oldInvites.complete([
      _invite('old', _person('old', 'Old invitation'), _me),
    ]);
    await tester.pumpAndSettle();
    expect(_key('circle-friend-new'), findsOneWidget);
    expect(_key('circle-friend-old'), findsNothing);
    await _tap(tester, 'circle-tab-requests');
    expect(_key('circle-accept-old'), findsNothing);
    expect(tester.takeException(), isNull);
  });

  for (final (width, scale) in [(1200.0, 1.0), (390.0, 1.0), (320.0, 2.0)]) {
    testWidgets(
      'Circle fits ${width.toInt()}px with ${scale}x text and real fonts',
      (tester) async {
        final boundary = await _mount(
          tester,
          _sample(),
          width: width,
          scale: scale,
        );
        await tester.pumpAndSettle();
        expect(_key('circle-find-people'), findsOneWidget);
        expect(_key('circle-search-toggle'), findsOneWidget);
        expect(tester.takeException(), isNull);
        await _capture(boundary, tester, '${width.toInt()}-${scale}x');
        await _tap(tester, 'circle-search-toggle');
        await tester.enterText(_key('circle-search'), 'oliver');
        await tester.pumpAndSettle();
        expect(_key('circle-friend-oliver'), findsOneWidget);
        expect(tester.takeException(), isNull);
        await tester.enterText(_key('circle-search'), '');
        await tester.pumpAndSettle();
        await _tap(tester, 'circle-tab-requests');
        await tester.ensureVisible(_key('circle-accept-incoming'));
        await tester.pumpAndSettle();
        expect(_key('circle-accept-incoming').hitTestable(), findsOneWidget);
        expect(_key('circle-decline-incoming').hitTestable(), findsOneWidget);
        expect(tester.takeException(), isNull);
        if (width == 390) await _capture(boundary, tester, '390-requests');
        await _tap(tester, 'circle-tab-sent');
        await tester.ensureVisible(_key('circle-cancel-outgoing'));
        await tester.pumpAndSettle();
        expect(_key('circle-cancel-outgoing').hitTestable(), findsOneWidget);
        expect(tester.takeException(), isNull);
      },
    );
  }
}
