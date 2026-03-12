class Spell {
  static const effectTypeUnlockLocks = 'unlock_locks';

  final String id;
  final String name;
  final String description;
  final String iconUrl;
  final String abilityType;
  final int abilityLevel;
  final int cooldownTurns;
  final int cooldownTurnsRemaining;
  final int cooldownSecondsRemaining;
  final DateTime? cooldownExpiresAt;
  final String effectText;
  final String schoolOfMagic;
  final int manaCost;
  final List<SpellEffect> effects;

  const Spell({
    required this.id,
    required this.name,
    this.description = '',
    this.iconUrl = '',
    this.abilityType = 'spell',
    this.abilityLevel = 1,
    this.cooldownTurns = 0,
    this.cooldownTurnsRemaining = 0,
    this.cooldownSecondsRemaining = 0,
    this.cooldownExpiresAt,
    this.effectText = '',
    this.schoolOfMagic = '',
    this.manaCost = 0,
    this.effects = const [],
  });

  factory Spell.fromJson(Map<String, dynamic> json) {
    return Spell(
      id: json['id']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      description: json['description']?.toString() ?? '',
      iconUrl: json['iconUrl']?.toString() ?? '',
      abilityType: json['abilityType']?.toString() ?? 'spell',
      abilityLevel: (json['abilityLevel'] as num?)?.toInt() ?? 1,
      cooldownTurns: (json['cooldownTurns'] as num?)?.toInt() ?? 0,
      cooldownTurnsRemaining:
          (json['cooldownTurnsRemaining'] as num?)?.toInt() ?? 0,
      cooldownSecondsRemaining:
          (json['cooldownSecondsRemaining'] as num?)?.toInt() ?? 0,
      cooldownExpiresAt: _parseDateTime(json['cooldownExpiresAt']),
      effectText: json['effectText']?.toString() ?? '',
      schoolOfMagic: json['schoolOfMagic']?.toString() ?? '',
      manaCost: (json['manaCost'] as num?)?.toInt() ?? 0,
      effects:
          (json['effects'] as List<dynamic>?)
              ?.map(
                (entry) => SpellEffect.fromJson(entry as Map<String, dynamic>),
              )
              .toList() ??
          const [],
    );
  }

  Map<String, dynamic> toJson() => {
    'id': id,
    'name': name,
    'description': description,
    'iconUrl': iconUrl,
    'abilityType': abilityType,
    'abilityLevel': abilityLevel,
    'cooldownTurns': cooldownTurns,
    'cooldownTurnsRemaining': cooldownTurnsRemaining,
    'cooldownSecondsRemaining': cooldownSecondsRemaining,
    'cooldownExpiresAt': cooldownExpiresAt?.toIso8601String(),
    'effectText': effectText,
    'schoolOfMagic': schoolOfMagic,
    'manaCost': manaCost,
    'effects': effects.map((effect) => effect.toJson()).toList(),
  };

  static DateTime? _parseDateTime(dynamic raw) {
    if (raw is String && raw.trim().isNotEmpty) {
      return DateTime.tryParse(raw.trim());
    }
    return null;
  }
}

class SpellEffect {
  final String type;
  final int amount;
  final int hits;

  const SpellEffect({required this.type, this.amount = 0, this.hits = 0});

  factory SpellEffect.fromJson(Map<String, dynamic> json) {
    return SpellEffect(
      type: json['type']?.toString() ?? '',
      amount: (json['amount'] as num?)?.toInt() ?? 0,
      hits: (json['hits'] as num?)?.toInt() ?? 0,
    );
  }

  Map<String, dynamic> toJson() => {
    'type': type,
    'amount': amount,
    'hits': hits,
  };
}
