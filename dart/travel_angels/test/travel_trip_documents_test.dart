import 'dart:async';
import 'dart:io';
import 'dart:typed_data';
import 'dart:ui' as ui;

import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/travel_trip_documents.dart';

import 'library_test_support.dart';
import 'support/app_fonts.dart';
import 'travel_trip_widgets_test.dart' show TripTestAuth;

Map<String, dynamic> _document({String kind = 'guide'}) => {
  'id': 'document-1',
  'title': 'A week in Kyoto',
  'kind': kind,
  'format': 'itinerary',
  'content': kind == 'guide'
      ? '# The first morning\nCoffee, temples, and a walk by the river.\n\n## Friday\nMeet at the station at 10.'
      : '',
  'createdAt': '2026-10-27T12:00:00Z',
  'addedBy': {'id': 'me', 'displayName': 'Max'},
  'canDelete': true,
  'canContribute': true,
  'commentCount': 1,
  if (kind == 'file') ...{
    'filename': 'kyoto.xlsx',
    'fileUrl': '/travel-angels/calendar/stops/trip/documents/document-1/file',
    'contentType':
        'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    'sizeBytes': 4096,
  },
};

Map<String, dynamic> _comment(String id, String body, {bool own = false}) => {
  'id': id,
  'body': body,
  'author': {
    'id': own ? 'me' : 'friend',
    'displayName': own ? 'Max' : 'Caroline',
  },
  'createdAt': '2026-10-28T12:00:00Z',
  'canEdit': own,
  'canDelete': true,
};

class _Service extends TravelCalendarService {
  Map<String, dynamic> document = _document();
  List<Map<String, dynamic>> comments = [
    _comment('comment-1', 'Could we add a tea ceremony?'),
  ];
  final calls = <Map<String, dynamic>>[];
  Object? readFailure;
  Object? writeFailure;
  Completer<Map<String, dynamic>>? readGate;
  Completer<Uint8List>? fileGate;
  bool deleted = false;
  bool hasOlder = false;
  bool canContribute = true;

  @override
  Future<Map<String, dynamic>> get(
    String path, {
    Map<String, dynamic>? query,
  }) async {
    calls.add({'method': 'GET', 'path': path, 'query': query});
    if (readFailure != null) throw readFailure!;
    if (path.endsWith('/documents')) {
      return {
        'documents': deleted ? [] : [document],
        'canView': true,
        'canContribute': canContribute,
      };
    }
    if (path.endsWith('/comments')) {
      if (query?['before'] != null) {
        return {
          'comments': [_comment('older', 'The earlier suggestion')],
          'canContribute': canContribute,
          'hasMore': false,
        };
      }
      return {
        'comments': comments,
        'canContribute': canContribute,
        'hasMore': hasOlder,
        if (hasOlder) 'nextBefore': 'comment-1',
      };
    }
    if (readGate != null) return readGate!.future;
    return {
      'document': {...document, 'canContribute': canContribute},
    };
  }

  @override
  Future<Map<String, dynamic>> post(
    String path, {
    Map<String, dynamic>? data,
    Map<String, dynamic>? query,
  }) async {
    calls.add({'method': 'POST', 'path': path, 'data': data, 'query': query});
    if (writeFailure != null) throw writeFailure!;
    if (path.endsWith('/comments')) {
      comments.add(_comment('mine', data!['body'] as String, own: true));
    }
    return {};
  }

  @override
  Future<Map<String, dynamic>> put(
    String path, {
    Map<String, dynamic>? data,
    Map<String, dynamic>? query,
  }) async {
    calls.add({'method': 'PUT', 'path': path, 'data': data, 'query': query});
    if (writeFailure != null) throw writeFailure!;
    comments = comments
        .map(
          (comment) => path.endsWith('/${comment['id']}')
              ? {...comment, 'body': data!['body']}
              : comment,
        )
        .toList();
    return {};
  }

