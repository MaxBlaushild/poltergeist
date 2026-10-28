/// Shared phone input rules for sign-in and claiming a travel invitation.
abstract final class AuthPhoneNumber {
  static final _formatting = RegExp(r'[\s().-]');
  static final _international = RegExp(r'^\+[1-9]\d{7,14}$');

  /// Accepts familiar formatting while requiring an explicit country code.
  /// This checks the input shape; the authentication service validates delivery.
  static String? normalize(String value) {
    final canonical = value.replaceAll(_formatting, '');
    return _international.hasMatch(canonical) ? canonical : null;
  }

  // Keep aligned with go/authenticator/cmd/server/test_auth_users.go. These
  // accounts skip requesting SMS only; their code is still checked by the API.
  static const _testAccounts = {'+14407858475', '+12025550101'};

  static bool isTestAccount(String value) =>
      _testAccounts.contains(normalize(value));
}
