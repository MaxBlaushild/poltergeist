import 'package:json_annotation/json_annotation.dart';

part 'quick_decision_request.g.dart';

@JsonSerializable()
class QuickDecisionRequest {
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

  @JsonKey(name: 'option1')
  final String option1;

  @JsonKey(name: 'option2')
  final String option2;

  @JsonKey(name: 'option3')
  final String? option3;

  QuickDecisionRequest({
    required this.id,
    this.createdAt,
    this.updatedAt,
    this.userId,
    required this.question,
    required this.option1,
    required this.option2,
    this.option3,
  });

  factory QuickDecisionRequest.fromJson(Map<String, dynamic> json) =>
      _$QuickDecisionRequestFromJson(json);

  Map<String, dynamic> toJson() => _$QuickDecisionRequestToJson(this);

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
