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
  void initState() {
    super.initState();
    debugPrint('RpgDialogueModal: initState character=${widget.character.id} action=${widget.action.id}');
  }

  @override
  void didUpdateWidget(covariant RpgDialogueModal oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.action.id != widget.action.id ||
        oldWidget.character.id != widget.character.id) {
      _currentIndex = 0;
    }
  }

  @override
  Widget build(BuildContext context) {
    debugPrint(
      'RpgDialogueModal: build character=${widget.character.id} action=${widget.action.id} dialogueCount=${widget.action.dialogue.length}',
    );
    final baseDialogue = widget.action.dialogue.isNotEmpty
        ? widget.action.dialogue
        : const [DialogueMessage(speaker: 'character', text: '...', order: 0)];
    final sorted = List<DialogueMessage>.from(baseDialogue)
      ..sort((a, b) => a.order.compareTo(b.order));
    final safeIndex = _currentIndex < sorted.length ? _currentIndex : 0;
    final current = sorted[safeIndex];
    final hasNext = safeIndex < sorted.length - 1;
    final imageUrl = widget.character.dialogueImageUrl ?? widget.character.mapIconUrl;

    return Material(
      color: Colors.black54,
      child: SafeArea(
        child: Center(
          child: Container(
            width: 700,
            constraints: const BoxConstraints(maxWidth: 900, maxHeight: 520),
            padding: const EdgeInsets.all(24),
            decoration: BoxDecoration(
              color: Theme.of(context).colorScheme.surface,
              borderRadius: BorderRadius.circular(16),
            ),
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
                        setState(() => _currentIndex = safeIndex + 1);
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
        ),
      ),
    );
  }
}
