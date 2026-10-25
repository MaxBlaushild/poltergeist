import 'package:flutter/material.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/document.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/document_service.dart';
import 'package:travel_angels/widgets/file_picker_widget.dart';
import 'package:travel_angels/widgets/google_drive_file_picker.dart';
import 'package:travel_angels/widgets/guide_details_fields.dart';

Future<Document?> showAddGuideFlow(BuildContext context) async {
  final source = await showModalBottomSheet<String>(
    context: context,
    isScrollControlled: true,
    builder: (context) => SafeArea(
      child: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Add a guide',
              style: Theme.of(context).textTheme.headlineMedium,
            ),
            const SizedBox(height: 8),
            const Text(
              'Start with a favorite file or write something worth passing along.',
            ),
            const SizedBox(height: 20),
            ListTile(
              contentPadding: EdgeInsets.zero,
              leading: const Icon(Icons.edit_note),
              title: const Text('Write a guide or itinerary'),
              subtitle: const Text(
                'Your recommendations, day by day or place by place',
              ),
              onTap: () => Navigator.pop(context, 'write'),
            ),
            ListTile(
              contentPadding: EdgeInsets.zero,
              leading: const Icon(Icons.add_to_drive_outlined),
              title: const Text('Google Docs or Sheets'),
              subtitle: const Text(
                'Import a copy or keep a link to the original',
              ),
              onTap: () => Navigator.pop(context, 'drive'),
            ),
            ListTile(
              contentPadding: EdgeInsets.zero,
              leading: const Icon(Icons.upload_file_outlined),
              title: const Text('Upload a PDF or Word file'),
              subtitle: const Text('Bring the text into your library'),
              onTap: () => Navigator.pop(context, 'file'),
            ),
            const Divider(height: 28),
            const Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Icon(Icons.lock_outline, size: 18),
                SizedBox(width: 8),
                Expanded(
                  child: Text(
                    'Only you can see a new guide. Choose who to share it with when you’re ready.',
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    ),
  );
  if (source == null || !context.mounted) {
    return null;
  }
  if (source == 'write') {
    return Navigator.push<Document>(
      context,
      MaterialPageRoute(builder: (_) => const WriteGuideScreen()),
    );
  }
  return showModalBottomSheet<Document>(
    context: context,
    isScrollControlled: true,
    builder: (_) => source == 'drive'
        ? const GoogleDriveFilePicker()
        : const FilePickerWidget(),
  );
}

class WriteGuideScreen extends StatefulWidget {
  const WriteGuideScreen({super.key, this.service});
  final DocumentService? service;

  @override
  State<WriteGuideScreen> createState() => _WriteGuideScreenState();
}

class _WriteGuideScreenState extends State<WriteGuideScreen> {
  final _form = GlobalKey<FormState>();
  final _details = GuideDetailsController();
  final _content = TextEditingController();
  late final _service =
      widget.service ?? DocumentService(APIClient(ApiConstants.baseUrl));
  bool _saving = false;
  String? _error;

  @override
  void dispose() {
    _details.dispose();
    _content.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    if (!_form.currentState!.validate()) {
      return;
    }
    setState(() {
      _saving = true;
      _error = null;
    });
    try {
      final result = await _service.createDocument(
        title: _details.title.text.trim(),
        provider: 'internal',
        format: _details.format,
        destination: _details.destination.text.trim(),
        personalNote: _details.personalNote.text.trim(),
        content: _content.text.trim(),
      );
      if (mounted) {
        Navigator.pop(context, Document.fromJson(result));
      }
    } catch (_) {
      if (mounted) {
        setState(() {
          _saving = false;
          _error = 'Your guide could not be saved. Please try again.';
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Write a guide')),
    body: SafeArea(
      child: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 760),
            child: Form(
              key: _form,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  GuideDetailsFields(controller: _details, enabled: !_saving),
                  const SizedBox(height: 24),
                  TextFormField(
                    controller: _content,
                    enabled: !_saving,
                    minLines: 10,
                    maxLines: null,
                    textCapitalization: TextCapitalization.sentences,
                    decoration: const InputDecoration(
                      labelText: 'Your guide',
                      alignLabelWithHint: true,
                      hintText:
                          'Day 1 — Arrive in Lisbon\n\nThe places you loved, what to book, and what you’d do again.',
                      helperText:
                          'Headings, lists, and Markdown tables are supported.',
                      helperMaxLines: 2,
                    ),
                    validator: (value) => value == null || value.trim().isEmpty
                        ? 'Add some details to your guide'
                        : null,
                  ),
                  const SizedBox(height: 24),
                  const Text('Only you can see this guide until you share it.'),
                  if (_error != null)
                    Padding(
                      padding: const EdgeInsets.only(top: 16),
                      child: Text(
                        _error!,
                        style: TextStyle(
                          color: Theme.of(context).colorScheme.error,
                        ),
                      ),
                    ),
                  const SizedBox(height: 16),
                  FilledButton.icon(
                    onPressed: _saving ? null : _save,
                    icon: _saving
                        ? const SizedBox(
                            width: 18,
                            height: 18,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Icon(Icons.add, size: 18),
                    label: Text(_saving ? 'Saving…' : 'Add to your library'),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    ),
  );
}
