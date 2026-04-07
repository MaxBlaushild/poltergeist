import '../models/user.dart';

const String dialogueUsernameToken = '{{username}}';

String interpolateDialogueText(String text, User? user) {
  final username = user?.username.trim() ?? '';
  final name = user?.name.trim() ?? '';
  final replacement = username.isNotEmpty
      ? username
      : (name.isNotEmpty ? name : 'you');

  return text.replaceAllMapped(
    RegExp(r'\{\{\s*(username|user\.username)\s*\}\}', caseSensitive: false),
    (_) => replacement,
  );
}
