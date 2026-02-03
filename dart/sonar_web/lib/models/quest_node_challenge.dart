class QuestNodeChallenge {
  final String id;
  final int tier;
  final String question;
  final int reward;
  final int? inventoryItemId;

  const QuestNodeChallenge({
    required this.id,
    required this.tier,
    required this.question,
    required this.reward,
    this.inventoryItemId,
  });

  factory QuestNodeChallenge.fromJson(Map<String, dynamic> json) {
    return QuestNodeChallenge(
      id: json['id'] as String? ?? '',
      tier: (json['tier'] as num?)?.toInt() ?? 0,
      question: json['question'] as String? ?? '',
      reward: (json['reward'] as num?)?.toInt() ?? 0,
      inventoryItemId: (json['inventoryItemId'] as num?)?.toInt(),
    );
  }
}
