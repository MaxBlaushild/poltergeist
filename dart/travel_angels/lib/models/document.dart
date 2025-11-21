import 'package:json_annotation/json_annotation.dart';

part 'document.g.dart';

enum CloudDocumentProvider {
  @JsonValue('unknown')
  unknown,
  @JsonValue('google_docs')
  googleDocs,
  @JsonValue('google_sheets')
  googleSheets,
  @JsonValue('internal')
  internal,
}

@JsonSerializable()
class Document {
  @JsonKey(name: 'id', fromJson: _idFromJson)
  final String id;

  @JsonKey(name: 'createdAt', fromJson: _dateTimeFromJson)
  final DateTime? createdAt;

  @JsonKey(name: 'updatedAt', fromJson: _dateTimeFromJson)
  final DateTime? updatedAt;

  @JsonKey(name: 'title')
  final String title;

  @JsonKey(name: 'provider', fromJson: _providerFromJson)
  final CloudDocumentProvider provider;

  @JsonKey(name: 'userId', fromJson: _idFromJson)
  final String? userId;

  @JsonKey(name: 'link')
  final String? link;

  @JsonKey(name: 'content')
  final String? content;

  Document({
    required this.id,
    this.createdAt,
    this.updatedAt,
    required this.title,
    required this.provider,
    this.userId,
    this.link,
    this.content,
  });

  factory Document.fromJson(Map<String, dynamic> json) =>
      _$DocumentFromJson(json);

  Map<String, dynamic> toJson() => _$DocumentToJson(this);

  static String _idFromJson(dynamic json) {
    if (json == null) return '';
    return json.toString();
  }

  static DateTime? _dateTimeFromJson(dynamic json) {
    if (json == null) return null;
    if (json is String) {
      try {
        return DateTime.parse(json);
      } catch (e) {
        return null;
      }
    }
    return null;
  }

  static CloudDocumentProvider _providerFromJson(dynamic json) {
    if (json == null) {
      return CloudDocumentProvider.unknown;
    }
    final str = json.toString();
    switch (str) {
      case 'google_docs':
        return CloudDocumentProvider.googleDocs;
      case 'google_sheets':
        return CloudDocumentProvider.googleSheets;
      case 'internal':
        return CloudDocumentProvider.internal;
      default:
        return CloudDocumentProvider.unknown;
    }
  }
}

