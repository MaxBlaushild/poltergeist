import 'dart:io';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/document_service.dart';

/// Widget for picking and importing PDF or Word files from device
class FilePickerWidget extends StatefulWidget {
  final VoidCallback? onImportComplete;

  const FilePickerWidget({
    super.key,
    this.onImportComplete,
  });

  @override
  State<FilePickerWidget> createState() => _FilePickerWidgetState();
}

class _FilePickerWidgetState extends State<FilePickerWidget> {
  final DocumentService _documentService = DocumentService(
    APIClient(ApiConstants.baseUrl),
  );

  bool _isLoading = false;
  String? _errorMessage;

  Future<void> _pickAndImportFile() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      // Pick file - only PDF and Word files
      final result = await FilePicker.platform.pickFiles(
        type: FileType.custom,
        allowedExtensions: ['pdf', 'docx'],
        allowMultiple: false,
      );

      if (result == null || result.files.isEmpty) {
        // User cancelled file selection
        if (mounted) {
          Navigator.of(context).pop();
        }
        return;
      }

      final pickedFile = result.files.single;
      if (pickedFile.path == null) {
        setState(() {
          _isLoading = false;
          _errorMessage = 'Failed to get file path';
        });
        return;
      }

      final file = File(pickedFile.path!);

      // Validate file extension
      final extension = pickedFile.extension?.toLowerCase();
      if (extension != 'pdf' && extension != 'docx') {
        setState(() {
          _isLoading = false;
          _errorMessage = 'Only PDF and Word (.docx) files are supported';
        });
        return;
      }

      // Show loading dialog
      if (mounted) {
        showDialog(
          context: context,
          barrierDismissible: false,
          builder: (context) => const Center(
            child: CircularProgressIndicator(),
          ),
        );
      }

      // Parse the document
      final parsedDoc = await _documentService.parseDocument(file);

      // Extract filename without extension for title
      final fileName = pickedFile.name;
      final title = fileName.replaceAll(RegExp(r'\.(pdf|docx)$', caseSensitive: false), '');

      // Create document with parsed content
      await _documentService.createDocument(
        title: title,
        provider: 'internal',
        content: parsedDoc['content'] as String?,
      );

      // Close loading dialog and file picker
      if (mounted) {
        Navigator.of(context).pop(); // Close loading
        Navigator.of(context).pop(); // Close file picker

        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Document imported successfully!'),
            backgroundColor: Colors.green,
          ),
        );

        // Notify parent that import is complete
        widget.onImportComplete?.call();
      }
    } catch (e) {
      // Close loading dialog if it's open
      if (mounted && Navigator.of(context).canPop()) {
        Navigator.of(context).pop();
      }

      setState(() {
        _isLoading = false;
        _errorMessage = e.toString().replaceFirst('Exception: ', '');
      });

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to import document: ${_errorMessage ?? 'Unknown error'}'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  @override
  void initState() {
    super.initState();
    // Automatically trigger file picker when widget is shown
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _pickAndImportFile();
    });
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
              'Import from File',
              style: theme.textTheme.titleLarge?.copyWith(
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 24),
            if (_isLoading)
              const Center(
                child: Padding(
                  padding: EdgeInsets.all(24.0),
                  child: CircularProgressIndicator(),
                ),
              )
            else if (_errorMessage != null)
              Container(
                padding: const EdgeInsets.all(12.0),
                decoration: BoxDecoration(
                  color: theme.colorScheme.errorContainer,
                  borderRadius: BorderRadius.circular(8.0),
                ),
                child: Row(
                  children: [
                    Icon(
                      Icons.error_outline,
                      color: theme.colorScheme.onErrorContainer,
                    ),
                    const SizedBox(width: 12.0),
                    Expanded(
                      child: Text(
                        _errorMessage!,
                        style: TextStyle(
                          color: theme.colorScheme.onErrorContainer,
                        ),
                      ),
                    ),
                  ],
                ),
              )
            else
              Text(
                'Select a PDF or Word (.docx) file to import',
                style: theme.textTheme.bodyMedium,
              ),
            const SizedBox(height: 16),
          ],
        ),
      ),
    );
  }
}

