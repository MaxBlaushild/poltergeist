class ResourceType {
  final String id;
  final String name;
  final String slug;
  final String description;
  final String mapIconUrl;
  final String mapIconPrompt;

  const ResourceType({
    required this.id,
    required this.name,
    required this.slug,
    required this.description,
    required this.mapIconUrl,
    required this.mapIconPrompt,
  });

  factory ResourceType.fromJson(Map<String, dynamic> json) {
    return ResourceType(
      id: json['id']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      slug: json['slug']?.toString() ?? '',
      description: json['description']?.toString() ?? '',
      mapIconUrl: json['mapIconUrl']?.toString() ?? '',
      mapIconPrompt: json['mapIconPrompt']?.toString() ?? '',
    );
  }
}
