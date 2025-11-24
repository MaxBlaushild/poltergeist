import 'package:json_annotation/json_annotation.dart';
import 'package:travel_angels/models/user.dart';

part 'friend_invite.g.dart';

@JsonSerializable()
class FriendInvite {
  @JsonKey(name: 'id', fromJson: _idFromJson)
  final String id;

  @JsonKey(name: 'createdAt', fromJson: _dateTimeFromJson)
  final DateTime? createdAt;

  @JsonKey(name: 'updatedAt', fromJson: _dateTimeFromJson)
  final DateTime? updatedAt;

  @JsonKey(name: 'inviterId', fromJson: _idFromJson)
  final String inviterId;

  @JsonKey(name: 'inviteeId', fromJson: _idFromJson)
  final String inviteeId;

  @JsonKey(name: 'inviter')
  final User? inviter;

  @JsonKey(name: 'invitee')
  final User? invitee;

  FriendInvite({
    required this.id,
    this.createdAt,
    this.updatedAt,
    required this.inviterId,
    required this.inviteeId,
    this.inviter,
    this.invitee,
  });

  factory FriendInvite.fromJson(Map<String, dynamic> json) =>
      _$FriendInviteFromJson(json);

  Map<String, dynamic> toJson() => _$FriendInviteToJson(this);

  static String _idFromJson(dynamic json) {
    if (json == null) return '';
    return json.toString();
  }

  static DateTime? _dateTimeFromJson(dynamic json) {
    if (json == null) return null;
    if (json is String) {
      try {
        return DateTime.parse(json);
      } catch (e) {
        return null;
      }
    }
    return null;
  }
}

