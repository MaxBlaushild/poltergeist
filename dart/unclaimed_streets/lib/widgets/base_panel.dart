import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';

import '../models/base.dart';
import '../models/base_progression.dart';
import '../providers/auth_provider.dart';
import '../services/base_service.dart';
import 'paper_texture.dart';

class BasePanel extends StatefulWidget {
  const BasePanel({
    super.key,
    required this.base,
    required this.onClose,
    this.onEnterBase,
  });

  final BasePin base;
  final VoidCallback onClose;
  final VoidCallback? onEnterBase;

  @override
  State<BasePanel> createState() => _BasePanelState();
}

class _BasePanelState extends State<BasePanel> {
  BaseProgressionSnapshot? _snapshot;
  bool _loading = true;
  String? _error;

  bool get _isOwner {
    final userId = context.read<AuthProvider>().user?.id ?? '';
    return userId.isNotEmpty && userId == widget.base.userId;
  }

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    if (!_isOwner) {
      if (!mounted) return;
      setState(() {
        _loading = false;
        _error = null;
      });
      return;
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final snapshot = await context.read<BaseService>().getBaseById(
        widget.base.id,
      );
      if (!mounted) return;
      setState(() {
        _snapshot = snapshot;
        _loading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _loading = false;
        _error = e.toString();
      });
    }
  }

  String _friendlyResourceName(String key) {
    switch (key) {
      case 'arcane_dust':
        return 'Arcane Dust';
      case 'monster_parts':
        return 'Monster Parts';
      case 'relic_shards':
        return 'Relic Shards';
      default:
        final text = key.replaceAll('_', ' ');
        if (text.isEmpty) return key;
        return text
            .split(' ')
            .map((part) {
              if (part.isEmpty) return part;
              return '${part[0].toUpperCase()}${part.substring(1)}';
            })
            .join(' ');
    }
  }

  String get _baseTitle {
    final owner = widget.base.owner;
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

  Widget _buildOwnerSection(BuildContext context) {
    final theme = Theme.of(context);
    if (_loading) {
      return const Padding(
        padding: EdgeInsets.symmetric(vertical: 24),
        child: Center(child: CircularProgressIndicator()),
      );
    }
    if (_error != null && _snapshot == null) {
      return Padding(
        padding: const EdgeInsets.only(top: 18),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'We could not load your base details.',
              style: theme.textTheme.titleMedium?.copyWith(
                fontWeight: FontWeight.w700,
              ),
            ),
            const SizedBox(height: 8),
            Text(_error!, style: theme.textTheme.bodyMedium),
            const SizedBox(height: 12),
            OutlinedButton(onPressed: _load, child: const Text('Try again')),
          ],
        ),
      );
    }

    final snapshot = _snapshot;
    final builtCount = snapshot?.structures.length ?? 0;
    final resources = (snapshot?.resources ?? const <BaseResourceBalanceData>[])
        .where((entry) => entry.amount > 0)
        .take(4)
        .toList();

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: 16),
        if (resources.isNotEmpty) ...[
          Text(
            'Materials On Hand',
            style: theme.textTheme.titleSmall?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: resources
                .map(
                  (resource) => Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 8,
                    ),
                    decoration: BoxDecoration(
                      color: theme.colorScheme.surface,
                      borderRadius: BorderRadius.circular(999),
                      border: Border.all(
                        color: theme.colorScheme.outlineVariant,
                      ),
                    ),
                    child: Text(
                      '${_friendlyResourceName(resource.resourceKey)}: ${resource.amount}',
                      style: theme.textTheme.bodySmall?.copyWith(
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ),
                )
                .toList(),
          ),
          const SizedBox(height: 16),
        ],
        Container(
          padding: const EdgeInsets.all(14),
          decoration: BoxDecoration(
            color: theme.colorScheme.surface,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: theme.colorScheme.outlineVariant),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Built rooms: $builtCount',
                style: theme.textTheme.titleSmall?.copyWith(
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: 12),
              SizedBox(
                width: double.infinity,
                child: FilledButton(
                  onPressed: _enterBase,
                  child: const Text('Enter Base'),
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  void _enterBase() {
    widget.onClose();
    final callback = widget.onEnterBase;
    if (callback != null) {
      callback();
    } else {
      context.push('/base-management/${widget.base.id}');
    }
  }

  Widget _buildFriendEnterSection(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(top: 16),
      child: SizedBox(
        width: double.infinity,
        child: FilledButton(
          onPressed: _enterBase,
          child: const Text('Enter Base'),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final maxHeight = MediaQuery.sizeOf(context).height * 0.82;
    final imageUrl = widget.base.imageUrl.trim().isNotEmpty
        ? widget.base.imageUrl.trim()
        : widget.base.thumbnailUrl;
    return PaperSheet(
      child: ConstrainedBox(
        constraints: BoxConstraints(maxHeight: maxHeight),
        child: SingleChildScrollView(
          padding: const EdgeInsets.fromLTRB(16, 16, 16, 24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    _baseTitle,
                    style: theme.textTheme.titleLarge?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  IconButton(
                    onPressed: widget.onClose,
                    icon: const Icon(Icons.close),
                  ),
                ],
              ),
              ClipRRect(
                borderRadius: BorderRadius.circular(14),
                child: SizedBox(
                  height: 220,
                  width: double.infinity,
                  child: Image.network(
                    imageUrl,
                    fit: BoxFit.cover,
                    errorBuilder: (_, _, _) => Container(
                      color: theme.colorScheme.surfaceContainerHighest,
                      child: const Icon(Icons.home_work_outlined, size: 48),
                    ),
                  ),
                ),
              ),
              if (widget.base.description.trim().isNotEmpty) ...[
                const SizedBox(height: 14),
                Text(
                  widget.base.description.trim(),
                  style: theme.textTheme.bodyMedium?.copyWith(height: 1.4),
                ),
              ],
              const SizedBox(height: 16),
              if (_isOwner)
                _buildOwnerSection(context)
              else
                _buildFriendEnterSection(context),
            ],
          ),
        ),
      ),
    );
  }
}
