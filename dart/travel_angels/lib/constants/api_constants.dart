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
  static const String documentsEndpoint = '/travel-angels/documents';
  static String updateDocumentEndpoint(String documentId) => '$documentsEndpoint/$documentId';
  static const String friendsDocumentsEndpoint = '/travel-angels/documents/friends';
  static const String parseDocumentEndpoint = '/travel-angels/documents/parse';

  // Google Drive endpoints
  static const String googleDriveStatusEndpoint = '/travel-angels/google-drive/status';
  static const String googleDriveAuthEndpoint = '/travel-angels/google-drive/auth';
  static const String googleDriveRevokeEndpoint = '/travel-angels/google-drive/revoke';
  static const String googleDriveFilesEndpoint = '/travel-angels/google-drive/files';
  static const String googleDriveImportDocumentEndpoint = '/travel-angels/google-drive/documents/import';

  // Friend endpoints
  static const String friendsEndpoint = '/travel-angels/friends';
  static const String friendInvitesEndpoint = '/travel-angels/friend-invites';
  static const String createFriendInviteEndpoint = '/travel-angels/friend-invites/create';
  static const String acceptFriendInviteEndpoint = '/travel-angels/friend-invites/accept';
  static String deleteFriendInviteEndpoint(String inviteId) => '/travel-angels/friend-invites/$inviteId';
  static String searchUsersEndpoint(String query) => '/travel-angels/users/search?query=$query';

  // Media endpoints
  static const String presignedUploadUrlEndpoint = '/travel-angels/media/uploadUrl';
  static const String profileEndpoint = '/travel-angels/profile';
  static String validateUsernameEndpoint(String username) => '/travel-angels/users/validate-username?username=$username';

  // Credits endpoints
  static const String creditsEndpoint = '/travel-angels/credits';
  static const String purchaseCreditsEndpoint = '/travel-angels/credits/purchase';
  static const String creditsWebhookEndpoint = '/travel-angels/credits/webhook';

  // App name for verification code requests
  static const String appName = 'travel-angels';
}


