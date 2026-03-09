class QuestAcceptanceV2 {
  final String id;
  final String userId;
  final String questId;
  final String acceptedAt;
  final String? turnedInAt;

  const QuestAcceptanceV2({
    required this.id,
    required this.userId,
    required this.questId,
    required this.acceptedAt,
    this.turnedInAt,
  });

  factory QuestAcceptanceV2.fromJson(Map<String, dynamic> json) {
    return QuestAcceptanceV2(
      id: json['id'] as String? ?? '',
      userId: json['userId'] as String? ?? '',
      questId: json['questId'] as String? ?? '',
      acceptedAt: json['acceptedAt']?.toString() ?? '',
      turnedInAt: json['turnedInAt']?.toString(),
    );
  }
}
