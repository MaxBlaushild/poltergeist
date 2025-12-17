import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/document_location.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/google_drive_service.dart';
import 'package:travel_angels/widgets/import_type_dialog.dart';
import 'package:travel_angels/widgets/location_selector.dart';

/// Widget for picking a Google Drive file (Doc or Sheet)
class GoogleDriveFilePicker extends StatefulWidget {
  final VoidCallback? onImportComplete;

  const GoogleDriveFilePicker({
    super.key,
    this.onImportComplete,
  });

  @override
  State<GoogleDriveFilePicker> createState() => _GoogleDriveFilePickerState();
}

class _GoogleDriveFilePickerState extends State<GoogleDriveFilePicker> {
  final GoogleDriveService _googleDriveService = GoogleDriveService(
    APIClient(ApiConstants.baseUrl),
  );

  List<Map<String, dynamic>> _files = [];
  bool _isLoading = true;
  String? _errorMessage;
  String? _nextPageToken;
  bool _isLoadingMore = false;

  @override
  void initState() {
    super.initState();
    _loadFiles();
  }

  Future<void> _loadFiles({bool loadMore = false}) async {
    if (loadMore && (_nextPageToken == null || _isLoadingMore)) {
      return;
    }

    setState(() {
      if (loadMore) {
        _isLoadingMore = true;
      } else {
        _isLoading = true;
        _errorMessage = null;
      }
    });

    try {
      print('[GoogleDriveFilePicker] Loading files (loadMore: $loadMore, pageToken: $_nextPageToken)');
      
      final response = await _googleDriveService.listFiles(
        pageSize: 50,
        pageToken: loadMore ? _nextPageToken : null,
      );

      print('[GoogleDriveFilePicker] Received response: ${response.keys}');
      print('[GoogleDriveFilePicker] Files count: ${(response['files'] as List?)?.length ?? 0}');

      final List<dynamic> filesList = response['files'] ?? [];
      final List<Map<String, dynamic>> newFiles = filesList
          .map((file) => Map<String, dynamic>.from(file))
          .toList();

      setState(() {
        if (loadMore) {
          _files.addAll(newFiles);
        } else {
          _files = newFiles;
        }
        _nextPageToken = response['nextPageToken'];
        _isLoading = false;
        _isLoadingMore = false;
        _errorMessage = null;
      });
      
      print('[GoogleDriveFilePicker] Successfully loaded ${newFiles.length} files');
    } catch (e) {
      print('[GoogleDriveFilePicker] Error loading files: $e');
      print('[GoogleDriveFilePicker] Error type: ${e.runtimeType}');
      
        String errorMsg = 'Failed to load files';
      String? detailedError;
      
        if (e is DioException) {
        print('[GoogleDriveFilePicker] DioException details:');
        print('  - Status code: ${e.response?.statusCode}');
        print('  - Status message: ${e.response?.statusMessage}');
        print('  - Response data: ${e.response?.data}');
        print('  - Request path: ${e.requestOptions.path}');
        print('  - Request method: ${e.requestOptions.method}');
        print('  - Error type: ${e.type}');
        print('  - Error message: ${e.message}');
        
          if (e.response != null) {
            errorMsg = '${errorMsg}: ${e.response?.statusCode} - ${e.response?.statusMessage}';
          
          // Try to extract detailed error message from response
          if (e.response?.data != null) {
            if (e.response!.data is Map) {
              final errorData = e.response!.data as Map<String, dynamic>;
              detailedError = errorData['error']?.toString();
              if (detailedError != null) {
                errorMsg = detailedError;
              }
            } else if (e.response!.data is String) {
              detailedError = e.response!.data as String;
              errorMsg = '$errorMsg\n$detailedError';
            }
          }
          } else {
            errorMsg = '${errorMsg}: ${e.message ?? e.toString()}';
          }
        } else {
          errorMsg = '$errorMsg: $e';
        }
      
      print('[GoogleDriveFilePicker] Final error message: $errorMsg');
      
      setState(() {
        _isLoading = false;
        _isLoadingMore = false;
        _errorMessage = errorMsg;
      });
    }
  }

