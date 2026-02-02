class Notification {
  final String? id;
  final String? userId;
  final String type;
  final String? actorId;
  final String? albumId;
  final String? postId;
  final String? inviteId;
  final DateTime? readAt;
  final DateTime? createdAt;
  final Map<String, dynamic>? actor;
  final Map<String, dynamic>? album;

  const Notification({
    this.id,
    this.userId,
    required this.type,
    this.actorId,
    this.albumId,
    this.postId,
    this.inviteId,
    this.readAt,
    this.createdAt,
    this.actor,
    this.album,
  });

  bool get isRead => readAt != null;

  factory Notification.fromJson(Map<String, dynamic> json) {
    return Notification(
      id: json['id']?.toString(),
      userId: json['userId']?.toString(),
      type: json['type'] as String? ?? '',
      actorId: json['actorId']?.toString(),
      albumId: json['albumId']?.toString(),
      postId: json['postId']?.toString(),
      inviteId: json['inviteId']?.toString(),
      readAt: json['readAt'] != null ? DateTime.tryParse(json['readAt'] as String) : null,
      createdAt: json['createdAt'] != null ? DateTime.tryParse(json['createdAt'] as String) : null,
      actor: json['actor'] as Map<String, dynamic>?,
      album: json['album'] as Map<String, dynamic>?,
    );
  }
}
