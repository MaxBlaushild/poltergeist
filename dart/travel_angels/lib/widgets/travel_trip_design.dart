import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:url_launcher/url_launcher.dart';

const tripInk = Color(0xFF163D39);
const tripMuted = Color(0xFF62746F);
const tripPaper = Color(0xFFF7F5EE);
const tripSand = Color(0xFFEDE9DE);
const tripOrange = Color(0xFFD16D4E);

Map<String, dynamic> tripMap(dynamic value) =>
    value is Map ? Map<String, dynamic>.from(value) : <String, dynamic>{};
List<Map<String, dynamic>> tripMaps(dynamic value) =>
    value is List ? value.whereType<Map>().map(tripMap).toList() : [];
Map<String, dynamic> tripCopy(Map<String, dynamic> value) =>
    tripMap(jsonDecode(jsonEncode(value)));
Map<String, dynamic> tripDraft(Map<String, dynamic> stop) =>
    tripCopy(stop['draft'] is Map ? tripMap(stop['draft']) : stop);
String tripError(Object error) => error is TravelCalendarException
    ? error.message
    : 'Something went wrong. Your changes are still here; please try again.';

Future<Map<String, dynamic>> saveTripDraft(
  TravelCalendarService service,
  Map<String, dynamic> stop,
  Map<String, dynamic> changes,
) => service.put(
  '/calendar/stops/${stop['id']}',
  data: {...tripDraft(stop), ...changes, 'revision': stop['revision']},
);

String tripParticipationLabel(dynamic status) => switch (status) {
  'joined' => 'Confirmed',
  'requested' => 'Awaiting approval',
  'waitlisted' => 'On the waitlist',
  'interested' => 'Interested',
  'needs_reconfirmation' => 'Please reconfirm',
  'declined' => 'Request declined',
  'withdrawn' => 'Withdrawn',
  'removed' => 'Attendance removed',
  'cancelled' => 'Cancelled',
  _ => 'Come along',
};

String tripDates(Map<String, dynamic> stop) {
  final start = DateTime.tryParse(stop['startDate'] as String? ?? '');
  final end = DateTime.tryParse(stop['endDate'] as String? ?? '');
  if (stop['datePrecision'] == 'tbd' || start == null) {
    return 'Dates to be decided';
  }
  final prefix = stop['datePrecision'] == 'approximate' ? 'Around ' : '';
  if (end == null || end == start) {
    return '$prefix${DateFormat.yMMMd().format(start)}';
  }
  if (start.year == end.year && start.month == end.month) {
    return '$prefix${DateFormat.MMMd().format(start)}–${end.day}, ${end.year}';
  }
  return '$prefix${DateFormat.yMMMd().format(start)} – ${DateFormat.yMMMd().format(end)}';
}

Future<void> openTripLink(BuildContext context, String raw) async {
  final uri = Uri.tryParse(raw);
  if (uri == null ||
      !['https', 'http'].contains(uri.scheme) ||
      uri.host.isEmpty) {
    return;
  }
  final opened = await launchUrl(uri, mode: LaunchMode.externalApplication);
  if (!opened && context.mounted) {
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('This link could not be opened.')),
    );
  }
}

class TripPill extends StatelessWidget {
  const TripPill(
    this.label, {
    super.key,
    this.icon,
    this.color = tripInk,
    this.background = tripSand,
  });
  final String label;
  final IconData? icon;
  final Color color;
  final Color background;

  @override
  Widget build(BuildContext context) => Container(
    padding: const EdgeInsets.symmetric(horizontal: 11, vertical: 6),
    decoration: BoxDecoration(
      color: background,
      borderRadius: BorderRadius.circular(30),
    ),
    child: Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        if (icon != null) ...[
          Icon(icon, size: 14, color: color),
          const SizedBox(width: 5),
        ],
        Flexible(
          child: Text(
            label,
            style: TextStyle(
              color: color,
              fontSize: 12,
              fontWeight: FontWeight.w600,
            ),
          ),
        ),
      ],
    ),
  );
}

class TripSection extends StatelessWidget {
  const TripSection({
    super.key,
    required this.eyebrow,
    required this.title,
    required this.child,
    this.action,
    this.subtitle,
  });
  final String eyebrow;
  final String title;
  final String? subtitle;
  final Widget child;
  final Widget? action;

  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.symmetric(vertical: 26),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          eyebrow.toUpperCase(),
          style: const TextStyle(
            color: tripMuted,
            letterSpacing: 2.2,
            fontSize: 10,
            fontWeight: FontWeight.w700,
          ),
        ),
        const SizedBox(height: 8),
        Wrap(
          spacing: 16,
          runSpacing: 8,
          crossAxisAlignment: WrapCrossAlignment.center,
          children: [
            Text(
              title,
              style: const TextStyle(
                color: tripInk,
                fontFamily: 'Georgia',
                fontSize: 30,
                height: 1.15,
              ),
            ),
            if (action != null) action!,
          ],
        ),
        if (subtitle != null) ...[
          const SizedBox(height: 8),
          Text(
            subtitle!,
            style: const TextStyle(color: tripMuted, height: 1.5),
          ),
        ],
        const SizedBox(height: 20),
        child,
      ],
    ),
  );
}

