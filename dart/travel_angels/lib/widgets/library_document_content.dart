import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:travel_angels/models/document.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:url_launcher/url_launcher.dart';

String libraryAuthor(Document document) {
  final user = document.user;
  return user?.name?.trim().isNotEmpty == true
      ? user!.name!.trim()
      : user?.username?.trim().isNotEmpty == true
      ? user!.username!.trim()
      : 'Travel Angels member';
}

String libraryDestination(Document document) =>
    document.destination.trim().isNotEmpty
    ? document.destination.trim()
    : (document.documentLocations ?? [])
          .map((location) => location.name)
          .where((name) => name.isNotEmpty)
          .join(' · ');

String libraryFormat(Document document) {
  if (document.format == 'video') return 'video';
  final path = Uri.tryParse(document.link ?? '')?.path.toLowerCase() ?? '';
  if (RegExp(r'\.(mp4|mov|m4v|webm)$').hasMatch(path) ||
      path.contains('/videos/')) {
    return 'video';
  }
  if (['spreadsheet', 'itinerary', 'guide'].contains(document.format)) {
    return document.format;
  }
  return document.provider == CloudDocumentProvider.googleSheets
      ? 'spreadsheet'
      : 'guide';
}

String libraryFormatLabel(String format) => switch (format) {
  'spreadsheet' => 'Spreadsheet',
  'itinerary' => 'Itinerary',
  'video' => 'Video',
  _ => 'Guide',
};

IconData libraryFormatIcon(String format) => switch (format) {
  'spreadsheet' => Icons.table_chart_outlined,
  'itinerary' => Icons.route_outlined,
  'video' => Icons.play_circle_outline,
  _ => Icons.menu_book_outlined,
};

Color libraryFormatColor(String format) => switch (format) {
  'spreadsheet' => const Color(0xFFE1E8D8),
  'itinerary' => const Color(0xFFF6DDC7),
  'video' => const Color(0xFFE8DDF0),
  _ => const Color(0xFFDCE8F2),
};

String librarySource(Document document) => switch (document.provider) {
  CloudDocumentProvider.googleDocs => 'Google Docs',
  CloudDocumentProvider.googleSheets => 'Google Sheets',
  CloudDocumentProvider.internal =>
    libraryFormat(document) == 'video' ? 'Video' : 'Travel Angels',
  CloudDocumentProvider.unknown => switch (Uri.tryParse(
    document.link ?? '',
  )?.path.toLowerCase().split('.').last) {
    'pdf' => 'PDF',
    'doc' || 'docx' => 'Word document',
    _ => 'Imported guide',
  },
};

String libraryAudience(String visibility) => switch (visibility) {
  'network' => 'Your network',
  'selected' => 'Selected friends',
  _ => 'Only you',
};

Uri? librarySafeUri(String? value) {
  final uri = Uri.tryParse(value ?? '');
  return uri != null &&
          (uri.scheme == 'https' || uri.scheme == 'http') &&
          uri.host.isNotEmpty &&
          uri.userInfo.isEmpty
      ? uri
      : null;
}

Future<void> openLibrarySource(BuildContext context, String url) async {
  final uri = librarySafeUri(url);
  try {
    if (uri == null ||
        !await launchUrl(uri, mode: LaunchMode.externalApplication)) {
      throw StateError('unavailable');
    }
  } catch (_) {
    if (context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('This source could not be opened.')),
      );
    }
  }
}

String libraryPlainText(String value) => value
    .replaceAllMapped(RegExp(r'\[([^\]]+)\]\([^)]*\)'), (match) => match[1]!)
    .replaceAll(RegExp(r'^\s{0,3}#{1,6}\s+', multiLine: true), '')
    .replaceAll('**', '')
    .replaceAll('`', '')
    .trim();

