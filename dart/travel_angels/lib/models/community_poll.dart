import 'package:json_annotation/json_annotation.dart';

part 'community_poll.g.dart';

@JsonSerializable()
class CommunityPoll {
  @JsonKey(name: 'id', fromJson: _idFromJson)
  final String id;

  @JsonKey(name: 'createdAt', fromJson: _dateTimeFromJson)
  final DateTime? createdAt;

  @JsonKey(name: 'updatedAt', fromJson: _dateTimeFromJson)
  final DateTime? updatedAt;

  @JsonKey(name: 'userId', fromJson: _idFromJson)
  final String? userId;

  @JsonKey(name: 'question')
  final String question;

  @JsonKey(name: 'options')
  final List<String> options;

  CommunityPoll({
    required this.id,
    this.createdAt,
    this.updatedAt,
    this.userId,
    required this.question,
    required this.options,
  });

  factory CommunityPoll.fromJson(Map<String, dynamic> json) =>
      _$CommunityPollFromJson(json);

  Map<String, dynamic> toJson() => _$CommunityPollToJson(this);

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
