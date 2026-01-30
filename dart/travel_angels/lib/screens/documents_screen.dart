import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/document.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/document_service.dart';
import 'package:travel_angels/screens/edit_document_screen.dart';
import 'package:travel_angels/screens/video_editor_screen.dart';
import 'package:travel_angels/services/media_service.dart';
import 'package:travel_angels/utils/document_utils.dart';
import 'package:travel_angels/utils/video_platform_utils.dart';
import 'package:travel_angels/widgets/documents_table.dart';
import 'package:travel_angels/widgets/import_document_bottom_sheet.dart';
import 'package:travel_angels/widgets/pagination_controls.dart';
import 'package:travel_angels/widgets/video_preview_dialog.dart';

/// Documents screen for managing travel documents
class DocumentsScreen extends StatefulWidget {
  const DocumentsScreen({super.key});

  @override
  State<DocumentsScreen> createState() => _DocumentsScreenState();
}

class _DocumentsScreenState extends State<DocumentsScreen> {
  final DocumentService _documentService = DocumentService(
    APIClient(ApiConstants.baseUrl),
  );
  final MediaService _mediaService = MediaService(APIClient(ApiConstants.baseUrl));

  List<Document> _allDocuments = [];
  List<Document> _sortedDocuments = [];
  List<Document> _paginatedDocuments = [];
  bool _isLoading = true;
  String? _errorMessage;

  // Sorting state
  int? _sortColumnIndex;
  bool _sortAscending = false;

  // Pagination state
  int _currentPage = 0;
  int _pageSize = 25;
  final List<int> _pageSizeOptions = [10, 25, 50, 100];

  // Selection state
  Set<String> _selectedDocumentIds = {};

  @override
  void initState() {
    super.initState();
    _loadDocuments();
  }

