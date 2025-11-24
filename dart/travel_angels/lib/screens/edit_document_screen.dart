import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/document.dart';
import 'package:travel_angels/models/document_tag.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/document_service.dart';

/// Screen for editing a document's title and content
class EditDocumentScreen extends StatefulWidget {
  final Document document;

  const EditDocumentScreen({
    super.key,
    required this.document,
  });

  @override
  State<EditDocumentScreen> createState() => _EditDocumentScreenState();
}

class _EditDocumentScreenState extends State<EditDocumentScreen> {
  final DocumentService _documentService = DocumentService(
    APIClient(ApiConstants.baseUrl),
  );

  late TextEditingController _titleController;
  late TextEditingController _contentController;
  late TextEditingController _tagInputController;
  List<DocumentTag> _tags = [];
  bool _isLoading = false;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _titleController = TextEditingController(text: widget.document.title);
    _contentController = TextEditingController(
      text: widget.document.content ?? '',
    );
    _tagInputController = TextEditingController();
    // Load tags from document
    _tags = List<DocumentTag>.from(widget.document.documentTags ?? []);
  }

  @override
  void dispose() {
    _titleController.dispose();
    _contentController.dispose();
    _tagInputController.dispose();
    super.dispose();
  }

  void _addTag(String tagText) {
    final trimmedText = tagText.trim();
    if (trimmedText.isEmpty) {
      return;
    }

    // Check for duplicates (case-insensitive)
    final lowerText = trimmedText.toLowerCase();
    if (_tags.any((tag) => tag.text.toLowerCase() == lowerText)) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Tag "$trimmedText" already exists'),
          backgroundColor: Theme.of(context).colorScheme.error,
        ),
      );
      return;
    }

    // Add new tag (without ID, will be created on backend)
    setState(() {
      _tags.add(DocumentTag(id: '', text: trimmedText));
      _tagInputController.clear();
    });
  }

  void _removeTag(DocumentTag tag) {
    setState(() {
      _tags.remove(tag);
    });
  }

  Future<void> _handleUpdate() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      // Calculate existing tag IDs and new tag texts
      final existingTagIds = _tags
          .where((tag) => tag.id.isNotEmpty)
          .map((tag) => tag.id)
          .toList();
      
      final newTagTexts = _tags
          .where((tag) => tag.id.isEmpty)
          .map((tag) => tag.text)
          .toList();

      // Send content if document has content (always allow content updates)
      // Always send tag arrays (empty if no tags) to replace all tags
      await _documentService.updateDocument(
        documentId: widget.document.id,
        title: _titleController.text.trim(),
        content: widget.document.content != null && widget.document.content!.isNotEmpty
            ? _contentController.text.trim()
            : null,
        existingTagIds: existingTagIds,
        newTagTexts: newTagTexts,
      );

      if (mounted) {
        Navigator.pop(context, true); // Return true to indicate successful update
      }
    } catch (e) {
      setState(() {
        _isLoading = false;
        _errorMessage = e.toString().replaceFirst('Exception: ', '');
      });
    }
  }

  void _handleCancel() {
    Navigator.pop(context);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Edit Document'),
        actions: [
          if (_isLoading)
            const Padding(
              padding: EdgeInsets.all(16.0),
              child: SizedBox(
                width: 20,
                height: 20,
                child: CircularProgressIndicator(strokeWidth: 2),
              ),
            )
          else
            TextButton(
              onPressed: _isLoading ? null : _handleUpdate,
              child: const Text('Update'),
            ),
        ],
        leading: IconButton(
          icon: const Icon(Icons.close),
          onPressed: _isLoading ? null : _handleCancel,
        ),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (_errorMessage != null) ...[
              Container(
                padding: const EdgeInsets.all(12.0),
                margin: const EdgeInsets.only(bottom: 16.0),
                decoration: BoxDecoration(
                  color: Theme.of(context).colorScheme.errorContainer,
                  borderRadius: BorderRadius.circular(8.0),
                ),
                child: Row(
                  children: [
                    Icon(
                      Icons.error_outline,
                      color: Theme.of(context).colorScheme.onErrorContainer,
                    ),
                    const SizedBox(width: 12.0),
                    Expanded(
                      child: Text(
                        _errorMessage!,
                        style: TextStyle(
                          color: Theme.of(context).colorScheme.onErrorContainer,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ],
            TextField(
              controller: _titleController,
              decoration: const InputDecoration(
                labelText: 'Document Title',
                hintText: 'Enter document title',
                border: OutlineInputBorder(),
              ),
              enabled: !_isLoading,
            ),
            const SizedBox(height: 16.0),
            // Tags section
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Tags',
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                ),
                const SizedBox(height: 8.0),
                // Display existing tags as chips
                if (_tags.isNotEmpty)
                  Wrap(
                    spacing: 8.0,
                    runSpacing: 8.0,
                    children: _tags.map((tag) {
                      return Chip(
                        label: Text(tag.text),
                        onDeleted: _isLoading
                            ? null
                            : () => _removeTag(tag),
                        deleteIcon: const Icon(Icons.close, size: 18),
                      );
                    }).toList(),
                  ),
                const SizedBox(height: 8.0),
                // Add tag input
                Row(
                  children: [
                    Expanded(
                      child: TextField(
                        controller: _tagInputController,
                        decoration: const InputDecoration(
                          labelText: 'Add Tag',
                          hintText: 'Enter tag name',
                          border: OutlineInputBorder(),
                        ),
                        enabled: !_isLoading,
                        onSubmitted: (value) {
                          if (value.trim().isNotEmpty) {
                            _addTag(value);
                          }
                        },
                      ),
                    ),
                    const SizedBox(width: 8.0),
                    IconButton(
                      onPressed: _isLoading
                          ? null
                          : () {
                              if (_tagInputController.text.trim().isNotEmpty) {
                                _addTag(_tagInputController.text);
                              }
                            },
                      icon: const Icon(Icons.add),
                      tooltip: 'Add Tag',
                    ),
                  ],
                ),
              ],
            ),
            const SizedBox(height: 16.0),
            if (widget.document.link != null && widget.document.link!.isNotEmpty) ...[
              // Show link when document has a link
              Container(
                padding: const EdgeInsets.all(16.0),
                decoration: BoxDecoration(
                  border: Border.all(
                    color: Theme.of(context).colorScheme.outline,
                  ),
                  borderRadius: BorderRadius.circular(8.0),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Icon(
                          Icons.link,
                          color: Theme.of(context).colorScheme.primary,
                        ),
                        const SizedBox(width: 8.0),
                        Text(
                          'Document Link',
                          style: Theme.of(context).textTheme.titleMedium?.copyWith(
                                fontWeight: FontWeight.bold,
                              ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12.0),
                    Text(
                      widget.document.link!,
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            color: Theme.of(context).colorScheme.primary,
                            decoration: TextDecoration.underline,
                          ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 12.0),
                    SizedBox(
                      width: double.infinity,
                      child: ElevatedButton.icon(
                        onPressed: _isLoading
                            ? null
                            : () async {
                                final uri = Uri.parse(widget.document.link!);
                                if (await canLaunchUrl(uri)) {
                                  await launchUrl(uri, mode: LaunchMode.externalApplication);
                                } else {
                                  if (mounted) {
                                    ScaffoldMessenger.of(context).showSnackBar(
                                      SnackBar(
                                        content: Text('Could not open link: ${widget.document.link}'),
                                        backgroundColor: Theme.of(context).colorScheme.error,
                                      ),
                                    );
                                  }
                                }
                              },
                        icon: const Icon(Icons.open_in_new),
                        label: const Text('Open Link'),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 16.0),
            ],
            if (widget.document.content != null && widget.document.content!.isNotEmpty) ...[
              // Show content field when document has content
              TextField(
                controller: _contentController,
                decoration: const InputDecoration(
                  labelText: 'Document Content',
                  hintText: 'Enter document content',
                  border: OutlineInputBorder(),
                  alignLabelWithHint: true,
                ),
                maxLines: null,
                minLines: 10,
                enabled: !_isLoading,
                textAlignVertical: TextAlignVertical.top,
              ),
            ],
          ],
        ),
      ),
      bottomNavigationBar: _isLoading
          ? null
          : SafeArea(
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: [
                    TextButton(
                      onPressed: _handleCancel,
                      child: const Text('Cancel'),
                    ),
                    const SizedBox(width: 8.0),
                    ElevatedButton(
                      onPressed: _handleUpdate,
                      child: const Text('Update'),
                    ),
                  ],
                ),
              ),
            ),
    );
  }
}

