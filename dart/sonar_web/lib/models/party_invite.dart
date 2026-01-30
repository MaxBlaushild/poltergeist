import 'user.dart';

class PartyInvite {
  final String id;
  final String createdAt;
  final String updatedAt;
  final String inviterId;
  final String inviteeId;
  final User invitee;
  final User inviter;

  const PartyInvite({
    required this.id,
    required this.createdAt,
    required this.updatedAt,
    required this.inviterId,
    required this.inviteeId,
    required this.invitee,
    required this.inviter,
  });

  factory PartyInvite.fromJson(Map<String, dynamic> json) {
    return PartyInvite(
      id: json['id'] as String,
      createdAt: json['createdAt'] as String? ?? '',
      updatedAt: json['updatedAt'] as String? ?? '',
      inviterId: json['inviterId'] as String,
      inviteeId: json['inviteeId'] as String,
      invitee: User.fromJson(json['invitee'] as Map<String, dynamic>),
      inviter: User.fromJson(json['inviter'] as Map<String, dynamic>),
    );
  }
}