  @override
  Future<Map<String, dynamic>> delete(
    String path, {
    Map<String, dynamic>? data,
    Map<String, dynamic>? query,
  }) async {
    calls.add({'method': 'DELETE', 'path': path, 'query': query});
    if (writeFailure != null) throw writeFailure!;
    if (path.endsWith('/document-1')) deleted = true;
    comments.removeWhere((comment) => path.endsWith('/${comment['id']}'));
    return {};
  }

  @override
  Future<Map<String, dynamic>> uploadTripDocument(
    String stopId,
    Uint8List bytes,
    String filename, {
    required String title,
    Map<String, dynamic>? query,
  }) async {
    calls.add({
      'method': 'UPLOAD',
      'bytes': bytes,
      'filename': filename,
      'title': title,
      'query': query,
    });
    if (writeFailure != null) throw writeFailure!;
    return {};
  }

  @override
  Future<Uint8List> loadTripDocumentFile(
    String url, {
    Map<String, dynamic>? query,
  }) async {
    calls.add({'method': 'FILE', 'url': url, 'query': query});
    return fileGate?.future ?? Uint8List.fromList([1, 2, 3]);
  }
}

class _Picker extends FilePicker {
  int size = 4;
  int saves = 0;
  @override
  Future<FilePickerResult?> pickFiles({
    String? dialogTitle,
    String? initialDirectory,
    FileType type = FileType.any,
    List<String>? allowedExtensions,
    Function(FilePickerStatus)? onFileLoading,
    bool allowCompression = true,
    int compressionQuality = 30,
    bool allowMultiple = false,
    bool withData = false,
    bool withReadStream = false,
    bool lockParentWindow = false,
    bool readSequential = false,
  }) async => FilePickerResult([
    PlatformFile(
      name: 'reservations.pdf',
      size: size,
      bytes: Uint8List.fromList([1, 2, 3, 4]),
    ),
  ]);

  @override
  Future<String?> saveFile({
    String? dialogTitle,
    String? fileName,
    String? initialDirectory,
    FileType type = FileType.any,
    List<String>? allowedExtensions,
    Uint8List? bytes,
    bool lockParentWindow = false,
  }) async {
    saves++;
    return null;
  }
}

Future<GlobalKey> _pump(
  WidgetTester tester,
  _Service service, {
  TripTestAuth? auth,
  bool canView = true,
  bool canContribute = true,
  LibraryFakeDocuments? library,
  double width = 390,
  double scale = 1,
}) async {
  final size = Size(width, 844);
  await tester.binding.setSurfaceSize(size);
  addTearDown(() => tester.binding.setSurfaceSize(null));
  final boundary = GlobalKey();
  final section = TravelTripDocuments(
    stopId: 'trip',
    userId: canView ? 'me' : null,
    service: service,
    canView: canView,
    canContribute: canContribute,
    accessQuery: const {'stopLinkToken': 'shared-link'},
    documentService: library,
  );
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
          child: auth == null
              ? child!
              : ChangeNotifierProvider<AuthProvider>.value(
                  value: auth,
                  child: child!,
                ),
        ),
        home: Scaffold(
          appBar: AppBar(title: const Text('Kyoto together')),
          body: SingleChildScrollView(
            padding: const EdgeInsets.all(20),
            child: section,
          ),
        ),
      ),
    ),
  );
  await tester.pumpAndSettle();
  return boundary;
}

Future<void> _open(WidgetTester tester) async {
  await tester.tap(find.text('A week in Kyoto'));
  await tester.pumpAndSettle();
}

Future<void> _add(WidgetTester tester, String kind) async {
  await tester.tap(find.byTooltip('Add document'));
  await tester.pumpAndSettle();
  await tester.tap(find.text(kind));
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
    final bytes = await image.toByteData(format: ui.ImageByteFormat.png);
    await File(
      '/private/tmp/travel-angels-documents-$name.png',
    ).writeAsBytes(bytes!.buffer.asUint8List());
    image.dispose();
  });
}

