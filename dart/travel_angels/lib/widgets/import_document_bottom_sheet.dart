import 'package:flutter/material.dart';
import 'package:travel_angels/widgets/google_drive_file_picker.dart';

/// Bottom sheet for selecting import source (Google Drive or File)
class ImportDocumentBottomSheet extends StatelessWidget {
  const ImportDocumentBottomSheet({super.key});

  void _handleGoogleDriveSelected(BuildContext context) {
    Navigator.pop(context); // Close bottom sheet
    // Show Google Drive file picker
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      builder: (context) => const GoogleDriveFilePicker(),
    );
  }

  void _handleFileSelected(BuildContext context) {
    // File import is not yet available
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('File import is not yet available'),
        backgroundColor: Colors.orange,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Container(
      padding: const EdgeInsets.all(16.0),
      decoration: BoxDecoration(
        color: theme.scaffoldBackgroundColor,
        borderRadius: const BorderRadius.vertical(
          top: Radius.circular(16.0),
        ),
      ),
      child: SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // Handle bar
            Center(
              child: Container(
                width: 40,
                height: 4,
                margin: const EdgeInsets.only(bottom: 16.0),
                decoration: BoxDecoration(
                  color: theme.colorScheme.onSurface.withOpacity(0.3),
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
            ),
            // Title
            Text(
              'Import Document',
              style: theme.textTheme.titleLarge?.copyWith(
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 24),
            // Google Drive option
            _ImportOption(
              icon: Icons.cloud,
              title: 'Import from Google Drive',
              subtitle: 'Select a Google Doc or Sheet',
              enabled: true,
              onTap: () => _handleGoogleDriveSelected(context),
            ),
            const SizedBox(height: 12),
            // File option (disabled)
            _ImportOption(
              icon: Icons.file_upload,
              title: 'Import from File',
              subtitle: 'Coming soon',
              enabled: false,
              onTap: () => _handleFileSelected(context),
            ),
            const SizedBox(height: 16),
          ],
        ),
      ),
    );
  }
}

class _ImportOption extends StatelessWidget {
  final IconData icon;
  final String title;
  final String subtitle;
  final bool enabled;
  final VoidCallback onTap;

  const _ImportOption({
    required this.icon,
    required this.title,
    required this.subtitle,
    required this.enabled,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return InkWell(
      onTap: enabled ? onTap : null,
      borderRadius: BorderRadius.circular(12),
      child: Container(
        padding: const EdgeInsets.all(16.0),
        decoration: BoxDecoration(
          border: Border.all(
            color: enabled
                ? theme.colorScheme.outline
                : theme.colorScheme.outline.withOpacity(0.3),
          ),
          borderRadius: BorderRadius.circular(12),
          color: enabled
              ? null
              : theme.colorScheme.surface.withOpacity(0.5),
        ),
        child: Row(
          children: [
            Icon(
              icon,
              color: enabled
                  ? theme.colorScheme.primary
                  : theme.colorScheme.onSurface.withOpacity(0.3),
              size: 32,
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w600,
                      color: enabled
                          ? theme.colorScheme.onSurface
                          : theme.colorScheme.onSurface.withOpacity(0.5),
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    subtitle,
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: enabled
                          ? theme.colorScheme.onSurface.withOpacity(0.7)
                          : theme.colorScheme.onSurface.withOpacity(0.4),
                    ),
                  ),
                ],
              ),
            ),
            if (!enabled)
              Icon(
                Icons.lock_outline,
                color: theme.colorScheme.onSurface.withOpacity(0.3),
                size: 20,
              ),
          ],
        ),
      ),
    );
  }
}