  bool _isSelectable(Map<String, dynamic> file) {
    final mimeType = file['mimeType'] as String?;
    return mimeType == 'application/vnd.google-apps.document' ||
        mimeType == 'application/vnd.google-apps.spreadsheet';
  }

  IconData _getFileIcon(Map<String, dynamic> file) {
    final mimeType = file['mimeType'] as String?;
    if (mimeType == 'application/vnd.google-apps.document') {
      return Icons.description;
    } else if (mimeType == 'application/vnd.google-apps.spreadsheet') {
      return Icons.table_chart;
    }
    return Icons.insert_drive_file;
  }

  String _getFileTypeLabel(Map<String, dynamic> file) {
    final mimeType = file['mimeType'] as String?;
    if (mimeType == 'application/vnd.google-apps.document') {
      return 'Google Doc';
    } else if (mimeType == 'application/vnd.google-apps.spreadsheet') {
      return 'Google Sheet';
    }
    return 'Other';
  }

  void _handleFileSelected(Map<String, dynamic> file) async {
    // Show import type dialog
    final importType = await showDialog<String>(
      context: context,
      builder: (context) => const ImportTypeDialog(),
    );

    if (importType == null || !mounted) {
      return;
    }

    // Show location selection dialog
    List<DocumentLocation> selectedLocations = [];
    final shouldImport = await showDialog<bool>(
      context: context,
      builder: (context) => _LocationSelectionDialog(
        fileName: file['name'] as String? ?? 'Document',
        onLocationsChanged: (locations) {
          selectedLocations = locations;
        },
      ),
    );

    if (shouldImport != true || !mounted) {
      return;
    }

    // Show loading indicator
    if (mounted) {
      showDialog(
        context: context,
        barrierDismissible: false,
        builder: (context) => const Center(
          child: CircularProgressIndicator(),
        ),
      );
    }

    try {
      // Import document with locations
      await _googleDriveService.importDocument(
        file['id'] as String,
        importType,
        locations: selectedLocations.isNotEmpty ? selectedLocations : null,
      );

      // Close loading and file picker
      if (mounted) {
        Navigator.of(context).pop(); // Close loading
        Navigator.of(context).pop(); // Close file picker
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              importType == 'import'
                  ? 'Document imported successfully!'
                  : 'Document reference created successfully!',
            ),
            backgroundColor: Colors.green,
          ),
        );

        // Notify parent that import is complete
        widget.onImportComplete?.call();
      }
    } catch (e) {
      // Close loading
      if (mounted) {
        Navigator.of(context).pop(); // Close loading
        String errorMsg = 'Failed to import document';
        if (e is DioException) {
          if (e.response != null) {
            errorMsg = '${errorMsg}: ${e.response?.statusCode} - ${e.response?.statusMessage}';
            if (e.response?.data != null && e.response?.data is Map) {
              final errorData = e.response?.data as Map<String, dynamic>;
              errorMsg = errorData['error']?.toString() ?? errorMsg;
            }
          } else {
            errorMsg = '${errorMsg}: ${e.message ?? e.toString()}';
          }
        } else {
          errorMsg = '$errorMsg: $e';
        }
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(errorMsg),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Container(
      height: MediaQuery.of(context).size.height * 0.8,
      padding: const EdgeInsets.all(16.0),
      decoration: BoxDecoration(
        color: theme.scaffoldBackgroundColor,
        borderRadius: const BorderRadius.vertical(
          top: Radius.circular(16.0),
        ),
      ),
      child: SafeArea(
        child: Column(
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
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  'Select a Document',
                  style: theme.textTheme.titleLarge?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.close),
                  onPressed: () => Navigator.of(context).pop(),
                ),
              ],
            ),
            const SizedBox(height: 16),
            // Content
            Expanded(
              child: _isLoading
                  ? const Center(child: CircularProgressIndicator())
                  : _errorMessage != null
                      ? Center(
                          child: Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Icon(
                                Icons.error_outline,
                                size: 48,
                                color: theme.colorScheme.error,
                              ),
                              const SizedBox(height: 16),
                              Text(
                                _errorMessage!,
                                style: theme.textTheme.bodyMedium?.copyWith(
                                  color: theme.colorScheme.error,
                                ),
                                textAlign: TextAlign.center,
                              ),
                              const SizedBox(height: 16),
                              ElevatedButton(
                                onPressed: () => _loadFiles(),
                                child: const Text('Retry'),
                              ),
                            ],
                          ),
                        )
                      : _files.isEmpty
                          ? Center(
                              child: Text(
                                'No files found',
                                style: theme.textTheme.bodyMedium?.copyWith(
                                  color: theme.colorScheme.onSurface.withOpacity(0.6),
                                ),
                              ),
                            )
                          : ListView.builder(
                              itemCount: _files.length + (_nextPageToken != null ? 1 : 0),
                              itemBuilder: (context, index) {
                                if (index == _files.length) {
                                  // Load more button
                                  if (_isLoadingMore) {
                                    return const Center(
                                      child: Padding(
                                        padding: EdgeInsets.all(16.0),
                                        child: CircularProgressIndicator(),
                                      ),
                                    );
                                  }
                                  return TextButton(
                                    onPressed: () => _loadFiles(loadMore: true),
                                    child: const Text('Load More'),
                                  );
                                }

                                final file = _files[index];
                                final isSelectable = _isSelectable(file);

                                return _FileItem(
                                  file: file,
                                  icon: _getFileIcon(file),
                                  typeLabel: _getFileTypeLabel(file),
                                  isSelectable: isSelectable,
                                  onTap: isSelectable
                                      ? () => _handleFileSelected(file)
                                      : null,
                                );
                              },
                            ),
            ),
          ],
        ),
      ),
    );
  }
}

