import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/document_location.dart';
import 'package:travel_angels/models/document.dart';
import 'package:travel_angels/widgets/guide_details_fields.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/document_service.dart';
import 'package:travel_angels/widgets/location_selector.dart';

/// Widget for picking and importing PDF or Word files from device
class FilePickerWidget extends StatefulWidget {
  final VoidCallback? onImportComplete;

  const FilePickerWidget({super.key, this.onImportComplete});

  @override
  State<FilePickerWidget> createState() => _FilePickerWidgetState();
}

class _FilePickerWidgetState extends State<FilePickerWidget> {
  final DocumentService _documentService = DocumentService(
    APIClient(ApiConstants.baseUrl),
  );

  bool _isLoading = false;
  String? _errorMessage;
  PlatformFile? _selectedFile;
  final _form = GlobalKey<FormState>();
  final _details = GuideDetailsController();
  String? _parsedContent;
  List<DocumentLocation> _selectedLocations = [];

  Future<void> _createDocument() async {
    if (_selectedFile == null || !_form.currentState!.validate()) {
      return;
    }

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    // Show loading dialog
    if (mounted) {
      showDialog(
        context: context,
        barrierDismissible: false,
        builder: (context) => const Center(child: CircularProgressIndicator()),
      );
    }

    try {
      // Create document with parsed content and locations
      final created = await _documentService.createDocument(
        title: _details.title.text.trim(),
        format: _details.format,
        personalNote: _details.personalNote.text.trim(),
        destination: _details.destination.text.trim(),
        provider: 'internal',
        content: _parsedContent,
        locations: _selectedLocations.isNotEmpty ? _selectedLocations : null,
      );

      // Close loading dialog and file picker
      if (mounted) {
        Navigator.of(context).pop(); // Close loading
        Navigator.of(
          context,
        ).pop(Document.fromJson(created)); // Close file picker

        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Document imported successfully!'),
            backgroundColor: AppTheme.success,
          ),
        );

        // Notify parent that import is complete
        widget.onImportComplete?.call();
      }
    } catch (e) {
      if (!mounted) return;
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
            content: Text(
              'Failed to import document: ${_errorMessage ?? 'Unknown error'}',
            ),
            backgroundColor: AppTheme.error,
          ),
        );
      }
    }
  }

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
        withData: true,
      );

      if (result == null || result.files.isEmpty) {
        // User cancelled file selection
        if (mounted) {
          Navigator.of(context).pop();
        }
        return;
      }

      if (!mounted) return;
      final pickedFile = result.files.single;
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
          builder: (context) =>
              const Center(child: CircularProgressIndicator()),
        );
      }

      // Parse the document
      final bytes = pickedFile.bytes;
      if (bytes == null) throw Exception('Unable to read the selected file');
      final parsedDoc = await _documentService.parseDocumentBytes(
        bytes,
        pickedFile.name,
      );
      if (!mounted) return;

      // Extract filename without extension for title
      final fileName = pickedFile.name;
      final title = fileName.replaceAll(
        RegExp(r'\.(pdf|docx)$', caseSensitive: false),
        '',
      );

      // Store file info for later use
      setState(() {
        _selectedFile = pickedFile;
        _parsedContent = parsedDoc['content'] as String?;
        _details.title.text = title;
        _isLoading = false;
      });

      // Close loading dialog
      if (mounted && Navigator.of(context).canPop()) {
        Navigator.of(context).pop();
      }
    } catch (e) {
      if (!mounted) return;
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
            content: Text(
              'Failed to import document: ${_errorMessage ?? 'Unknown error'}',
            ),
            backgroundColor: AppTheme.error,
          ),
        );
      }
    }
  }

  @override
  void dispose() {
    _details.dispose();
    super.dispose();
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

    return SingleChildScrollView(
      padding: EdgeInsets.only(bottom: MediaQuery.viewInsetsOf(context).bottom),
      child: Form(
        key: _form,
        child: Container(
          padding: const EdgeInsets.all(24.0),
          decoration: BoxDecoration(
            color: theme.scaffoldBackgroundColor,
            borderRadius: const BorderRadius.vertical(top: Radius.circular(8)),
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
                      color: theme.colorScheme.onSurface.withValues(alpha: 0.3),
                      borderRadius: BorderRadius.circular(2),
                    ),
                  ),
                ),
                // Title
                Text('Import from File', style: theme.textTheme.titleLarge),
                const SizedBox(height: 24),
                if (_selectedFile == null)
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
                        borderRadius: BorderRadius.circular(6),
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
                    )
                else ...[
                  // File selected, show location selector and create button
                  Text(
                    'File selected: ${_selectedFile!.name}',
                    style: theme.textTheme.bodyMedium?.copyWith(
                      fontWeight: FontWeight.w400,
                    ),
                  ),
                  const SizedBox(height: 16),
                  GuideDetailsFields(
                    controller: _details,
                    enabled: !_isLoading,
                  ),
                  const SizedBox(height: 16),
                  const Text('Only you can see this guide until you share it.'),
                  const SizedBox(height: 16),
                  Text(
                    'Optionally add locations to tag this document:',
                    style: theme.textTheme.bodySmall,
                  ),
                  const SizedBox(height: 12),
                  LocationSelector(
                    initialLocations: _selectedLocations,
                    onLocationsChanged: (locations) {
                      setState(() {
                        _selectedLocations = locations;
                      });
                    },
                  ),
                  const SizedBox(height: 16),
                  ElevatedButton.icon(
                    onPressed: _isLoading ? null : _createDocument,
                    icon: _isLoading
                        ? const SizedBox(
                            width: 16,
                            height: 16,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Icon(Icons.upload_file),
                    label: Text(
                      _isLoading ? 'Importing...' : 'Import Document',
                    ),
                    style: ElevatedButton.styleFrom(
                      padding: const EdgeInsets.symmetric(vertical: 16),
                    ),
                  ),
                ],
                const SizedBox(height: 16),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
