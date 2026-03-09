import 'user.dart';

class Party {
  final String id;
  final String createdAt;
  final String updatedAt;
  final String leaderId;
  final User leader;
  final List<User> members;

  const Party({
    required this.id,
    required this.createdAt,
    required this.updatedAt,
    required this.leaderId,
    required this.leader,
    required this.members,
  });

  factory Party.fromJson(Map<String, dynamic> json) {
    return Party(
      id: json['id'] as String,
      createdAt: json['createdAt'] as String? ?? '',
      updatedAt: json['updatedAt'] as String? ?? '',
      leaderId: json['leaderId'] as String,
      leader: User.fromJson(json['leader'] as Map<String, dynamic>),
      members: (json['members'] as List<dynamic>?)
              ?.map((e) => User.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}
