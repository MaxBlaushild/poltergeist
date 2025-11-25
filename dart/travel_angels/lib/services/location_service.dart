import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/services/api_client.dart';

class LocationCandidate {
  final String placeId;
  final String name;
  final String formattedAddress;
  final double latitude;
  final double longitude;

  LocationCandidate({
    required this.placeId,
    required this.name,
    required this.formattedAddress,
    required this.latitude,
    required this.longitude,
  });

  factory LocationCandidate.fromJson(Map<String, dynamic> json) {
    return LocationCandidate(
      placeId: json['placeId'] as String,
      name: json['name'] as String,
      formattedAddress: json['formattedAddress'] as String,
      latitude: (json['latitude'] as num).toDouble(),
      longitude: (json['longitude'] as num).toDouble(),
    );
  }
}

class LocationService {
  final APIClient _apiClient;

  LocationService(this._apiClient);

  Future<List<LocationCandidate>> searchLocations(String query) async {
    try {
      final response = await _apiClient.get<List<dynamic>>(
        ApiConstants.locationSearchEndpoint(query),
      );
      
      if (response.isEmpty) {
        return [];
      }

      return response
          .map((json) => LocationCandidate.fromJson(json as Map<String, dynamic>))
          .toList();
    } catch (e) {
      rethrow;
    }
  }
}

