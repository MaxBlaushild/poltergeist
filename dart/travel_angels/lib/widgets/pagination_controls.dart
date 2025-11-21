import 'package:flutter/material.dart';

/// Widget for pagination controls (page navigation and page size selector)
class PaginationControls extends StatelessWidget {
  final int currentPage;
  final int totalPages;
  final int totalItems;
  final int pageSize;
  final List<int> pageSizeOptions;
  final Function(int page) onPageChanged;
  final Function(int pageSize) onPageSizeChanged;

  const PaginationControls({
    super.key,
    required this.currentPage,
    required this.totalPages,
    required this.totalItems,
    required this.pageSize,
    required this.pageSizeOptions,
    required this.onPageChanged,
    required this.onPageSizeChanged,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final startItem = totalItems == 0 ? 0 : (currentPage * pageSize) + 1;
    final endItem = ((currentPage + 1) * pageSize).clamp(0, totalItems);

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 8.0),
      decoration: BoxDecoration(
        border: Border(
          top: BorderSide(
            color: theme.colorScheme.outline.withOpacity(0.2),
          ),
        ),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          // Items per page selector
          Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                'Items per page:',
                style: theme.textTheme.bodySmall,
              ),
              const SizedBox(width: 8),
              DropdownButton<int>(
                value: pageSize,
                items: pageSizeOptions.map((size) {
                  return DropdownMenuItem<int>(
                    value: size,
                    child: Text(size.toString()),
                  );
                }).toList(),
                onChanged: (value) {
                  if (value != null) {
                    onPageSizeChanged(value);
                  }
                },
                underline: Container(),
              ),
            ],
          ),
          // Page info and navigation
          Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              // Item range display
              Text(
                totalItems == 0
                    ? 'No items'
                    : '$startItem-$endItem of $totalItems',
                style: theme.textTheme.bodySmall,
              ),
              const SizedBox(width: 16),
              // Previous button
              IconButton(
                icon: const Icon(Icons.chevron_left),
                onPressed: currentPage > 0
                    ? () => onPageChanged(currentPage - 1)
                    : null,
                tooltip: 'Previous page',
              ),
              // Page number display
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 12),
                child: Text(
                  totalPages == 0
                      ? '0 / 0'
                      : '${currentPage + 1} / $totalPages',
                  style: theme.textTheme.bodyMedium?.copyWith(
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ),
              // Next button
              IconButton(
                icon: const Icon(Icons.chevron_right),
                onPressed: currentPage < totalPages - 1
                    ? () => onPageChanged(currentPage + 1)
                    : null,
                tooltip: 'Next page',
              ),
            ],
          ),
        ],
      ),
    );
  }
}

