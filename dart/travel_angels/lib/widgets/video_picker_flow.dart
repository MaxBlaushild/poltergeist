import 'dart:io';

import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:travel_angels/screens/video_editor_screen.dart';
import 'package:travel_angels/utils/video_platform_utils.dart';

/// Bottom sheet that lets the user pick a video from gallery, camera, or file,
/// then navigates to [VideoEditorScreen]. Only shown when
/// [supportsFullVideoEditing] is true.
void showVideoPickerFlow(BuildContext context, {VoidCallback? onComplete}) {
  if (!supportsFullVideoEditing) {
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text(
          'Full video editing is available on Android, iOS, and macOS. '
          'Use the mobile app for the full experience.',
        ),
      ),
    );
    return;
  }

  final navigator = Navigator.of(context);
  navigator.pop(); // Close import bottom sheet if open
  showModalBottomSheet<void>(
    context: navigator.context,
    isScrollControlled: true,
    builder: (ctx) => _VideoPickerSheet(onComplete: onComplete),
  );
}

class _VideoPickerSheet extends StatelessWidget {
  final VoidCallback? onComplete;

  const _VideoPickerSheet({this.onComplete});

  Future<void> _pickFromGallery(BuildContext context) async {
    final picker = ImagePicker();
    final xFile = await picker.pickVideo(source: ImageSource.gallery);
    if (!context.mounted) return;
    Navigator.of(context).pop();
    if (xFile == null) return;
    final file = File(xFile.path);
    if (!context.mounted) return;
    _openEditor(context, file);
  }

  Future<void> _pickFromCamera(BuildContext context) async {
    final picker = ImagePicker();
    final xFile = await picker.pickVideo(source: ImageSource.camera);
    if (!context.mounted) return;
    Navigator.of(context).pop();
    if (xFile == null) return;
    final file = File(xFile.path);
    if (!context.mounted) return;
    _openEditor(context, file);
  }

  Future<void> _pickFromFile(BuildContext context) async {
    final result = await FilePicker.platform.pickFiles(
      type: FileType.custom,
      allowedExtensions: ['mp4', 'mov', 'm4v'],
      allowMultiple: false,
    );
    if (!context.mounted) return;
    Navigator.of(context).pop();
    if (result == null || result.files.isEmpty) return;
    final f = result.files.single;
    final path = f.path;
    if (path == null) return;
    final file = File(path);
    if (!context.mounted) return;
    _openEditor(context, file);
  }

  void _openEditor(BuildContext context, File file) {
    Navigator.of(context).push<void>(
      MaterialPageRoute(
        builder: (context) => VideoEditorScreen(
          videoFile: file,
          onComplete: onComplete,
        ),
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
        borderRadius: const BorderRadius.vertical(top: Radius.circular(16.0)),
      ),
      child: SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Center(
              child: Container(
                width: 40,
                height: 4,
                margin: const EdgeInsets.only(bottom: 16.0),
                decoration: BoxDecoration(
                  color: theme.colorScheme.onSurface.withValues(alpha: 0.3),
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
            ),
            Text(
              'Import Video',
              style: theme.textTheme.titleLarge?.copyWith(
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              'Choose a video to edit',
              style: theme.textTheme.bodyMedium?.copyWith(
                color: theme.colorScheme.onSurface.withValues(alpha: 0.7),
              ),
            ),
            const SizedBox(height: 24),
            _Option(
              icon: Icons.photo_library_outlined,
              title: 'Gallery',
              subtitle: 'Pick a video from your library',
              onTap: () => _pickFromGallery(context),
            ),
            const SizedBox(height: 12),
            _Option(
              icon: Icons.videocam_outlined,
              title: 'Camera',
              subtitle: 'Record a new video',
              onTap: () => _pickFromCamera(context),
            ),
            const SizedBox(height: 12),
            _Option(
              icon: Icons.folder_outlined,
              title: 'From file',
              subtitle: 'Select MP4, MOV, or M4V',
              onTap: () => _pickFromFile(context),
            ),
            const SizedBox(height: 16),
          ],
        ),
      ),
    );
  }
}

class _Option extends StatelessWidget {
  final IconData icon;
  final String title;
  final String subtitle;
  final VoidCallback onTap;

  const _Option({
    required this.icon,
    required this.title,
    required this.subtitle,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(12),
      child: Container(
        padding: const EdgeInsets.all(16.0),
        decoration: BoxDecoration(
          border: Border.all(color: theme.colorScheme.outline),
          borderRadius: BorderRadius.circular(12),
        ),
        child: Row(
          children: [
            Icon(icon, color: theme.colorScheme.primary, size: 32),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    subtitle,
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: theme.colorScheme.onSurface.withValues(alpha: 0.7),
                    ),
                  ),
                ],
              ),
            ),
            Icon(
              Icons.chevron_right,
              color: theme.colorScheme.onSurface.withValues(alpha: 0.5),
            ),
          ],
        ),
      ),
    );
  }
}