class TripEmptyState extends StatelessWidget {
  const TripEmptyState({
    super.key,
    required this.icon,
    required this.title,
    required this.message,
    this.action,
  });
  final IconData icon;
  final String title;
  final String message;
  final Widget? action;

  @override
  Widget build(BuildContext context) => Container(
    width: double.infinity,
    padding: const EdgeInsets.all(24),
    decoration: BoxDecoration(
      color: Colors.white.withValues(alpha: .65),
      border: Border.all(color: tripSand),
      borderRadius: BorderRadius.circular(18),
    ),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(icon, size: 28, color: tripMuted),
        const SizedBox(height: 14),
        Text(
          title,
          style: const TextStyle(
            color: tripInk,
            fontSize: 17,
            fontWeight: FontWeight.w600,
          ),
        ),
        const SizedBox(height: 6),
        Text(message, style: const TextStyle(color: tripMuted, height: 1.5)),
        if (action != null) ...[const SizedBox(height: 12), action!],
      ],
    ),
  );
}

class TripAvatar extends StatelessWidget {
  const TripAvatar({super.key, required this.name, this.url, this.radius = 20});
  final String name;
  final String? url;
  final double radius;

  @override
  Widget build(BuildContext context) => CircleAvatar(
    radius: radius,
    backgroundColor: const Color(0xFFDFE7DE),
    foregroundImage: url != null && Uri.tryParse(url!)?.scheme == 'https'
        ? NetworkImage(url!)
        : null,
    child: Text(
      name.trim().isEmpty ? '?' : name.trim().substring(0, 1).toUpperCase(),
      style: TextStyle(
        color: tripInk,
        fontWeight: FontWeight.w600,
        fontSize: radius * .75,
      ),
    ),
  );
}

/// A deliberately small Markdown subset: paragraphs, bullets and safe links.
/// Rich HTML and arbitrary URL schemes are never rendered.
class TripDescription extends StatelessWidget {
  const TripDescription(this.text, {super.key});
  final String text;

  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: text.split('\n').map((line) {
      if (line.trim().isEmpty) return const SizedBox(height: 10);
      final bullet = RegExp(r'^\s*[-*]\s+').hasMatch(line);
      final content = bullet
          ? line.replaceFirst(RegExp(r'^\s*[-*]\s+'), '')
          : line;
      final links = RegExp(
        r'\[([^\]]+)\]\((https?://[^\s)]+)\)|(https?://[^\s]+)',
      );
      final parts = <Widget>[];
      var offset = 0;
      for (final match in links.allMatches(content)) {
        if (match.start > offset) {
          parts.add(
            Text(
              content.substring(offset, match.start),
              style: const TextStyle(
                color: tripInk,
                fontSize: 16,
                height: 1.65,
              ),
            ),
          );
        }
        final url = match.group(2) ?? match.group(3)!;
        parts.add(
          TextButton(
            style: TextButton.styleFrom(
              padding: EdgeInsets.zero,
              minimumSize: const Size(0, 28),
              tapTargetSize: MaterialTapTargetSize.shrinkWrap,
            ),
            onPressed: () => openTripLink(context, url),
            child: Text(
              match.group(1) ?? url,
              style: const TextStyle(decoration: TextDecoration.underline),
            ),
          ),
        );
        offset = match.end;
      }
      if (offset < content.length) {
        parts.add(
          Text(
            content.substring(offset),
            style: const TextStyle(color: tripInk, fontSize: 16, height: 1.65),
          ),
        );
      }
      return Padding(
        padding: const EdgeInsets.only(bottom: 4),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (bullet)
              const Padding(
                padding: EdgeInsets.only(right: 10, top: 4),
                child: Text('•', style: TextStyle(color: tripOrange)),
              ),
            Expanded(
              child: Wrap(
                crossAxisAlignment: WrapCrossAlignment.center,
                children: parts,
              ),
            ),
          ],
        ),
      );
    }).toList(),
  );
}

class TripFormDialog extends StatelessWidget {
  const TripFormDialog({
    super.key,
    required this.title,
    required this.children,
    required this.onSave,
    this.busy = false,
    this.error,
    this.saveLabel = 'Save changes',
  });
  final String title;
  final List<Widget> children;
  final VoidCallback? onSave;
  final bool busy;
  final String? error;
  final String saveLabel;

  @override
  Widget build(BuildContext context) => PopScope(
    canPop: !busy,
    child: AlertDialog(
      title: Text(title),
      content: SizedBox(
        width: 580,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              ...children,
              if (error != null)
                Padding(
                  padding: const EdgeInsets.only(top: 16),
                  child: Text(
                    error!,
                    style: TextStyle(
                      color: Theme.of(context).colorScheme.error,
                    ),
                  ),
                ),
            ],
          ),
        ),
      ),
      actions: [
        TextButton(
          onPressed: busy ? null : () => Navigator.pop(context),
          child: const Text('Cancel'),
        ),
        FilledButton(
          onPressed: busy ? null : onSave,
          child: Text(busy ? 'Saving…' : saveLabel),
        ),
      ],
    ),
  );
}
