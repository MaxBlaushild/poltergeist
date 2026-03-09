class QuestNodeProgress {
  final String id;
  final String questAcceptanceId;
  final String questNodeId;
  final String? completedAt;

  const QuestNodeProgress({
    required this.id,
    required this.questAcceptanceId,
    required this.questNodeId,
    this.completedAt,
  });

  factory QuestNodeProgress.fromJson(Map<String, dynamic> json) {
    return QuestNodeProgress(
      id: json['id'] as String? ?? '',
      questAcceptanceId: json['questAcceptanceId'] as String? ?? '',
      questNodeId: json['questNodeId'] as String? ?? '',
      completedAt: json['completedAt']?.toString(),
    );
  }
}
