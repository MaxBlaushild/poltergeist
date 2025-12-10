// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'quick_decision_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

QuickDecisionRequest _$QuickDecisionRequestFromJson(
  Map<String, dynamic> json,
) => QuickDecisionRequest(
  id: QuickDecisionRequest._idFromJson(json['id']),
  createdAt: QuickDecisionRequest._dateTimeFromJson(json['createdAt']),
  updatedAt: QuickDecisionRequest._dateTimeFromJson(json['updatedAt']),
  userId: QuickDecisionRequest._idFromJson(json['userId']),
  question: json['question'] as String,
  option1: json['option1'] as String,
  option2: json['option2'] as String,
  option3: json['option3'] as String?,
);

Map<String, dynamic> _$QuickDecisionRequestToJson(
  QuickDecisionRequest instance,
) => <String, dynamic>{
  'id': instance.id,
  'createdAt': instance.createdAt?.toIso8601String(),
  'updatedAt': instance.updatedAt?.toIso8601String(),
  'userId': instance.userId,
  'question': instance.question,
  'option1': instance.option1,
  'option2': instance.option2,
  'option3': instance.option3,
};
