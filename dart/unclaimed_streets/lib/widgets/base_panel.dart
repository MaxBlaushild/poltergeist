import 'package:flutter/material.dart';

import '../models/base.dart';
import '../screens/base_management_screen.dart';
import 'paper_texture.dart';

class BasePanel extends StatelessWidget {
  const BasePanel({
    super.key,
    required this.base,
    required this.onClose,
  });

  final BasePin base;
  final VoidCallback onClose;

  String get _baseTitle {
    final owner = base.owner;
    final preferredName = owner.username.trim().isNotEmpty
        ? owner.username.trim()
        : owner.name.trim().isNotEmpty
        ? owner.name.trim()
        : owner.displayName.replaceFirst('@', '').trim();
    if (preferredName.isEmpty) {
      return 'Base';
    }
    return "$preferredName's Base";
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final maxHeight = MediaQuery.sizeOf(context).height * 0.82;

    return PaperSheet(
      child: ConstrainedBox(
        constraints: BoxConstraints(maxHeight: maxHeight),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Expanded(
                    child: Text(
                      _baseTitle,
                      style: theme.textTheme.titleLarge?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  IconButton(
                    onPressed: onClose,
                    icon: const Icon(Icons.close),
                  ),
                ],
              ),
            ),
            Expanded(
              child: BaseManagementContent(
                baseId: base.id,
                padding: const EdgeInsets.fromLTRB(16, 8, 16, 24),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
