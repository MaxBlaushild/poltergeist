import 'movement_pattern.dart';

class Character {
  final String id;
  final String name;
  final String? description;
  final String? mapIconUrl;
  final String? dialogueImageUrl;
  final MovementPattern? movementPattern;

  const Character({
    required this.id,
    required this.name,
    this.description,
    this.mapIconUrl,
    this.dialogueImageUrl,
    this.movementPattern,
  });

  factory Character.fromJson(Map<String, dynamic> json) {
    return Character(
      id: json['id'] as String,
      name: json['name'] as String? ?? '',
      description: json['description'] as String?,
      mapIconUrl: json['mapIconUrl'] as String?,
      dialogueImageUrl: json['dialogueImageUrl'] as String?,
      movementPattern: json['movementPattern'] != null
          ? MovementPattern.fromJson(json['movementPattern'] as Map<String, dynamic>)
          : null,
    );
  }

  double get lat => movementPattern?.startingLatitude ?? 0;
  double get lng => movementPattern?.startingLongitude ?? 0;
}
