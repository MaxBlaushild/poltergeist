import 'package:dio/dio.dart';
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
  /// [manifestUrl] - S3 URL of the C2PA manifest
  /// [manifestHash] - SHA-256 hash of manifest bytes (hex string)
  /// [certFingerprint] - Certificate fingerprint (hex string)
  /// [assetId] - Optional C2PA asset identifier
  /// 
  /// Returns the created post
  Future<Post> createPost(
    String imageUrl, {
    String? caption,
    String? manifestUrl,
    String? manifestHash,
    String? certFingerprint,
    String? assetId,
  }) async {
    try {
      final data = <String, dynamic>{
        'imageUrl': imageUrl,
      };
      
      if (caption != null && caption.isNotEmpty) {
        data['caption'] = caption;
      }

      if (manifestUrl != null) {
        data['manifestUrl'] = manifestUrl;
      }

      if (manifestHash != null) {
        data['manifestHash'] = manifestHash;
      }

      if (certFingerprint != null) {
        data['certFingerprint'] = certFingerprint;
      }

      if (assetId != null) {
        data['assetId'] = assetId;
      }

      final response = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.createPostEndpoint,
        data: data,
      );

      return Post.fromJson(response);
    } catch (e) {
      // Log additional error details for debugging
      if (e is DioException) {
        print('Post creation error - Status: ${e.response?.statusCode}');
        print('Post creation error - Response: ${e.response?.data}');
        print('Post creation error - Request data: ${e.requestOptions.data}');
      }
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

  /// Gets a single post by ID
  /// 
  /// [postId] - The post ID
  /// 
  /// Returns the post with user information and reactions
  Future<Post> getPost(String postId) async {
    try {
      final response = await _apiClient.get<Map<String, dynamic>>(
        ApiConstants.getPostEndpoint(postId),
      );

      return Post.fromJson(response);
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

  /// Reacts to a post with an emoji
  /// 
  /// [postId] - The post ID
  /// [emoji] - The emoji to react with
  /// 
  /// Returns the created/updated reaction
  Future<Map<String, dynamic>> reactToPost(String postId, String emoji) async {
    try {
      final response = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.reactToPostEndpoint(postId),
        data: {'emoji': emoji},
      );
      return response;
    } catch (e) {
      rethrow;
    }
  }

  /// Removes a reaction from a post
  /// 
  /// [postId] - The post ID
  /// 
  /// Returns true if successful
  Future<bool> removeReaction(String postId) async {
    try {
      await _apiClient.delete(
        ApiConstants.removeReactionEndpoint(postId),
      );
      return true;
    } catch (e) {
      rethrow;
    }
  }

  /// Gets all comments for a post
  /// 
  /// [postId] - The post ID
  /// 
  /// Returns list of comments with user information
  Future<List<Comment>> getComments(String postId) async {
    try {
      final response = await _apiClient.get<List<dynamic>>(
        ApiConstants.getCommentsEndpoint(postId),
      );

      return response
          .map((json) => Comment.fromJson(json as Map<String, dynamic>))
          .toList();
    } catch (e) {
      rethrow;
    }
  }

  /// Creates a comment on a post
  /// 
  /// [postId] - The post ID
  /// [text] - The comment text
  /// 
  /// Returns the created comment with user information
  Future<Comment> createComment(String postId, String text) async {
    try {
      final response = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.createCommentEndpoint(postId),
        data: {'text': text},
      );

      return Comment.fromJson(response);
    } catch (e) {
      rethrow;
    }
  }

  /// Deletes a comment
  /// 
  /// [postId] - The post ID
  /// [commentId] - The comment ID
  /// 
  /// Returns true if successful
  Future<bool> deleteComment(String postId, String commentId) async {
    try {
      await _apiClient.delete(
        ApiConstants.deleteCommentEndpoint(postId, commentId),
      );
      return true;
    } catch (e) {
      rethrow;
    }
  }

  /// Gets the blockchain transaction for a post's manifest
  /// 
  /// [postId] - The post ID
  /// 
  /// Returns the blockchain transaction or null if not found
  Future<Map<String, dynamic>?> getBlockchainTransaction(String postId) async {
    try {
      final response = await _apiClient.get<Map<String, dynamic>>(
        ApiConstants.getBlockchainTransactionEndpoint(postId),
      );
      return response;
    } catch (e) {
      // Return null if transaction not found (404) or post has no manifest
      if (e is DioException && e.response?.statusCode == 404) {
        return null;
      }
      rethrow;
    }
  }
}

