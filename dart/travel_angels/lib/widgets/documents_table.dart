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

  const DocumentsTable({
    super.key,
    required this.documents,
    this.sortColumnIndex,
    required this.sortAscending,
    required this.onSort,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final dateFormat = DateFormat('MMM d, yyyy h:mm a');

    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: SingleChildScrollView(
        child: DataTable(
          sortColumnIndex: sortColumnIndex,
          sortAscending: sortAscending,
          columns: [
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
          ],
          rows: documents.map((document) {
            return DataRow(
              cells: [
                DataCell(
                  Tooltip(
                    message: document.title,
                    child: Text(
                      document.title,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(fontWeight: FontWeight.w500),
                    ),
                  ),
                ),
                DataCell(
                  Row(
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
                DataCell(
                  Text(
                    document.createdAt != null
                        ? dateFormat.format(document.createdAt!)
                        : 'N/A',
                    style: const TextStyle(fontSize: 13),
                  ),
                ),
                DataCell(
                  Text(
                    document.updatedAt != null
                        ? dateFormat.format(document.updatedAt!)
                        : 'N/A',
                    style: const TextStyle(fontSize: 13),
                  ),
                ),
              ],
            );
          }).toList(),
        ),
      ),
    );
  }
}

