import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/fetch_quest_turn_in.dart';
import '../models/inventory_item.dart';
import '../providers/quest_log_provider.dart';
import '../widgets/inventory_requirement_chip.dart';
import '../widgets/paper_texture.dart';

class FetchQuestTurnInScreen extends StatefulWidget {
  const FetchQuestTurnInScreen({super.key, required this.questId});

  final String questId;

  @override
  State<FetchQuestTurnInScreen> createState() => _FetchQuestTurnInScreenState();
}

class _FetchQuestTurnInScreenState extends State<FetchQuestTurnInScreen> {
  FetchQuestTurnInDetails? _details;
  bool _loading = true;
  bool _submitting = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    if (!mounted) return;
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final details = await context
          .read<QuestLogProvider>()
          .getFetchQuestTurnIn(widget.questId);
      if (!mounted) return;
      setState(() {
        _details = details;
        _loading = false;
      });
    } catch (error) {
      if (!mounted) return;
      setState(() {
        _loading = false;
        _error = _errorMessage(error);
      });
    }
  }

  String _errorMessage(Object error) {
    if (error is DioException && error.response?.data is Map) {
      final data = error.response!.data as Map<String, dynamic>;
      final message = data['error'] ?? data['message'];
      if (message != null && message.toString().trim().isNotEmpty) {
        return message.toString().trim();
      }
    }
    return error.toString();
  }

  InventoryItem _fallbackItem(FetchQuestTurnInRequirement requirement) {
    return InventoryItem(
      id: requirement.inventoryItemId,
      name: 'Item ${requirement.inventoryItemId}',
      imageUrl: '',
      flavorText: '',
      effectText: '',
    );
  }

  Future<void> _submit() async {
    final details = _details;
    if (_submitting || details == null || !details.canDeliver) return;

    setState(() => _submitting = true);
    try {
      final questLog = context.read<QuestLogProvider>();
      final response = await questLog.submitFetchQuestTurnIn(widget.questId);
      await questLog.refresh();
      if (!mounted) return;
      final questCompleted = response['questCompleted'] == true;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            questCompleted
                ? 'Items delivered. Return to the quest giver.'
                : 'Items delivered.',
          ),
        ),
      );
      Navigator.of(context).pop(true);
    } catch (error) {
      if (!mounted) return;
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(_errorMessage(error))));
      setState(() => _submitting = false);
      return;
    }

    if (mounted) {
      setState(() => _submitting = false);
    }
  }

  Widget _buildHeaderCard(
    BuildContext context,
    FetchQuestTurnInDetails details,
  ) {
    final theme = Theme.of(context);
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: theme.colorScheme.outlineVariant),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            details.questName.trim().isNotEmpty
                ? details.questName.trim()
                : 'Deliver Items',
            style: theme.textTheme.headlineSmall?.copyWith(
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'Bring these items to ${details.characterName}.',
            style: theme.textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.w600,
            ),
          ),
          if (details.questDescription.trim().isNotEmpty) ...[
            const SizedBox(height: 10),
            Text(
              details.questDescription.trim(),
              style: theme.textTheme.bodyMedium?.copyWith(
                color: theme.colorScheme.onSurface.withValues(alpha: 0.78),
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildRequirementsCard(
    BuildContext context,
    FetchQuestTurnInDetails details,
  ) {
    final theme = Theme.of(context);
    final allReady = details.canDeliver;
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: theme.colorScheme.outlineVariant),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Required Items',
            style: theme.textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            allReady
                ? 'You have everything this quest step needs.'
                : 'You still need some of the required items before this character can accept them.',
            style: theme.textTheme.bodyMedium?.copyWith(
              color: allReady
                  ? theme.colorScheme.primary
                  : theme.colorScheme.onSurface.withValues(alpha: 0.72),
              fontWeight: allReady ? FontWeight.w700 : FontWeight.w500,
            ),
          ),
          const SizedBox(height: 14),
          Wrap(
            spacing: 10,
            runSpacing: 10,
            children: details.requirements
                .map(
                  (requirement) => InventoryRequirementChip(
                    item:
                        requirement.inventoryItem ?? _fallbackItem(requirement),
                    quantity: requirement.quantity,
                    ownedQuantity: requirement.ownedQuantity,
                  ),
                )
                .toList(),
          ),
        ],
      ),
    );
  }

  Widget _buildLoadedState(
    BuildContext context,
    FetchQuestTurnInDetails details,
  ) {
    return ListView(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 28),
      children: [
        _buildHeaderCard(context, details),
        const SizedBox(height: 14),
        _buildRequirementsCard(context, details),
        const SizedBox(height: 18),
        FilledButton(
          onPressed: details.canDeliver && !_submitting ? _submit : null,
          style: FilledButton.styleFrom(minimumSize: const Size.fromHeight(52)),
          child: Text(
            _submitting
                ? 'Delivering...'
                : details.canDeliver
                ? 'Deliver Items'
                : 'Not Enough Items',
          ),
        ),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final title = _details?.questName.trim();
    return Scaffold(
      backgroundColor: theme.colorScheme.surface,
      appBar: AppBar(
        centerTitle: true,
        title: Text(
          title != null && title.isNotEmpty ? title : 'Deliver Items',
          overflow: TextOverflow.ellipsis,
        ),
      ),
      body: PaperSheet(
        child: _loading
            ? const Center(child: CircularProgressIndicator())
            : _error != null
            ? Center(
                child: Padding(
                  padding: const EdgeInsets.all(24),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(
                        _error!,
                        textAlign: TextAlign.center,
                        style: theme.textTheme.bodyLarge,
                      ),
                      const SizedBox(height: 16),
                      FilledButton(
                        onPressed: _load,
                        child: const Text('Retry'),
                      ),
                    ],
                  ),
                ),
              )
            : _details == null
            ? const SizedBox.shrink()
            : _buildLoadedState(context, _details!),
      ),
    );
  }
}
