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
    return {'album': album, 'posts': posts};
  }

  Future<void> deleteAlbum(String albumId) async {
    await _api.delete(ApiConstants.albumEndpoint(albumId));
  }
}