/// Reads quoted CSV cells and GFM escapes without treating literal pipes as
/// column boundaries. Other backslashes remain part of the cell's text.
List<String> _cells(String line, String delimiter) {
  var value = line.trim();
  if (delimiter == '|') {
    if (value.startsWith('|')) value = value.substring(1);
    if (value.endsWith('|')) {
      var backslashes = 0;
      for (var i = value.length - 2; i >= 0 && value[i] == '\\'; i--) {
        backslashes++;
      }
      if (backslashes.isEven) value = value.substring(0, value.length - 1);
    }
  }
  final cells = <String>[];
  final cell = StringBuffer();
  var quoted = false;
  for (var i = 0; i < value.length; i++) {
    final char = value[i];
    if (delimiter == '|' &&
        char == '\\' &&
        i + 1 < value.length &&
        (value[i + 1] == '|' || value[i + 1] == '\\')) {
      cell.write(value[++i]);
    } else if (char == '"' && delimiter != '|') {
      if (quoted && i + 1 < value.length && value[i + 1] == '"') {
        cell.write('"');
        i++;
      } else {
        quoted = !quoted;
      }
    } else if (char == delimiter && !quoted) {
      cells.add(cell.toString().trim());
      cell.clear();
    } else {
      cell.write(char);
    }
  }
  cells.add(cell.toString().trim());
  return cells
      .map(
        (value) =>
            value.replaceAll(RegExp(r'<br\s*/?>', caseSensitive: false), '\n'),
      )
      .toList(growable: false);
}

bool _tableSeparator(List<String> cells) =>
    cells.isNotEmpty &&
    cells.every((cell) => RegExp(r'^:?-+:?$').hasMatch(cell));

class _ContentBlock {
  const _ContentBlock.text(this.text) : rows = null;
  const _ContentBlock.table(this.rows) : text = '';
  final String text;
  final List<List<String>>? rows;
}

List<_ContentBlock> _blocks(String content, String format) {
  final lines = content.replaceAll('\r\n', '\n').split('\n');
  final result = <_ContentBlock>[];
  for (var index = 0; index < lines.length; index++) {
    final line = lines[index];
    String? delimiter;
    final next = index + 1 < lines.length ? lines[index + 1] : '';
    final nextPipeCells = _cells(next, '|');
    final markdownTable =
        line.contains('|') &&
        next.contains('|') &&
        _cells(line, '|').length == nextPipeCells.length &&
        _tableSeparator(nextPipeCells);
    if (markdownTable) {
      delimiter = '|';
    } else if (format == 'spreadsheet') {
      for (final candidate in ['\t', ',', '|']) {
        if (_cells(line, candidate).length > 1 &&
            _cells(line, candidate).length == _cells(next, candidate).length) {
          delimiter = candidate;
          break;
        }
      }
    }
    if (delimiter == null) {
      result.add(_ContentBlock.text(line));
      continue;
    }
    final rows = <List<String>>[_cells(line, delimiter)];
    final columns = rows.first.length;
    if (markdownTable) index++;
    while (index + 1 < lines.length) {
      final nextCells = _cells(lines[index + 1], delimiter);
      if (nextCells.length != columns) break;
      index++;
      rows.add(nextCells);
    }
    result.add(_ContentBlock.table(rows));
  }
  return result;
}

class LibraryDocumentContent extends StatelessWidget {
  const LibraryDocumentContent({
    super.key,
    required this.content,
    required this.format,
    this.preview = false,
  });
  final String content;
  final String format;
  final bool preview;

