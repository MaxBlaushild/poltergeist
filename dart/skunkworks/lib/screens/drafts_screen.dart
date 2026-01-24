import 'dart:io';
import 'package:flutter/material.dart';
import 'package:skunkworks/constants/app_colors.dart';
import 'package:skunkworks/models/draft.dart';
import 'package:skunkworks/services/draft_service.dart';

class DraftsScreen extends StatefulWidget {
  const DraftsScreen({super.key});

  @override
  State<DraftsScreen> createState() => _DraftsScreenState();
}

class _DraftsScreenState extends State<DraftsScreen> {
  final DraftService _draftService = DraftService();
  Future<List<Draft>>? _draftsFuture;

  @override
  void initState() {
    super.initState();
    _draftsFuture = _draftService.getDrafts();
  }

  Future<void> _refreshDrafts() async {
    setState(() {
      _draftsFuture = _draftService.getDrafts();
    });
  }

  Future<void> _deleteDraft(Draft draft) async {
    await _draftService.deleteDraft(draft.id);
    await _refreshDrafts();
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Draft deleted')),
      );
    }
  }

  String _formatDate(DateTime d) {
    final now = DateTime.now();
    final diff = now.difference(d);
    if (diff.inDays > 0) return '${diff.inDays}d ago';
    if (diff.inHours > 0) return '${diff.inHours}h ago';
    if (diff.inMinutes > 0) return '${diff.inMinutes}m ago';
    return 'Just now';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.warmWhite,
      appBar: AppBar(
        backgroundColor: AppColors.warmWhite,
        elevation: 0,
        leading: IconButton(
          icon: Icon(Icons.arrow_back, color: AppColors.graphiteInk),
          onPressed: () => Navigator.of(context).pop(),
        ),
        title: const Text(
          'Drafts',
          style: TextStyle(
            color: AppColors.graphiteInk,
            fontWeight: FontWeight.w600,
            fontSize: 18,
          ),
        ),
      ),
      body: FutureBuilder<List<Draft>>(
        future: _draftsFuture,
        builder: (context, snapshot) {
          if (snapshot.connectionState == ConnectionState.waiting) {
            return const Center(child: CircularProgressIndicator());
          }
          if (snapshot.hasError) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(
                    'Error loading drafts: ${snapshot.error}',
                    style: const TextStyle(color: AppColors.coralPop),
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: _refreshDrafts,
                    child: const Text('Retry'),
                  ),
                ],
              ),
            );
          }
          final drafts = snapshot.data ?? [];
          if (drafts.isEmpty) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    Icons.drafts_outlined,
                    size: 64,
                    color: Colors.grey.shade400,
                  ),
                  const SizedBox(height: 16),
                  Text(
                    'No drafts',
                    style: TextStyle(
                      color: Colors.grey.shade600,
                      fontSize: 16,
                    ),
                  ),
                ],
              ),
            );
          }
          return RefreshIndicator(
            onRefresh: _refreshDrafts,
            child: ListView.builder(
              padding: const EdgeInsets.symmetric(vertical: 8, horizontal: 16),
              itemCount: drafts.length,
              itemBuilder: (context, index) {
                final draft = drafts[index];
                return Dismissible(
                  key: Key(draft.id),
                  direction: DismissDirection.endToStart,
                  background: Container(
                    color: AppColors.coralPop,
                    alignment: Alignment.centerRight,
                    padding: const EdgeInsets.only(right: 20),
                    child: const Icon(Icons.delete, color: Colors.white, size: 28),
                  ),
                  confirmDismiss: (direction) async {
                    return await showDialog<bool>(
                      context: context,
                      builder: (ctx) => AlertDialog(
                        title: const Text('Delete draft?'),
                        content: const Text(
                          'This draft will be permanently deleted.',
                        ),
                        actions: [
                          TextButton(
                            onPressed: () => Navigator.pop(ctx, false),
                            child: const Text('Cancel'),
                          ),
                          TextButton(
                            onPressed: () => Navigator.pop(ctx, true),
                            style: TextButton.styleFrom(
                              foregroundColor: AppColors.coralPop,
                            ),
                            child: const Text('Delete'),
                          ),
                        ],
                      ),
                    );
                  },
                  onDismissed: (_) => _deleteDraft(draft),
                  child: Card(
                    margin: const EdgeInsets.only(bottom: 12),
                    clipBehavior: Clip.antiAlias,
                    child: InkWell(
                      onTap: () => Navigator.of(context).pop(draft),
                      child: Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          SizedBox(
                            width: 80,
                            height: 80,
                            child: Image.file(
                              File(draft.imagePath),
                              fit: BoxFit.cover,
                              errorBuilder: (_, __, ___) => Container(
                                color: Colors.grey.shade300,
                                child: Icon(
                                  Icons.broken_image,
                                  color: Colors.grey.shade600,
                                ),
                              ),
                            ),
                          ),
                          Expanded(
                            child: Padding(
                              padding: const EdgeInsets.all(12),
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    draft.caption?.isNotEmpty == true
                                        ? (draft.caption!.length > 60
                                            ? '${draft.caption!.substring(0, 60)}...'
                                            : draft.caption!)
                                        : 'No caption',
                                    style: TextStyle(
                                      color: AppColors.graphiteInk,
                                      fontSize: 14,
                                    ),
                                    maxLines: 2,
                                    overflow: TextOverflow.ellipsis,
                                  ),
                                  const SizedBox(height: 4),
                                  Text(
                                    _formatDate(draft.createdAt),
                                    style: TextStyle(
                                      color: Colors.grey.shade600,
                                      fontSize: 12,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                );
              },
            ),
          );
        },
      ),
    );
  }
}
