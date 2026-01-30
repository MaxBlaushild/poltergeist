/// Activity feed item from GET /sonar/activities.
class ActivityFeed {
  final String id;
  final String userId;
  final String activityType;
  final Map<String, dynamic> data;
  final bool seen;
  final String createdAt;
  final String updatedAt;

  const ActivityFeed({
    required this.id,
    required this.userId,
    required this.activityType,
    required this.data,
    required this.seen,
    required this.createdAt,
    required this.updatedAt,
  });

  factory ActivityFeed.fromJson(Map<String, dynamic> json) {
    return ActivityFeed(
      id: json['id'] as String,
      userId: json['userId'] as String? ?? '',
      activityType: json['activityType'] as String? ?? '',
      data: json['data'] is Map<String, dynamic>
          ? json['data'] as Map<String, dynamic>
          : {},
      seen: json['seen'] as bool? ?? false,
      createdAt: json['createdAt']?.toString() ?? '',
      updatedAt: json['updatedAt']?.toString() ?? '',
    );
  }
}
