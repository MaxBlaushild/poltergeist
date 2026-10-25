import 'dart:async';

import 'package:travel_angels/models/document.dart';
import 'package:travel_angels/models/document_location.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/document_service.dart';
import 'package:travel_angels/services/friend_service.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';

Map<String, dynamic> libraryGuide(
  String id,
  String title, {
  String owner = 'friend-1',
  String format = 'guide',
  String destination = 'Lisbon, Portugal',
  String visibility = 'network',
  String content =
      '# A slow afternoon\nStart at the river and walk through Alfama.',
  bool saved = false,
}) => {
  'id': id,
  'title': title,
  'provider': format == 'spreadsheet' ? 'google_sheets' : 'internal',
  'userId': owner,
  'user': {'id': owner, 'name': owner == 'me' ? 'You' : 'Maya Chen'},
  'format': format,
  'destination': destination,
  'visibility': visibility,
  'content': content,
  'personalNote': 'Places worth taking your time to enjoy.',
  'createdAt': '2026-10-10T12:00:00Z',
  'updatedAt': '2026-10-20T12:00:00Z',
  'isSaved': saved,
  'sharedWithUserIds': <String>[],
  'savedTripIds': <String>[],
};

class LibraryFakeDocuments extends DocumentService {
  LibraryFakeDocuments([List<Map<String, dynamic>>? initial])
    : super(APIClient('https://example.test')) {
    for (final row in initial ?? <Map<String, dynamic>>[]) {
      documents[row['id'] as String] = Map<String, dynamic>.from(row);
    }
  }

  final documents = <String, Map<String, dynamic>>{};
  final calls = <Map<String, dynamic>>[];
  Object? listFailure;
  Object? readFailure;
  Object? saveFailure;
  Object? sharingFailure;
  Object? tripFailure;
  Object? createFailure;
  Completer<void>? listGate;
  Completer<void>? readGate;
  Completer<void>? saveGate;

  @override
  Future<List<Map<String, dynamic>>> getLibraryDocuments(String scope) async {
    calls.add({'action': 'list', 'scope': scope});
    if (listGate != null) await listGate!.future;
    if (listFailure != null) throw listFailure!;
    return documents.values
        .where(
          (row) => switch (scope) {
            'yours' => row['userId'] == 'me',
            'saved' => row['isSaved'] == true,
            _ => row['userId'] != 'me',
          },
        )
        .map(Map<String, dynamic>.from)
        .toList();
  }

  @override
  Future<List<Map<String, dynamic>>> getTripDocuments(String tripId) async {
    calls.add({'action': 'trip-library', 'tripId': tripId});
    if (listFailure != null) throw listFailure!;
    return documents.values
        .where((row) => (row['savedTripIds'] as List).contains(tripId))
        .map(Map<String, dynamic>.from)
        .toList();
  }

  @override
  Future<Map<String, dynamic>> getDocument(String id) async {
    calls.add({'action': 'read', 'id': id});
    if (readGate != null) await readGate!.future;
    if (readFailure != null) throw readFailure!;
    return Map<String, dynamic>.from(documents[id]!);
  }

  @override
  Future<Map<String, dynamic>> setDocumentSaved(String id, bool saved) async {
    calls.add({'action': 'save', 'id': id, 'saved': saved});
    if (saveGate != null) await saveGate!.future;
    if (saveFailure != null) throw saveFailure!;
    documents[id]!['isSaved'] = saved;
    return Map<String, dynamic>.from(documents[id]!);
  }

  @override
  Future<Map<String, dynamic>> updateDocumentSharing(
    String id, {
    required String visibility,
    List<String> userIds = const [],
  }) async {
    calls.add({
      'action': 'share',
      'id': id,
      'visibility': visibility,
      'userIds': List<String>.from(userIds),
    });
    if (sharingFailure != null) throw sharingFailure!;
    documents[id]!['visibility'] = visibility;
    documents[id]!['sharedWithUserIds'] = List<String>.from(userIds);
    return Map<String, dynamic>.from(documents[id]!);
  }

  @override
  Future<List<Map<String, dynamic>>> getDocumentTrips(String id) async {
    calls.add({'action': 'trips', 'id': id});
    if (tripFailure != null) throw tripFailure!;
    final ids = List<String>.from(documents[id]!['savedTripIds'] as List);
    return [
      {
        'id': 'trip-1',
        'title': 'Our Portugal trip',
        'destination': 'Lisbon',
        'isSaved': ids.contains('trip-1'),
      },
      {
        'id': 'trip-2',
        'title': 'Autumn in Japan',
        'destination': 'Kyoto',
        'isSaved': ids.contains('trip-2'),
      },
    ];
  }

  @override
  Future<Map<String, dynamic>> setDocumentTrip(
    String id,
    String tripId,
    bool saved,
  ) async {
    calls.add({'action': 'trip', 'id': id, 'tripId': tripId, 'saved': saved});
    if (tripFailure != null) throw tripFailure!;
    final ids = List<String>.from(documents[id]!['savedTripIds'] as List);
    if (saved && !ids.contains(tripId)) ids.add(tripId);
    if (!saved) ids.remove(tripId);
    documents[id]!['savedTripIds'] = ids;
    return Map<String, dynamic>.from(documents[id]!);
  }

  @override
  Future<Map<String, dynamic>> createDocument({
    required String title,
    required String provider,
    String? link,
    String? content,
    List<String>? tagIds,
    List<String>? newTags,
    List<DocumentLocation>? locations,
    String? format,
    String? personalNote,
    String? destination,
  }) async {
    calls.add({
      'action': 'create',
      'title': title,
      'provider': provider,
      'content': content,
      'format': format,
      'personalNote': personalNote,
      'destination': destination,
    });
    if (createFailure != null) throw createFailure!;
    final row = libraryGuide(
      'new-guide',
      title,
      owner: 'me',
      visibility: 'private',
      format: format ?? 'guide',
      content: content ?? '',
      destination: destination ?? '',
    );
    row['personalNote'] = personalNote ?? '';
    documents['new-guide'] = row;
    return Map<String, dynamic>.from(row);
  }

  @override
  Future<void> deleteDocument(String documentId) async {
    calls.add({'action': 'delete', 'id': documentId});
    documents.remove(documentId);
  }

  Document document(String id) => Document.fromJson(documents[id]!);
}

class LibraryFakeFriends extends FriendService {
  LibraryFakeFriends() : super(APIClient('https://example.test'));

  @override
  Future<List<User>> getFriends() async => [
    User(id: 'friend-1', name: 'Maya Chen', username: 'maya'),
    User(id: 'friend-2', name: 'Nora Patel', username: 'nora'),
  ];
}

class LibraryFakeCalendar extends TravelCalendarService {
  final calls = <String>[];

  @override
  Future<Map<String, dynamic>> get(
    String path, {
    Map<String, dynamic>? query,
  }) async {
    calls.add(path);
    return {
      'calendar': {'id': 'calendar-me', 'title': 'My calendar'},
      'stops': [
        {
          'id': 'trip-1',
          'title': 'Our Portugal trip',
          'destination': 'Lisbon',
          'startDate': '2028-02-10',
          'endDate': '2028-02-17',
          'datePrecision': 'fixed',
          'published': true,
        },
        {
          'id': 'trip-2',
          'title': 'Autumn in Japan',
          'destination': 'Kyoto',
          'datePrecision': 'tbd',
          'published': true,
        },
      ],
    };
  }
}
