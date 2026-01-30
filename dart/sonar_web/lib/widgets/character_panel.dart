import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/character.dart';
import '../models/character_action.dart';
import '../providers/auth_provider.dart';
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
  final void Function(Character, CharacterAction)? onStartDialogue;
  final void Function(Character, CharacterAction)? onStartShop;

  @override
  State<CharacterPanel> createState() => _CharacterPanelState();
}

class _CharacterPanelState extends State<CharacterPanel> {
  List<CharacterAction> _actions = [];
  bool _loading = true;
  bool _acceptingQuest = false;

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
    if (mounted) setState(() => _loading = false);
  }

  Future<void> _handleAction(CharacterAction action) async {
    if (action.actionType == 'shop' && widget.onStartShop != null) {
      widget.onStartShop!(widget.character, action);
      widget.onClose();
    } else if (action.actionType == 'talk' && widget.onStartDialogue != null) {
      widget.onStartDialogue!(widget.character, action);
      widget.onClose();
    } else if (action.actionType == 'giveQuest') {
      final questId = action.pointOfInterestGroupId;
      if (questId == null) return;
      setState(() => _acceptingQuest = true);
      try {
        await context.read<PoiService>().acceptQuest(
              characterId: widget.character.id,
              pointOfInterestGroupId: questId,
            );
        if (mounted) widget.onClose();
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
  }

  @override
  Widget build(BuildContext context) {
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
                          backgroundImage: widget.character.dialogueImageUrl != null
                              ? NetworkImage(widget.character.dialogueImageUrl!)
                              : widget.character.mapIconUrl != null
                                  ? NetworkImage(widget.character.mapIconUrl!)
                                  : null,
                          child: widget.character.dialogueImageUrl == null &&
                                  widget.character.mapIconUrl == null
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
                      : ListView.builder(
                          controller: scrollController,
                          itemCount: _actions.length,
                          itemBuilder: (_, i) {
                            final a = _actions[i];
                            String label;
                            if (a.actionType == 'talk') {
                              label = 'Talk';
                            } else if (a.actionType == 'shop') {
                              final count = a.shopInventory?.length ?? 0;
                              label = 'Shop${count > 0 ? ' ($count items)' : ''}';
                            } else if (a.actionType == 'giveQuest') {
                              label = 'Give Quest';
                            } else {
                              label = a.actionType;
                            }
                            return ListTile(
                              title: Text(label),
                              trailing: a.actionType == 'giveQuest' && _acceptingQuest
                                  ? const SizedBox(
                                      width: 20,
                                      height: 20,
                                      child: CircularProgressIndicator(strokeWidth: 2),
                                    )
                                  : const Icon(Icons.arrow_forward),
                              onTap: () => _handleAction(a),
                            );
                          },
                        ),
            ),
          ],
        ),
      ),
    );
  }
}
