class ApiConstants {
  // Base URL - should be configured via environment or config
  static const String baseUrl = 'https://api.unclaimedstreets.com';

  // App name for verification code requests
  static const String appName = 'skunkworks';

  // Authenticator endpoints (shared)
  static const String verificationCodeEndpoint = '/authenticator/text/verification-code';
  static const String verifyTokenEndpoint = '/authenticator/token/verify';

  // Verifiable-sn endpoints
  static const String loginEndpoint = '/verifiable-sn/login';
  static const String registerEndpoint = '/verifiable-sn/register';
  
  // Post endpoints
  static const String createPostEndpoint = '/verifiable-sn/posts';
  static const String postTagSuggestionsEndpoint = '/verifiable-sn/post-tag-suggestions';
  static const String feedEndpoint = '/verifiable-sn/posts/feed';
  static String userPostsEndpoint(String userId) => '/verifiable-sn/posts/user/$userId';
  static String getPostEndpoint(String postId) => '/verifiable-sn/posts/$postId';
  static String postTagsEndpoint(String postId) => '/verifiable-sn/posts/$postId/tags';
  static String deletePostEndpoint(String postId) => '/verifiable-sn/posts/$postId';
  static String reactToPostEndpoint(String postId) => '/verifiable-sn/posts/$postId/reactions';
  static String removeReactionEndpoint(String postId) => '/verifiable-sn/posts/$postId/reactions';
  static String getCommentsEndpoint(String postId) => '/verifiable-sn/posts/$postId/comments';
  static String createCommentEndpoint(String postId) => '/verifiable-sn/posts/$postId/comments';
  static String deleteCommentEndpoint(String postId, String commentId) => '/verifiable-sn/posts/$postId/comments/$commentId';
  static String getBlockchainTransactionEndpoint(String postId) => '/verifiable-sn/posts/$postId/blockchain-transaction';
  static String flagPostEndpoint(String postId) => '/verifiable-sn/posts/$postId/flag';

  // Album endpoints
  static const String albumsEndpoint = '/verifiable-sn/albums';
  static String albumEndpoint(String albumId) => '/verifiable-sn/albums/$albumId';
  static String albumTagsEndpoint(String albumId) => '/verifiable-sn/albums/$albumId/tags';
  static String albumInviteEndpoint(String albumId) => '/verifiable-sn/albums/$albumId/invite';
  static String albumMembersEndpoint(String albumId) => '/verifiable-sn/albums/$albumId/members';
  static String albumInvitesEndpoint(String albumId) => '/verifiable-sn/albums/$albumId/invites';
  static const String albumInvitesListEndpoint = '/verifiable-sn/album-invites';
  static String acceptAlbumInviteEndpoint(String inviteId) => '/verifiable-sn/album-invites/$inviteId/accept';
  static String rejectAlbumInviteEndpoint(String inviteId) => '/verifiable-sn/album-invites/$inviteId/reject';
  static String createAlbumShareEndpoint(String albumId) => '/verifiable-sn/albums/$albumId/share';
  static String albumShareEndpoint(String token) => '/verifiable-sn/album-shares/$token';

  // Notification endpoints
  static const String notificationsEndpoint = '/verifiable-sn/notifications';
  static String notificationReadEndpoint(String id) => '/verifiable-sn/notifications/$id/read';
  static const String notificationsReadAllEndpoint = '/verifiable-sn/notifications/read-all';
  static const String deviceTokensEndpoint = '/verifiable-sn/device-tokens';

  // Media endpoints
  static const String presignedUploadUrlEndpoint = '/verifiable-sn/media/uploadUrl';

  // Social endpoints
  static const String socialAccountsEndpoint = '/verifiable-sn/social/accounts';
  static String socialAuthEndpoint(String provider) => '/verifiable-sn/social/$provider/auth';
  static String socialCallbackEndpoint(String provider) => '/verifiable-sn/social/$provider/callback';
  static String socialRevokeEndpoint(String provider) => '/verifiable-sn/social/$provider/revoke';
  static String socialPostEndpoint(String provider) => '/verifiable-sn/social/$provider/post';
  
  // Friend endpoints
  static const String friendsEndpoint = '/verifiable-sn/friends';
  static const String friendInvitesEndpoint = '/verifiable-sn/friend-invites';
  static const String createFriendInviteEndpoint = '/verifiable-sn/friend-invites/create';
  static const String acceptFriendInviteEndpoint = '/verifiable-sn/friend-invites/accept';
  static String deleteFriendInviteEndpoint(String inviteId) => '/verifiable-sn/friend-invites/$inviteId';
  
  // User endpoints
  static String searchUsersEndpoint(String query) => '/verifiable-sn/users/search?query=$query';
  static const String updateProfileEndpoint = '/verifiable-sn/users/profile';
  
  // S3 bucket for posts
  static const String postsBucket = 'verifiable-sn-posts';

  /// Public TestFlight beta link for sharing posts. Append ?post=<id> to
  /// associate shared link with a post (for future deeplink / analytics).
  static const String shareTestFlightUrl = 'https://testflight.apple.com/join/XTzctGYr';

  static String sharePostUrl(String postId) =>
      '$shareTestFlightUrl?post=${Uri.encodeComponent(postId)}';

  static String shareAlbumUrl(String token) =>
      '$baseUrl/verifiable-sn/album-shares/${Uri.encodeComponent(token)}/open';

  /// Builds vera:// deep link for export QR code.
  /// m=manifestHash, t=txHash (blockchain). Omit params if null.
  static String exportPostDeepLink(String postId, {String? manifestHash, String? txHash}) {
    final uri = Uri(scheme: 'vera', host: 'post', pathSegments: [postId]);
    final query = <String, String>{};
    if (manifestHash != null && manifestHash.isNotEmpty) query['m'] = manifestHash;
    if (txHash != null && txHash.isNotEmpty) query['t'] = txHash;
    return query.isEmpty ? uri.toString() : uri.replace(queryParameters: query).toString();
  }

  // Certificate endpoints
  static const String checkCertificateEndpoint = '/verifiable-sn/certificate/check';
  static const String enrollCertificateEndpoint = '/verifiable-sn/certificate/enroll';
  static const String getCertificateEndpoint = '/verifiable-sn/certificate';
  static String getUserCertificateEndpoint(String userId) => '/verifiable-sn/certificate/user/$userId';
}
