import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:travel_angels/models/document.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/library_document_content.dart';

class LibraryDocumentCard extends StatelessWidget {
  const LibraryDocumentCard({
    super.key,
    required this.document,
    required this.onOpen,
    required this.onSave,
    this.isOwner = false,
    this.saving = false,
  });
  final Document document;
  final VoidCallback onOpen;
  final VoidCallback? onSave;
  final bool isOwner;
  final bool saving;

  @override
  Widget build(BuildContext context) {
    final format = libraryFormat(document);
    final destination = libraryDestination(document);
    final author = libraryAuthor(document);
    final content = document.content?.trim() ?? '';
    final note = document.personalNote.trim();
    return Material(
      color: AppTheme.paper,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(3),
        side: const BorderSide(color: AppTheme.outline),
      ),
      clipBehavior: Clip.antiAlias,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          InkWell(
            onTap: onOpen,
            child: Semantics(
              label: 'Preview ${document.title}',
              button: true,
              child: ExcludeSemantics(
                child: Container(
                  color: libraryFormatColor(format),
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      Wrap(
                        spacing: 12,
                        runSpacing: 4,
                        alignment: WrapAlignment.spaceBetween,
                        children: [
                          Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Icon(libraryFormatIcon(format), size: 16),
                              const SizedBox(width: 6),
                              Text(
                                libraryFormatLabel(format),
                                style: Theme.of(context).textTheme.labelSmall,
                              ),
                            ],
                          ),
                          Text(
                            librarySource(document),
                            style: Theme.of(context).textTheme.labelSmall,
                          ),
                        ],
                      ),
                      const SizedBox(height: 14),
                      if (format == 'video')
                        const Padding(
                          padding: EdgeInsets.all(20),
                          child: Icon(
                            Icons.play_circle_outline,
                            size: 44,
                            color: AppTheme.ink,
                          ),
                        )
                      else if (content.isNotEmpty)
                        LibraryDocumentContent(
                          content: content,
                          format: format,
                          preview: true,
                        )
                      else
                        Padding(
                          padding: const EdgeInsets.symmetric(vertical: 20),
                          child: Text(
                            'Open this guide to read more.',
                            style: Theme.of(context).textTheme.bodySmall,
                          ),
                        ),
                    ],
                  ),
                ),
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                if (destination.isNotEmpty)
                  Text(
                    destination,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: Theme.of(context).textTheme.labelSmall?.copyWith(
                      color: AppTheme.rust,
                      letterSpacing: .8,
                    ),
                  ),
                const SizedBox(height: 6),
                InkWell(
                  onTap: onOpen,
                  child: Padding(
                    padding: const EdgeInsets.symmetric(vertical: 4),
                    child: Text(
                      document.title,
                      style: Theme.of(
                        context,
                      ).textTheme.headlineSmall?.copyWith(fontSize: 24),
                    ),
                  ),
                ),
                if (note.isNotEmpty)
                  Padding(
                    padding: const EdgeInsets.only(top: 8),
                    child: Text(
                      note,
                      maxLines: 3,
                      overflow: TextOverflow.ellipsis,
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        height: 1.6,
                        color: AppTheme.muted,
                      ),
                    ),
                  ),
                const SizedBox(height: 14),
                Row(
                  children: [
                    Expanded(
                      child: Row(
                        children: [
                          CircleAvatar(
                            radius: 17,
                            backgroundColor: AppTheme.peach,
                            foregroundColor: AppTheme.ink,
                            child: Text(
                              author.characters.first.toUpperCase(),
                              style: Theme.of(context).textTheme.bodyMedium,
                            ),
                          ),
                          const SizedBox(width: 9),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  isOwner ? 'You' : author,
                                  style: Theme.of(context).textTheme.bodySmall,
                                ),
                                if (document.updatedAt != null)
                                  Text(
                                    DateFormat.yMMMd().format(
                                      document.updatedAt!.toLocal(),
                                    ),
                                    style: Theme.of(
                                      context,
                                    ).textTheme.labelSmall,
                                  ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
                    IconButton(
                      key: ValueKey('library-save-${document.id}'),
                      tooltip: document.isSaved
                          ? 'Unsave ${document.title}'
                          : 'Save ${document.title}',
                      onPressed: saving ? null : onSave,
                      isSelected: document.isSaved,
                      icon: const Icon(Icons.bookmark_border),
                      selectedIcon: const Icon(Icons.bookmark),
                    ),
                  ],
                ),
                if (isOwner)
                  Padding(
                    padding: const EdgeInsets.only(top: 12),
                    child: Row(
                      children: [
                        Icon(
                          document.visibility == 'private'
                              ? Icons.lock_outline
                              : Icons.people_outline,
                          size: 16,
                          color: AppTheme.muted,
                        ),
                        const SizedBox(width: 6),
                        Expanded(
                          child: Text(
                            libraryAudience(document.visibility),
                            style: Theme.of(context).textTheme.labelSmall,
                          ),
                        ),
                      ],
                    ),
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
