import 'package:flutter/material.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/theme/app_theme.dart';

String circlePersonName(User? user) {
  for (final value in [user?.name, user?.username]) {
    if (value?.trim().isNotEmpty == true) return value!.trim();
  }
  return 'Traveler';
}

/// A shared identity row; personal profile details are deliberately not included.
class CirclePersonRow extends StatelessWidget {
  const CirclePersonRow({
    super.key,
    required this.user,
    this.trailing,
    this.detail,
  });

  final User? user;
  final Widget? trailing;
  final String? detail;

  Widget _avatar(String name) {
    final initials = name
        .split(RegExp(r'\s+'))
        .take(2)
        .map((word) => word.characters.first)
        .join()
        .toUpperCase();
    final fallback = ColoredBox(
      color: AppTheme.peach,
      child: Padding(
        padding: const EdgeInsets.all(8),
        child: Center(
          child: FittedBox(
            fit: BoxFit.scaleDown,
            child: Text(
              initials,
              style: const TextStyle(
                fontFamily: AppTheme.displayFont,
                fontSize: 20,
                color: AppTheme.ink,
              ),
            ),
          ),
        ),
      ),
    );
    final uri = Uri.tryParse(user?.profilePictureUrl?.trim() ?? '');
    final valid =
        uri != null &&
        (uri.scheme == 'https' || uri.scheme == 'http') &&
        uri.host.isNotEmpty &&
        uri.userInfo.isEmpty;
    return ExcludeSemantics(
      child: ClipOval(
        child: SizedBox.square(
          dimension: 40,
          child: valid
              ? Image.network(
                  uri.toString(),
                  fit: BoxFit.cover,
                  cacheWidth: 120,
                  errorBuilder: (_, _, _) => fallback,
                )
              : fallback,
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final name = circlePersonName(user);
    final username = user?.username?.trim() ?? '';
    final text = Theme.of(context).textTheme;
    final person = Row(
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        _avatar(name),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                name,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: text.titleSmall,
              ),
              if (username.isNotEmpty)
                Text(
                  '@$username',
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: text.bodySmall,
                ),
              if (detail?.trim().isNotEmpty == true)
                Text(detail!, style: text.bodySmall),
            ],
          ),
        ),
      ],
    );
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 10),
      child: LayoutBuilder(
        builder: (context, constraints) {
          if (trailing == null) return person;
          final stack =
              constraints.maxWidth < 300 ||
              MediaQuery.textScalerOf(context).scale(14) > 20;
          if (stack) {
            return Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                person,
                const SizedBox(height: 4),
                Align(alignment: Alignment.centerRight, child: trailing),
              ],
            );
          }
          return Row(
            children: [
              Expanded(child: person),
              const SizedBox(width: 12),
              ConstrainedBox(
                constraints: BoxConstraints(
                  maxWidth: constraints.maxWidth * .45,
                ),
                child: trailing,
              ),
            ],
          );
        },
      ),
    );
  }
}
