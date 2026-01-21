import 'package:skunkworks/constants/api_constants.dart';
import 'package:skunkworks/models/post.dart';
import 'package:skunkworks/services/api_client.dart';

class PostService {
  final APIClient _apiClient;

  PostService(this._apiClient);

  /// Creates a new post
  /// 
  /// [imageUrl] - The S3 URL of the uploaded image
  /// [caption] - Optional caption text
  /// 
  /// Returns the created post
  Future<Post> createPost(String imageUrl, {String? caption}) async {
    try {
      final data = <String, dynamic>{
        'imageUrl': imageUrl,
      };
      
      if (caption != null && caption.isNotEmpty) {
        data['caption'] = caption;
      }

      final response = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.createPostEndpoint,
        data: data,
      );

      return Post.fromJson(response);
    } catch (e) {
      rethrow;
    }
  }

  /// Gets the feed of posts from friends (reverse chronological)
  /// 
  /// Returns list of posts with user information
  Future<List<Post>> getFeed() async {
    try {
      final response = await _apiClient.get<List<dynamic>>(
        ApiConstants.feedEndpoint,
      );

      return response
          .map((json) => Post.fromJson(json as Map<String, dynamic>))
          .toList();
    } catch (e) {
      rethrow;
    }
  }

  /// Gets all posts from a specific user
  /// 
  /// [userId] - The user ID
  /// 
  /// Returns list of posts from the user
  Future<List<Post>> getUserPosts(String userId) async {
    try {
      final response = await _apiClient.get<List<dynamic>>(
        ApiConstants.userPostsEndpoint(userId),
      );

      return response
          .map((json) => Post.fromJson(json as Map<String, dynamic>))
          .toList();
    } catch (e) {
      rethrow;
    }
  }

  /// Deletes a post
  /// 
  /// [postId] - The post ID
  /// 
  /// Returns true if successful
  Future<bool> deletePost(String postId) async {
    try {
      await _apiClient.delete(
        ApiConstants.deletePostEndpoint(postId),
      );
      return true;
    } catch (e) {
      rethrow;
    }
  }
}

