/// Shared links use URL fragments so static web hosting needs no path rewrites.
/// Path and native-app forms remain accepted for deep-link navigation.
class TravelCalendarRoute {
  const TravelCalendarRoute(this.kind, this.token);

  final String kind;
  final String token;

  String get path => '/$kind/${Uri.encodeComponent(token)}';

  static TravelCalendarRoute? parse(String location) {
    final uri = Uri.tryParse(location);
    if (uri == null) return null;
    final target = uri.fragment.startsWith('/')
        ? Uri.tryParse(uri.fragment)
        : uri;
    if (target == null) return null;
    final parts = [
      if (uri.scheme == 'travelangels' && uri.host.isNotEmpty) uri.host,
      ...target.pathSegments.where((part) => part.isNotEmpty),
    ];
    if (parts.length != 2 ||
        !{'calendar', 'stops', 'preferences'}.contains(parts[0]) ||
        !RegExp(r'^[A-Za-z0-9_-]{16,256}$').hasMatch(parts[1])) {
      return null;
    }
    return TravelCalendarRoute(parts[0], parts[1]);
  }
}
