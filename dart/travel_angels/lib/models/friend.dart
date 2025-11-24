import 'package:json_annotation/json_annotation.dart';
import 'package:travel_angels/models/user.dart';

part 'friend.g.dart';

@JsonSerializable()
class Friend {
  @JsonKey(name: 'id', fromJson: _idFromJson)
  final String id;

  @JsonKey(name: 'createdAt', fromJson: _dateTimeFromJson)
  final DateTime? createdAt;

  @JsonKey(name: 'updatedAt', fromJson: _dateTimeFromJson)
  final DateTime? updatedAt;

  @JsonKey(name: 'firstUserId', fromJson: _idFromJson)
  final String firstUserId;

  @JsonKey(name: 'secondUserId', fromJson: _idFromJson)
  final String secondUserId;

  @JsonKey(name: 'firstUser')
  final User? firstUser;

  @JsonKey(name: 'secondUser')
  final User? secondUser;

  Friend({
    required this.id,
    this.createdAt,
    this.updatedAt,
    required this.firstUserId,
    required this.secondUserId,
    this.firstUser,
    this.secondUser,
  });

  factory Friend.fromJson(Map<String, dynamic> json) =>
      _$FriendFromJson(json);

  Map<String, dynamic> toJson() => _$FriendToJson(this);

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

