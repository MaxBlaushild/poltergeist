import 'package:skunkworks/models/user.dart';

class Post {
  final String? id;
  final DateTime? createdAt;
  final DateTime? updatedAt;
  final String? userId;
  final String? imageUrl;
  final String? caption;
  final String? manifestUri;
  final String? manifestHash;
  final String? certFingerprint;
  final String? assetId;
  final User? user;

  Post({
    this.id,
    this.createdAt,
    this.updatedAt,
    this.userId,
    this.imageUrl,
    this.caption,
    this.manifestUri,
    this.manifestHash,
    this.certFingerprint,
    this.assetId,
    this.user,
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
      caption: json['caption'] as String?,
      manifestUri: json['manifestUri'] as String?,
      manifestHash: json['manifestHash'] as String?,
      certFingerprint: json['certFingerprint'] as String?,
      assetId: json['assetId'] as String?,
      user: json['user'] != null
          ? User.fromJson(json['user'] as Map<String, dynamic>)
          : null,
    );
  }
}

