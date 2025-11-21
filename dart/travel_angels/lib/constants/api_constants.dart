class ApiConstants {
  // Base URL - should be configured via environment or config
  static const String baseUrl = 'https://api.unclaimedstreets.com';

  // Authenticator endpoints
  static const String verificationCodeEndpoint = '/authenticator/text/verification-code';
  static const String verifyTokenEndpoint = '/authenticator/token/verify';

  // Travel Angels endpoints
  static const String loginEndpoint = '/travel-angels/login';
  static const String registerEndpoint = '/travel-angels/register';
  static const String whoamiEndpoint = '/travel-angels/whoami';
  static const String levelEndpoint = '/travel-angels/level';

  // Google Drive endpoints
  static const String googleDriveStatusEndpoint = '/travel-angels/google-drive/status';
  static const String googleDriveAuthEndpoint = '/travel-angels/google-drive/auth';
  static const String googleDriveRevokeEndpoint = '/travel-angels/google-drive/revoke';

  // App name for verification code requests
  static const String appName = 'travel-angels';
}


