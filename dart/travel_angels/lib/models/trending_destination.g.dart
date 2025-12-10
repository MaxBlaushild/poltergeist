// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'trending_destination.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

TrendingDestination _$TrendingDestinationFromJson(Map<String, dynamic> json) =>
    TrendingDestination(
      id: TrendingDestination._idFromJson(json['id']),
      name: json['name'] as String,
      formattedAddress: json['formattedAddress'] as String,
      documentCount: (json['documentCount'] as num).toInt(),
      placeId: json['placeId'] as String,
      latitude: (json['latitude'] as num).toDouble(),
      longitude: (json['longitude'] as num).toDouble(),
      locationType: TrendingDestination._locationTypeFromJson(
        json['locationType'],
      ),
      rank: (json['rank'] as num).toInt(),
    );

Map<String, dynamic> _$TrendingDestinationToJson(
  TrendingDestination instance,
) => <String, dynamic>{
  'id': instance.id,
  'name': instance.name,
  'formattedAddress': instance.formattedAddress,
  'documentCount': instance.documentCount,
  'placeId': instance.placeId,
  'latitude': instance.latitude,
  'longitude': instance.longitude,
  'locationType': _$LocationTypeEnumMap[instance.locationType]!,
  'rank': instance.rank,
};

const _$LocationTypeEnumMap = {
  LocationType.city: 'city',
  LocationType.country: 'country',
  LocationType.continent: 'continent',
};
