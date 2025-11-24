import 'package:json_annotation/json_annotation.dart';

part 'document_tag.g.dart';

@JsonSerializable()
class DocumentTag {
  @JsonKey(name: 'id', fromJson: _idFromJson)
  final String id;

  @JsonKey(name: 'text')
  final String text;

  DocumentTag({
    required this.id,
    required this.text,
  });

  factory DocumentTag.fromJson(Map<String, dynamic> json) =>
      _$DocumentTagFromJson(json);

  Map<String, dynamic> toJson() => _$DocumentTagToJson(this);

  static String _idFromJson(dynamic json) {
    if (json == null) return '';
    return json.toString();
  }
}

