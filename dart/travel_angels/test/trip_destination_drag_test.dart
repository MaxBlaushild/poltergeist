import 'dart:ui' as ui;

import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/models/travel_destination.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/auth_service.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/trip_destination_picker.dart';

import 'support/app_fonts.dart';

const _locations = [
  TravelDestination(id: 'lisbon', name: 'Lisbon'),
  TravelDestination(id: 'tokyo', name: 'Tokyo'),
  TravelDestination(id: 'san-francisco', name: 'San Francisco'),
  TravelDestination(id: 'mexico-city', name: 'Mexico City'),
];

Finder _chip(String id) => find.byKey(ValueKey('trip-location-$id'));

class _DragScenario extends ChangeNotifier {
  _DragScenario([List<TravelDestination>? value])
    : value = value ?? List.of(_locations);

  List<TravelDestination> value;
  bool enabled = true;
  int changes = 0;
  List<String?> get ids => value.map((location) => location.id).toList();

  void select(List<TravelDestination> selected) {
    value = selected;
    ++changes;
    notifyListeners();
  }

  void disable() {
    enabled = false;
    notifyListeners();
  }
}

class _Destinations extends TravelCalendarService {
  @override
  Future<Map<String, dynamic>> get(
    String path, {
    Map<String, dynamic>? query,
  }) async => {'destinations': []};
}

class _Identity extends AuthProvider {
  _Identity() : super(AuthService(APIClient('https://example.test')));
  User? current = User(id: 'traveler');
  @override
  User? get user => current;
  @override
  Future<void> verifyToken() async {}
  void switchAccount() {
    current = User(id: 'someone-else');
    notifyListeners();
  }
}

Future<void> _open(
  WidgetTester tester,
  _DragScenario scenario, {
  double width = 260,
  TargetPlatform platform = TargetPlatform.android,
  _Identity? auth,
}) async {
  await tester.binding.setSurfaceSize(const Size(400, 800));
  addTearDown(() => tester.binding.setSurfaceSize(null));
  addTearDown(scenario.dispose);
  final service = _Destinations();
  Widget app = MaterialApp(
    theme: AppTheme.lightTheme.copyWith(platform: platform),
    home: Scaffold(
      body: Center(
        child: SizedBox(
          width: width,
          child: AnimatedBuilder(
            animation: scenario,
            builder: (_, _) => TripDestinationPicker(
              service: service,
              value: scenario.value,
              enabled: scenario.enabled,
              onChanged: scenario.select,
            ),
          ),
        ),
      ),
    ),
  );
  if (auth != null) {
    app = ChangeNotifierProvider<AuthProvider>.value(value: auth, child: app);
  }
  await tester.pumpWidget(app);
  await tester.pumpAndSettle();
}

Future<TestGesture> _beginDrag(
  WidgetTester tester,
  String id, {
  ui.PointerDeviceKind kind = ui.PointerDeviceKind.touch,
}) async {
  final rect = tester.getRect(_chip(id));
  final gesture = await tester.startGesture(
    rect.centerLeft + const Offset(24, 0),
    kind: kind,
  );
  await tester.pump(
    kind == ui.PointerDeviceKind.touch
        ? kLongPressTimeout + const Duration(milliseconds: 100)
        : const Duration(milliseconds: 16),
  );
  return gesture;
}

Future<void> _dropOn(
  WidgetTester tester,
  TestGesture gesture,
  Offset target,
) async {
  await gesture.moveTo(target);
  await tester.pump();
  await gesture.up();
  await tester.pumpAndSettle();
}

