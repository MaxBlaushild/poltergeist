/// Log/chat item from GET /sonar/chat.
class AuditItem {
  final String id;
  final String matchId;
  final String createdAt;
  final String updatedAt;
  final String message;

  const AuditItem({
    required this.id,
    required this.matchId,
    required this.createdAt,
    required this.updatedAt,
    required this.message,
  });

  factory AuditItem.fromJson(Map<String, dynamic> json) {
    return AuditItem(
      id: json['id'] as String,
      matchId: json['matchId'] as String? ?? '',
      createdAt: json['createdAt']?.toString() ?? '',
      updatedAt: json['updatedAt']?.toString() ?? '',
      message: json['message'] as String? ?? '',
    );
  }
}
