// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'document_location.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

DocumentLocation _$DocumentLocationFromJson(Map<String, dynamic> json) =>
    DocumentLocation(
      id: DocumentLocation._idFromJson(json['id']),
      documentId: DocumentLocation._idFromJson(json['documentId']),
      placeId: json['placeId'] as String,
      name: json['name'] as String,
      formattedAddress: json['formattedAddress'] as String,
      latitude: (json['latitude'] as num).toDouble(),
      longitude: (json['longitude'] as num).toDouble(),
      locationType: DocumentLocation._locationTypeFromJson(
        json['locationType'],
      ),
    );

Map<String, dynamic> _$DocumentLocationToJson(DocumentLocation instance) =>
    <String, dynamic>{
      'id': instance.id,
      'documentId': instance.documentId,
      'placeId': instance.placeId,
      'name': instance.name,
      'formattedAddress': instance.formattedAddress,
      'latitude': instance.latitude,
      'longitude': instance.longitude,
      'locationType': _$LocationTypeEnumMap[instance.locationType]!,
    };

const _$LocationTypeEnumMap = {
  LocationType.city: 'city',
  LocationType.country: 'country',
  LocationType.continent: 'continent',
};
