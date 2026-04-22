import 'package:flutter/material.dart';

class ZoneKindVisualProfile {
  final String label;
  final IconData icon;
  final String atmosphereLabel;
  final String focusLabel;
  final String lineJoin;
  final double discoveredFillOpacity;
  final double undiscoveredFillOpacity;
  final double outerLineWidth;
  final double undiscoveredOuterLineWidth;
  final double innerLineWidth;
  final double undiscoveredInnerLineWidth;
  final double outerLineOpacity;
  final double undiscoveredOuterLineOpacity;
  final double innerLineOpacity;
  final double outerLineBlur;
  final double undiscoveredOuterLineBlur;
  final Color panelSurface;
  final Color panelBorder;
  final Color panelAccent;
  final Color panelText;
  final Color panelSubtext;
  final Color panelChipBackground;
  final Color panelChipText;
  final Color panelShadow;

  const ZoneKindVisualProfile({
    required this.label,
    required this.icon,
    required this.atmosphereLabel,
    required this.focusLabel,
    required this.lineJoin,
    required this.discoveredFillOpacity,
    required this.undiscoveredFillOpacity,
    required this.outerLineWidth,
    required this.undiscoveredOuterLineWidth,
    required this.innerLineWidth,
    required this.undiscoveredInnerLineWidth,
    required this.outerLineOpacity,
    required this.undiscoveredOuterLineOpacity,
    required this.innerLineOpacity,
    required this.outerLineBlur,
    required this.undiscoveredOuterLineBlur,
    required this.panelSurface,
    required this.panelBorder,
    required this.panelAccent,
    required this.panelText,
    required this.panelSubtext,
    required this.panelChipBackground,
    required this.panelChipText,
    required this.panelShadow,
  });

  ZoneKindVisualProfile copyWith({
    String? label,
    String? atmosphereLabel,
    String? focusLabel,
  }) {
    return ZoneKindVisualProfile(
      label: label ?? this.label,
      icon: icon,
      atmosphereLabel: atmosphereLabel ?? this.atmosphereLabel,
      focusLabel: focusLabel ?? this.focusLabel,
      lineJoin: lineJoin,
      discoveredFillOpacity: discoveredFillOpacity,
      undiscoveredFillOpacity: undiscoveredFillOpacity,
      outerLineWidth: outerLineWidth,
      undiscoveredOuterLineWidth: undiscoveredOuterLineWidth,
      innerLineWidth: innerLineWidth,
      undiscoveredInnerLineWidth: undiscoveredInnerLineWidth,
      outerLineOpacity: outerLineOpacity,
      undiscoveredOuterLineOpacity: undiscoveredOuterLineOpacity,
      innerLineOpacity: innerLineOpacity,
      outerLineBlur: outerLineBlur,
      undiscoveredOuterLineBlur: undiscoveredOuterLineBlur,
      panelSurface: panelSurface,
      panelBorder: panelBorder,
      panelAccent: panelAccent,
      panelText: panelText,
      panelSubtext: panelSubtext,
      panelChipBackground: panelChipBackground,
      panelChipText: panelChipText,
      panelShadow: panelShadow,
    );
  }

  Color surfaceColor({required bool undiscovered}) {
    if (!undiscovered) return panelSurface;
    return Color.lerp(panelSurface, const Color(0xFF091018), 0.42)!;
  }

  Color borderColor({required bool undiscovered}) {
    if (!undiscovered) return panelBorder;
    return Color.lerp(panelBorder, const Color(0xFF0A1018), 0.18)!;
  }

  Color accentColor({required bool undiscovered}) {
    if (!undiscovered) return panelAccent;
    return Color.lerp(panelAccent, Colors.white, 0.12)!;
  }

  Color chipBackgroundColor({required bool undiscovered}) {
    if (!undiscovered) return panelChipBackground;
    return Color.lerp(panelChipBackground, const Color(0xFF101723), 0.32)!;
  }

  Color chipTextColor({required bool undiscovered}) {
    if (!undiscovered) return panelChipText;
    return Color.lerp(panelChipText, Colors.white, 0.08)!;
  }

