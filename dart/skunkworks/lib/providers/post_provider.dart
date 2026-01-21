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
  /// [imageUrl] - The S3 URL of the uploaded image
  /// [caption] - Optional caption text
  Future<void> createPost(String imageUrl, {String? caption}) async {
    _error = null;
    notifyListeners();

    try {
      await _postService.createPost(imageUrl, caption: caption);
      // Refresh feed after creating post
      await loadFeed();
    } catch (e) {
      _error = e.toString();
      notifyListeners();
      rethrow;
    }
  }
}

