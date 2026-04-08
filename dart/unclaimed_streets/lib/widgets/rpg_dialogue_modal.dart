import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';

import '../models/character.dart';
import '../models/character_action.dart';
import '../providers/auth_provider.dart';
import '../utils/dialogue_template.dart';

class RpgDialogueModal extends StatefulWidget {
  const RpgDialogueModal({
    super.key,
    required this.character,
    required this.action,
    required this.onClose,
    this.dialogueOverride,
    this.footerContent,
    this.primaryActionLabel,
    this.secondaryActionLabel,
    this.onPrimaryAction,
    this.onSecondaryAction,
    this.showCloseButton = true,
    this.finalStepLabel,
    this.speakerCharacters,
  });

  final Character character;
  final CharacterAction action;
  final VoidCallback onClose;
  final List<DialogueMessage>? dialogueOverride;
  final Widget? footerContent;
  final String? primaryActionLabel;
  final String? secondaryActionLabel;
  final VoidCallback? onPrimaryAction;
  final VoidCallback? onSecondaryAction;
  final bool showCloseButton;
  final String? finalStepLabel;
  final Map<String, Character>? speakerCharacters;

  @override
  State<RpgDialogueModal> createState() => _RpgDialogueModalState();
}

class _RpgDialogueModalState extends State<RpgDialogueModal>
    with SingleTickerProviderStateMixin {
  int _currentIndex = 0;
  late final AnimationController _effectController;
  String _activeEffectKey = '';

  String _dialogueEffectName(String? value) =>
      (value ?? '').trim().toLowerCase();

  Duration _effectDurationFor(String effect) {
    switch (effect) {
      case 'angry':
        return const Duration(milliseconds: 360);
      case 'surprised':
        return const Duration(milliseconds: 460);
      case 'whisper':
        return const Duration(milliseconds: 700);
      case 'shout':
        return const Duration(milliseconds: 280);
      case 'mysterious':
        return const Duration(milliseconds: 1100);
      case 'determined':
        return const Duration(milliseconds: 520);
      default:
        return const Duration(milliseconds: 240);
    }
  }

  Color _effectAccentColor(ColorScheme colorScheme, String effect) {
    switch (effect) {
      case 'angry':
        return const Color(0xFFC62828);
      case 'surprised':
        return const Color(0xFFFFF3C4);
      case 'whisper':
        return const Color(0xFF455A64);
      case 'shout':
        return const Color(0xFFE65100);
      case 'mysterious':
        return const Color(0xFF3949AB);
      case 'determined':
        return const Color(0xFFC9A227);
      default:
        return colorScheme.primary;
    }
  }

  @override
  void initState() {
    super.initState();
    _effectController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 360),
    );
    debugPrint(
      'RpgDialogueModal: initState character=${widget.character.id} action=${widget.action.id}',
    );
  }

  @override
  void dispose() {
    _effectController.dispose();
    super.dispose();
  }

  @override
  void didUpdateWidget(covariant RpgDialogueModal oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.action.id != widget.action.id ||
        oldWidget.character.id != widget.character.id) {
      _currentIndex = 0;
      _activeEffectKey = '';
    }
  }

  void _triggerEffectFor(DialogueMessage message) {
    final effect = _dialogueEffectName(message.effect);
    final nextKey = '${message.order}|${message.text}|$effect';
    if (_activeEffectKey == nextKey) {
      return;
    }
    _activeEffectKey = nextKey;
    _effectController.duration = _effectDurationFor(effect);
    switch (effect) {
      case 'angry':
      case 'surprised':
      case 'whisper':
      case 'shout':
      case 'mysterious':
      case 'determined':
        _effectController
          ..stop()
          ..reset()
          ..forward();
        return;
      default:
        _effectController
          ..stop()
          ..reset();
    }
  }

  TextStyle? _messageTextStyle(
    BuildContext context,
    String effect,
    double progress,
  ) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;
    final base = theme.textTheme.bodyLarge?.copyWith(
      height: 1.4,
      color: colorScheme.onSurface,
    );
    switch (effect) {
      case 'whisper':
        return base?.copyWith(
          color: colorScheme.onSurface.withValues(alpha: 0.76),
          fontStyle: FontStyle.italic,
          letterSpacing: 0.2,
        );
      case 'shout':
        return base?.copyWith(
          color: const Color(0xFF8C2500),
          fontWeight: FontWeight.w800,
          letterSpacing: 0.15,
        );
      case 'mysterious':
        return base?.copyWith(
          color: const Color(0xFF283593),
          letterSpacing: 0.25,
        );
      case 'determined':
        return base?.copyWith(
          color: Color.lerp(
            colorScheme.onSurface,
            const Color(0xFF6D4C00),
            0.25 + (progress * 0.35),
          ),
          fontWeight: FontWeight.w700,
        );
      case 'surprised':
        return base?.copyWith(fontWeight: FontWeight.w700);
      default:
        return base;
    }
  }

  Color _portraitTintFor(String effect, double progress) {
    switch (effect) {
      case 'angry':
        return Color.lerp(
              Colors.transparent,
              Colors.red.withValues(alpha: 0.52),
              0.35 + (progress * 0.65),
            ) ??
            Colors.red.withValues(alpha: 0.45);
      case 'shout':
        return Color.lerp(
              Colors.transparent,
              const Color(0xFFFF6F00).withValues(alpha: 0.58),
              0.4 + (progress * 0.6),
            ) ??
            const Color(0xFFFF6F00);
      case 'mysterious':
        return Color.lerp(
              Colors.transparent,
              const Color(0xFF3949AB).withValues(alpha: 0.38),
              0.2 + (progress * 0.55),
            ) ??
            const Color(0xFF3949AB);
      case 'determined':
        return Color.lerp(
              Colors.transparent,
              const Color(0xFFD4AF37).withValues(alpha: 0.34),
              0.25 + (progress * 0.45),
            ) ??
            const Color(0xFFD4AF37);
      case 'surprised':
        return Color.lerp(
              Colors.transparent,
              Colors.white.withValues(alpha: 0.4),
              0.3 + (progress * 0.5),
            ) ??
            Colors.white;
      default:
        return Colors.transparent;
    }
  }

  double _portraitScaleFor(String effect, double progress) {
    switch (effect) {
      case 'surprised':
        return 1.0 + (math.sin(progress * math.pi) * 0.08);
      case 'shout':
        return 1.0 + (math.sin(progress * math.pi) * 0.05);
      case 'determined':
        return 1.0 + (math.sin(progress * math.pi) * 0.035);
      default:
        return 1.0;
    }
  }

  double _portraitGlowOpacityFor(String effect, double progress) {
    switch (effect) {
      case 'surprised':
        return 0.22 * math.sin(progress * math.pi);
      case 'shout':
        return 0.18 * math.sin(progress * math.pi);
      case 'mysterious':
        return 0.16 * (0.3 + progress * 0.7);
      case 'determined':
        return 0.2 * math.sin(progress * math.pi);
      default:
        return 0.0;
    }
  }

  Character _characterForDialogueMessage(DialogueMessage message) {
    final characterId = message.characterId?.trim() ?? '';
    if (message.speaker == 'character' &&
        characterId.isNotEmpty &&
        widget.speakerCharacters != null) {
      final resolved = widget.speakerCharacters![characterId];
      if (resolved != null) {
        return resolved;
      }
    }
    return widget.character;
  }

  List<BoxShadow> _portraitShadowFor(String effect, double progress) {
    final glow = _portraitGlowOpacityFor(effect, progress);
    if (glow <= 0) {
      return const [];
    }
    return [
      BoxShadow(
        color: _effectAccentColor(
          Theme.of(context).colorScheme,
          effect,
        ).withValues(alpha: glow),
        blurRadius: 18 + (progress * 12),
        spreadRadius: 1.5 + (progress * 2),
      ),
    ];
  }

  Color _dialoguePanelColor(
    ColorScheme colorScheme,
    String effect,
    double progress,
  ) {
    switch (effect) {
      case 'whisper':
        return Color.lerp(
              colorScheme.surfaceContainerHighest,
              const Color(0xFFECEFF1),
              0.45 + (progress * 0.2),
            ) ??
            colorScheme.surfaceContainerHighest;
      case 'mysterious':
        return Color.lerp(
              colorScheme.surfaceContainerHighest,
              const Color(0xFFE8EAF6),
              0.3 + (progress * 0.25),
            ) ??
            colorScheme.surfaceContainerHighest;
      case 'determined':
        return Color.lerp(
              colorScheme.surfaceContainerHighest,
              const Color(0xFFFFF8E1),
              0.25 + (progress * 0.25),
            ) ??
            colorScheme.surfaceContainerHighest;
      default:
        return colorScheme.surfaceContainerHighest;
    }
  }

  double _shakeOffsetFor(String effect, double progress) {
    switch (effect) {
      case 'angry':
        return math.sin(progress * math.pi * 10.0) *
            (12.0 * (1.0 - progress * 0.35));
      case 'shout':
        return math.sin(progress * math.pi * 14.0) *
            (18.0 * (1.0 - progress * 0.28));
      case 'afraid':
        return math.sin(progress * math.pi * 18.0) * 4.0;
      default:
        return 0.0;
    }
  }

  double _cardScaleFor(String effect, double progress) {
    switch (effect) {
      case 'surprised':
        return 1.0 + (math.sin(progress * math.pi) * 0.02);
      case 'shout':
        return 1.0 + (math.sin(progress * math.pi) * 0.015);
      case 'determined':
        return 1.0 + (math.sin(progress * math.pi) * 0.012);
      default:
        return 1.0;
    }
  }

  double _cardYOffsetFor(String effect, double progress) {
    switch (effect) {
      case 'whisper':
        return 4.0 * math.sin(progress * math.pi);
      case 'mysterious':
        return -6.0 * math.sin(progress * math.pi);
      case 'surprised':
        return -5.0 * math.sin(progress * math.pi);
      default:
        return 0.0;
    }
  }

  Widget _buildEffectAura(String effect, double progress) {
    if (effect.isEmpty) {
      return const SizedBox.shrink();
    }
    final color = _effectAccentColor(Theme.of(context).colorScheme, effect);
    double opacity;
    switch (effect) {
      case 'whisper':
        opacity = 0.1 * (0.4 + progress * 0.6);
        break;
      case 'mysterious':
        opacity = 0.12 * (0.5 + progress * 0.5);
        break;
      case 'surprised':
      case 'shout':
      case 'determined':
      case 'angry':
        opacity = 0.08 * math.sin(progress * math.pi);
        break;
      default:
        opacity = 0.0;
    }
    if (opacity <= 0) {
      return const SizedBox.shrink();
    }
    return Positioned.fill(
      child: IgnorePointer(
        child: DecoratedBox(
          decoration: BoxDecoration(
            gradient: RadialGradient(
              center: const Alignment(0, -0.35),
              radius: 1.1,
              colors: [
                color.withValues(alpha: opacity),
                Colors.transparent,
              ],
            ),
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    debugPrint(
      'RpgDialogueModal: build character=${widget.character.id} action=${widget.action.id} dialogueCount=${widget.action.dialogue.length}',
    );
    final baseDialogue =
        (widget.dialogueOverride != null && widget.dialogueOverride!.isNotEmpty)
        ? widget.dialogueOverride!
        : (widget.action.dialogue.isNotEmpty
              ? widget.action.dialogue
              : const [
                  DialogueMessage(speaker: 'character', text: '...', order: 0),
                ]);
    final sorted = List<DialogueMessage>.from(baseDialogue)
      ..sort((a, b) => a.order.compareTo(b.order));
    final safeIndex = _currentIndex < sorted.length ? _currentIndex : 0;
    final current = sorted[safeIndex];
    final activeCharacter = _characterForDialogueMessage(current);
    final hasNext = safeIndex < sorted.length - 1;
    final hasDecisionActions =
        !hasNext &&
        (widget.onPrimaryAction != null || widget.onSecondaryAction != null);
    final imageUrl =
        activeCharacter.dialogueImageUrl ?? activeCharacter.mapIconUrl;
    final currentEffect = _dialogueEffectName(current.effect);
    final viewer = context.watch<AuthProvider>().user;
    final renderedText = interpolateDialogueText(current.text, viewer);
    _triggerEffectFor(current);

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
                  ? activeCharacter.name
                  : 'You';

              return AnimatedBuilder(
                animation: _effectController,
                builder: (context, _) {
                  final progress = _effectController.value;
                  final shakeOffset = _shakeOffsetFor(currentEffect, progress);
                  final cardScale = _cardScaleFor(currentEffect, progress);
                  final cardYOffset = _cardYOffsetFor(currentEffect, progress);
                  final accentColor = _effectAccentColor(
                    colorScheme,
                    currentEffect,
                  );
                  final panelColor = _dialoguePanelColor(
                    colorScheme,
                    currentEffect,
                    progress,
                  );

                  return Transform.translate(
                    offset: Offset(shakeOffset, cardYOffset),
                    child: Transform.scale(
                      scale: cardScale,
                      child: Stack(
                        children: [
                          _buildEffectAura(currentEffect, progress),
                          Container(
                            width: maxWidth,
                            constraints: BoxConstraints(maxHeight: maxHeight),
                            padding: const EdgeInsets.fromLTRB(24, 22, 24, 20),
                            decoration: BoxDecoration(
                              color: colorScheme.surface,
                              borderRadius: BorderRadius.circular(18),
                              border: Border.all(
                                color: currentEffect.isEmpty
                                    ? colorScheme.outlineVariant
                                    : Color.lerp(
                                            colorScheme.outlineVariant,
                                            accentColor.withValues(alpha: 0.65),
                                            0.35 +
                                                (math.sin(progress * math.pi) *
                                                    0.25),
                                          ) ??
                                          colorScheme.outlineVariant,
                              ),
                              boxShadow: [
                                BoxShadow(
                                  color: Colors.black.withValues(alpha: 0.2),
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
                                    _buildPortrait(
                                      context,
                                      imageUrl,
                                      effect: currentEffect,
                                      progress: progress,
                                    ),
                                    const SizedBox(width: 16),
                                    Expanded(
                                      child: Column(
                                        crossAxisAlignment:
                                            CrossAxisAlignment.start,
                                        children: [
                                          Text(
                                            speakerName,
                                            style: GoogleFonts.cinzel(
                                              textStyle:
                                                  theme.textTheme.titleLarge,
                                              fontWeight: FontWeight.w700,
                                              letterSpacing: 0.6,
                                              color: colorScheme.onSurface,
                                            ),
                                          ),
                                          if (currentEffect.isNotEmpty) ...[
                                            const SizedBox(height: 6),
                                            Container(
                                              padding:
                                                  const EdgeInsets.symmetric(
                                                    horizontal: 8,
                                                    vertical: 4,
                                                  ),
                                              decoration: BoxDecoration(
                                                color: accentColor.withValues(
                                                  alpha: 0.12,
                                                ),
                                                borderRadius:
                                                    BorderRadius.circular(999),
                                              ),
                                              child: Text(
                                                currentEffect
                                                    .replaceAll('_', ' ')
                                                    .toUpperCase(),
                                                style: theme
                                                    .textTheme
                                                    .labelSmall
                                                    ?.copyWith(
                                                      letterSpacing: 0.8,
                                                      fontWeight:
                                                          FontWeight.w800,
                                                      color: accentColor,
                                                    ),
                                              ),
                                            ),
                                          ],
                                        ],
                                      ),
                                    ),
                                    if (widget.showCloseButton)
                                      IconButton(
                                        onPressed: widget.onClose,
                                        icon: const Icon(Icons.close),
                                        style: IconButton.styleFrom(
                                          backgroundColor: colorScheme
                                              .surfaceContainerHighest,
                                          shape: const CircleBorder(),
                                        ),
                                      ),
                                  ],
                                ),
                                const SizedBox(height: 18),
                                Expanded(
                                  child: Container(
                                    padding: const EdgeInsets.fromLTRB(
                                      18,
                                      16,
                                      18,
                                      16,
                                    ),
                                    decoration: BoxDecoration(
                                      color: panelColor,
                                      borderRadius: BorderRadius.circular(16),
                                      border: Border.all(
                                        color: currentEffect.isEmpty
                                            ? colorScheme.outlineVariant
                                            : accentColor.withValues(
                                                alpha: 0.24,
                                              ),
                                      ),
                                    ),
                                    child: Column(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        Text(
                                          'Message',
                                          style: theme.textTheme.labelLarge
                                              ?.copyWith(
                                                letterSpacing: 0.8,
                                                fontWeight: FontWeight.w700,
                                                color: colorScheme
                                                    .onSurfaceVariant,
                                              ),
                                        ),
                                        const SizedBox(height: 10),
                                        Expanded(
                                          child: SingleChildScrollView(
                                            child: Text(
                                              renderedText,
                                              style: _messageTextStyle(
                                                context,
                                                currentEffect,
                                                progress,
                                              ),
                                            ),
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                ),
                                if (widget.footerContent != null) ...[
                                  const SizedBox(height: 14),
                                  widget.footerContent!,
                                ],
                                const SizedBox(height: 18),
                                Align(
                                  alignment: Alignment.centerRight,
                                  child: hasDecisionActions
                                      ? Row(
                                          mainAxisSize: MainAxisSize.min,
                                          children: [
                                            if (widget.onSecondaryAction !=
                                                null)
                                              OutlinedButton.icon(
                                                onPressed:
                                                    widget.onSecondaryAction,
                                                icon: const Icon(
                                                  Icons.close,
                                                  size: 18,
                                                ),
                                                label: Text(
                                                  widget.secondaryActionLabel ??
                                                      'Decline',
                                                ),
                                                style: OutlinedButton.styleFrom(
                                                  minimumSize: const Size(
                                                    120,
                                                    44,
                                                  ),
                                                ),
                                              ),
                                            if (widget.onSecondaryAction !=
                                                null)
                                              const SizedBox(width: 12),
                                            FilledButton.icon(
                                              onPressed: widget.onPrimaryAction,
                                              icon: const Icon(
                                                Icons.check,
                                                size: 18,
                                              ),
                                              label: Text(
                                                widget.primaryActionLabel ??
                                                    'Accept',
                                              ),
                                              style: FilledButton.styleFrom(
                                                minimumSize: const Size(
                                                  140,
                                                  44,
                                                ),
                                              ),
                                            ),
                                          ],
                                        )
                                      : FilledButton.icon(
                                          onPressed: () {
                                            if (hasNext) {
                                              setState(
                                                () => _currentIndex =
                                                    safeIndex + 1,
                                              );
                                            } else {
                                              widget.onClose();
                                            }
                                          },
                                          icon: Icon(
                                            hasNext
                                                ? Icons.arrow_forward
                                                : Icons.check,
                                            size: 18,
                                          ),
                                          label: Text(
                                            hasNext
                                                ? 'Next'
                                                : (widget.finalStepLabel ??
                                                      'Close'),
                                          ),
                                          style: FilledButton.styleFrom(
                                            minimumSize: const Size(140, 44),
                                          ),
                                        ),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
                  );
                },
              );
            },
          ),
        ),
      ),
    );
  }

  Widget _buildPortrait(
    BuildContext context,
    String? imageUrl, {
    required String effect,
    required double progress,
  }) {
    final theme = Theme.of(context);
    final portraitTint = _portraitTintFor(effect, progress);
    final portraitChild = imageUrl != null && imageUrl.isNotEmpty
        ? Image.network(
            imageUrl,
            width: 96,
            height: 96,
            fit: BoxFit.cover,
            errorBuilder: (_, _, _) => const Icon(Icons.person, size: 48),
          )
        : const Icon(Icons.person, size: 48);
    return Container(
      width: 96,
      height: 96,
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: theme.dividerColor),
        boxShadow: _portraitShadowFor(effect, progress),
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(16),
        child: Transform.scale(
          scale: _portraitScaleFor(effect, progress),
          child: portraitTint.a == 0
              ? portraitChild
              : ColorFiltered(
                  colorFilter: ColorFilter.mode(
                    portraitTint,
                    BlendMode.modulate,
                  ),
                  child: portraitChild,
                ),
        ),
      ),
    );
  }
}
