class QuestNodeObjective {
  static const typeChallenge = 'challenge';
  static const typeScenario = 'scenario';
  static const typeMonsterEncounter = 'monster_encounter';
  static const typeMonster = 'monster';

  final String id;
  final String type;
  final String prompt;
  final String description;
  final String imageUrl;
  final String thumbnailUrl;
  final int reward;
  final int? inventoryItemId;
  final String submissionType;
  final int difficulty;
  final List<String> statTags;
  final String? proficiency;

  const QuestNodeObjective({
    required this.id,
    required this.type,
    required this.prompt,
    this.description = '',
    this.imageUrl = '',
    this.thumbnailUrl = '',
    this.reward = 0,
    this.inventoryItemId,
    this.submissionType = 'photo',
    this.difficulty = 0,
    this.statTags = const [],
    this.proficiency,
  });

  factory QuestNodeObjective.fromJson(Map<String, dynamic> json) {
    return QuestNodeObjective(
      id: json['id'] as String? ?? '',
      type: json['type'] as String? ?? '',
      prompt: json['prompt'] as String? ?? '',
      description: json['description'] as String? ?? '',
      imageUrl: json['imageUrl'] as String? ?? '',
      thumbnailUrl: json['thumbnailUrl'] as String? ?? '',
      reward: (json['reward'] as num?)?.toInt() ?? 0,
      inventoryItemId: (json['inventoryItemId'] as num?)?.toInt(),
      submissionType: json['submissionType'] as String? ?? 'photo',
      difficulty: (json['difficulty'] as num?)?.toInt() ?? 0,
      statTags:
          (json['statTags'] as List<dynamic>?)
              ?.map((tag) => tag.toString())
              .toList() ??
          const [],
      proficiency: json['proficiency'] as String?,
    );
  }
}