void main() {
  setUpAll(loadAppFonts);

  testWidgets('touch drag reorders across wrapped rows only when dropped', (
    tester,
  ) async {
    final scenario = _DragScenario();
    await _open(tester, scenario);
    expect(
      tester.getTopLeft(_chip('mexico-city')).dy,
      greaterThan(tester.getTopLeft(_chip('tokyo')).dy),
    );
    final target = tester.getCenter(_chip('tokyo'));
    final gesture = await _beginDrag(tester, 'mexico-city');
    await gesture.moveTo(target);
    await tester.pump();
    expect(scenario.changes, 0);
    expect(scenario.ids, ['lisbon', 'tokyo', 'san-francisco', 'mexico-city']);
    await gesture.up();
    await tester.pumpAndSettle();
    expect(scenario.ids, ['lisbon', 'mexico-city', 'tokyo', 'san-francisco']);
    expect(scenario.changes, 1);
    expect(find.byType(PopupMenuButton<String>), findsNothing);
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'cancelling a hovered drag preserves order and keeps removal usable',
    (tester) async {
      final scenario = _DragScenario();
      await _open(tester, scenario);
      final target = tester.getCenter(_chip('lisbon'));
      final gesture = await _beginDrag(tester, 'mexico-city');
      await gesture.moveTo(target);
      await tester.pump();
      await gesture.cancel();
      await tester.pumpAndSettle();
      expect(scenario.ids, ['lisbon', 'tokyo', 'san-francisco', 'mexico-city']);
      expect(scenario.changes, 0);
      await tester.tap(find.byTooltip('Remove Tokyo'));
      await tester.pumpAndSettle();
      expect(scenario.ids, ['lisbon', 'san-francisco', 'mexico-city']);
      expect(scenario.changes, 1);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets('dropping outside the location pills does not reorder', (
    tester,
  ) async {
    final scenario = _DragScenario();
    await _open(tester, scenario);
    final gesture = await _beginDrag(tester, 'tokyo');
    await _dropOn(tester, gesture, const Offset(12, 12));
    expect(scenario.ids, ['lisbon', 'tokyo', 'san-francisco', 'mexico-city']);
    expect(scenario.changes, 0);
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'desktop mouse drag starts immediately and can move a pill later',
    (tester) async {
      final scenario = _DragScenario();
      await _open(tester, scenario, platform: TargetPlatform.macOS);
      final target = tester.getCenter(_chip('mexico-city'));
      final gesture = await _beginDrag(
        tester,
        'lisbon',
        kind: ui.PointerDeviceKind.mouse,
      );
      await _dropOn(tester, gesture, target);
      expect(scenario.ids, ['tokyo', 'san-francisco', 'mexico-city', 'lisbon']);
      expect(scenario.changes, 1);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets('disabling the picker during a drag prevents a stale drop', (
    tester,
  ) async {
    final scenario = _DragScenario();
    await _open(tester, scenario);
    final target = tester.getCenter(_chip('lisbon'));
    final gesture = await _beginDrag(tester, 'mexico-city');
    await gesture.moveTo(target);
    await tester.pump();
    scenario.disable();
    await tester.pump();
    await gesture.up();
    await tester.pumpAndSettle();
    expect(scenario.ids, ['lisbon', 'tokyo', 'san-francisco', 'mexico-city']);
    expect(scenario.changes, 0);
    expect(
      tester
          .widget<IconButton>(
            find.widgetWithIcon(IconButton, Icons.close).first,
          )
          .onPressed,
      isNull,
    );
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'account invalidation during a drag prevents applying its old payload',
    (tester) async {
      final scenario = _DragScenario();
      final auth = _Identity();
      addTearDown(auth.dispose);
      await _open(tester, scenario, auth: auth);
      final target = tester.getCenter(_chip('lisbon'));
      final gesture = await _beginDrag(tester, 'mexico-city');
      await gesture.moveTo(target);
      await tester.pump();
      auth.switchAccount();
      await tester.pump();
      await gesture.up();
      await tester.pumpAndSettle();
      expect(scenario.ids, ['lisbon', 'tokyo', 'san-francisco', 'mexico-city']);
      expect(scenario.changes, 0);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'keyboard shortcuts reorder the focused pill without opening a menu',
    (tester) async {
      final scenario = _DragScenario();
      await _open(tester, scenario, platform: TargetPlatform.macOS);
      Focus.of(tester.element(find.text('Tokyo'))).requestFocus();
      await tester.pump();
      await tester.sendKeyDownEvent(LogicalKeyboardKey.altLeft);
      await tester.sendKeyEvent(LogicalKeyboardKey.arrowLeft);
      await tester.sendKeyUpEvent(LogicalKeyboardKey.altLeft);
      await tester.pumpAndSettle();
      expect(scenario.ids, ['tokyo', 'lisbon', 'san-francisco', 'mexico-city']);
      expect(scenario.changes, 1);
      expect(find.byType(PopupMenuButton<String>), findsNothing);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets('screen-reader move actions remain available without dot menus', (
    tester,
  ) async {
    final semantics = tester.ensureSemantics();
    try {
      final scenario = _DragScenario();
      await _open(tester, scenario);
      final node = tester.getSemantics(_chip('tokyo'));
      final actions = node.getSemanticsData().customSemanticsActionIds ?? [];
      final earlier = actions.singleWhere(
        (id) => CustomSemanticsAction.getAction(id)?.label == 'Move earlier',
      );
      final later = actions.singleWhere(
        (id) => CustomSemanticsAction.getAction(id)?.label == 'Move later',
      );
      expect(earlier, isNot(later));
      tester
          .renderObject(_chip('tokyo'))
          .owner!
          .semanticsOwner!
          .performAction(node.id, ui.SemanticsAction.customAction, earlier);
      await tester.pumpAndSettle();
      expect(scenario.ids, ['tokyo', 'lisbon', 'san-francisco', 'mexico-city']);
      expect(tester.takeException(), isNull);
    } finally {
      semantics.dispose();
    }
  });
}
