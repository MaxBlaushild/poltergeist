class Friend {
  final String? id;
  final DateTime? createdAt;
  final DateTime? updatedAt;
  final String? firstUserId;
  final String? secondUserId;

  Friend({
    this.id,
    this.createdAt,
    this.updatedAt,
    this.firstUserId,
    this.secondUserId,
  });

  factory Friend.fromJson(Map<String, dynamic> json) {
    return Friend(
      id: json['id']?.toString(),
      createdAt: json['createdAt'] != null
          ? DateTime.parse(json['createdAt'])
          : null,
      updatedAt: json['updatedAt'] != null
          ? DateTime.parse(json['updatedAt'])
          : null,
      firstUserId: json['firstUserId']?.toString(),
      secondUserId: json['secondUserId']?.toString(),
    );
  }
}

class FriendInvite {
  final String? id;
  final DateTime? createdAt;
  final DateTime? updatedAt;
  final String? inviterID;
  final String? inviteeID;

  FriendInvite({
    this.id,
    this.createdAt,
    this.updatedAt,
    this.inviterID,
    this.inviteeID,
  });

  factory FriendInvite.fromJson(Map<String, dynamic> json) {
    return FriendInvite(
      id: json['id']?.toString(),
      createdAt: json['createdAt'] != null
          ? DateTime.parse(json['createdAt'])
          : null,
      updatedAt: json['updatedAt'] != null
          ? DateTime.parse(json['updatedAt'])
          : null,
      inviterID: json['inviterId']?.toString() ?? json['inviterID']?.toString(),
      inviteeID: json['inviteeId']?.toString() ?? json['inviteeID']?.toString(),
    );
  }
}

