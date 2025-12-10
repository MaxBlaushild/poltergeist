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
      print('LocationService: Searching for "$query"');
      final endpoint = ApiConstants.locationSearchEndpoint(query);
      print('LocationService: Calling endpoint: $endpoint');
      
      final response = await _apiClient.get<List<dynamic>>(
        endpoint,
      );
      
      print('LocationService: Received response: ${response.length} items');
      print('LocationService: Response data: $response');
      
      if (response.isEmpty) {
        print('LocationService: Empty response, returning empty list');
        return [];
      }

      final candidates = response
          .map((json) {
            try {
              return LocationCandidate.fromJson(json as Map<String, dynamic>);
            } catch (e) {
              print('LocationService: Error parsing candidate: $e, JSON: $json');
              rethrow;
            }
          })
          .toList();
      
      print('LocationService: Successfully parsed ${candidates.length} candidates');
      return candidates;
    } catch (e, stackTrace) {
      // Log the error for debugging
      print('LocationService.searchLocations error: $e');
      print('LocationService.searchLocations stackTrace: $stackTrace');
      rethrow;
    }
  }
}