  Color shadowColor({required bool undiscovered}) {
    final base = undiscovered
        ? Color.lerp(panelShadow, Colors.black, 0.25)!
        : panelShadow;
    return base.withValues(alpha: undiscovered ? 0.52 : 0.38);
  }
}

const ZoneKindVisualProfile defaultZoneKindVisualProfile =
    ZoneKindVisualProfile(
      label: 'Frontier',
      icon: Icons.explore,
      atmosphereLabel: 'Unknown lands',
      focusLabel: 'Mixed territory',
      lineJoin: 'round',
      discoveredFillOpacity: 0.4,
      undiscoveredFillOpacity: 0.68,
      outerLineWidth: 7.0,
      undiscoveredOuterLineWidth: 8.0,
      innerLineWidth: 2.8,
      undiscoveredInnerLineWidth: 2.2,
      outerLineOpacity: 0.18,
      undiscoveredOuterLineOpacity: 0.42,
      innerLineOpacity: 0.95,
      outerLineBlur: 1.6,
      undiscoveredOuterLineBlur: 2.1,
      panelSurface: Color(0xFF1C2430),
      panelBorder: Color(0xFFD3BF88),
      panelAccent: Color(0xFFF0D48B),
      panelText: Color(0xFFF6E7B8),
      panelSubtext: Color(0xFFD7DDE8),
      panelChipBackground: Color(0xFF2A3644),
      panelChipText: Color(0xFFF6E7B8),
      panelShadow: Color(0xFF101723),
    );

