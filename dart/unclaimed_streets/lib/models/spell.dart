class Spell {
  final String id;
  final String name;
  final String description;
  final String iconUrl;
  final String abilityType;
  final int abilityLevel;
  final int cooldownTurns;
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
    'effectText': effectText,
    'schoolOfMagic': schoolOfMagic,
    'manaCost': manaCost,
    'effects': effects.map((effect) => effect.toJson()).toList(),
  };
}

class SpellEffect {
  final String type;
  final int amount;

  const SpellEffect({required this.type, this.amount = 0});

  factory SpellEffect.fromJson(Map<String, dynamic> json) {
    return SpellEffect(
      type: json['type']?.toString() ?? '',
      amount: (json['amount'] as num?)?.toInt() ?? 0,
    );
  }

  Map<String, dynamic> toJson() => {'type': type, 'amount': amount};
}
