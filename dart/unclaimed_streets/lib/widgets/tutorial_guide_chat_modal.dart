import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import '../models/character.dart';
import '../models/tutorial_guide_chat.dart';

class TutorialGuideChatModal extends StatefulWidget {
  const TutorialGuideChatModal({
    super.key,
    required this.character,
    required this.onClose,
    required this.onSendMessage,
    required this.initialAssistantMessage,
  });

  final Character character;
  final VoidCallback onClose;
  final Future<String> Function(
    String message,
    List<TutorialGuideChatTurn> history,
  )
  onSendMessage;
  final String initialAssistantMessage;

  @override
  State<TutorialGuideChatModal> createState() => _TutorialGuideChatModalState();
}

class _TutorialGuideChatModalState extends State<TutorialGuideChatModal> {
  late final TextEditingController _composerController;
  late final ScrollController _scrollController;
  late final List<TutorialGuideChatTurn> _messages;
  bool _composerHasText = false;
  bool _sending = false;
  String? _errorText;

  @override
  void initState() {
    super.initState();
    _composerController = TextEditingController();
    _composerController.addListener(_handleComposerChanged);
    _scrollController = ScrollController();
    _messages = [
      TutorialGuideChatTurn(
        role: 'assistant',
        content: widget.initialAssistantMessage,
      ),
    ];
  }