const Map<String, ZoneKindVisualProfile> _zoneKindVisualProfiles = {
  'city': ZoneKindVisualProfile(
    label: 'City',
    icon: Icons.location_city,
    atmosphereLabel: 'Dense intrigue',
    focusLabel: 'Scenario rich',
    lineJoin: 'miter',
    discoveredFillOpacity: 0.32,
    undiscoveredFillOpacity: 0.63,
    outerLineWidth: 8.0,
    undiscoveredOuterLineWidth: 8.6,
    innerLineWidth: 2.2,
    undiscoveredInnerLineWidth: 2.0,
    outerLineOpacity: 0.24,
    undiscoveredOuterLineOpacity: 0.44,
    innerLineOpacity: 0.92,
    outerLineBlur: 0.8,
    undiscoveredOuterLineBlur: 1.5,
    panelSurface: Color(0xFF20252E),
    panelBorder: Color(0xFFD2B57A),
    panelAccent: Color(0xFFF2D18A),
    panelText: Color(0xFFF3EDDE),
    panelSubtext: Color(0xFFD3DCE5),
    panelChipBackground: Color(0xFF2D3743),
    panelChipText: Color(0xFFF4E7C2),
    panelShadow: Color(0xFF18130C),
  ),
  'village': ZoneKindVisualProfile(
    label: 'Village',
    icon: Icons.cottage,
    atmosphereLabel: 'Neighborly lanes',
    focusLabel: 'Quest friendly',
    lineJoin: 'round',
    discoveredFillOpacity: 0.35,
    undiscoveredFillOpacity: 0.64,
    outerLineWidth: 7.3,
    undiscoveredOuterLineWidth: 8.0,
    innerLineWidth: 2.1,
    undiscoveredInnerLineWidth: 1.9,
    outerLineOpacity: 0.22,
    undiscoveredOuterLineOpacity: 0.4,
    innerLineOpacity: 0.93,
    outerLineBlur: 0.9,
    undiscoveredOuterLineBlur: 1.4,
    panelSurface: Color(0xFF2C241B),
    panelBorder: Color(0xFFE1B772),
    panelAccent: Color(0xFFF1C66F),
    panelText: Color(0xFFF8F1DE),
    panelSubtext: Color(0xFFE0D4BE),
    panelChipBackground: Color(0xFF403122),
    panelChipText: Color(0xFFF7E2B8),
    panelShadow: Color(0xFF20160E),
  ),
  'academy': ZoneKindVisualProfile(
    label: 'Academy',
    icon: Icons.menu_book,
    atmosphereLabel: 'Arcane study',
    focusLabel: 'Lore heavy',
    lineJoin: 'round',
    discoveredFillOpacity: 0.36,
    undiscoveredFillOpacity: 0.66,
    outerLineWidth: 7.6,
    undiscoveredOuterLineWidth: 8.2,
    innerLineWidth: 2.6,
    undiscoveredInnerLineWidth: 2.2,
    outerLineOpacity: 0.24,
    undiscoveredOuterLineOpacity: 0.45,
    innerLineOpacity: 0.98,
    outerLineBlur: 1.6,
    undiscoveredOuterLineBlur: 2.2,
    panelSurface: Color(0xFF241F33),
    panelBorder: Color(0xFFD8C28F),
    panelAccent: Color(0xFFBDA8F5),
    panelText: Color(0xFFF7F1FF),
    panelSubtext: Color(0xFFDCCFF3),
    panelChipBackground: Color(0xFF342B48),
    panelChipText: Color(0xFFE9DDFE),
    panelShadow: Color(0xFF16101E),
  ),
  'forest': ZoneKindVisualProfile(
    label: 'Forest',
    icon: Icons.park,
    atmosphereLabel: 'Wild canopy',
    focusLabel: 'Herbalism rich',
    lineJoin: 'round',
    discoveredFillOpacity: 0.42,
    undiscoveredFillOpacity: 0.7,
    outerLineWidth: 7.4,
    undiscoveredOuterLineWidth: 8.4,
    innerLineWidth: 2.9,
    undiscoveredInnerLineWidth: 2.3,
    outerLineOpacity: 0.2,
    undiscoveredOuterLineOpacity: 0.4,
    innerLineOpacity: 0.98,
    outerLineBlur: 1.7,
    undiscoveredOuterLineBlur: 2.3,
    panelSurface: Color(0xFF1B291E),
    panelBorder: Color(0xFF9CC07A),
    panelAccent: Color(0xFFC5D97A),
    panelText: Color(0xFFF0F8E6),
    panelSubtext: Color(0xFFCBE1C7),
    panelChipBackground: Color(0xFF273A2A),
    panelChipText: Color(0xFFE2F0BE),
    panelShadow: Color(0xFF10170F),
  ),
  'swamp': ZoneKindVisualProfile(
    label: 'Swamp',
    icon: Icons.water,
    atmosphereLabel: 'Murk and mist',
    focusLabel: 'Strange encounters',
    lineJoin: 'round',
    discoveredFillOpacity: 0.44,
    undiscoveredFillOpacity: 0.72,
    outerLineWidth: 7.8,
    undiscoveredOuterLineWidth: 8.8,
    innerLineWidth: 2.6,
    undiscoveredInnerLineWidth: 2.2,
    outerLineOpacity: 0.22,
    undiscoveredOuterLineOpacity: 0.43,
    innerLineOpacity: 0.96,
    outerLineBlur: 2.0,
    undiscoveredOuterLineBlur: 2.6,
    panelSurface: Color(0xFF1F2826),
    panelBorder: Color(0xFF8EB39D),
    panelAccent: Color(0xFFB6D08B),
    panelText: Color(0xFFEFF7F1),
    panelSubtext: Color(0xFFCCE0D6),
    panelChipBackground: Color(0xFF2C3935),
    panelChipText: Color(0xFFE2EDC5),
    panelShadow: Color(0xFF121816),
  ),
  'badlands': ZoneKindVisualProfile(
    label: 'Badlands',
    icon: Icons.landscape,
    atmosphereLabel: 'Cracked frontier',
    focusLabel: 'Boss pressure',
    lineJoin: 'bevel',
    discoveredFillOpacity: 0.37,
    undiscoveredFillOpacity: 0.67,
    outerLineWidth: 8.4,
    undiscoveredOuterLineWidth: 9.0,
    innerLineWidth: 2.4,
    undiscoveredInnerLineWidth: 2.0,
    outerLineOpacity: 0.24,
    undiscoveredOuterLineOpacity: 0.45,
    innerLineOpacity: 0.92,
    outerLineBlur: 1.0,
    undiscoveredOuterLineBlur: 1.6,
    panelSurface: Color(0xFF2E211A),
    panelBorder: Color(0xFFE0AF7A),
    panelAccent: Color(0xFFF0A96A),
    panelText: Color(0xFFF8EEE4),
    panelSubtext: Color(0xFFE3CBB7),
    panelChipBackground: Color(0xFF443028),
    panelChipText: Color(0xFFF4D2B2),
    panelShadow: Color(0xFF20150F),
  ),
  'farmland': ZoneKindVisualProfile(
    label: 'Farmland',
    icon: Icons.agriculture,
    atmosphereLabel: 'Cultivated calm',
    focusLabel: 'Foraging lanes',
    lineJoin: 'miter',
    discoveredFillOpacity: 0.34,
    undiscoveredFillOpacity: 0.62,
    outerLineWidth: 7.3,
    undiscoveredOuterLineWidth: 8.0,
    innerLineWidth: 2.0,
    undiscoveredInnerLineWidth: 1.9,
    outerLineOpacity: 0.21,
    undiscoveredOuterLineOpacity: 0.39,
    innerLineOpacity: 0.91,
    outerLineBlur: 0.8,
    undiscoveredOuterLineBlur: 1.3,
    panelSurface: Color(0xFF2A2918),
    panelBorder: Color(0xFFDCC678),
    panelAccent: Color(0xFFC8D878),
    panelText: Color(0xFFF7F4E2),
    panelSubtext: Color(0xFFE2D9B7),
    panelChipBackground: Color(0xFF3D3B23),
    panelChipText: Color(0xFFEEF0BF),
    panelShadow: Color(0xFF1A170F),
  ),
  'highlands': ZoneKindVisualProfile(
    label: 'Highlands',
    icon: Icons.filter_hdr,
    atmosphereLabel: 'Wind-scoured rise',
    focusLabel: 'Mixed frontier',
    lineJoin: 'round',
    discoveredFillOpacity: 0.35,
    undiscoveredFillOpacity: 0.64,
    outerLineWidth: 7.6,
    undiscoveredOuterLineWidth: 8.4,
    innerLineWidth: 2.3,
    undiscoveredInnerLineWidth: 2.0,
    outerLineOpacity: 0.22,
    undiscoveredOuterLineOpacity: 0.41,
    innerLineOpacity: 0.94,
    outerLineBlur: 1.3,
    undiscoveredOuterLineBlur: 1.8,
    panelSurface: Color(0xFF23281F),
    panelBorder: Color(0xFFD7C890),
    panelAccent: Color(0xFFBFD39A),
    panelText: Color(0xFFF3F4EB),
    panelSubtext: Color(0xFFD6D8C4),
    panelChipBackground: Color(0xFF31372C),
    panelChipText: Color(0xFFE8E7CB),
    panelShadow: Color(0xFF151912),
  ),
  'mountain': ZoneKindVisualProfile(
    label: 'Mountain',
    icon: Icons.terrain,
    atmosphereLabel: 'High danger',
    focusLabel: 'Mining rich',
    lineJoin: 'miter',
    discoveredFillOpacity: 0.31,
    undiscoveredFillOpacity: 0.61,
    outerLineWidth: 8.2,
    undiscoveredOuterLineWidth: 8.9,
    innerLineWidth: 2.5,
    undiscoveredInnerLineWidth: 2.1,
    outerLineOpacity: 0.23,
    undiscoveredOuterLineOpacity: 0.43,
    innerLineOpacity: 0.94,
    outerLineBlur: 1.0,
    undiscoveredOuterLineBlur: 1.5,
    panelSurface: Color(0xFF232730),
    panelBorder: Color(0xFFC4CCD6),
    panelAccent: Color(0xFFBAC7D9),
    panelText: Color(0xFFF1F4F8),
    panelSubtext: Color(0xFFD4DCE6),
    panelChipBackground: Color(0xFF323844),
    panelChipText: Color(0xFFE1E9F2),
    panelShadow: Color(0xFF14181D),
  ),
  'ruins': ZoneKindVisualProfile(
    label: 'Ruins',
    icon: Icons.account_balance,
    atmosphereLabel: 'Shattered relics',
    focusLabel: 'Treasure rich',
    lineJoin: 'bevel',
    discoveredFillOpacity: 0.38,
    undiscoveredFillOpacity: 0.68,
    outerLineWidth: 7.9,
    undiscoveredOuterLineWidth: 8.6,
    innerLineWidth: 2.4,
    undiscoveredInnerLineWidth: 2.1,
    outerLineOpacity: 0.23,
    undiscoveredOuterLineOpacity: 0.44,
    innerLineOpacity: 0.95,
    outerLineBlur: 1.4,
    undiscoveredOuterLineBlur: 1.9,
    panelSurface: Color(0xFF2C241F),
    panelBorder: Color(0xFFE1BD84),
    panelAccent: Color(0xFFCEB080),
    panelText: Color(0xFFF7EEE4),
    panelSubtext: Color(0xFFE2D0C2),
    panelChipBackground: Color(0xFF3E322B),
    panelChipText: Color(0xFFF0DEBF),
    panelShadow: Color(0xFF1D1613),
  ),
  'graveyard': ZoneKindVisualProfile(
    label: 'Graveyard',
    icon: Icons.nightlight_round,
    atmosphereLabel: 'Haunted hush',
    focusLabel: 'Combat heavy',
    lineJoin: 'round',
    discoveredFillOpacity: 0.36,
    undiscoveredFillOpacity: 0.69,
    outerLineWidth: 8.0,
    undiscoveredOuterLineWidth: 8.9,
    innerLineWidth: 2.3,
    undiscoveredInnerLineWidth: 2.1,
    outerLineOpacity: 0.24,
    undiscoveredOuterLineOpacity: 0.46,
    innerLineOpacity: 0.96,
    outerLineBlur: 1.8,
    undiscoveredOuterLineBlur: 2.5,
    panelSurface: Color(0xFF222631),
    panelBorder: Color(0xFFC3CAD8),
    panelAccent: Color(0xFFA8B7D6),
    panelText: Color(0xFFF1F3F8),
    panelSubtext: Color(0xFFD2D7E1),
    panelChipBackground: Color(0xFF313747),
    panelChipText: Color(0xFFDFE4F2),
    panelShadow: Color(0xFF141721),
  ),
  'industrial': ZoneKindVisualProfile(
    label: 'Industrial',
    icon: Icons.precision_manufacturing,
    atmosphereLabel: 'Forged pressure',
    focusLabel: 'Raid heavy',
    lineJoin: 'miter',
    discoveredFillOpacity: 0.34,
    undiscoveredFillOpacity: 0.65,
    outerLineWidth: 8.3,
    undiscoveredOuterLineWidth: 9.0,
    innerLineWidth: 2.2,
    undiscoveredInnerLineWidth: 2.0,
    outerLineOpacity: 0.24,
    undiscoveredOuterLineOpacity: 0.45,
    innerLineOpacity: 0.92,
    outerLineBlur: 0.7,
    undiscoveredOuterLineBlur: 1.2,
    panelSurface: Color(0xFF2B231D),
    panelBorder: Color(0xFFD0B397),
    panelAccent: Color(0xFFCF946B),
    panelText: Color(0xFFF7EEE8),
    panelSubtext: Color(0xFFE0D1C6),
    panelChipBackground: Color(0xFF3D3028),
    panelChipText: Color(0xFFF0DDCF),
    panelShadow: Color(0xFF1B1511),
  ),
  'desert': ZoneKindVisualProfile(
    label: 'Desert',
    icon: Icons.wb_sunny_outlined,
    atmosphereLabel: 'Sun-baked reach',
    focusLabel: 'Hidden caches',
    lineJoin: 'round',
    discoveredFillOpacity: 0.32,
    undiscoveredFillOpacity: 0.63,
    outerLineWidth: 7.5,
    undiscoveredOuterLineWidth: 8.3,
    innerLineWidth: 2.1,
    undiscoveredInnerLineWidth: 1.9,
    outerLineOpacity: 0.22,
    undiscoveredOuterLineOpacity: 0.41,
    innerLineOpacity: 0.92,
    outerLineBlur: 1.1,
    undiscoveredOuterLineBlur: 1.7,
    panelSurface: Color(0xFF312516),
    panelBorder: Color(0xFFE3BE74),
    panelAccent: Color(0xFFF0C36A),
    panelText: Color(0xFFF9F0E0),
    panelSubtext: Color(0xFFE6D2B5),
    panelChipBackground: Color(0xFF453120),
    panelChipText: Color(0xFFF6E0B6),
    panelShadow: Color(0xFF22170D),
  ),
  'temple-grounds': ZoneKindVisualProfile(
    label: 'Temple Grounds',
    icon: Icons.temple_buddhist,
    atmosphereLabel: 'Sacred calm',
    focusLabel: 'Restorative',
    lineJoin: 'round',
    discoveredFillOpacity: 0.33,
    undiscoveredFillOpacity: 0.62,
    outerLineWidth: 8.0,
    undiscoveredOuterLineWidth: 8.5,
    innerLineWidth: 2.7,
    undiscoveredInnerLineWidth: 2.2,
    outerLineOpacity: 0.24,
    undiscoveredOuterLineOpacity: 0.44,
    innerLineOpacity: 0.99,
    outerLineBlur: 1.6,
    undiscoveredOuterLineBlur: 2.0,
    panelSurface: Color(0xFF2A241F),
    panelBorder: Color(0xFFE0C487),
    panelAccent: Color(0xFFE8D9A0),
    panelText: Color(0xFFF8F2E6),
    panelSubtext: Color(0xFFE3D6C4),
    panelChipBackground: Color(0xFF3A312B),
    panelChipText: Color(0xFFF0E8C8),
    panelShadow: Color(0xFF1B1612),
  ),
  'volcanic': ZoneKindVisualProfile(
    label: 'Volcanic',
    icon: Icons.local_fire_department,
    atmosphereLabel: 'Molten edge',
    focusLabel: 'Extreme danger',
    lineJoin: 'bevel',
    discoveredFillOpacity: 0.39,
    undiscoveredFillOpacity: 0.71,
    outerLineWidth: 8.6,
    undiscoveredOuterLineWidth: 9.2,
    innerLineWidth: 2.8,
    undiscoveredInnerLineWidth: 2.3,
    outerLineOpacity: 0.26,
    undiscoveredOuterLineOpacity: 0.47,
    innerLineOpacity: 0.99,
    outerLineBlur: 1.4,
    undiscoveredOuterLineBlur: 2.0,
    panelSurface: Color(0xFF2B1E1A),
    panelBorder: Color(0xFFE1A17A),
    panelAccent: Color(0xFFF06A49),
    panelText: Color(0xFFFAEEE8),
    panelSubtext: Color(0xFFE4C7BF),
    panelChipBackground: Color(0xFF402924),
    panelChipText: Color(0xFFF7D1C2),
    panelShadow: Color(0xFF1D110F),
  ),
};

