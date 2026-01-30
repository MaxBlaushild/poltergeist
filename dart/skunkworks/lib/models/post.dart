import 'package:skunkworks/models/user.dart';

class ReactionSummary {
  final String emoji;
  final int count;
  final bool userReacted;

  ReactionSummary({
    required this.emoji,
    required this.count,
    required this.userReacted,
  });

  factory ReactionSummary.fromJson(Map<String, dynamic> json) {
    return ReactionSummary(
      emoji: json['emoji'] as String,
      count: json['count'] as int,
      userReacted: json['userReacted'] as bool? ?? false,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'emoji': emoji,
      'count': count,
      'userReacted': userReacted,
    };
  }
}

class Comment {
  final String? id;
  final DateTime? createdAt;
  final DateTime? updatedAt;
  final String? postId;
  final String? userId;
  final String? text;
  final User? user;

  Comment({
    this.id,
    this.createdAt,
    this.updatedAt,
    this.postId,
    this.userId,
    this.text,
    this.user,
  });

  factory Comment.fromJson(Map<String, dynamic> json) {
    return Comment(
      id: json['id']?.toString(),
      createdAt: json['createdAt'] != null
          ? DateTime.parse(json['createdAt'])
          : null,
      updatedAt: json['updatedAt'] != null
          ? DateTime.parse(json['updatedAt'])
          : null,
      postId: json['postId']?.toString(),
      userId: json['userId']?.toString(),
      text: json['text'] as String?,
      user: json['user'] != null
          ? User.fromJson(json['user'] as Map<String, dynamic>)
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'createdAt': createdAt?.toIso8601String(),
      'updatedAt': updatedAt?.toIso8601String(),
      'postId': postId,
      'userId': userId,
      'text': text,
      // Note: user is not serialized in toJson as it's typically loaded separately
    };
  }
}

class Post {
  final String? id;
  final DateTime? createdAt;
  final DateTime? updatedAt;
  final String? userId;
  final String? imageUrl;
  final String? mediaType; // "image" or "video", defaults to "image" if null
  final String? caption;
  final String? manifestUri;
  final String? manifestHash;
  final String? certFingerprint;
  final String? assetId;
  final User? user;
  final List<ReactionSummary>? reactions;
  final int? commentCount;
  final List<Comment>? comments;
  final List<String>? tags;

  Post({
    this.id,
    this.createdAt,
    this.updatedAt,
    this.userId,
    this.imageUrl,
    this.mediaType,
    this.caption,
    this.manifestUri,
    this.manifestHash,
    this.certFingerprint,
    this.assetId,
    this.user,
    this.reactions,
    this.commentCount,
    this.comments,
    this.tags,
  });

  factory Post.fromJson(Map<String, dynamic> json) {
    return Post(
      id: json['id']?.toString(),
      createdAt: json['createdAt'] != null
          ? DateTime.parse(json['createdAt'])
          : null,
      updatedAt: json['updatedAt'] != null
          ? DateTime.parse(json['updatedAt'])
          : null,
      userId: json['userId']?.toString(),
      imageUrl: json['imageUrl'] as String?,
      mediaType: json['mediaType'] as String?,
      caption: json['caption'] as String?,
      manifestUri: json['manifestUri'] as String?,
      manifestHash: json['manifestHash'] as String?,
      certFingerprint: json['certFingerprint'] as String?,
      assetId: json['assetId'] as String?,
      user: json['user'] != null
          ? User.fromJson(json['user'] as Map<String, dynamic>)
          : null,
      reactions: json['reactions'] != null
          ? (json['reactions'] as List<dynamic>)
              .map((r) => ReactionSummary.fromJson(r as Map<String, dynamic>))
              .toList()
          : null,
      commentCount: json['commentCount'] != null
          ? (json['commentCount'] as int)
          : null,
      comments: json['comments'] != null
          ? (json['comments'] as List<dynamic>)
              .map((c) => Comment.fromJson(c as Map<String, dynamic>))
              .toList()
          : null,
      tags: json['tags'] != null
          ? (json['tags'] as List<dynamic>).map((t) => t.toString()).toList()
          : null,
    );
  }

  /// Returns true if this post is a video
  bool get isVideo => mediaType == 'video' || _isVideoUrl(imageUrl);

  /// Returns true if this post is an image (default)
  bool get isImage => !isVideo;

  /// Helper to detect video from URL extension
  static bool _isVideoUrl(String? url) {
    if (url == null) return false;
    final lowerUrl = url.toLowerCase();
    return lowerUrl.endsWith('.mp4') ||
        lowerUrl.endsWith('.mov') ||
        lowerUrl.endsWith('.avi') ||
        lowerUrl.endsWith('.mkv') ||
        lowerUrl.contains('.mp4?') ||
        lowerUrl.contains('.mov?') ||
        lowerUrl.contains('.avi?') ||
        lowerUrl.contains('.mkv?');
  }
}

