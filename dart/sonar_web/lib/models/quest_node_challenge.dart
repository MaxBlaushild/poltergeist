class QuestNodeChallenge {
  final String id;
  final int tier;
  final String question;
  final int reward;
  final int? inventoryItemId;
  final int difficulty;
  final List<String> statTags;
  final String? proficiency;

  const QuestNodeChallenge({
    required this.id,
    required this.tier,
    required this.question,
    required this.reward,
    this.inventoryItemId,
    this.difficulty = 0,
    this.statTags = const [],
    this.proficiency,
  });

  factory QuestNodeChallenge.fromJson(Map<String, dynamic> json) {
    return QuestNodeChallenge(
      id: json['id'] as String? ?? '',
      tier: (json['tier'] as num?)?.toInt() ?? 0,
      question: json['question'] as String? ?? '',
      reward: (json['reward'] as num?)?.toInt() ?? 0,
      inventoryItemId: (json['inventoryItemId'] as num?)?.toInt(),
      difficulty: (json['difficulty'] as num?)?.toInt() ?? 0,
      statTags: (json['statTags'] as List<dynamic>?)
              ?.map((tag) => tag.toString())
              .toList() ??
          const [],
      proficiency: json['proficiency'] as String?,
    );
  }
}
