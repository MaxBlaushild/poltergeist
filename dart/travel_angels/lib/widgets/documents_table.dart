import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:travel_angels/models/document.dart';
import 'package:travel_angels/utils/document_utils.dart';

/// Widget for displaying a sortable table of documents
class DocumentsTable extends StatelessWidget {
  final List<Document> documents;
  final int? sortColumnIndex;
  final bool sortAscending;
  final Function(int columnIndex, bool ascending) onSort;
  final Function(Document document)? onDocumentTap;
  final Function(Document document)? onEditTap;
  final Set<String> selectedDocumentIds;
  final Function(String documentId, bool selected)? onSelectionChanged;

  const DocumentsTable({
    super.key,
    required this.documents,
    this.sortColumnIndex,
    required this.sortAscending,
    required this.onSort,
    this.onDocumentTap,
    this.onEditTap,
    this.selectedDocumentIds = const {},
    this.onSelectionChanged,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final dateFormat = DateFormat('MMM d, yyyy h:mm a');

    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: SingleChildScrollView(
        child: ConstrainedBox(
          constraints: const BoxConstraints(minWidth: 800),
          child: DataTable(
            sortColumnIndex: sortColumnIndex,
            sortAscending: sortAscending,
            columns: [
            const DataColumn(
              label: SizedBox.shrink(), // Checkbox column
            ),
            DataColumn(
              label: const Text('Title'),
              onSort: (columnIndex, ascending) => onSort(columnIndex, ascending),
            ),
            DataColumn(
              label: const Text('Provider'),
              onSort: (columnIndex, ascending) => onSort(columnIndex, ascending),
            ),
            DataColumn(
              label: const Text('Date Created'),
              onSort: (columnIndex, ascending) => onSort(columnIndex, ascending),
            ),
            DataColumn(
              label: const Text('Date Edited'),
              onSort: (columnIndex, ascending) => onSort(columnIndex, ascending),
            ),
            DataColumn(
              label: const Text('Actions'),
              tooltip: 'Edit or view document',
            ),
          ],
          rows: documents.map((document) {
            final isSelected = selectedDocumentIds.contains(document.id);
            final isVideo = DocumentUtils.isVideo(document);
            print('[DocumentsTable] Building row for document: id=${document.id}, title="${document.title}", link=${document.link}, isVideo=$isVideo, onEditTap=${onEditTap != null}');
            return DataRow(
              selected: isSelected,
              onSelectChanged: null, // Disable row-level selection
              cells: [
                // Checkbox cell
                DataCell(
                  Checkbox(
                    value: isSelected,
                    onChanged: onSelectionChanged != null
                        ? (bool? value) {
                            if (value != null) {
                              onSelectionChanged!(document.id, value);
                            }
                          }
                        : null,
                  ),
                ),
                // Title cell - clickable to navigate
                DataCell(
                  GestureDetector(
                    onTap: onDocumentTap != null
                        ? () => onDocumentTap!(document)
                        : null,
                    child: Tooltip(
                      message: document.title,
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          if (DocumentUtils.isVideo(document))
                            Padding(
                              padding: const EdgeInsets.only(right: 8),
                              child: Icon(
                                Icons.videocam,
                                size: 18,
                                color: theme.colorScheme.primary,
                              ),
                            ),
                          Flexible(
                            child: Text(
                              document.title,
                              overflow: TextOverflow.ellipsis,
                              style: const TextStyle(fontWeight: FontWeight.w500),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
                // Provider cell - clickable to navigate
                DataCell(
                  GestureDetector(
                    onTap: onDocumentTap != null
                        ? () => onDocumentTap!(document)
                        : null,
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(
                          DocumentUtils.getProviderIcon(document.provider),
                          size: 18,
                          color: theme.colorScheme.primary,
                        ),
                        const SizedBox(width: 8),
                        Text(
                          DocumentUtils.getProviderLabel(document.provider),
                          style: const TextStyle(fontSize: 14),
                        ),
                      ],
                    ),
                  ),
                ),
                // Date Created cell - clickable to navigate
                DataCell(
                  GestureDetector(
                    onTap: onDocumentTap != null
                        ? () => onDocumentTap!(document)
                        : null,
                    child: Text(
                      document.createdAt != null
                          ? dateFormat.format(document.createdAt!)
                          : 'N/A',
                      style: const TextStyle(fontSize: 13),
                    ),
                  ),
                ),
                // Date Edited cell - clickable to navigate
                DataCell(
                  GestureDetector(
                    onTap: onDocumentTap != null
                        ? () => onDocumentTap!(document)
                        : null,
                    child: Text(
                      document.updatedAt != null
                          ? dateFormat.format(document.updatedAt!)
                          : 'N/A',
                      style: const TextStyle(fontSize: 13),
                    ),
                  ),
                ),
                // Actions cell - edit button
                DataCell(
                  Builder(
                    builder: (context) {
                      final isVideoDoc = DocumentUtils.isVideo(document);
                      print('[DocumentsTable] Actions cell - onEditTap=${onEditTap != null}, isVideo=$isVideoDoc');
                      return Container(
                        width: 60,
                        alignment: Alignment.center,
                        child: onEditTap != null
                            ? IconButton(
                                icon: Icon(
                                  isVideoDoc 
                                      ? Icons.video_library 
                                      : Icons.edit,
                                  size: 22,
                                  color: theme.colorScheme.primary,
                                ),
                                onPressed: () {
                                  print('[DocumentsTable] Edit button pressed for document: id=${document.id}, title="${document.title}", isVideo=$isVideoDoc');
                                  onEditTap!(document);
                                },
                                tooltip: isVideoDoc 
                                    ? 'Edit Video' 
                                    : 'Edit Document',
                                padding: const EdgeInsets.all(8),
                                constraints: const BoxConstraints(
                                  minWidth: 48,
                                  minHeight: 48,
                                ),
                              )
                            : const SizedBox(width: 48),
                      );
                    },
                  ),
                ),
              ],
            );
          }).toList(),
          ),
        ),
      ),
    );
  }
}

