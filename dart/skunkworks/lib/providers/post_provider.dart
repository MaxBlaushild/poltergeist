import 'package:flutter/material.dart';
import 'package:skunkworks/models/post.dart';
import 'package:skunkworks/services/post_service.dart';

class PostProvider extends ChangeNotifier {
  final PostService _postService;
  List<Post> _feedPosts = [];
  bool _loading = false;
  String? _error;

  PostProvider(this._postService);

  List<Post> get feedPosts => _feedPosts;
  bool get loading => _loading;
  String? get error => _error;

  /// Gets tag suggestions for posts: album tags and recently used tags.
  Future<Map<String, dynamic>> getPostTagSuggestions() async {
    return _postService.getPostTagSuggestions();
  }

  /// Loads the feed of posts from friends
  Future<void> loadFeed() async {
    _loading = true;
    _error = null;
    notifyListeners();

    try {
      _feedPosts = await _postService.getFeed();
    } catch (e) {
      _error = e.toString();
    } finally {
      _loading = false;
      notifyListeners();
    }
  }

  /// Creates a new post and refreshes the feed
  /// 
  /// [imageUrl] - The S3 URL of the uploaded image or video
  /// [caption] - Optional caption text
  /// [manifestUrl] - S3 URL of the C2PA manifest
  /// [manifestHash] - SHA-256 hash of manifest bytes (hex string)
  /// [certFingerprint] - Certificate fingerprint (hex string)
  /// [assetId] - Optional C2PA asset identifier
  /// [mediaType] - Optional media type ("image" or "video"), defaults to "image"
  /// [tags] - Optional list of tags for the post
  Future<void> createPost(
    String imageUrl, {
    String? caption,
    String? manifestUrl,
    String? manifestHash,
    String? certFingerprint,
    String? assetId,
    String? mediaType,
    List<String>? tags,
  }) async {
    _error = null;
    notifyListeners();

    try {
      await _postService.createPost(
        imageUrl,
        caption: caption,
        manifestUrl: manifestUrl,
        manifestHash: manifestHash,
        certFingerprint: certFingerprint,
        assetId: assetId,
        mediaType: mediaType,
        tags: tags,
      );
      // Refresh feed after creating post
      await loadFeed();
    } catch (e) {
      _error = e.toString();
      notifyListeners();
      rethrow;
    }
  }

  /// Reacts to a post with an emoji
  /// 
  /// [postId] - The post ID
  /// [emoji] - The emoji to react with
  Future<void> reactToPost(String postId, String emoji) async {
    try {
      await _postService.reactToPost(postId, emoji);
      // Update local state optimistically
      _updatePostReaction(postId, emoji, true);
      notifyListeners();
    } catch (e) {
      _error = e.toString();
      notifyListeners();
      rethrow;
    }
  }

  /// Removes a reaction from a post
  /// 
  /// [postId] - The post ID
  Future<void> removeReaction(String postId) async {
    try {
      await _postService.removeReaction(postId);
      // Update local state optimistically
      _updatePostReaction(postId, null, false);
      notifyListeners();
    } catch (e) {
      _error = e.toString();
      notifyListeners();
      rethrow;
    }
  }

  /// Updates post reaction in local state
  void _updatePostReaction(String postId, String? emoji, bool add) {
    for (int i = 0; i < _feedPosts.length; i++) {
      if (_feedPosts[i].id == postId) {
        final post = _feedPosts[i];
        final reactions = post.reactions ?? [];
        
        if (add && emoji != null) {
          // Find existing reaction from current user
          final userReactionIndex = reactions.indexWhere((r) => r.userReacted);
          
          if (userReactionIndex >= 0) {
            // User already reacted, update the emoji
            final existingReaction = reactions[userReactionIndex];
            if (existingReaction.emoji == emoji) {
              // Same emoji, remove it
              if (existingReaction.count == 1) {
                reactions.removeAt(userReactionIndex);
              } else {
                reactions[userReactionIndex] = ReactionSummary(
                  emoji: existingReaction.emoji,
                  count: existingReaction.count - 1,
                  userReacted: false,
                );
              }
            } else {
              // Different emoji, update
              // Decrease count of old emoji
              if (existingReaction.count == 1) {
                reactions.removeAt(userReactionIndex);
              } else {
                reactions[userReactionIndex] = ReactionSummary(
                  emoji: existingReaction.emoji,
                  count: existingReaction.count - 1,
                  userReacted: false,
                );
              }
              
              // Add or increase count of new emoji
              final newReactionIndex = reactions.indexWhere((r) => r.emoji == emoji);
              if (newReactionIndex >= 0) {
                reactions[newReactionIndex] = ReactionSummary(
                  emoji: emoji,
                  count: reactions[newReactionIndex].count + 1,
                  userReacted: true,
                );
              } else {
                reactions.add(ReactionSummary(
                  emoji: emoji,
                  count: 1,
                  userReacted: true,
                ));
              }
            }
          } else {
            // User hasn't reacted yet, add new reaction
            final emojiIndex = reactions.indexWhere((r) => r.emoji == emoji);
            if (emojiIndex >= 0) {
              reactions[emojiIndex] = ReactionSummary(
                emoji: emoji,
                count: reactions[emojiIndex].count + 1,
                userReacted: true,
              );
            } else {
              reactions.add(ReactionSummary(
                emoji: emoji,
                count: 1,
                userReacted: true,
              ));
            }
          }
        } else {
          // Remove reaction
          final userReactionIndex = reactions.indexWhere((r) => r.userReacted);
          if (userReactionIndex >= 0) {
            final existingReaction = reactions[userReactionIndex];
            if (existingReaction.count == 1) {
              reactions.removeAt(userReactionIndex);
            } else {
              reactions[userReactionIndex] = ReactionSummary(
                emoji: existingReaction.emoji,
                count: existingReaction.count - 1,
                userReacted: false,
              );
            }
          }
        }
        
        _feedPosts[i] = Post(
          id: post.id,
          createdAt: post.createdAt,
          updatedAt: post.updatedAt,
          userId: post.userId,
          imageUrl: post.imageUrl,
          caption: post.caption,
          manifestUri: post.manifestUri,
          manifestHash: post.manifestHash,
          certFingerprint: post.certFingerprint,
          assetId: post.assetId,
          user: post.user,
          reactions: reactions,
          commentCount: post.commentCount,
          comments: post.comments,
        );
        break;
      }
    }
  }

