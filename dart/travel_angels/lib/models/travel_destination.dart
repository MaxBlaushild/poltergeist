/// A catalog location or a trip-scoped legacy/custom location.
class TravelDestination {
  const TravelDestination({
    this.id,
    required this.name,
    this.kind = 'place',
    this.countryCode = '',
    this.countryName = '',
    this.parentId,
    this.formattedAddress = '',
    this.latitude,
    this.longitude,
    this.timeZone = '',
    this.aliases = const [],
    this.recent = false,
    this.frequency = 0,
  });

  final String? id;
  final String name;
  final String kind;
  final String countryCode;
  final String countryName;
  final String? parentId;
  final String formattedAddress;
  final double? latitude;
  final double? longitude;
  final String timeZone;
  final List<String> aliases;
  final bool recent;
  final int frequency;

  factory TravelDestination.fromJson(Map<String, dynamic> json) {
    String string(String key) => json[key] is String ? json[key] as String : '';
    final id = string('id');
    return TravelDestination(
      id: id.isEmpty ? null : id,
      name: string('name'),
      kind: string('kind').isEmpty ? 'place' : string('kind'),
      countryCode: string('countryCode'),
      countryName: string('countryName'),
      parentId: string('parentId').isEmpty ? null : string('parentId'),
      formattedAddress: string('formattedAddress'),
      latitude: (json['latitude'] as num?)?.toDouble(),
      longitude: (json['longitude'] as num?)?.toDouble(),
      timeZone: string('timeZone'),
      aliases: json['aliases'] is List
          ? (json['aliases'] as List).whereType<String>().toList()
          : const [],
      recent: json['recent'] == true,
      frequency: (json['frequency'] as num?)?.toInt() ?? 0,
    );
  }

  String get label => name.isNotEmpty ? name : id ?? '';
  String get identity => id ?? 'custom:${name.trim().toLowerCase()}';
  String get contextLabel {
    final type = switch (kind) {
      'city' => 'City',
      'country' => 'Country',
      'region' => 'Region',
      _ => 'Place',
    };
    final detail = countryName.isNotEmpty && countryName != name
        ? countryName
        : formattedAddress != name
        ? formattedAddress
        : '';
    return detail.isEmpty ? type : '$type · $detail';
  }

  String get searchText => [
    name,
    countryCode,
    countryName,
    formattedAddress,
    ...aliases,
  ].join(' ').toLowerCase();

  Map<String, dynamic> toSelectionJson() =>
      id != null ? {'id': id} : {'name': name.trim()};
}

/// Only the absence of the new field invokes the legacy fallback. An explicit
/// empty selection must stay empty so validation can catch it.
List<TravelDestination> travelDestinations(Map<String, dynamic> trip) {
  if (trip['destinations'] is List) {
    return (trip['destinations'] as List)
        .whereType<Map>()
        .map(
          (item) => TravelDestination.fromJson(Map<String, dynamic>.from(item)),
        )
        .where((item) => item.label.isNotEmpty)
        .toList();
  }
  final legacy = trip['destination'] is String
      ? (trip['destination'] as String).trim()
      : '';
  return legacy.isEmpty ? [] : [TravelDestination(name: legacy)];
}

String travelDestinationLabel(Map<String, dynamic> trip) => travelDestinations(
  trip,
).map((destination) => destination.label).join(' → ');