void main() {
  setUpAll(loadAppFonts);
  setUp(() => FilePicker.platform = _Picker());

  testWidgets(
    'guests never fetch shared documents or receive attachment controls',
    (tester) async {
      final service = _Service();
      await _pump(tester, service, canView: false, canContribute: false);
      expect(service.calls, isEmpty);
      expect(find.byTooltip('Add document'), findsNothing);
      expect(find.text('A week in Kyoto'), findsNothing);
      expect(
        find.text('Shared documents are for hosts and joined travelers.'),
        findsOneWidget,
      );
    },
  );

  testWidgets(
    'reader rechecks access before showing content and carries trip access query',
    (tester) async {
      final service = _Service()..readGate = Completer();
      await _pump(tester, service);
      await tester.tap(find.text('A week in Kyoto'));
      await tester.pump();
      expect(find.text('The first morning'), findsNothing);
      service.readGate!.complete({'document': service.document});
      await tester.pumpAndSettle();
      expect(find.text('The first morning'), findsOneWidget);
      expect(
        service.calls.every(
          (call) => call['query']?['stopLinkToken'] == 'shared-link',
        ),
        isTrue,
      );
    },
  );

  testWidgets(
    'link attachment validates scheme and retains fields after a failed save',
    (tester) async {
      final service = _Service();
      await _pump(tester, service);
      await _add(tester, 'Add a link');
      await tester.enterText(
        find.widgetWithText(TextField, 'Title'),
        'Train times',
      );
      await tester.enterText(
        find.widgetWithText(TextField, 'Link'),
        'javascript:alert(1)',
      );
      await tester.tap(find.text('Share with trip'));
      await tester.pumpAndSettle();
      expect(find.text('Enter a complete http or https link.'), findsOneWidget);
      expect(service.calls.where((call) => call['method'] == 'POST'), isEmpty);
      await tester.enterText(
        find.widgetWithText(TextField, 'Link'),
        'https://rail.example/timetable',
      );
      service.writeFailure = const TravelCalendarException('Please try again.');
      await tester.tap(find.text('Share with trip'));
      await tester.pumpAndSettle();
      expect(find.text('Train times'), findsOneWidget);
      expect(find.text('Please try again.'), findsOneWidget);
      service.writeFailure = null;
      await tester.tap(find.text('Share with trip'));
      await tester.pumpAndSettle();
      expect(find.byType(TripDocumentAddDialog), findsNothing);
      final request = service.calls.lastWhere(
        (call) => call['method'] == 'POST',
      );
      expect(request['data'], {
        'title': 'Train times',
        'link': 'https://rail.example/timetable',
      });
    },
  );

  testWidgets(
    'copies only an owned library guide without modifying its audience',
    (tester) async {
      final service = _Service();
      final library = LibraryFakeDocuments([
        libraryGuide(
          'mine',
          'My private itinerary',
          owner: 'me',
          visibility: 'private',
        ),
        libraryGuide('theirs', 'Someone else’s guide'),
      ]);
      await _pump(tester, service, library: library);
      await _add(tester, 'Copy a guide from your library');
      expect(find.text('Someone else’s guide'), findsNothing);
      expect(
        find.text(
          'A copy is shared with this trip. Your library guide stays unchanged.',
        ),
        findsOneWidget,
      );
      await tester.tap(find.text('My private itinerary'));
      await tester.pump();
      await tester.tap(find.text('Share with trip'));
      await tester.pumpAndSettle();
      expect(
        service.calls.singleWhere((call) => call['method'] == 'POST')['data'],
        {'title': 'My private itinerary', 'sourceDocumentId': 'mine'},
      );
      expect(library.documents['mine']!['visibility'], 'private');
      expect(library.calls.every((call) => call['action'] == 'list'), isTrue);
    },
  );

  testWidgets(
    'uploads original bytes and rejects oversized files before a request',
    (tester) async {
      final oldPicker = FilePicker.platform;
      final picker = _Picker()..size = 21 * 1024 * 1024;
      FilePicker.platform = picker;
      addTearDown(() => FilePicker.platform = oldPicker);
      final service = _Service();
      await _pump(tester, service);
      await _add(tester, 'Upload a file');
      await tester.tap(find.text('Choose file'));
      await tester.pumpAndSettle();
      expect(find.text('Choose a nonempty file up to 20 MB.'), findsOneWidget);
      expect(
        service.calls.where((call) => call['method'] == 'UPLOAD'),
        isEmpty,
      );
      picker.size = 4;
      await tester.tap(find.text('Choose file'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Share with trip'));
      await tester.pumpAndSettle();
      final upload = service.calls.singleWhere(
        (call) => call['method'] == 'UPLOAD',
      );
      expect(upload['bytes'], [1, 2, 3, 4]);
      expect(upload['filename'], 'reservations.pdf');
    },
  );

  testWidgets(
    'comments preserve a failed draft and support own edits and host moderation',
    (tester) async {
      final service = _Service();
      await _pump(tester, service);
      await _open(tester);
      await tester.ensureVisible(find.byType(TextField));
      await tester.enterText(
        find.byType(TextField),
        'I booked the tea ceremony.',
      );
      service.writeFailure = const TravelCalendarException(
        'The connection was interrupted.',
      );
      await tester.tap(find.byTooltip('Post comment'));
      await tester.pumpAndSettle();
      expect(find.text('I booked the tea ceremony.'), findsOneWidget);
      service.writeFailure = null;
      await tester.tap(find.byTooltip('Post comment'));
      await tester.pumpAndSettle();
      expect(find.text('I booked the tea ceremony.'), findsOneWidget);
      await tester.ensureVisible(find.byTooltip('Comment options').last);
      await tester.tap(find.byTooltip('Comment options').last);
      await tester.pumpAndSettle();
      await tester.tap(find.text('Edit comment'));
      await tester.pumpAndSettle();
      await tester.ensureVisible(find.byType(TextField));
      await tester.enterText(
        find.byType(TextField),
        'Tea is booked for Friday.',
      );
      await tester.tap(find.byTooltip('Save comment'));
      await tester.pumpAndSettle();
      expect(
        service.calls.lastWhere((call) => call['method'] == 'PUT')['data'],
        {'body': 'Tea is booked for Friday.'},
      );
      await tester.ensureVisible(find.byTooltip('Comment options').first);
      await tester.tap(find.byTooltip('Comment options').first);
      await tester.pumpAndSettle();
      expect(find.text('Edit comment'), findsNothing);
      await tester.tap(find.text('Remove comment'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Remove'));
      await tester.pumpAndSettle();
      expect(find.text('Could we add a tea ceremony?'), findsNothing);
    },
  );

  testWidgets('earlier comments are loaded within the attachment thread', (
    tester,
  ) async {
    final service = _Service()..hasOlder = true;
    await _pump(tester, service);
    await _open(tester);
    await tester.ensureVisible(find.text('Earlier comments'));
    await tester.tap(find.text('Earlier comments'));
    await tester.pumpAndSettle();
    expect(find.text('The earlier suggestion'), findsOneWidget);
    expect(find.text('Earlier comments'), findsNothing);
    expect(service.calls.last['query'], {
      'stopLinkToken': 'shared-link',
      'before': 'comment-1',
    });
  });

  testWidgets(
    'read-only trips retain removal but hide contribution and comment editing',
    (tester) async {
      final service = _Service()..canContribute = false;
      await _pump(tester, service, canContribute: false);
      expect(find.byTooltip('Add document'), findsNothing);
      await _open(tester);
      expect(find.byType(TextField), findsNothing);
      await tester.tap(find.byTooltip('Document options'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Remove from trip'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Remove'));
      await tester.pumpAndSettle();
      expect(service.deleted, isTrue);
      expect(find.text('A week in Kyoto'), findsNothing);
      expect(
        service.calls
            .where((call) => call['method'] == 'DELETE')
            .single['path'],
        '/calendar/stops/trip/documents/document-1',
      );
    },
  );

  testWidgets(
    'account switch clears a reader and ignores its late protected response',
    (tester) async {
      final auth = TripTestAuth(User(id: 'me'));
      final service = _Service()..readGate = Completer();
      await _pump(tester, service, auth: auth);
      await tester.tap(find.text('A week in Kyoto'));
      await tester.pump();
      auth.change(User(id: 'other'));
      await tester.pumpAndSettle();
      service.readGate!.complete({'document': service.document});
      await tester.pumpAndSettle();
      expect(find.text('The first morning'), findsNothing);
      expect(find.text('A week in Kyoto'), findsNothing);
      expect(find.byTooltip('Add document'), findsNothing);
    },
  );

  testWidgets(
    'account switch closes attachment modal and discards a pending download',
    (tester) async {
      final auth = TripTestAuth(User(id: 'me'));
      final service = _Service()
        ..document = _document(kind: 'file')
        ..fileGate = Completer();
      final oldPicker = FilePicker.platform;
      final picker = _Picker();
      FilePicker.platform = picker;
      addTearDown(() => FilePicker.platform = oldPicker);
      await _pump(tester, service, auth: auth);
      await _add(tester, 'Add a link');
      await tester.enterText(
        find.widgetWithText(TextField, 'Title'),
        'Private draft',
      );
      auth.change(null);
      await tester.pumpAndSettle();
      expect(find.byType(TripDocumentAddDialog), findsNothing);
      expect(find.text('Private draft'), findsNothing);
      auth.change(User(id: 'me'));
      await _pump(tester, service, auth: auth);
      await _open(tester);
      await tester.tap(find.text('Download'));
      await tester.pump();
      auth.change(null);
      await tester.pumpAndSettle();
      service.fileGate!.complete(Uint8List.fromList([1, 2, 3]));
      await tester.pumpAndSettle();
      expect(picker.saves, 0);
    },
  );

  testWidgets(
    'access denial during comment save closes all protected content',
    (tester) async {
      final service = _Service();
      await _pump(tester, service);
      await _open(tester);
      await tester.ensureVisible(find.byType(TextField));
      await tester.enterText(find.byType(TextField), 'Pending private message');
      service.writeFailure = const TravelCalendarException(
        'Access denied',
        statusCode: 403,
      );
      await tester.tap(find.byTooltip('Post comment'));
      await tester.pumpAndSettle();
      expect(find.text('The first morning'), findsNothing);
      expect(find.text('Pending private message'), findsNothing);
      expect(find.text('A week in Kyoto'), findsNothing);
    },
  );

  for (final layout in [(390.0, 1.0), (320.0, 2.0)]) {
    testWidgets(
      'section, reader, and add dialog fit ${layout.$1} at ${layout.$2}x',
      (tester) async {
        final service = _Service();
        final boundary = await _pump(
          tester,
          service,
          width: layout.$1,
          scale: layout.$2,
        );
        await _capture(
          tester,
          boundary,
          'section-${layout.$1.toInt()}-${layout.$2}x',
        );
        expect(tester.takeException(), isNull);
        await _open(tester);
        await _capture(
          tester,
          boundary,
          'reader-${layout.$1.toInt()}-${layout.$2}x',
        );
        await tester.ensureVisible(find.byType(TextField));
        await tester.pumpAndSettle();
        await _capture(
          tester,
          boundary,
          'comments-${layout.$1.toInt()}-${layout.$2}x',
        );
        expect(tester.takeException(), isNull);
        await tester.pageBack();
        await tester.pumpAndSettle();
        await _add(tester, 'Upload a file');
        expect(tester.takeException(), isNull);
        await _capture(
          tester,
          boundary,
          'add-${layout.$1.toInt()}-${layout.$2}x',
        );
      },
    );
  }
}
