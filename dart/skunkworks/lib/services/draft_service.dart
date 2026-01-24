import 'dart:convert';
import 'dart:io';
import 'dart:math';
import 'package:path_provider/path_provider.dart';
import 'package:skunkworks/models/draft.dart';

class DraftService {
  static const String _draftsDirName = 'drafts';
  static const String _indexFileName = 'drafts.json';

  Future<Directory> _getDraftsDir() async {
    final appDir = await getApplicationDocumentsDirectory();
    final draftsDir = Directory('${appDir.path}/$_draftsDirName');
    if (!await draftsDir.exists()) {
      await draftsDir.create(recursive: true);
    }
    return draftsDir;
  }

  Future<File> _getIndexFile() async {
    final dir = await _getDraftsDir();
    return File('${dir.path}/$_indexFileName');
  }

  String _generateId() {
    final r = Random();
    return 'draft_${DateTime.now().millisecondsSinceEpoch}_${r.nextInt(9999)}';
  }

  Future<List<Draft>> _readIndex() async {
    final indexFile = await _getIndexFile();
    if (!await indexFile.exists()) {
      return [];
    }
    try {
      final contents = await indexFile.readAsString();
      final list = jsonDecode(contents) as List<dynamic>?;
      if (list == null) return [];
      return list
          .map((e) => Draft.fromJson(e as Map<String, dynamic>))
          .toList();
    } catch (e) {
      return [];
    }
  }

  Future<void> _writeIndex(List<Draft> drafts) async {
    final indexFile = await _getIndexFile();
    final list = drafts.map((d) => d.toJson()).toList();
    await indexFile.writeAsString(jsonEncode(list));
  }

  /// Saves a new draft. Copies [image] to drafts dir and appends to index.
  Future<Draft> saveDraft(File image, String? caption) async {
    final dir = await _getDraftsDir();
    final id = _generateId();
    final destPath = '${dir.path}/$id.jpg';
    await image.copy(destPath);

    final draft = Draft(
      id: id,
      imagePath: destPath,
      caption: caption?.trim().isEmpty ?? true ? null : caption!.trim(),
      createdAt: DateTime.now(),
    );

    final drafts = await _readIndex();
    drafts.add(draft);
    await _writeIndex(drafts);
    return draft;
  }

  /// Returns all drafts, newest first. Skips entries whose image file is missing.
  Future<List<Draft>> getDrafts() async {
    final drafts = await _readIndex();
    final valid = <Draft>[];
    final missingIds = <String>[];

    for (final d in drafts) {
      final f = File(d.imagePath);
      if (await f.exists()) {
        valid.add(d);
      } else {
        missingIds.add(d.id);
      }
    }

    if (missingIds.isNotEmpty) {
      final cleaned = drafts.where((d) => !missingIds.contains(d.id)).toList();
      await _writeIndex(cleaned);
    }

    valid.sort((a, b) => b.createdAt.compareTo(a.createdAt));
    return valid;
  }

  /// Returns a single draft by id, or null if not found.
  Future<Draft?> getDraft(String id) async {
    final drafts = await _readIndex();
    try {
      return drafts.firstWhere((d) => d.id == id);
    } catch (_) {
      return null;
    }
  }

  /// Overwrites the draft's image and caption. Use when "Save draft" on existing draft.
  Future<void> updateDraft(String id, File image, String? caption) async {
    final drafts = await _readIndex();
    final i = drafts.indexWhere((d) => d.id == id);
    if (i < 0) return;

    final dir = await _getDraftsDir();
    final destPath = '${dir.path}/$id.jpg';
    final dest = File(destPath);
    if (await dest.exists()) await dest.delete();
    await image.copy(destPath);

    drafts[i] = Draft(
      id: id,
      imagePath: destPath,
      caption: caption?.trim().isEmpty ?? true ? null : caption!.trim(),
      createdAt: drafts[i].createdAt,
    );
    await _writeIndex(drafts);
  }

  /// Removes draft from index and deletes its image file.
  Future<void> deleteDraft(String id) async {
    final drafts = await _readIndex();
    final i = drafts.indexWhere((x) => x.id == id);
    if (i >= 0) {
      final d = drafts[i];
      final f = File(d.imagePath);
      if (await f.exists()) await f.delete();
      drafts.removeAt(i);
      await _writeIndex(drafts);
    }
  }
}