class _LocationSelectionDialog extends StatefulWidget {
  final String fileName;
  final Function(List<DocumentLocation>) onLocationsChanged;

  const _LocationSelectionDialog({
    required this.fileName,
    required this.onLocationsChanged,
  });

  @override
  State<_LocationSelectionDialog> createState() => _LocationSelectionDialogState();
}

class _LocationSelectionDialogState extends State<_LocationSelectionDialog> {
  List<DocumentLocation> _selectedLocations = [];

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return AlertDialog(
      title: Text('Tag "${widget.fileName}"'),
      content: SizedBox(
        width: double.maxFinite,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(
                'Optionally add locations to tag this document:',
                style: theme.textTheme.bodyMedium,
              ),
              const SizedBox(height: 16),
              LocationSelector(
                initialLocations: _selectedLocations,
                onLocationsChanged: (locations) {
                  setState(() {
                    _selectedLocations = locations;
                  });
                  widget.onLocationsChanged(locations);
                },
              ),
            ],
          ),
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(false),
          child: const Text('Cancel'),
        ),
        ElevatedButton(
          onPressed: () => Navigator.of(context).pop(true),
          child: const Text('Import'),
        ),
      ],
    );
  }
}

class _FileItem extends StatelessWidget {
  final Map<String, dynamic> file;
  final IconData icon;
  final String typeLabel;
  final bool isSelectable;
  final VoidCallback? onTap;

  const _FileItem({
    required this.file,
    required this.icon,
    required this.typeLabel,
    required this.isSelectable,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final fileName = file['name'] as String? ?? 'Unknown';

    return InkWell(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 12.0),
        decoration: BoxDecoration(
          border: Border(
            bottom: BorderSide(
              color: theme.colorScheme.outline.withOpacity(0.2),
            ),
          ),
        ),
        child: Row(
          children: [
            Icon(
              icon,
              color: isSelectable
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
                    fileName,
                    style: theme.textTheme.bodyLarge?.copyWith(
                      fontWeight: FontWeight.w500,
                      color: isSelectable
                          ? theme.colorScheme.onSurface
                          : theme.colorScheme.onSurface.withOpacity(0.5),
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    typeLabel,
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: isSelectable
                          ? theme.colorScheme.onSurface.withOpacity(0.7)
                          : theme.colorScheme.onSurface.withOpacity(0.4),
                    ),
                  ),
                ],
              ),
            ),
            if (!isSelectable)
              Icon(
                Icons.block,
                color: theme.colorScheme.onSurface.withOpacity(0.3),
                size: 20,
              ),
          ],
        ),
      ),
    );
  }
}

