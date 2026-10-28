import 'dart:io';
import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/international_phone_field.dart';

import 'support/app_fonts.dart';

Finder _phone() => find.descendant(
  of: find.byType(InternationalPhoneField),
  matching: find.byType(TextField),
);
Finder _search() => find.byWidgetPredicate(
  (widget) =>
      widget is TextField && widget.decoration?.labelText == 'Search countries',
);
Finder _countryButton() => find.descendant(
  of: find.byType(InternationalPhoneField),
  matching: find.byType(TextButton),
);
Finder _countryName(String name) =>
    find.descendant(of: find.byType(ListTile), matching: find.text(name));

Future<GlobalKey> _show(
  WidgetTester tester, {
  required ValueChanged<String?> onChanged,
  String? initialCountry,
  ValueNotifier<bool>? enabled,
  ValueNotifier<double>? keyboardInset,
  double width = 390,
  double height = 844,
  double textScale = 1,
}) async {
  await tester.binding.setSurfaceSize(Size(width, height));
  addTearDown(() => tester.binding.setSurfaceSize(null));
  final enabledState = enabled ?? ValueNotifier(true);
  final keyboardState = keyboardInset ?? ValueNotifier(0.0);
  if (enabled == null) addTearDown(enabledState.dispose);
  if (keyboardInset == null) addTearDown(keyboardState.dispose);
  final boundary = GlobalKey();
  await tester.pumpWidget(
    RepaintBoundary(
      key: boundary,
      child: MaterialApp(
        debugShowCheckedModeBanner: false,
        theme: AppTheme.lightTheme,
        builder: (context, child) => ValueListenableBuilder<double>(
          valueListenable: keyboardState,
          builder: (context, bottom, _) => MediaQuery(
            data: MediaQuery.of(context).copyWith(
              textScaler: TextScaler.linear(textScale),
              viewInsets: EdgeInsets.only(bottom: bottom),
            ),
            child: child!,
          ),
        ),
        home: Scaffold(
          appBar: AppBar(title: const TravelAngelsWordmark()),
          body: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const Text(
                  'PLANS WITH FRIENDS',
                  style: TextStyle(fontSize: 12, letterSpacing: 1.5),
                ),
                const SizedBox(height: 16),
                const Text(
                  'Sign in or sign up',
                  style: TextStyle(
                    fontFamily: AppTheme.displayFont,
                    fontSize: 32,
                  ),
                ),
                const SizedBox(height: 28),
                ValueListenableBuilder<bool>(
                  valueListenable: enabledState,
                  builder: (_, value, _) => InternationalPhoneField(
                    onChanged: onChanged,
                    enabled: value,
                    initialCountryCode: initialCountry,
                  ),
                ),
                const SizedBox(height: 20),
                FilledButton(onPressed: () {}, child: const Text('Continue')),
              ],
            ),
          ),
        ),
      ),
    ),
  );
  await tester.pumpAndSettle();
  return boundary;
}

Future<void> _openCountries(WidgetTester tester) async {
  await tester.ensureVisible(_countryButton());
  await tester.tap(_countryButton());
  await tester.pumpAndSettle();
  expect(_search(), findsOneWidget);
}

Future<void> _searchFor(WidgetTester tester, String text) async {
  await tester.enterText(_search(), text);
  await tester.pumpAndSettle();
}

Future<void> _select(WidgetTester tester, String country) async {
  await _searchFor(tester, country);
  await tester.tap(_countryName(country));
  await tester.pumpAndSettle();
}

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
    try {
      final data = await image.toByteData(format: ui.ImageByteFormat.png);
      await File(
        '/private/tmp/travel-angels-phone-$name.png',
      ).writeAsBytes(data!.buffer.asUint8List());
    } finally {
      image.dispose();
    }
  });
}

