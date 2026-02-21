class ApiConstants {
  static const String baseUrl = 'https://api.unclaimedstreets.com';
  static const String appName = 'Sonar';

  // Authenticator (shared)
  static const String verificationCodeEndpoint =
      '/authenticator/text/verification-code';
  static const String verifyTokenEndpoint = '/authenticator/token/verify';

  // Sonar auth
  static const String loginEndpoint = '/sonar/login';
  static const String registerEndpoint = '/sonar/register';

  // Sonar user/profile
  static const String whoamiEndpoint = '/sonar/whoami';
  static const String profileEndpoint = '/sonar/profile';
  static const String profilePictureEndpoint = '/sonar/profilePicture';
  static const String mediaUploadUrlEndpoint = '/sonar/media/uploadUrl';

  static const String crewProfileBucket = 'crew-profile-icons';
  static const String crewPointsOfInterestBucket = 'crew-points-of-interest';

  // Party
  static const String partyEndpoint = '/sonar/party';
  static const String partyLeaveEndpoint = '/sonar/party/leave';
  static const String partySetLeaderEndpoint = '/sonar/party/setLeader';
  static const String partyInvitesEndpoint = '/sonar/partyInvites';
  static const String partyInvitesAcceptEndpoint = '/sonar/partyInvites/accept';
  static const String partyInvitesRejectEndpoint = '/sonar/partyInvites/reject';

  // Friends
  static const String usersSearchEndpoint = '/sonar/users/search';
  static const String friendsEndpoint = '/sonar/friends';
  static const String friendInvitesEndpoint = '/sonar/friendInvites';
  static const String friendInvitesCreateEndpoint = '/sonar/friendInvites/create';
  static const String friendInvitesAcceptEndpoint = '/sonar/friendInvites/accept';

  static String userCharacterEndpoint(String userId) =>
      '/sonar/users/$userId/character';
}
