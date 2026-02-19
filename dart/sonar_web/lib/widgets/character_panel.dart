import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/character.dart';
import '../models/character_action.dart';
import '../models/quest.dart';
import '../providers/auth_provider.dart';
import '../providers/completed_task_provider.dart';
import '../providers/quest_log_provider.dart';
import '../services/poi_service.dart';
import '../widgets/paper_texture.dart';
import 'rpg_dialogue_modal.dart';

class CharacterPanel extends StatefulWidget {
  const CharacterPanel({
    super.key,
    required this.character,
    required this.onClose,
    this.onQuestAccepted,
    this.onStartDialogue,
    this.onStartShop,
  });

  final Character character;
  final VoidCallback onClose;
  final VoidCallback? onQuestAccepted;
  final void Function(BuildContext, Character, CharacterAction)? onStartDialogue;
  final void Function(BuildContext, Character, CharacterAction)? onStartShop;

  @override
  State<CharacterPanel> createState() => _CharacterPanelState();
}

class _CharacterPanelState extends State<CharacterPanel> {
  List<CharacterAction> _actions = [];
  bool _loading = true;
  bool _acceptingQuest = false;
  bool _turningInQuest = false;

  @override
  void initState() {
    super.initState();
    _loadActions();
  }

  Future<void> _loadActions() async {
    setState(() => _loading = true);
    try {
      final svc = context.read<PoiService>();
      _actions = await svc.getCharacterActions(widget.character.id);
    } catch (_) {
      _actions = [];
    }
    debugPrint('CharacterPanel: loaded ${_actions.length} actions for ${widget.character.id}');
    final hasTalkAction = _actions.any((action) => action.actionType == 'talk');
    if (!hasTalkAction) {
      final fallbackTalk = CharacterAction(
        id: 'local-talk-${widget.character.id}',
        createdAt: DateTime.now().toIso8601String(),
        updatedAt: DateTime.now().toIso8601String(),
        characterId: widget.character.id,
        actionType: 'talk',
        dialogue: const [
          DialogueMessage(speaker: 'character', text: '...', order: 0),
        ],
      );
      _actions = [fallbackTalk, ..._actions];
    }
    if (mounted) setState(() => _loading = false);
  }

  CharacterAction? _firstActionOfType(String type) {
    for (final action in _actions) {
      if (action.actionType == type) return action;
    }
    return null;
  }

  CharacterAction? _firstActionOfTypes(List<String> types) {
    for (final action in _actions) {
      if (types.contains(action.actionType)) return action;
    }
    return null;
  }

  Quest? _questForAction(CharacterAction action) {
    final questId = action.questId;
    if (questId == null || questId.isEmpty) return null;
    final quests = context.read<QuestLogProvider>().quests;
    try {
      return quests.firstWhere((q) => q.id == questId);
    } catch (_) {
      return null;
    }
  }

  List<DialogueMessage> _buildQuestAcceptanceDialogue(Quest? quest, CharacterAction action) {
    final questLines = (quest?.acceptanceDialogue ?? const [])
        .map((line) => line.trim())
        .where((line) => line.isNotEmpty)
        .toList();
    if (questLines.isNotEmpty) {
      return [
        for (var i = 0; i < questLines.length; i++)
          DialogueMessage(speaker: 'character', text: questLines[i], order: i),
      ];
    }

    if (action.dialogue.isNotEmpty) {
      return action.dialogue;
    }

    final actionLines = action.questAcceptanceDialogue
        .map((line) => line.trim())
        .where((line) => line.isNotEmpty)
        .toList();
    if (actionLines.isNotEmpty) {
      return [
        for (var i = 0; i < actionLines.length; i++)
          DialogueMessage(speaker: 'character', text: actionLines[i], order: i),
      ];
    }

    final fallback = quest?.description.trim() ??
        action.questDescription?.trim() ??
        '';
    if (fallback.isNotEmpty) {
      return [DialogueMessage(speaker: 'character', text: fallback, order: 0)];
    }

    return const [DialogueMessage(speaker: 'character', text: '...', order: 0)];
  }

  Future<void> _showQuestAcceptanceDialog(CharacterAction action) async {
    if (_acceptingQuest) return;
    final quest = _questForAction(action);
    final accepted = await showDialog<bool>(
      context: context,
      useRootNavigator: true,
      barrierDismissible: true,
      builder: (dialogContext) {
        return RpgDialogueModal(
          character: widget.character,
          action: action,
          dialogueOverride: _buildQuestAcceptanceDialogue(quest, action),
          primaryActionLabel: 'Accept quest',
          secondaryActionLabel: 'Decline',
          onPrimaryAction: () => Navigator.of(dialogContext).pop(true),
          onSecondaryAction: () => Navigator.of(dialogContext).pop(false),
          onClose: () => Navigator.of(dialogContext).pop(false),
        );
      },
    );
    if (accepted == true) {
      await _handleQuest(action);
    }
  }

