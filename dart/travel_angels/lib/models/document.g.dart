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
  documentLocations: (json['documentLocations'] as List<dynamic>?)
      ?.map((e) => DocumentLocation.fromJson(e as Map<String, dynamic>))
      .toList(),
  visibility: json['visibility'] as String? ?? 'private',
  format: json['format'] as String? ?? '',
  personalNote: json['personalNote'] as String? ?? '',
  destination: json['destination'] as String? ?? '',
  isSaved: json['isSaved'] as bool? ?? false,
  sharedWithUserIds:
      (json['sharedWithUserIds'] as List<dynamic>?)
          ?.map((e) => e as String)
          .toList() ??
      const [],
  savedTripIds:
      (json['savedTripIds'] as List<dynamic>?)
          ?.map((e) => e as String)
          .toList() ??
      const [],
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
  'documentLocations': instance.documentLocations,
  'user': instance.user,
  'visibility': instance.visibility,
  'format': instance.format,
  'personalNote': instance.personalNote,
  'destination': instance.destination,
  'isSaved': instance.isSaved,
  'sharedWithUserIds': instance.sharedWithUserIds,
  'savedTripIds': instance.savedTripIds,
};

const _$CloudDocumentProviderEnumMap = {
  CloudDocumentProvider.unknown: 'unknown',
  CloudDocumentProvider.googleDocs: 'google_docs',
  CloudDocumentProvider.googleSheets: 'google_sheets',
  CloudDocumentProvider.internal: 'internal',
};
