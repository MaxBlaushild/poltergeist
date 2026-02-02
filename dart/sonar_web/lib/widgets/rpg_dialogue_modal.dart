import 'package:flutter/material.dart';

import '../models/character.dart';
import '../models/character_action.dart';

class RpgDialogueModal extends StatefulWidget {
  const RpgDialogueModal({
    super.key,
    required this.character,
    required this.action,
    required this.onClose,
  });

  final Character character;
  final CharacterAction action;
  final VoidCallback onClose;

  @override
  State<RpgDialogueModal> createState() => _RpgDialogueModalState();
}

class _RpgDialogueModalState extends State<RpgDialogueModal> {
  int _currentIndex = 0;

  @override
  Widget build(BuildContext context) {
    final sorted = List<DialogueMessage>.from(widget.action.dialogue)
      ..sort((a, b) => a.order.compareTo(b.order));
    if (sorted.isEmpty) {
      WidgetsBinding.instance.addPostFrameCallback((_) => widget.onClose());
      return const SizedBox.shrink();
    }
    final current = sorted[_currentIndex];
    final hasNext = _currentIndex < sorted.length - 1;
    final imageUrl = widget.character.dialogueImageUrl ?? widget.character.mapIconUrl;

    return Dialog(
      child: Container(
        width: 700,
        height: 400,
        padding: const EdgeInsets.all(24),
        child: Column(
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                if (imageUrl != null)
                  Image.network(
                    imageUrl,
                    width: 100,
                    height: 100,
                    fit: BoxFit.cover,
                    errorBuilder: (_, __, ___) => const Icon(Icons.person, size: 100),
                  )
                else
                  const Icon(Icons.person, size: 100),
                Expanded(
                  child: Container(
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: Colors.black87,
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: Colors.white70, width: 2),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          current.speaker == 'character'
                              ? widget.character.name
                              : 'You',
                          style: Theme.of(context).textTheme.titleMedium?.copyWith(
                                fontWeight: FontWeight.bold,
                                color: Colors.white,
                              ),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          current.text,
                          style: const TextStyle(color: Colors.white70),
                        ),
                      ],
                    ),
                  ),
                ),
                IconButton(
                  onPressed: widget.onClose,
                  icon: const Icon(Icons.close),
                ),
              ],
            ),
            const Spacer(),
            Align(
              alignment: Alignment.centerRight,
              child: FilledButton(
                onPressed: () {
                  if (hasNext) {
                    setState(() => _currentIndex++);
                  } else {
                    widget.onClose();
                  }
                },
                child: Text(hasNext ? 'Next' : 'Close'),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
