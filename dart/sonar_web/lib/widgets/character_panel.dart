import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/character.dart';
import '../models/character_action.dart';
import '../models/quest.dart';
import '../providers/completed_task_provider.dart';
import '../providers/quest_log_provider.dart';
import '../services/poi_service.dart';

class CharacterPanel extends StatefulWidget {
  const CharacterPanel({
    super.key,
    required this.character,
    required this.onClose,
    this.onStartDialogue,
    this.onStartShop,
  });

  final Character character;
  final VoidCallback onClose;
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

  Future<void> _handleQuest(CharacterAction action) async {
    final questId = action.questId;
    if (questId == null) return;
    setState(() => _acceptingQuest = true);
    try {
      await context.read<PoiService>().acceptQuest(
            characterId: widget.character.id,
            questId: questId,
          );
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Quest accepted')),
        );
        widget.onClose();
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
    final questAction = _firstActionOfType('giveQuest');
    final hasQuest = questAction?.questId != null;
    final questReadyToTurnIn = questAction != null ? _questReadyToTurnIn(questAction) : null;
    final imageUrl = widget.character.dialogueImageUrl ?? widget.character.mapIconUrl;

    return DraggableScrollableSheet(
      initialChildSize: 0.6,
      minChildSize: 0.3,
      maxChildSize: 0.9,
      builder: (_, scrollController) => Container(
        decoration: BoxDecoration(
          color: Theme.of(context).colorScheme.surface,
          borderRadius: const BorderRadius.vertical(top: Radius.circular(16)),
        ),
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
                                  if (questReadyToTurnIn != null)
                                    _DialogueChoiceButton(
                                      label: _turningInQuest ? 'Turning in…' : 'Turn in',
                                      icon: Icons.assignment_turned_in,
                                      onTap: _turningInQuest
                                          ? null
                                          : () => _handleTurnIn(questReadyToTurnIn, questAction!),
                                    )
                                  else if (hasQuest)
                                    _DialogueChoiceButton(
                                      label: _acceptingQuest ? 'Accepting quest…' : 'Quest',
                                      icon: Icons.assignment_turned_in,
                                      onTap: _acceptingQuest
                                          ? null
                                          : () => _handleQuest(questAction!),
                                    ),
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
