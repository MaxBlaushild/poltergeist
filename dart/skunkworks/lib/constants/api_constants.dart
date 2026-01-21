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
  static const String feedEndpoint = '/verifiable-sn/posts/feed';
  static String userPostsEndpoint(String userId) => '/verifiable-sn/posts/user/$userId';
  static String deletePostEndpoint(String postId) => '/verifiable-sn/posts/$postId';
  
  // Media endpoints
  static const String presignedUploadUrlEndpoint = '/verifiable-sn/media/uploadUrl';
  
  // Friend endpoints
  static const String friendsEndpoint = '/verifiable-sn/friends';
  static const String friendInvitesEndpoint = '/verifiable-sn/friend-invites';
  static const String createFriendInviteEndpoint = '/verifiable-sn/friend-invites/create';
  static const String acceptFriendInviteEndpoint = '/verifiable-sn/friend-invites/accept';
  static String deleteFriendInviteEndpoint(String inviteId) => '/verifiable-sn/friend-invites/$inviteId';
  
  // User endpoints
  static String searchUsersEndpoint(String query) => '/verifiable-sn/users/search?query=$query';
  
  // S3 bucket for posts
  static const String postsBucket = 'verifiable-sn-posts';
  
  // Certificate endpoints
  static const String checkCertificateEndpoint = '/verifiable-sn/certificate/check';
  static const String enrollCertificateEndpoint = '/verifiable-sn/certificate/enroll';
  static const String getCertificateEndpoint = '/verifiable-sn/certificate';
}