  @override
  Widget build(BuildContext context) {
    final blocks = _blocks(
      preview ? content.substring(0, content.length.clamp(0, 12000)) : content,
      format,
    );
    final shown = preview
        ? blocks
              .where(
                (block) => block.rows != null || block.text.trim().isNotEmpty,
              )
              .take(3)
        : blocks;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        for (final block in shown)
          if (block.rows != null)
            _table(context, block.rows!)
          else
            _line(context, block.text),
      ],
    );
  }

  Widget _table(BuildContext context, List<List<String>> rows) {
    final visibleRows = preview ? rows.take(4).toList() : rows;
    final columns = rows.first.length;
    final table = Table(
      border: TableBorder.all(color: AppTheme.outline.withValues(alpha: .55)),
      defaultVerticalAlignment: TableCellVerticalAlignment.middle,
      children: [
        for (var row = 0; row < visibleRows.length; row++)
          TableRow(
            decoration: row == 0
                ? BoxDecoration(color: AppTheme.paper.withValues(alpha: .55))
                : null,
            children: [
              for (final value in visibleRows[row])
                Padding(
                  padding: EdgeInsets.all(preview ? 5 : 10),
                  child: Semantics(
                    header: row == 0,
                    child: Text(
                      libraryPlainText(value),
                      maxLines: preview ? 1 : null,
                      overflow: preview ? TextOverflow.ellipsis : null,
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        fontSize: preview ? 11 : 14,
                        height: 1.4,
                        color: AppTheme.ink,
                      ),
                    ),
                  ),
                ),
            ],
          ),
      ],
    );
    return Padding(
      padding: EdgeInsets.only(bottom: preview ? 0 : 18),
      child: preview
          ? table
          : LayoutBuilder(
              builder: (context, constraints) => SingleChildScrollView(
                scrollDirection: Axis.horizontal,
                child: SizedBox(
                  width: constraints.maxWidth > columns * 150
                      ? constraints.maxWidth
                      : columns * 150.0,
                  child: table,
                ),
              ),
            ),
    );
  }

  Widget _line(BuildContext context, String line) {
    if (line.trim().isEmpty) return SizedBox(height: preview ? 0 : 12);
    if (preview) {
      return Padding(
        padding: const EdgeInsets.only(bottom: 7),
        child: Text(
          libraryPlainText(line),
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
          style: Theme.of(
            context,
          ).textTheme.bodySmall?.copyWith(color: AppTheme.ink, height: 1.4),
        ),
      );
    }
    final heading = RegExp(r'^\s{0,3}(#{1,6})\s+(.+)$').firstMatch(line);
    final bullet = RegExp(r'^\s*([-*+]\s+|\d+[.)]\s+)(.*)$').firstMatch(line);
    if (heading != null) {
      return Padding(
        padding: const EdgeInsets.only(top: 12, bottom: 8),
        child: LibraryRichText(
          heading[2]!,
          style: heading[1]!.length <= 2
              ? Theme.of(context).textTheme.headlineSmall
              : Theme.of(context).textTheme.titleLarge,
        ),
      );
    }
    return Padding(
      padding: const EdgeInsets.only(bottom: 5),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (bullet != null)
            const Padding(
              padding: EdgeInsets.only(right: 10, top: 2),
              child: Text('•', style: TextStyle(color: AppTheme.rust)),
            ),
          Expanded(child: LibraryRichText(bullet?[2] ?? line)),
        ],
      ),
    );
  }
}

/// Safe inline Markdown: links, emphasis and code. HTML remains inert text.
class LibraryRichText extends StatefulWidget {
  const LibraryRichText(this.text, {super.key, this.style});
  final String text;
  final TextStyle? style;
  @override
  State<LibraryRichText> createState() => _LibraryRichTextState();
}

class _LibraryRichTextState extends State<LibraryRichText> {
  final _recognizers = <TapGestureRecognizer>[];
  void _clear() {
    for (final recognizer in _recognizers) {
      recognizer.dispose();
    }
    _recognizers.clear();
  }

  @override
  void dispose() {
    _clear();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    _clear();
    final spans = <InlineSpan>[];
    final pattern = RegExp(
      r'\[([^\]]+)\]\((https?://[^\s)]+)\)|\*\*([^*]+)\*\*|`([^`]+)`',
    );
    var position = 0;
    for (final match in pattern.allMatches(widget.text)) {
      if (match.start > position) {
        spans.add(TextSpan(text: widget.text.substring(position, match.start)));
      }
      if (match[2] != null) {
        final recognizer = TapGestureRecognizer()
          ..onTap = () => openLibrarySource(context, match[2]!);
        _recognizers.add(recognizer);
        spans.add(
          TextSpan(
            text: match[1],
            style: const TextStyle(
              decoration: TextDecoration.underline,
              color: AppTheme.rust,
            ),
            recognizer: recognizer,
          ),
        );
      } else {
        spans.add(
          TextSpan(
            text: match[3] ?? match[4],
            style: TextStyle(
              fontWeight: match[3] != null ? FontWeight.bold : null,
              fontFamily: match[4] != null ? 'monospace' : null,
            ),
          ),
        );
      }
      position = match.end;
    }
    spans.add(TextSpan(text: widget.text.substring(position)));
    return Text.rich(
      TextSpan(children: spans),
      style:
          widget.style ??
          Theme.of(
            context,
          ).textTheme.bodyLarge?.copyWith(height: 1.65, color: AppTheme.ink),
    );
  }
}