void main() {
  setUpAll(loadAppFonts);

  testWidgets(
    'country search supports names, ISO, calling codes, and no results',
    (tester) async {
      await _show(tester, onChanged: (_) {}, initialCountry: 'US');
      await _openCountries(tester);
      await _searchFor(tester, 'kingdom');
      expect(_countryName('United Kingdom'), findsOneWidget);
      await _searchFor(tester, 'gb');
      expect(_countryName('United Kingdom'), findsOneWidget);
      expect(find.text('GB'), findsOneWidget);
      await _searchFor(tester, '+44');
      expect(_countryName('United Kingdom'), findsOneWidget);
      expect(_countryName('Jersey'), findsOneWidget);
      await _searchFor(tester, 'not-a-country');
      expect(find.text('No matching countries'), findsOneWidget);
      await _select(tester, 'United Kingdom');
      expect(find.text('GB +44'), findsOneWidget);
      expect(_search(), findsNothing);
    },
  );

  testWidgets(
    'UK national prefix produces canonical destination and clearing invalidates it',
    (tester) async {
      final changes = <String?>[];
      await _show(tester, onChanged: changes.add, initialCountry: 'GB');
      await tester.enterText(_phone(), '020 7946 0018');
      await tester.pumpAndSettle();
      expect(changes, ['+442079460018']);
      await tester.enterText(_phone(), '');
      await tester.pumpAndSettle();
      expect(changes.last, isNull);
    },
  );

  for (final paste in ['+44 (20) 7946-0018', '(+44) 20 7946 0018']) {
    testWidgets('international paste $paste updates country and emits once', (
      tester,
    ) async {
      final changes = <String?>[];
      await _show(tester, onChanged: changes.add, initialCountry: 'US');
      await tester.enterText(_phone(), paste);
      await tester.pumpAndSettle();
      expect(changes, ['+442079460018']);
      expect(find.text('GB +44'), findsOneWidget);
      final displayed = tester.widget<TextField>(_phone()).controller!.text;
      expect(displayed, isNot(contains('+44')));
      expect(displayed.replaceAll(RegExp(r'\D'), ''), '2079460018');
    });
  }

  testWidgets('changing country cannot retain a previous valid destination', (
    tester,
  ) async {
    final changes = <String?>[];
    await _show(tester, onChanged: changes.add, initialCountry: 'GB');
    await tester.enterText(_phone(), '020 7946 0018');
    await tester.pumpAndSettle();
    expect(changes.last, '+442079460018');
    await _openCountries(tester);
    await _select(tester, 'Singapore');
    expect(find.text('SG +65'), findsOneWidget);
    expect(changes.last, isNull);
    await tester.enterText(_phone(), '8123 4567');
    await tester.pumpAndSettle();
    expect(changes.last, '+6581234567');
  });

  testWidgets(
    'shared calling codes keep distinct countries without duplicate rows or area prefixes',
    (tester) async {
      final changes = <String?>[];
      await _show(tester, onChanged: changes.add, initialCountry: 'US');
      await _openCountries(tester);
      await _searchFor(tester, '+1');
      final samoa = find.ancestor(
        of: _countryName('American Samoa'),
        matching: find.byType(ListTile),
      );
      expect(samoa, findsOneWidget);
      expect(
        find.descendant(of: samoa, matching: find.text('+1')),
        findsOneWidget,
      );
      expect(find.text('+1684'), findsNothing);
      await _searchFor(tester, 'Dominican Republic');
      expect(_countryName('Dominican Republic'), findsOneWidget);
      await _searchFor(tester, 'United States');
      expect(_countryName('United States'), findsOneWidget);
      await _select(tester, 'Canada');
      expect(find.text('CA +1'), findsOneWidget);
      await tester.enterText(_phone(), '+1 800 555 1234');
      await tester.pumpAndSettle();
      expect(find.text('CA +1'), findsOneWidget);
      expect(changes.last, '+18005551234');
    },
  );

  for (final (locale, expected) in [
    (const Locale('en', 'GB'), 'GB +44'),
    (const Locale('en'), 'US +1'),
  ]) {
    testWidgets('device locale $locale defaults country to $expected', (
      tester,
    ) async {
      tester.binding.platformDispatcher.localesTestValue = [locale];
      addTearDown(tester.binding.platformDispatcher.clearLocalesTestValue);
      await _show(tester, onChanged: (_) {});
      expect(find.text(expected), findsOneWidget);
    });
  }

  testWidgets('disabled field locks typing and the country selector', (
    tester,
  ) async {
    final enabled = ValueNotifier(false);
    addTearDown(enabled.dispose);
    final changes = <String?>[];
    await _show(
      tester,
      onChanged: changes.add,
      initialCountry: 'US',
      enabled: enabled,
    );
    expect(tester.widget<TextField>(_phone()).enabled, isFalse);
    expect(tester.widget<TextButton>(_countryButton()).onPressed, isNull);
    await tester.tap(_countryButton());
    await tester.pumpAndSettle();
    expect(_search(), findsNothing);
    expect(changes, isEmpty);

    enabled.value = true;
    await tester.pumpAndSettle();
    await _openCountries(tester);
    enabled.value = false;
    await tester.pumpAndSettle();
    await _select(tester, 'United Kingdom');
    expect(find.text('US +1'), findsOneWidget);
    expect(changes, isEmpty);
  });

  testWidgets('screen readers can activate the labeled country selector', (
    tester,
  ) async {
    final semantics = tester.ensureSemantics();
    try {
      await _show(tester, onChanged: (_) {}, initialCountry: 'US');
      final node = tester.getSemantics(
        find.bySemanticsLabel('Country code: United States, +1'),
      );
      expect(node.getSemanticsData().hasAction(ui.SemanticsAction.tap), isTrue);
      node.owner!.performAction(node.id, ui.SemanticsAction.tap);
      await tester.pumpAndSettle();
      expect(_search(), findsOneWidget);
    } finally {
      semantics.dispose();
    }
  });

  for (final (width, scale) in [(390.0, 1.0), (320.0, 2.0)]) {
    testWidgets(
      'field and keyboard-open picker fit ${width}px at ${scale}x text',
      (tester) async {
        final keyboard = ValueNotifier(0.0);
        addTearDown(keyboard.dispose);
        final boundary = await _show(
          tester,
          onChanged: (_) {},
          initialCountry: 'US',
          width: width,
          textScale: scale,
          keyboardInset: keyboard,
        );
        await tester.enterText(_phone(), '2025550128');
        await tester.pumpAndSettle();
        expect(tester.takeException(), isNull);
        await _capture(tester, boundary, 'field-${width.toInt()}-${scale}x');
        await _openCountries(tester);
        await _searchFor(tester, '+44');
        keyboard.value = 300;
        await tester.pumpAndSettle();
        expect(tester.takeException(), isNull);
        expect(_search().hitTestable(), findsOneWidget);
        expect(
          find.byTooltip('Close country picker').hitTestable(),
          findsOneWidget,
        );
        await _capture(tester, boundary, 'picker-${width.toInt()}-${scale}x');
        await tester.tap(find.byTooltip('Close country picker'));
        await tester.pumpAndSettle();
        keyboard.value = 0;
        await tester.pumpAndSettle();
        expect(tester.takeException(), isNull);
      },
    );
  }
}
