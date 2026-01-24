class Draft {
  final String id;
  final String imagePath;
  final String? caption;
  final DateTime createdAt;

  Draft({
    required this.id,
    required this.imagePath,
    this.caption,
    required this.createdAt,
  });

  factory Draft.fromJson(Map<String, dynamic> json) {
    return Draft(
      id: json['id'] as String,
      imagePath: json['imagePath'] as String,
      caption: json['caption'] as String?,
      createdAt: json['createdAt'] != null
          ? DateTime.parse(json['createdAt'] as String)
          : DateTime.now(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'imagePath': imagePath,
      'caption': caption,
      'createdAt': createdAt.toIso8601String(),
    };
  }
}
