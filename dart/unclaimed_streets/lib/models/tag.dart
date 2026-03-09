/// Tag from GET /sonar/tags (returns tag groups).
class Tag {
  final String id;
  final String name;
  final String? tagGroupId;

  const Tag({required this.id, required this.name, this.tagGroupId});

  factory Tag.fromJson(Map<String, dynamic> json) {
    return Tag(
      id: json['id'] as String,
      name: json['name'] as String? ?? '',
      tagGroupId: json['tagGroupId'] as String?,
    );
  }
}

/// Tag group from GET /sonar/tagGroups.
class TagGroup {
  final String id;
  final String name;

  const TagGroup({required this.id, required this.name});

  factory TagGroup.fromJson(Map<String, dynamic> json) {
    return TagGroup(
      id: json['id'] as String,
      name: json['name'] as String? ?? '',
    );
  }
}
