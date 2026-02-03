import 'package:skunkworks/constants/api_constants.dart';
import 'package:skunkworks/models/album.dart';
import 'package:skunkworks/models/post.dart';
import 'package:skunkworks/services/api_client.dart';

class AlbumService {
  final APIClient _api;

  AlbumService(this._api);

  Future<Album> createAlbum(String name, List<String> tags) async {
    final response = await _api.post<Map<String, dynamic>>(
      ApiConstants.albumsEndpoint,
      data: {'name': name, 'tags': tags},
    );
    return Album.fromJson(response);
  }

  Future<List<Album>> getAlbums() async {
    final response = await _api.get<List<dynamic>>(ApiConstants.albumsEndpoint);
    return response
        .map((j) => Album.fromJson(j as Map<String, dynamic>))
        .toList();
  }

  Future<Map<String, dynamic>> getAlbum(String albumId) async {
    final response = await _api.get<Map<String, dynamic>>(
      ApiConstants.albumEndpoint(albumId),
    );
    final album = Album.fromJson(response['album'] as Map<String, dynamic>);
    final postsRaw = response['posts'] as List<dynamic>? ?? [];
    final posts = postsRaw
        .map((p) => Post.fromJson(p as Map<String, dynamic>))
        .toList();
    final role = response['role'] as String?;
    final members = response['members'] as List<dynamic>?;
    final pendingInvites = response['pendingInvites'] as List<dynamic>?;
    return {
      'album': album,
      'posts': posts,
      'role': role,
      'members': members ?? [],
      'pendingInvites': pendingInvites ?? [],
    };
  }

  Future<void> deleteAlbum(String albumId) async {
    await _api.delete(ApiConstants.albumEndpoint(albumId));
  }

  Future<void> addAlbumTag(String albumId, String tag) async {
    await _api.post(ApiConstants.albumTagsEndpoint(albumId), data: {'tag': tag});
  }

  Future<void> removeAlbumTag(String albumId, String tag) async {
    await _api.delete(ApiConstants.albumTagsEndpoint(albumId), data: {'tag': tag});
  }

  Future<Map<String, dynamic>> inviteToAlbum(String albumId, String userId, String role) async {
    final response = await _api.post<Map<String, dynamic>>(
      ApiConstants.albumInviteEndpoint(albumId),
      data: {'userId': userId, 'role': role},
    );
    return response;
  }

  Future<List<dynamic>> getAlbumMembers(String albumId) async {
    final response = await _api.get<List<dynamic>>(ApiConstants.albumMembersEndpoint(albumId));
    return response;
  }

  Future<void> removeAlbumMember(String albumId, String userId) async {
    await _api.delete(ApiConstants.albumMembersEndpoint(albumId), data: {'userId': userId});
  }

  Future<void> updateAlbumMemberRole(String albumId, String userId, String role) async {
    await _api.patch(ApiConstants.albumMembersEndpoint(albumId), data: {'userId': userId, 'role': role});
  }

  Future<List<dynamic>> getAlbumPendingInvites(String albumId) async {
    final response = await _api.get<List<dynamic>>(ApiConstants.albumInvitesEndpoint(albumId));
    return response;
  }

  Future<List<dynamic>> getMyAlbumInvites() async {
    final response = await _api.get<List<dynamic>>(ApiConstants.albumInvitesListEndpoint);
    return response;
  }

  Future<void> acceptAlbumInvite(String inviteId) async {
    await _api.post(ApiConstants.acceptAlbumInviteEndpoint(inviteId));
  }

  Future<void> rejectAlbumInvite(String inviteId) async {
    await _api.post(ApiConstants.rejectAlbumInviteEndpoint(inviteId));
  }
}
