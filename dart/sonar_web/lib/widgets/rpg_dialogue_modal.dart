import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import '../models/character.dart';
import '../models/character_action.dart';

class RpgDialogueModal extends StatefulWidget {
  const RpgDialogueModal({
    super.key,
    required this.character,
    required this.action,
    required this.onClose,
    this.dialogueOverride,
    this.primaryActionLabel,
    this.secondaryActionLabel,
    this.onPrimaryAction,
    this.onSecondaryAction,
  });

  final Character character;
  final CharacterAction action;
  final VoidCallback onClose;
  final List<DialogueMessage>? dialogueOverride;
  final String? primaryActionLabel;
  final String? secondaryActionLabel;
  final VoidCallback? onPrimaryAction;
  final VoidCallback? onSecondaryAction;

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
    final baseDialogue = (widget.dialogueOverride != null &&
            widget.dialogueOverride!.isNotEmpty)
        ? widget.dialogueOverride!
        : (widget.action.dialogue.isNotEmpty
            ? widget.action.dialogue
            : const [DialogueMessage(speaker: 'character', text: '...', order: 0)]);
    final sorted = List<DialogueMessage>.from(baseDialogue)
      ..sort((a, b) => a.order.compareTo(b.order));
    final safeIndex = _currentIndex < sorted.length ? _currentIndex : 0;
    final current = sorted[safeIndex];
    final hasNext = safeIndex < sorted.length - 1;
    final hasDecisionActions = !hasNext &&
        (widget.onPrimaryAction != null || widget.onSecondaryAction != null);
    final imageUrl = widget.character.dialogueImageUrl ?? widget.character.mapIconUrl;

    return Material(
      color: Colors.black54,
      child: SafeArea(
        child: Center(
          child: LayoutBuilder(
            builder: (context, constraints) {
              final theme = Theme.of(context);
              final colorScheme = theme.colorScheme;
              final maxWidth = math.min(720.0, constraints.maxWidth - 32);
              final maxHeight = math.min(560.0, constraints.maxHeight - 32);
              final speakerName = current.speaker == 'character'
                  ? widget.character.name
                  : 'You';

              return Container(
                width: maxWidth,
                constraints: BoxConstraints(maxHeight: maxHeight),
                padding: const EdgeInsets.fromLTRB(24, 22, 24, 20),
                decoration: BoxDecoration(
                  color: colorScheme.surface,
                  borderRadius: BorderRadius.circular(18),
                  border: Border.all(color: colorScheme.outlineVariant),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withOpacity(0.2),
                      blurRadius: 18,
                      offset: const Offset(0, 10),
                    ),
                  ],
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        _buildPortrait(context, imageUrl),
                        const SizedBox(width: 16),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                speakerName,
                                style: GoogleFonts.cinzel(
                                  textStyle: theme.textTheme.titleLarge,
                                  fontWeight: FontWeight.w700,
                                  letterSpacing: 0.6,
                                  color: colorScheme.onSurface,
                                ),
                              ),
                            ],
                          ),
                        ),
                        IconButton(
                          onPressed: widget.onClose,
                          icon: const Icon(Icons.close),
                          style: IconButton.styleFrom(
                            backgroundColor: colorScheme.surfaceContainerHighest,
                            shape: const CircleBorder(),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 18),
                    Expanded(
                      child: Container(
                        padding: const EdgeInsets.fromLTRB(18, 16, 18, 16),
                        decoration: BoxDecoration(
                          color: colorScheme.surfaceContainerHighest,
                          borderRadius: BorderRadius.circular(16),
                          border: Border.all(color: colorScheme.outlineVariant),
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Message',
                              style: theme.textTheme.labelLarge?.copyWith(
                                letterSpacing: 0.8,
                                fontWeight: FontWeight.w700,
                                color: colorScheme.onSurfaceVariant,
                              ),
                            ),
                            const SizedBox(height: 10),
                            Expanded(
                              child: SingleChildScrollView(
                                child: Text(
                                  current.text,
                                  style: theme.textTheme.bodyLarge?.copyWith(
                                    height: 1.4,
                                    color: colorScheme.onSurface,
                                  ),
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                    const SizedBox(height: 18),
                    Align(
                      alignment: Alignment.centerRight,
                      child: hasDecisionActions
                          ? Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                if (widget.onSecondaryAction != null)
                                  OutlinedButton.icon(
                                    onPressed: widget.onSecondaryAction,
                                    icon: const Icon(Icons.close, size: 18),
                                    label:
                                        Text(widget.secondaryActionLabel ?? 'Decline'),
                                    style: OutlinedButton.styleFrom(
                                      minimumSize: const Size(120, 44),
                                    ),
                                  ),
                                if (widget.onSecondaryAction != null)
                                  const SizedBox(width: 12),
                                FilledButton.icon(
                                  onPressed: widget.onPrimaryAction,
                                  icon: const Icon(Icons.check, size: 18),
                                  label: Text(widget.primaryActionLabel ?? 'Accept'),
                                  style: FilledButton.styleFrom(
                                    minimumSize: const Size(140, 44),
                                  ),
                                ),
                              ],
                            )
                          : FilledButton.icon(
                              onPressed: () {
                                if (hasNext) {
                                  setState(() => _currentIndex = safeIndex + 1);
                                } else {
                                  widget.onClose();
                                }
                              },
                              icon: Icon(
                                hasNext ? Icons.arrow_forward : Icons.check,
                                size: 18,
                              ),
                              label: Text(hasNext ? 'Next' : 'Close'),
                              style: FilledButton.styleFrom(
                                minimumSize: const Size(140, 44),
                              ),
                            ),
                    ),
                  ],
                ),
              );
            },
          ),
        ),
      ),
    );
  }

  Widget _buildPortrait(BuildContext context, String? imageUrl) {
    final theme = Theme.of(context);
    return Container(
      width: 96,
      height: 96,
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: theme.dividerColor),
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(16),
        child: imageUrl != null && imageUrl.isNotEmpty
            ? Image.network(
                imageUrl,
                width: 96,
                height: 96,
                fit: BoxFit.cover,
                errorBuilder: (_, __, ___) => const Icon(Icons.person, size: 48),
              )
            : const Icon(Icons.person, size: 48),
      ),
    );
  }
}
