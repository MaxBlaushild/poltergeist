import 'package:flutter/material.dart';

/// The same descriptive information is used for authored and imported guides.
class GuideDetailsController {
  GuideDetailsController({
    String title = '',
    String destination = '',
    String personalNote = '',
    this.format = 'guide',
  }) : title = TextEditingController(text: title),
       destination = TextEditingController(text: destination),
       personalNote = TextEditingController(text: personalNote);

  final TextEditingController title;
  final TextEditingController destination;
  final TextEditingController personalNote;
  String format;

  void dispose() {
    title.dispose();
    destination.dispose();
    personalNote.dispose();
  }
}

class GuideDetailsFields extends StatefulWidget {
  const GuideDetailsFields({
    super.key,
    required this.controller,
    this.enabled = true,
  });
  final GuideDetailsController controller;
  final bool enabled;

  @override
  State<GuideDetailsFields> createState() => _GuideDetailsFieldsState();
}

class _GuideDetailsFieldsState extends State<GuideDetailsFields> {
  @override
  Widget build(BuildContext context) {
    final fields = widget.controller;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        TextFormField(
          controller: fields.title,
          enabled: widget.enabled,
          textCapitalization: TextCapitalization.sentences,
          decoration: const InputDecoration(
            labelText: 'Guide title',
            hintText: 'A week in Portugal',
          ),
          validator: (value) =>
              value == null || value.trim().isEmpty ? 'Add a title' : null,
        ),
        const SizedBox(height: 16),
        TextFormField(
          controller: fields.destination,
          enabled: widget.enabled,
          textCapitalization: TextCapitalization.words,
          decoration: const InputDecoration(
            labelText: 'Destination',
            hintText: 'Lisbon & Porto',
          ),
        ),
        const SizedBox(height: 16),
        DropdownButtonFormField<String>(
          initialValue: fields.format,
          isExpanded: true,
          decoration: const InputDecoration(labelText: 'Format'),
          items: [
            DropdownMenuItem(value: 'guide', child: Text('Guide')),
            DropdownMenuItem(value: 'itinerary', child: Text('Itinerary')),
            DropdownMenuItem(value: 'spreadsheet', child: Text('Spreadsheet')),
            if (fields.format == 'video')
              const DropdownMenuItem(value: 'video', child: Text('Video')),
          ],
          onChanged: widget.enabled
              ? (value) {
                  if (value != null) setState(() => fields.format = value);
                }
              : null,
        ),
        const SizedBox(height: 16),
        TextFormField(
          controller: fields.personalNote,
          enabled: widget.enabled,
          minLines: 2,
          maxLines: 4,
          textCapitalization: TextCapitalization.sentences,
          decoration: const InputDecoration(
            labelText: 'A note from you',
            hintText: 'What made this trip special? What should friends know?',
            helperText: 'Visible to everyone you share this guide with.',
            helperMaxLines: 2,
          ),
        ),
      ],
    );
  }
}
