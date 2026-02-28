class Spell {
  final String id;
  final String name;
  final String description;
  final String iconUrl;
  final String effectText;
  final String schoolOfMagic;
  final int manaCost;

  const Spell({
    required this.id,
    required this.name,
    this.description = '',
    this.iconUrl = '',
    this.effectText = '',
    this.schoolOfMagic = '',
    this.manaCost = 0,
  });

  factory Spell.fromJson(Map<String, dynamic> json) {
    return Spell(
      id: json['id']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      description: json['description']?.toString() ?? '',
      iconUrl: json['iconUrl']?.toString() ?? '',
      effectText: json['effectText']?.toString() ?? '',
      schoolOfMagic: json['schoolOfMagic']?.toString() ?? '',
      manaCost: (json['manaCost'] as num?)?.toInt() ?? 0,
    );
  }

  Map<String, dynamic> toJson() => {
    'id': id,
    'name': name,
    'description': description,
    'iconUrl': iconUrl,
    'effectText': effectText,
    'schoolOfMagic': schoolOfMagic,
    'manaCost': manaCost,
  };
}