  Future<void> _loadDocuments() async {
    final authProvider = context.read<AuthProvider>();
    final user = authProvider.user;

    if (!authProvider.isAuthenticated || user?.id == null) {
      setState(() {
        _isLoading = false;
        _errorMessage = 'User not authenticated or user ID missing';
      });
      return;
    }

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final documentsJson = await _documentService.getDocumentsByUserId(user!.id!);
      final documents = documentsJson
          .map((json) => Document.fromJson(json))
          .toList();

      print('[DocumentsScreen._loadDocuments] Loaded ${documents.length} documents');
      for (final doc in documents) {
        final isVideo = DocumentUtils.isVideo(doc);
        print('[DocumentsScreen._loadDocuments] Document: id=${doc.id}, title="${doc.title}", provider=${doc.provider}, link=${doc.link}, isVideo=$isVideo');
      }

      _allDocuments = documents;
      _applySortAndPagination();

      setState(() {
        _isLoading = false;
        _errorMessage = null;
      });
    } catch (e) {
      setState(() {
        _isLoading = false;
        String errorMsg = 'Failed to load documents';
        if (e is DioException) {
          if (e.response != null) {
            errorMsg = '$errorMsg: ${e.response?.statusCode} - ${e.response?.statusMessage}';
            if (e.response?.data != null && e.response?.data is Map) {
              final errorData = e.response?.data as Map<String, dynamic>;
              errorMsg = errorData['error']?.toString() ?? errorMsg;
            }
          } else {
            errorMsg = '$errorMsg: ${e.message ?? e.toString()}';
          }
        } else {
          errorMsg = '$errorMsg: $e';
        }
        _errorMessage = errorMsg;
      });
    }
  }

  void _applySortAndPagination() {
    // Apply sorting
    List<Document> sorted = List.from(_allDocuments);

    if (_sortColumnIndex != null) {
      sorted.sort((a, b) {
        int comparison = 0;
        // Adjust for checkbox column (index 0 is now checkbox, so subtract 1)
        final adjustedIndex = _sortColumnIndex! - 1;
        switch (adjustedIndex) {
          case 0: // Title (was index 1, now index 0 after checkbox)
            comparison = a.title.compareTo(b.title);
            break;
          case 1: // Provider (was index 2, now index 1 after checkbox)
            comparison = a.provider.name.compareTo(b.provider.name);
            break;
          case 2: // Created Date (was index 3, now index 2 after checkbox)
            final aDate = a.createdAt ?? DateTime(0);
            final bDate = b.createdAt ?? DateTime(0);
            comparison = aDate.compareTo(bDate);
            break;
          case 3: // Updated Date (was index 4, now index 3 after checkbox)
            final aDate = a.updatedAt ?? DateTime(0);
            final bDate = b.updatedAt ?? DateTime(0);
            comparison = aDate.compareTo(bDate);
            break;
        }
        return _sortAscending ? comparison : -comparison;
      });
    } else {
      // Default sort: by created date (newest first)
      sorted.sort((a, b) {
        final aDate = a.createdAt ?? DateTime(0);
        final bDate = b.createdAt ?? DateTime(0);
        return bDate.compareTo(aDate);
      });
    }

    _sortedDocuments = sorted;

    // Apply pagination
    final totalPages = (_sortedDocuments.length / _pageSize).ceil();
    int currentPage = _currentPage;
    if (currentPage >= totalPages && totalPages > 0) {
      currentPage = totalPages - 1;
    }
    if (currentPage < 0) {
      currentPage = 0;
    }

    final startIndex = currentPage * _pageSize;
    final endIndex = (startIndex + _pageSize).clamp(0, _sortedDocuments.length);
    _paginatedDocuments = _sortedDocuments.sublist(
      startIndex.clamp(0, _sortedDocuments.length),
      endIndex,
    );

    // Update current page if it was adjusted
    if (currentPage != _currentPage) {
      _currentPage = currentPage;
    }
  }

  void _onSort(int columnIndex, bool ascending) {
    _sortColumnIndex = columnIndex;
    _sortAscending = ascending;
    _applySortAndPagination();
    setState(() {});
  }

  void _onPageChanged(int page) {
    _currentPage = page;
    _applySortAndPagination();
    setState(() {});
  }

  void _onPageSizeChanged(int pageSize) {
    _pageSize = pageSize;
    _currentPage = 0; // Reset to first page
    _applySortAndPagination();
    setState(() {});
  }

  void _showImportBottomSheet() async {
    await showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      builder: (context) => ImportDocumentBottomSheet(
        onImportComplete: () {
          // Refresh documents after import
          _loadDocuments();
        },
      ),
    );
  }

  Future<void> _handleDocumentTap(Document document) async {
    print('[DocumentsScreen._handleDocumentTap] Called for document: id=${document.id}, title="${document.title}", link=${document.link}');
    final isVideo = DocumentUtils.isVideo(document);
    print('[DocumentsScreen._handleDocumentTap] isVideo=$isVideo');
    
    // If it's a video, show preview instead of edit screen
    if (isVideo) {
      print('[DocumentsScreen._handleDocumentTap] Handling as VIDEO');
      if (document.link == null) {
        print('[DocumentsScreen._handleDocumentTap] ❌ Video link is null!');
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Video link is missing')),
        );
        return;
      }
      
      print('[DocumentsScreen._handleDocumentTap] Opening video preview dialog with URL: ${document.link}');
      // Show video preview dialog
      if (!mounted) return;
      showDialog(
        context: context,
        builder: (context) => VideoPreviewDialog(videoUrl: document.link!),
      );
      return;
    }
    
    print('[DocumentsScreen._handleDocumentTap] Handling as regular document');
    // Navigate to edit screen for non-video documents
    final result = await Navigator.push<bool>(
      context,
      MaterialPageRoute(
        builder: (context) => EditDocumentScreen(document: document),
      ),
    );

    // Refresh documents if update was successful
    if (result == true) {
      _loadDocuments();
    }
  }

  Future<void> _handleEditTap(Document document) async {
    print('[DocumentsScreen._handleEditTap] Called for document: id=${document.id}, title="${document.title}", link=${document.link}');
    final isVideo = DocumentUtils.isVideo(document);
    print('[DocumentsScreen._handleEditTap] isVideo=$isVideo');
    
    if (isVideo) {
      print('[DocumentsScreen._handleEditTap] Handling as VIDEO');
      // Handle video editing
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

      if (document.link == null) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Video link is missing')),
        );
        return;
      }

      // Show loading dialog
      if (!mounted) return;
      showDialog(
        context: context,
        barrierDismissible: false,
        builder: (context) => const Center(child: CircularProgressIndicator()),
      );

      try {
        // Download video from URL
        final videoFile = await _mediaService.downloadVideo(document.link!);
        if (!mounted) return;
        Navigator.of(context).pop(); // Close loading dialog

        if (videoFile == null) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Failed to download video')),
          );
          return;
        }

        // Open video editor
        await Navigator.push<void>(
          context,
          MaterialPageRoute(
            builder: (context) => VideoEditorScreen(
              videoFile: videoFile,
              onComplete: () {
                _loadDocuments();
              },
            ),
          ),
        );
      } catch (e) {
        if (mounted) {
          Navigator.of(context).pop(); // Close loading dialog if still open
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Failed to open video editor: $e')),
          );
        }
      }
    } else {
      // Handle document editing
      _handleDocumentTap(document);
    }
  }

  void _toggleDocumentSelection(String documentId, bool selected) {
    setState(() {
      if (selected) {
        _selectedDocumentIds.add(documentId);
      } else {
        _selectedDocumentIds.remove(documentId);
      }
    });
  }

  void _selectAll() {
    setState(() {
      _selectedDocumentIds = _paginatedDocuments.map((doc) => doc.id).toSet();
    });
  }

  void _deselectAll() {
    setState(() {
      _selectedDocumentIds.clear();
    });
  }

  Future<void> _deleteSingleDocument(Document document) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete Document'),
        content: Text('Are you sure you want to delete "${document.title}"? This action cannot be undone.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            style: TextButton.styleFrom(
              foregroundColor: Theme.of(context).colorScheme.error,
            ),
            child: const Text('Delete'),
          ),
        ],
      ),
    );

    if (confirmed != true) return;

    try {
      await _documentService.deleteDocument(document.id);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Document "${document.title}" deleted successfully')),
        );
        _loadDocuments();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to delete document: ${e.toString().replaceFirst('Exception: ', '')}'),
            backgroundColor: Theme.of(context).colorScheme.error,
          ),
        );
      }
    }
  }

  Future<void> _deleteSelectedDocuments() async {
    if (_selectedDocumentIds.isEmpty) return;

    final selectedDocs = _allDocuments
        .where((doc) => _selectedDocumentIds.contains(doc.id))
        .toList();

    final count = selectedDocs.length;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete Documents'),
        content: Text(
          count == 1
              ? 'Are you sure you want to delete "${selectedDocs.first.title}"? This action cannot be undone.'
              : 'Are you sure you want to delete $count documents? This action cannot be undone.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            style: TextButton.styleFrom(
              foregroundColor: Theme.of(context).colorScheme.error,
            ),
            child: const Text('Delete'),
          ),
        ],
      ),
    );

    if (confirmed != true) return;

    int successCount = 0;
    int failCount = 0;
    final failedTitles = <String>[];

    for (final doc in selectedDocs) {
      try {
        await _documentService.deleteDocument(doc.id);
        successCount++;
      } catch (e) {
        failCount++;
        failedTitles.add(doc.title);
      }
    }

    if (mounted) {
      if (failCount == 0) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              successCount == 1
                  ? 'Document deleted successfully'
                  : '$successCount documents deleted successfully',
            ),
          ),
        );
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              'Deleted $successCount document(s), but failed to delete $failCount document(s)',
            ),
            backgroundColor: Theme.of(context).colorScheme.error,
            duration: const Duration(seconds: 5),
          ),
        );
      }

      setState(() {
        _selectedDocumentIds.clear();
      });

      _loadDocuments();
    }
  }

  int get _totalPages => (_sortedDocuments.length / _pageSize).ceil();

  @override
  Widget build(BuildContext context) {
    final authProvider = context.watch<AuthProvider>();
    final user = authProvider.user;

    if (!authProvider.isAuthenticated) {
      return Scaffold(
        appBar: AppBar(
          title: const Text('Documents'),
        ),
        body: const Center(
          child: Text('Please log in to view documents'),
        ),
      );
    }

    // If authenticated but user ID is missing, show error
    if (user?.id == null) {
      return Scaffold(
        appBar: AppBar(
          title: const Text('Documents'),
          actions: [
            IconButton(
              icon: const Icon(Icons.refresh),
              onPressed: () {
                authProvider.verifyToken();
                _loadDocuments();
              },
              tooltip: 'Refresh',
            ),
          ],
        ),
        body: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                Icons.error_outline,
                size: 48,
                color: Theme.of(context).colorScheme.error,
              ),
              const SizedBox(height: 16),
              Text(
                'Unable to load user information',
                style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      color: Theme.of(context).colorScheme.error,
                    ),
              ),
              const SizedBox(height: 8),
              const Text('Please try refreshing or log out and log back in'),
              const SizedBox(height: 16),
              ElevatedButton(
                onPressed: () {
                  authProvider.verifyToken();
                  _loadDocuments();
                },
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
      );
    }

    final theme = Theme.of(context);
    
    return Scaffold(
      body: SafeArea(
        child: Column(
          children: [
            // Documents header with actions
            Padding(
              padding: const EdgeInsets.fromLTRB(16.0, 16.0, 8.0, 8.0),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  // Documents header
                  Text(
                    _selectedDocumentIds.isEmpty
                        ? 'Documents'
                        : '${_selectedDocumentIds.length} selected',
                    style: theme.textTheme.headlineMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  // Actions
                  Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
          if (_selectedDocumentIds.isNotEmpty) ...[
            IconButton(
              icon: const Icon(Icons.select_all),
              onPressed: _selectedDocumentIds.length == _paginatedDocuments.length
                  ? _deselectAll
                  : _selectAll,
              tooltip: _selectedDocumentIds.length == _paginatedDocuments.length
                  ? 'Deselect all'
                  : 'Select all',
            ),
            IconButton(
              icon: const Icon(Icons.delete),
              onPressed: _deleteSelectedDocuments,
              tooltip: 'Delete selected',
            ),
          ],
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadDocuments,
            tooltip: 'Refresh',
          ),
          IconButton(
            icon: const Icon(Icons.add),
            onPressed: _showImportBottomSheet,
            tooltip: 'Import Document',
          ),
        ],
      ),
                ],
              ),
            ),
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
                        onPressed: _loadDocuments,
                        child: const Text('Retry'),
                      ),
                    ],
                  ),
                )
              : _allDocuments.isEmpty
                  ? Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(
                            Icons.description_outlined,
                            size: 64,
                                    color: theme.colorScheme.onSurface.withOpacity(0.4),
                          ),
                          const SizedBox(height: 16),
                          Text(
                            'No documents yet',
                                    style: theme.textTheme.titleLarge?.copyWith(
                                          color: theme.colorScheme.onSurface.withOpacity(0.6),
                                ),
                          ),
                          const SizedBox(height: 8),
                          Text(
                            'Import your first document to get started',
                                    style: theme.textTheme.bodyMedium?.copyWith(
                                          color: theme.colorScheme.onSurface.withOpacity(0.5),
                                ),
                          ),
                          const SizedBox(height: 24),
                          ElevatedButton.icon(
                            onPressed: _showImportBottomSheet,
                            icon: const Icon(Icons.add),
                            label: const Text('Import Document'),
                          ),
                        ],
                      ),
                    )
                  : Column(
                      children: [
                        Expanded(
                          child: Builder(
                            builder: (context) {
                              print('[DocumentsScreen.build] Creating DocumentsTable with ${_paginatedDocuments.length} documents');
                              print('[DocumentsScreen.build] onEditTap callback: ${_handleEditTap != null ? "provided" : "null"}');
                              print('[DocumentsScreen.build] onDocumentTap callback: ${_handleDocumentTap != null ? "provided" : "null"}');
                              return DocumentsTable(
                                documents: _paginatedDocuments,
                                sortColumnIndex: _sortColumnIndex != null ? _sortColumnIndex! + 1 : null, // Adjust for checkbox column
                                sortAscending: _sortAscending,
                                onSort: (columnIndex, ascending) {
                                  // Adjust back for checkbox column
                                  _onSort(columnIndex - 1, ascending);
                                },
                                onDocumentTap: _handleDocumentTap,
                                onEditTap: _handleEditTap,
                                selectedDocumentIds: _selectedDocumentIds,
                                onSelectionChanged: _toggleDocumentSelection,
                              );
                            },
                          ),
                        ),
                        PaginationControls(
                          currentPage: _currentPage,
                          totalPages: _totalPages,
                          totalItems: _sortedDocuments.length,
                          pageSize: _pageSize,
                          pageSizeOptions: _pageSizeOptions,
                          onPageChanged: _onPageChanged,
                          onPageSizeChanged: _onPageSizeChanged,
                        ),
                      ],
                            ),
            ),
          ],
        ),
                    ),
    );
  }
}

