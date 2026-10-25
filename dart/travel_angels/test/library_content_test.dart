import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/library_document_content.dart';

import 'support/app_fonts.dart';

void main() {
  setUpAll(loadAppFonts);

  test(
    'source links allow web navigation without executable or local schemes',
    () {
      expect(
        librarySafeUri('https://example.com/guide?q=lisbon')?.host,
        'example.com',
      );
      for (final source in [
        'javascript:alert(1)',
        'data:text/html,<script>alert(1)</script>',
        'file:///private/travel-notes',
        '//example.com/guide',
        'https://',
        '',
      ]) {
        expect(librarySafeUri(source), isNull, reason: source);
      }
    },
  );

  testWidgets('a guide reads as headings lists and safe inline links', (
    tester,
  ) async {
    await tester.pumpWidget(
      MaterialApp(
        theme: AppTheme.lightTheme,
        home: const Scaffold(
          body: SingleChildScrollView(
            child: LibraryDocumentContent(
              format: 'guide',
              content:
                  '# A day in Lisbon\n- Walk through **Alfama**\n[Opening hours](https://example.com/hours)\n<script>untrusted()</script>',
            ),
          ),
        ),
      ),
    );
    expect(find.text('A day in Lisbon', findRichText: true), findsOneWidget);
    expect(
      find.text('Walk through Alfama', findRichText: true),
      findsOneWidget,
    );
    expect(find.text('Opening hours', findRichText: true), findsOneWidget);
    expect(
      find.text('<script>untrusted()</script>', findRichText: true),
      findsOneWidget,
    );
    expect(find.text('•'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'imported Markdown tables preserve escapes alignment and safe line breaks',
    (tester) async {
      final semantics = tester.ensureSemantics();
      try {
        await tester.pumpWidget(
          MaterialApp(
            theme: AppTheme.lightTheme,
            home: const Scaffold(
              body: LibraryDocumentContent(
                format: 'spreadsheet',
                content: r'''| Name \| type | Path | Notes |
| :-- | --: | :-: |
| A \| B | C:\\Trips\\2026 | First<br>Second<BR />Third<br/>Fourth |
| Backslash | \\ | <b>Inert markup</b> |
| --- | --- | --- |''',
              ),
            ),
          ),
        );
        final table = tester.widget<Table>(find.byType(Table));
        expect(table.children, hasLength(4));
        expect(table.children.every((row) => row.children.length == 3), isTrue);
        expect(find.text('Name | type'), findsOneWidget);
        expect(find.text('A | B'), findsOneWidget);
        expect(find.text(r'C:\Trips\2026'), findsOneWidget);
        expect(find.text('\\'), findsOneWidget);
        expect(find.text('First\nSecond\nThird\nFourth'), findsOneWidget);
        expect(find.text('<b>Inert markup</b>'), findsOneWidget);
        expect(find.text('---'), findsNWidgets(3));
        expect(
          tester.getSemantics(find.text('Name | type')),
          matchesSemantics(label: 'Name | type', isHeader: true),
        );
        expect(tester.takeException(), isNull);
      } finally {
        semantics.dispose();
      }
    },
  );

  testWidgets(
    'borderless tables keep terminal pipes and ordinary backslashes',
    (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          theme: AppTheme.lightTheme,
          home: const Scaffold(
            body: LibraryDocumentContent(
              format: 'guide',
              content: r'''Name | Notes
--- | ---
Pipe | Ends with \|
Path | C:\Trips\file
Both | A \\\| B''',
            ),
          ),
        ),
      );
      final table = tester.widget<Table>(find.byType(Table));
      expect(table.children, hasLength(4));
      expect(table.children.every((row) => row.children.length == 2), isTrue);
      expect(find.text('Ends with |'), findsOneWidget);
      expect(find.text(r'C:\Trips\file'), findsOneWidget);
      expect(find.text(r'A \| B'), findsOneWidget);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'spreadsheet cells retain quoted commas and remain readable at enlarged text',
    (tester) async {
      await tester.binding.setSurfaceSize(const Size(320, 740));
      addTearDown(() => tester.binding.setSurfaceSize(null));
      await tester.pumpWidget(
        MaterialApp(
          theme: AppTheme.lightTheme,
          builder: (_, child) => MediaQuery(
            data: const MediaQueryData(
              size: Size(320, 740),
              textScaler: TextScaler.linear(2),
            ),
            child: child!,
          ),
          home: const Scaffold(
            body: SingleChildScrollView(
              padding: EdgeInsets.all(16),
              child: LibraryDocumentContent(
                format: 'spreadsheet',
                content:
                    'Place,Notes,Cost\nAlfama,"Quiet streets, river views",Free\nBelem,"Try the ""pastel""",5 EUR',
              ),
            ),
          ),
        ),
      );
      expect(find.byType(Table), findsOneWidget);
      expect(find.text('Quiet streets, river views'), findsOneWidget);
      expect(find.text('Try the "pastel"'), findsOneWidget);
      await tester.drag(find.byType(Table), const Offset(-330, 0));
      await tester.pumpAndSettle();
      expect(find.text('Cost').hitTestable(), findsOneWidget);
      expect(find.text('5 EUR').hitTestable(), findsOneWidget);
      expect(tester.takeException(), isNull);
    },
  );
}
