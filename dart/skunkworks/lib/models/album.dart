class Album {
  final String? id;
  final DateTime? createdAt;
  final DateTime? updatedAt;
  final String? userId;
  final String name;
  final List<String> tags;

  const Album({
    this.id,
    this.createdAt,
    this.updatedAt,
    this.userId,
    required this.name,
    this.tags = const [],
  });

  factory Album.fromJson(Map<String, dynamic> json) {
    return Album(
      id: json['id']?.toString(),
      createdAt: json['createdAt'] != null
          ? DateTime.tryParse(json['createdAt'] as String)
          : null,
      updatedAt: json['updatedAt'] != null
          ? DateTime.tryParse(json['updatedAt'] as String)
          : null,
      userId: json['userId']?.toString(),
      name: json['name'] as String? ?? '',
      tags: json['tags'] != null
          ? (json['tags'] as List<dynamic>).map((t) => t.toString()).toList()
          : [],
    );
  }
}
