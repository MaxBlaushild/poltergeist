import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/models/document.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/screens/edit_document_screen.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/auth_service.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/add_guide_flow.dart';
import 'package:travel_angels/widgets/travel_trip_library.dart';

import 'library_test_support.dart';
import 'support/app_fonts.dart';

class _EditorAuth extends AuthProvider {
  _EditorAuth() : super(AuthService(APIClient('https://example.test')));

  User? current = User(id: 'me', name: 'You');

  @override
  User? get user => current;

  @override
  Future<void> verifyToken() async {}

  void change(User? next) {
    current = next;
    notifyListeners();
  }
}

Finder _field(String label) => find.byWidgetPredicate(
  (widget) => widget is TextField && widget.decoration?.labelText == label,
);

Future<void> _tap(WidgetTester tester, Finder finder) async {
  await tester.ensureVisible(finder);
  await tester.pumpAndSettle();
  await tester.tap(finder);
  await tester.pumpAndSettle();
}

void main() {
  setUpAll(loadAppFonts);

  for (final scenario in ['logout', 'account switch']) {
    testWidgets(
      'private editor clears its text and save controls after $scenario',
      (tester) async {
        final auth = _EditorAuth();
        final document = Document.fromJson(
          libraryGuide(
            'private-guide',
            'Private Portugal notes',
            owner: 'me',
            visibility: 'private',
            content: 'Private door code: 1234',
          ),
        );
        await tester.pumpWidget(
          ChangeNotifierProvider<AuthProvider>.value(
            value: auth,
            child: MaterialApp(
              theme: AppTheme.lightTheme,
              home: EditDocumentScreen(document: document),
            ),
          ),
        );
        await tester.pumpAndSettle();
        expect(find.text('Private Portugal notes'), findsOneWidget);
        expect(find.text('Private door code: 1234'), findsOneWidget);
        expect(find.text('Update'), findsNWidgets(2));
        final controllers = tester
            .widgetList<TextField>(find.byType(TextField))
            .map((field) => field.controller)
            .whereType<TextEditingController>()
            .toList();
        expect(
          controllers.any(
            (controller) => controller.text.contains('Private door code'),
          ),
          isTrue,
        );

        auth.change(
          scenario == 'logout' ? null : User(id: 'another-user', name: 'Nora'),
        );
        await tester.pumpAndSettle();
        expect(
          find.text('This guide is not available to your current account.'),
          findsOneWidget,
        );
        expect(find.text('Private Portugal notes'), findsNothing);
        expect(find.text('Private door code: 1234'), findsNothing);
        expect(find.byType(TextField), findsNothing);
        expect(find.text('Update'), findsNothing);
        expect(
          controllers.every((controller) => controller.text.isEmpty),
          isTrue,
        );
        expect(tester.takeException(), isNull);
        await tester.pumpWidget(const SizedBox());
        auth.dispose();
      },
    );
  }

  testWidgets(
    'writing validates required content, recovers from errors and returns a private guide',
    (tester) async {
      final service = LibraryFakeDocuments();
      Document? created;
      await tester.pumpWidget(
        MaterialApp(
          theme: AppTheme.lightTheme,
          home: Scaffold(
            body: Builder(
              builder: (context) => TextButton(
                onPressed: () async {
                  created = await Navigator.push<Document>(
                    context,
                    MaterialPageRoute(
                      builder: (_) => WriteGuideScreen(service: service),
                    ),
                  );
                },
                child: const Text('Write'),
              ),
            ),
          ),
        ),
      );
      await _tap(tester, find.text('Write'));
      await _tap(tester, find.text('Add to your library'));
      expect(find.text('Add a title'), findsOneWidget);
      expect(find.text('Add some details to your guide'), findsOneWidget);
      expect(service.calls, isEmpty);
      for (final entry in {
        'Guide title': '  A week in Portugal  ',
        'Destination': '  Lisbon & Porto  ',
        'A note from you': '  Book the train early.  ',
        'Your guide': '  # Day 1\nWalk along the river.  ',
      }.entries) {
        await tester.ensureVisible(_field(entry.key));
        await tester.enterText(_field(entry.key), entry.value);
      }
      await tester.testTextInput.receiveAction(TextInputAction.done);
      service.createFailure = StateError('offline');
      await _tap(tester, find.text('Add to your library'));
      expect(
        find.text('Your guide could not be saved. Please try again.'),
        findsOneWidget,
      );
      expect(
        tester.widget<TextField>(_field('Guide title')).controller!.text,
        '  A week in Portugal  ',
      );
      service.createFailure = null;
      await _tap(tester, find.text('Add to your library'));
      expect(created?.visibility, 'private');
      expect(created?.title, 'A week in Portugal');
      expect(created?.destination, 'Lisbon & Porto');
      expect(created?.personalNote, 'Book the train early.');
      expect(created?.content, '# Day 1\nWalk along the river.');
      expect(service.calls.last['provider'], 'internal');
      expect(service.calls.last['format'], 'guide');
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'trip library loads only on demand and rechecks document access when opened',
    (tester) async {
      final service = LibraryFakeDocuments([
        libraryGuide('for-trip', 'Portugal packing guide'),
        libraryGuide('elsewhere', 'Japan guide'),
      ]);
      service.documents['for-trip']!['savedTripIds'] = ['trip-1'];
      service.documents['elsewhere']!['savedTripIds'] = ['trip-2'];
      await tester.pumpWidget(
        MaterialApp(
          theme: AppTheme.lightTheme,
          home: Scaffold(
            body: SingleChildScrollView(
              child: TravelTripLibrary(
                tripId: 'trip-1',
                userId: 'me',
                service: service,
              ),
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(service.calls, isEmpty);
      await _tap(tester, find.text('Your trip library'));
      expect(service.calls.single, {
        'action': 'trip-library',
        'tripId': 'trip-1',
      });
      expect(find.text('Portugal packing guide'), findsOneWidget);
      expect(find.text('Japan guide'), findsNothing);
      expect(
        find.textContaining('This collection is just for you.'),
        findsOneWidget,
      );
      service.readFailure = StateError('offline');
      await _tap(tester, find.text('Portugal packing guide'));
      expect(
        service.calls.any(
          (call) => call['action'] == 'read' && call['id'] == 'for-trip',
        ),
        isTrue,
      );
      expect(
        find.byKey(const ValueKey('library-reader-content')),
        findsNothing,
      );
      expect(
        find.text('This guide could not be loaded. Please try again.'),
        findsOneWidget,
      );
      await tester.pumpWidget(const SizedBox());
    },
  );
}