  /// Loads comments for a post
  /// 
  /// [postId] - The post ID
  /// 
  /// Returns list of comments
  Future<List<Comment>> loadComments(String postId) async {
    try {
      final comments = await _postService.getComments(postId);
      // Update local state with loaded comments
      _updatePostComments(postId, comments);
      notifyListeners();
      return comments;
    } catch (e) {
      _error = e.toString();
      notifyListeners();
      rethrow;
    }
  }

  /// Creates a comment on a post
  /// 
  /// [postId] - The post ID
  /// [text] - The comment text
  Future<Comment> createComment(String postId, String text) async {
    try {
      final comment = await _postService.createComment(postId, text);
      // Update local state optimistically
      _addCommentToPost(postId, comment);
      notifyListeners();
      return comment;
    } catch (e) {
      _error = e.toString();
      notifyListeners();
      rethrow;
    }
  }

  /// Deletes a comment
  /// 
  /// [postId] - The post ID
  /// [commentId] - The comment ID
  Future<void> deleteComment(String postId, String commentId) async {
    try {
      await _postService.deleteComment(postId, commentId);
      // Update local state optimistically
      _removeCommentFromPost(postId, commentId);
      notifyListeners();
    } catch (e) {
      _error = e.toString();
      notifyListeners();
      rethrow;
    }
  }

  /// Updates post comments in local state
  void _updatePostComments(String postId, List<Comment> comments) {
    for (int i = 0; i < _feedPosts.length; i++) {
      if (_feedPosts[i].id == postId) {
        final post = _feedPosts[i];
        _feedPosts[i] = Post(
          id: post.id,
          createdAt: post.createdAt,
          updatedAt: post.updatedAt,
          userId: post.userId,
          imageUrl: post.imageUrl,
          caption: post.caption,
          manifestUri: post.manifestUri,
          manifestHash: post.manifestHash,
          certFingerprint: post.certFingerprint,
          assetId: post.assetId,
          user: post.user,
          reactions: post.reactions,
          commentCount: comments.length,
          comments: comments,
        );
        break;
      }
    }
  }

  /// Adds a comment to a post in local state
  void _addCommentToPost(String postId, Comment comment) {
    for (int i = 0; i < _feedPosts.length; i++) {
      if (_feedPosts[i].id == postId) {
        final post = _feedPosts[i];
        final comments = List<Comment>.from(post.comments ?? []);
        comments.add(comment);
        _feedPosts[i] = Post(
          id: post.id,
          createdAt: post.createdAt,
          updatedAt: post.updatedAt,
          userId: post.userId,
          imageUrl: post.imageUrl,
          caption: post.caption,
          manifestUri: post.manifestUri,
          manifestHash: post.manifestHash,
          certFingerprint: post.certFingerprint,
          assetId: post.assetId,
          user: post.user,
          reactions: post.reactions,
          commentCount: comments.length,
          comments: comments,
        );
        break;
      }
    }
  }

  /// Removes a comment from a post in local state
  void _removeCommentFromPost(String postId, String commentId) {
    for (int i = 0; i < _feedPosts.length; i++) {
      if (_feedPosts[i].id == postId) {
        final post = _feedPosts[i];
        final comments = List<Comment>.from(post.comments ?? []);
        comments.removeWhere((c) => c.id == commentId);
        _feedPosts[i] = Post(
          id: post.id,
          createdAt: post.createdAt,
          updatedAt: post.updatedAt,
          userId: post.userId,
          imageUrl: post.imageUrl,
          caption: post.caption,
          manifestUri: post.manifestUri,
          manifestHash: post.manifestHash,
          certFingerprint: post.certFingerprint,
          assetId: post.assetId,
          user: post.user,
          reactions: post.reactions,
          commentCount: comments.length,
          comments: comments,
        );
        break;
      }
    }
  }

  /// Flags a post for review
  ///
  /// [postId] - The post ID
  Future<void> flagPost(String postId) async {
    await _postService.flagPost(postId);
  }

  /// Deletes a post
  /// 
  /// [postId] - The post ID
  Future<void> deletePost(String postId) async {
    try {
      await _postService.deletePost(postId);
      // Update local state optimistically
      _removePostFromFeed(postId);
      notifyListeners();
    } catch (e) {
      _error = e.toString();
      notifyListeners();
      rethrow;
    }
  }

  /// Removes a post from the feed in local state
  void _removePostFromFeed(String postId) {
    _feedPosts.removeWhere((post) => post.id == postId);
  }
}

