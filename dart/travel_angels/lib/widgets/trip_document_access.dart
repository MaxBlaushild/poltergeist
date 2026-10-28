import 'package:flutter/foundation.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';

/// One access lifetime shared by the section and every reader or modal it opens.
/// Pending file reads and mutations must check this lifetime before continuing.
class TripDocumentAccess extends ChangeNotifier {
  TripDocumentAccess({
    required this.service,
    required this.stopId,
    required this.userId,
    required this.query,
    required this.canContribute,
  });

  final TravelCalendarService service;
  final String stopId;
  final String userId;
  final Map<String, dynamic> query;
  bool canContribute;
  bool active = true;
  int generation = 0;

  String get path => '/calendar/stops/${Uri.encodeComponent(stopId)}/documents';
  String documentPath(String id) => '$path/${Uri.encodeComponent(id)}';
  bool current(int captured) => active && captured == generation;

  void invalidate() {
    if (!active) return;
    active = false;
    generation++;
    notifyListeners();
  }

  void updateContribution(bool value) {
    if (canContribute == value) return;
    canContribute = value;
    generation++;
    notifyListeners();
  }
}

bool tripDocumentDenied(Object error) =>
    error is TravelCalendarException &&
    [401, 403, 404, 410].contains(error.statusCode);

String tripDocumentKind(Map<String, dynamic> document) {
  if (document['kind'] == 'guide') return 'Guide';
  if (document['kind'] == 'link') return 'Link';
  final filename = document['filename'] as String? ?? '';
  final extension = filename.contains('.') ? filename.split('.').last : 'File';
  return extension.toUpperCase();
}
