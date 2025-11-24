import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/services/api_client.dart';

class CreditsService {
  final APIClient _apiClient;

  CreditsService(this._apiClient);

  /// Gets the current user's credits balance
  /// 
  /// Returns the number of credits the user has
  Future<int> getCredits() async {
    try {
      final response = await _apiClient.get<Map<String, dynamic>>(
        ApiConstants.creditsEndpoint,
      );
      return response['credits'] as int? ?? 0;
    } catch (e) {
      rethrow;
    }
  }

  /// Initiates a credit purchase
  /// 
  /// [amountInDollars] - The amount in dollars to purchase (1 dollar = 1 credit)
  /// 
  /// Returns the Stripe checkout URL
  Future<String> purchaseCredits(int amountInDollars) async {
    try {
      final response = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.purchaseCreditsEndpoint,
        data: {
          'amountInDollars': amountInDollars,
        },
      );
      return response['checkoutUrl'] as String;
    } catch (e) {
      rethrow;
    }
  }
}