  @override
  void dispose() {
    _composerController.removeListener(_handleComposerChanged);
    _composerController.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  void _handleComposerChanged() {
    final hasText = _composerController.text.trim().isNotEmpty;
    if (hasText == _composerHasText || !mounted) {
      return;
    }
    setState(() => _composerHasText = hasText);
  }

  String _portraitUrl() {
    final dialogue = widget.character.dialogueImageUrl?.trim() ?? '';
    if (dialogue.isNotEmpty) return dialogue;
    final thumbnail = widget.character.thumbnailUrl?.trim() ?? '';
    if (thumbnail.isNotEmpty) return thumbnail;
    final mapIcon = widget.character.mapIconUrl?.trim() ?? '';
    return mapIcon;
  }

  List<TutorialGuideChatTurn> _requestHistory() {
    if (_messages.length <= 1) {
      return const [];
    }
    return List<TutorialGuideChatTurn>.from(_messages.skip(1));
  }

  void _scheduleScrollToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!_scrollController.hasClients) return;
      _scrollController.animateTo(
        _scrollController.position.maxScrollExtent,
        duration: const Duration(milliseconds: 220),
        curve: Curves.easeOutCubic,
      );
    });
  }

  Future<void> _sendCurrentMessage() async {
    final message = _composerController.text.trim();
    if (message.isEmpty || _sending) return;

    final history = _requestHistory();
    setState(() {
      _messages.add(TutorialGuideChatTurn(role: 'user', content: message));
      _composerController.clear();
      _composerHasText = false;
      _sending = true;
      _errorText = null;
    });
    _scheduleScrollToBottom();

    try {
      final answer = await widget.onSendMessage(message, history);
      if (!mounted) return;
      setState(() {
        _messages.add(
          TutorialGuideChatTurn(role: 'assistant', content: answer),
        );
      });
      _scheduleScrollToBottom();
    } catch (error) {
      if (!mounted) return;
      final rawError = error.toString().trim();
      final friendlyError = rawError.startsWith('Exception: ')
          ? rawError.substring('Exception: '.length).trim()
          : rawError;
      setState(() {
        _errorText = friendlyError.isEmpty
            ? 'The guide could not answer right now.'
            : friendlyError;
      });
    } finally {
      if (mounted) {
        setState(() => _sending = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;
    final imageUrl = _portraitUrl();
    final canSendMessage = _composerHasText && !_sending;

    return Material(
      color: Colors.black54,
      child: SafeArea(
        child: Center(
          child: LayoutBuilder(
            builder: (context, constraints) {
              final maxWidth = math.min(760.0, constraints.maxWidth - 32);
              final maxHeight = math.min(620.0, constraints.maxHeight - 32);

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
                        _ChatPortrait(imageUrl: imageUrl),
                        const SizedBox(width: 16),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                widget.character.name,
                                style: GoogleFonts.cinzel(
                                  textStyle: theme.textTheme.titleLarge,
                                  fontWeight: FontWeight.w700,
                                  letterSpacing: 0.6,
                                  color: colorScheme.onSurface,
                                ),
                              ),
                              const SizedBox(height: 6),
                              Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 8,
                                  vertical: 4,
                                ),
                                decoration: BoxDecoration(
                                  color: const Color(
                                    0xFFD4AF37,
                                  ).withValues(alpha: 0.12),
                                  borderRadius: BorderRadius.circular(999),
                                ),
                                child: Text(
                                  'GUIDE SUPPORT',
                                  style: theme.textTheme.labelSmall?.copyWith(
                                    letterSpacing: 0.8,
                                    fontWeight: FontWeight.w800,
                                    color: const Color(0xFF8C5A14),
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ),
                        IconButton(
                          onPressed: widget.onClose,
                          icon: const Icon(Icons.close),
                          style: IconButton.styleFrom(
                            backgroundColor:
                                colorScheme.surfaceContainerHighest,
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
                              'Conversation',
                              style: theme.textTheme.labelLarge?.copyWith(
                                letterSpacing: 0.8,
                                fontWeight: FontWeight.w700,
                                color: colorScheme.onSurfaceVariant,
                              ),
                            ),
                            const SizedBox(height: 10),
                            Expanded(
                              child: ListView.separated(
                                controller: _scrollController,
                                itemCount:
                                    _messages.length + (_sending ? 1 : 0),
                                separatorBuilder: (_, _) =>
                                    const SizedBox(height: 10),
                                itemBuilder: (context, index) {
                                  if (_sending && index == _messages.length) {
                                    return const _TypingBubble();
                                  }
                                  final turn = _messages[index];
                                  final isAssistant = turn.role == 'assistant';
                                  return _ChatBubble(
                                    isAssistant: isAssistant,
                                    speaker: isAssistant
                                        ? widget.character.name
                                        : 'You',
                                    message: turn.content,
                                  );
                                },
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                    const SizedBox(height: 14),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        AnimatedContainer(
                          duration: const Duration(milliseconds: 180),
                          curve: Curves.easeOutCubic,
                          padding: const EdgeInsets.fromLTRB(8, 8, 8, 8),
                          decoration: BoxDecoration(
                            color: colorScheme.surface,
                            borderRadius: BorderRadius.circular(24),
                            border: Border.all(
                              color: _errorText == null
                                  ? colorScheme.outlineVariant
                                  : colorScheme.error.withValues(alpha: 0.7),
                            ),
                            boxShadow: [
                              BoxShadow(
                                color: Colors.black.withValues(alpha: 0.08),
                                blurRadius: 12,
                                offset: const Offset(0, 4),
                              ),
                            ],
                          ),
                          child: Row(
                            crossAxisAlignment: CrossAxisAlignment.end,
                            children: [
                              Expanded(
                                child: Padding(
                                  padding: const EdgeInsets.fromLTRB(
                                    14,
                                    10,
                                    10,
                                    10,
                                  ),
                                  child: TextField(
                                    controller: _composerController,
                                    enabled: !_sending,
                                    minLines: 1,
                                    maxLines: 4,
                                    textCapitalization:
                                        TextCapitalization.sentences,
                                    textInputAction: TextInputAction.send,
                                    style: theme.textTheme.bodyMedium?.copyWith(
                                      height: 1.35,
                                      color: colorScheme.onSurface,
                                    ),
                                    decoration: InputDecoration.collapsed(
                                      hintText: 'Message',
                                      hintStyle: theme.textTheme.bodyMedium
                                          ?.copyWith(
                                            color: colorScheme.onSurfaceVariant
                                                .withValues(alpha: 0.72),
                                          ),
                                    ),
                                    onSubmitted: (_) => _sendCurrentMessage(),
                                  ),
                                ),
                              ),
                              Container(
                                margin: const EdgeInsets.only(left: 6),
                                decoration: BoxDecoration(
                                  color: canSendMessage
                                      ? colorScheme.primary
                                      : colorScheme.surfaceContainerHighest,
                                  shape: BoxShape.circle,
                                  boxShadow: [
                                    BoxShadow(
                                      color:
                                          (canSendMessage
                                                  ? colorScheme.primary
                                                  : colorScheme.outlineVariant)
                                              .withValues(alpha: 0.18),
                                      blurRadius: 10,
                                      offset: const Offset(0, 4),
                                    ),
                                  ],
                                ),
                                child: IconButton(
                                  onPressed: canSendMessage
                                      ? _sendCurrentMessage
                                      : null,
                                  tooltip: 'Send message',
                                  style: IconButton.styleFrom(
                                    backgroundColor: canSendMessage
                                        ? colorScheme.primary
                                        : colorScheme.surfaceContainerHighest,
                                    foregroundColor: canSendMessage
                                        ? colorScheme.onPrimary
                                        : colorScheme.onSurfaceVariant,
                                    disabledBackgroundColor:
                                        colorScheme.surfaceContainerHighest,
                                    disabledForegroundColor:
                                        colorScheme.onSurfaceVariant,
                                    minimumSize: const Size(46, 46),
                                  ),
                                  icon: _sending
                                      ? SizedBox(
                                          width: 18,
                                          height: 18,
                                          child: CircularProgressIndicator(
                                            strokeWidth: 2.2,
                                            color: colorScheme.onPrimary,
                                          ),
                                        )
                                      : const Icon(
                                          Icons.arrow_upward_rounded,
                                          size: 20,
                                        ),
                                ),
                              ),
                            ],
                          ),
                        ),
                        if (_errorText != null) ...[
                          const SizedBox(height: 10),
                          Padding(
                            padding: const EdgeInsets.symmetric(horizontal: 6),
                            child: Text(
                              _errorText!,
                              style: theme.textTheme.bodySmall?.copyWith(
                                color: colorScheme.error,
                                height: 1.35,
                              ),
                            ),
                          ),
                        ],
                      ],
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
}

class _ChatPortrait extends StatelessWidget {
  const _ChatPortrait({required this.imageUrl});

  final String imageUrl;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      width: 96,
      height: 96,
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: theme.dividerColor),
        boxShadow: const [
          BoxShadow(
            color: Colors.black26,
            blurRadius: 12,
            offset: Offset(0, 6),
          ),
        ],
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(16),
        child: imageUrl.isNotEmpty
            ? Image.network(
                imageUrl,
                width: 96,
                height: 96,
                fit: BoxFit.cover,
                errorBuilder: (_, _, _) => const Icon(Icons.person, size: 48),
              )
            : const Icon(Icons.person, size: 48),
      ),
    );
  }
}

class _ChatBubble extends StatelessWidget {
  const _ChatBubble({
    required this.isAssistant,
    required this.speaker,
    required this.message,
  });

  final bool isAssistant;
  final String speaker;
  final String message;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;
    final bubbleColor = isAssistant
        ? const Color(0xFFF7EBD1)
        : colorScheme.primaryContainer;
    final borderColor = isAssistant
        ? const Color(0xFFD9C89C)
        : colorScheme.primary.withValues(alpha: 0.28);
    final textColor = isAssistant
        ? const Color(0xFF4F3B1D)
        : colorScheme.onPrimaryContainer;

    return Align(
      alignment: isAssistant ? Alignment.centerLeft : Alignment.centerRight,
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 520),
        child: Container(
          padding: const EdgeInsets.fromLTRB(14, 12, 14, 12),
          decoration: BoxDecoration(
            color: bubbleColor,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: borderColor),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                speaker,
                style: theme.textTheme.labelMedium?.copyWith(
                  fontWeight: FontWeight.w700,
                  color: textColor.withValues(alpha: 0.86),
                ),
              ),
              const SizedBox(height: 6),
              Text(
                message,
                style: theme.textTheme.bodyMedium?.copyWith(
                  color: textColor,
                  height: 1.4,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _TypingBubble extends StatelessWidget {
  const _TypingBubble();

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Align(
      alignment: Alignment.centerLeft,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
        decoration: BoxDecoration(
          color: const Color(0xFFF7EBD1),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: const Color(0xFFD9C89C)),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            SizedBox(
              width: 16,
              height: 16,
              child: CircularProgressIndicator(
                strokeWidth: 2.2,
                color: theme.colorScheme.primary,
              ),
            ),
            const SizedBox(width: 10),
            Text(
              'Thinking...',
              style: theme.textTheme.bodyMedium?.copyWith(
                color: const Color(0xFF4F3B1D),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
