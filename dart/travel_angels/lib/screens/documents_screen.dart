import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/document.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/document_service.dart';
import 'package:travel_angels/widgets/documents_table.dart';
import 'package:travel_angels/widgets/import_document_bottom_sheet.dart';
import 'package:travel_angels/widgets/pagination_controls.dart';

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
        switch (_sortColumnIndex) {
          case 0: // Title
            comparison = a.title.compareTo(b.title);
            break;
          case 1: // Provider
            comparison = a.provider.name.compareTo(b.provider.name);
            break;
          case 2: // Created Date
            final aDate = a.createdAt ?? DateTime(0);
            final bDate = b.createdAt ?? DateTime(0);
            comparison = aDate.compareTo(bDate);
            break;
          case 3: // Updated Date
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
      builder: (context) => const ImportDocumentBottomSheet(),
    );
    // Refresh documents after import
    _loadDocuments();
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

    return Scaffold(
      appBar: AppBar(
        title: const Text('Documents'),
        actions: [
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
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _errorMessage != null
              ? Center(
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
                        _errorMessage!,
                        style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                              color: Theme.of(context).colorScheme.error,
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
                            color: Theme.of(context).colorScheme.onSurface.withOpacity(0.4),
                          ),
                          const SizedBox(height: 16),
                          Text(
                            'No documents yet',
                            style: Theme.of(context).textTheme.titleLarge?.copyWith(
                                  color: Theme.of(context).colorScheme.onSurface.withOpacity(0.6),
                                ),
                          ),
                          const SizedBox(height: 8),
                          Text(
                            'Import your first document to get started',
                            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                                  color: Theme.of(context).colorScheme.onSurface.withOpacity(0.5),
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
                          child: DocumentsTable(
                            documents: _paginatedDocuments,
                            sortColumnIndex: _sortColumnIndex,
                            sortAscending: _sortAscending,
                            onSort: _onSort,
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
    );
  }
}

