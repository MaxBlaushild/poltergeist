import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/trending_destination.dart';
import 'package:travel_angels/services/api_client.dart';

class TrendingService {
  final APIClient _apiClient;

  TrendingService(this._apiClient);

  /// Gets trending destinations (top 5 cities and top 5 countries)
  /// 
  /// Returns a map with "cities" and "countries" lists
  Future<Map<String, List<TrendingDestination>>> getTrendingDestinations() async {
    try {
      final response = await _apiClient.get<Map<String, dynamic>>(
        ApiConstants.trendingDestinationsEndpoint,
      );

      final citiesList = (response['cities'] as List<dynamic>?)
              ?.map((item) => TrendingDestination.fromJson(Map<String, dynamic>.from(item)))
              .toList() ??
          <TrendingDestination>[];

      final countriesList = (response['countries'] as List<dynamic>?)
              ?.map((item) => TrendingDestination.fromJson(Map<String, dynamic>.from(item)))
              .toList() ??
          <TrendingDestination>[];

      return {
        'cities': citiesList,
        'countries': countriesList,
      };
    } catch (e) {
      rethrow;
    }
  }
}