  Future<void> _handleQuest(CharacterAction action) async {
    final questId = action.questId;
    if (questId == null) return;
    setState(() => _acceptingQuest = true);
    try {
      await context.read<PoiService>().acceptQuest(
            characterId: widget.character.id,
            questId: questId,
          );
      await context.read<QuestLogProvider>().refresh();
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Quest accepted')),
        );
        widget.onClose();
        widget.onQuestAccepted?.call();
      }
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Failed to accept quest')),
        );
      }
    } finally {
      if (mounted) setState(() => _acceptingQuest = false);
    }
  }

  Future<void> _handleTurnIn(Quest quest, CharacterAction action) async {
    final questId = action.questId ?? quest.id;
    if (questId.isEmpty) return;
    setState(() => _turningInQuest = true);
    try {
      final resp = await context.read<QuestLogProvider>().turnInQuest(questId);
      if (mounted) {
        try {
          await context.read<AuthProvider>().refresh();
        } catch (_) {}
        context.read<CompletedTaskProvider>().showModal('questCompleted', data: {
          'questName': quest.name,
          ...resp,
        });
        widget.onClose();
      }
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Failed to turn in quest')),
        );
      }
    } finally {
      if (mounted) setState(() => _turningInQuest = false);
    }
  }

  Quest? _questReadyToTurnIn(CharacterAction action) {
    final questId = action.questId;
    if (questId == null || questId.isEmpty) return null;
    final quests = context.read<QuestLogProvider>().quests;
    try {
      return quests.firstWhere(
        (q) => q.id == questId && q.readyToTurnIn,
      );
    } catch (_) {
      return null;
    }
  }

  @override
  Widget build(BuildContext context) {
    final talkAction = _firstActionOfType('talk');
    final shopAction = _firstActionOfType('shop');
    final questActions = _actions
        .where((action) => ['giveQuest', 'quest', 'quests'].contains(action.actionType))
        .where((action) => action.questId != null && action.questId!.isNotEmpty)
        .toList();
    final imageUrl = widget.character.dialogueImageUrl ?? widget.character.mapIconUrl;

    return DraggableScrollableSheet(
      initialChildSize: 0.9,
      minChildSize: 0.4,
      maxChildSize: 0.95,
      builder: (_, scrollController) => PaperSheet(
        child: Column(
          children: [
            Container(
              padding: const EdgeInsets.all(16),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Expanded(
                    child: Row(
                      children: [
                        CircleAvatar(
                          radius: 24,
                          backgroundColor: Colors.grey.shade300,
                          backgroundImage: imageUrl != null
                              ? NetworkImage(imageUrl)
                              : null,
                          child: imageUrl == null
                              ? const Icon(Icons.person)
                              : null,
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                widget.character.name,
                                style: Theme.of(context).textTheme.titleLarge,
                              ),
                              if (widget.character.description != null)
                                Text(
                                  widget.character.description!,
                                  style: Theme.of(context).textTheme.bodySmall,
                                ),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),
                  IconButton(
                    onPressed: widget.onClose,
                    icon: const Icon(Icons.close),
                  ),
                ],
              ),
            ),
            Expanded(
              child: _loading
                  ? const Center(child: CircularProgressIndicator())
                  : _actions.isEmpty
                      ? const Center(child: Text('No actions available'))
                      : ListView(
                          controller: scrollController,
                          padding: const EdgeInsets.fromLTRB(16, 0, 16, 24),
                          children: [
                            Container(
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
                                    '${widget.character.name}:',
                                    style: const TextStyle(
                                      color: Colors.white,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                  const SizedBox(height: 8),
                                  const Text(
                                    'Choose an action:',
                                    style: TextStyle(color: Colors.white70),
                                  ),
                                  const SizedBox(height: 12),
                                  ...questActions.map((action) {
                                    final quest = _questForAction(action);
                                    final questReadyToTurnIn = _questReadyToTurnIn(action);
                                    if (questReadyToTurnIn != null) {
                                      return _DialogueChoiceButton(
                                        label: _turningInQuest
                                            ? 'Turning in…'
                                            : 'Turn in: ${questReadyToTurnIn.name}',
                                        icon: Icons.assignment_turned_in,
                                        onTap: _turningInQuest
                                            ? null
                                            : () => _handleTurnIn(questReadyToTurnIn, action),
                                      );
                                    }
                                    if (quest?.isAccepted == true) {
                                      return const SizedBox.shrink();
                                    }
                                    return _DialogueChoiceButton(
                                      label: _acceptingQuest
                                          ? 'Accepting quest…'
                                          : 'Accept: ${quest?.name ?? action.questName ?? 'Quest'}',
                                      icon: Icons.assignment_turned_in,
                                      onTap: _acceptingQuest
                                          ? null
                                          : () => _showQuestAcceptanceDialog(action),
                                    );
                                  }),
                                  if (shopAction != null)
                                    _DialogueChoiceButton(
                                      label: 'Shop',
                                      icon: Icons.storefront,
                                      onTap: widget.onStartShop == null
                                          ? null
                                          : () {
                                              widget.onStartShop!(context, widget.character, shopAction);
                                              widget.onClose();
                                            },
                                    ),
                                  if (talkAction != null)
                                    _DialogueChoiceButton(
                                      label: 'Dialogue',
                                      icon: Icons.chat_bubble_outline,
                                      onTap: widget.onStartDialogue == null
                                          ? null
                                          : () {
                                              widget.onStartDialogue!(context, widget.character, talkAction);
                                            },
                                    ),
                                ],
                              ),
                            ),
                            if (widget.character.description != null &&
                                widget.character.description!.isNotEmpty) ...[
                              const SizedBox(height: 16),
                              Text(
                                widget.character.description!,
                                style: Theme.of(context).textTheme.bodyMedium,
                              ),
                            ],
                          ],
                        ),
            ),
          ],
        ),
      ),
    );
  }
}

class _DialogueChoiceButton extends StatelessWidget {
  const _DialogueChoiceButton({
    required this.label,
    required this.icon,
    this.onTap,
  });

  final String label;
  final IconData icon;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(8),
        child: Container(
          padding: const EdgeInsets.symmetric(vertical: 10, horizontal: 12),
          decoration: BoxDecoration(
            color: onTap == null ? Colors.white12 : Colors.white10,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: Colors.white30),
          ),
          child: Row(
            children: [
              Icon(icon, color: Colors.white),
              const SizedBox(width: 10),
              Text(
                label,
                style: TextStyle(
                  color: onTap == null ? Colors.white38 : Colors.white,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
