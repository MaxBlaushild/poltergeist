class TutorialGuideChatTurn {
  final String role;
  final String content;

  const TutorialGuideChatTurn({required this.role, required this.content});

  Map<String, dynamic> toJson() => {'role': role, 'content': content};
}

class TutorialGuideChatResponse {
  final String message;

  const TutorialGuideChatResponse({required this.message});

  factory TutorialGuideChatResponse.fromJson(Map<String, dynamic> json) {
    return TutorialGuideChatResponse(
      message: json['message']?.toString().trim() ?? '',
    );
  }
}