String normalizeZoneKindSlug(String? rawSlug) {
  final trimmed = rawSlug?.trim().toLowerCase() ?? '';
  if (trimmed.isEmpty) {
    return '';
  }
  return trimmed
      .replaceAll(RegExp(r'[_\s]+'), '-')
      .replaceAll(RegExp(r'-+'), '-')
      .replaceAll(RegExp(r'^-|-$'), '');
}

String humanizeZoneKindSlug(String? rawSlug) {
  final normalized = normalizeZoneKindSlug(rawSlug);
  if (normalized.isEmpty) {
    return defaultZoneKindVisualProfile.label;
  }
  return normalized
      .split('-')
      .where((segment) => segment.isNotEmpty)
      .map((segment) => '${segment[0].toUpperCase()}${segment.substring(1)}')
      .join(' ');
}

ZoneKindVisualProfile zoneKindVisualProfileForSlug(String? rawSlug) {
  final normalized = normalizeZoneKindSlug(rawSlug);
  if (normalized.isEmpty) {
    return defaultZoneKindVisualProfile;
  }
  final profile = _zoneKindVisualProfiles[normalized];
  if (profile != null) {
    return profile;
  }
  return defaultZoneKindVisualProfile.copyWith(
    label: humanizeZoneKindSlug(normalized),
    atmosphereLabel: 'Unmapped character',
    focusLabel: 'Mixed territory',
  );
}
