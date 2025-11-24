// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'document.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

Document _$DocumentFromJson(Map<String, dynamic> json) => Document(
  id: Document._idFromJson(json['id']),
  createdAt: Document._dateTimeFromJson(json['createdAt']),
  updatedAt: Document._dateTimeFromJson(json['updatedAt']),
  title: json['title'] as String,
  provider: Document._providerFromJson(json['provider']),
  userId: Document._idFromJson(json['userId']),
  link: json['link'] as String?,
  content: json['content'] as String?,
  documentTags: (json['documentTags'] as List<dynamic>?)
      ?.map((e) => DocumentTag.fromJson(e as Map<String, dynamic>))
      .toList(),
  user: json['user'] == null
      ? null
      : User.fromJson(json['user'] as Map<String, dynamic>),
);

Map<String, dynamic> _$DocumentToJson(Document instance) => <String, dynamic>{
  'id': instance.id,
  'createdAt': instance.createdAt?.toIso8601String(),
  'updatedAt': instance.updatedAt?.toIso8601String(),
  'title': instance.title,
  'provider': _$CloudDocumentProviderEnumMap[instance.provider]!,
  'userId': instance.userId,
  'link': instance.link,
  'content': instance.content,
  'documentTags': instance.documentTags,
  'user': instance.user,
};

const _$CloudDocumentProviderEnumMap = {
  CloudDocumentProvider.unknown: 'unknown',
  CloudDocumentProvider.googleDocs: 'google_docs',
  CloudDocumentProvider.googleSheets: 'google_sheets',
  CloudDocumentProvider.internal: 'internal',
};
