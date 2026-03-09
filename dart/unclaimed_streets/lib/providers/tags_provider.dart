import 'package:flutter/foundation.dart';

import '../models/tag.dart';
import '../services/tags_service.dart';

class TagsProvider with ChangeNotifier {
  final TagsService _service;

  TagsProvider(this._service);

  List<Tag> _tags = [];
  List<TagGroup> _tagGroups = [];
  Set<String> _selectedTagIds = {};
  bool _loading = false;
  bool _hasInitializedSelection = false;

  List<Tag> get tags => _tags;
  List<TagGroup> get tagGroups => _tagGroups;
  Set<String> get selectedTagIds => Set.from(_selectedTagIds);
  bool get loading => _loading;

  Future<void> refresh() async {
    _loading = true;
    notifyListeners();
    try {
      _tags = await _service.getTags();
      _tagGroups = await _service.getTagGroups();
      if (!_hasInitializedSelection) {
        _selectedTagIds = {};
        _hasInitializedSelection = true;
      }
    } catch (_) {
      _tags = [];
      _tagGroups = [];
    }
    _loading = false;
    notifyListeners();
  }

  void toggleTag(String id) {
    if (_selectedTagIds.contains(id)) {
      _selectedTagIds.remove(id);
    } else {
      _selectedTagIds.add(id);
    }
    notifyListeners();
  }

  void clearFilters() {
    _selectedTagIds.clear();
    notifyListeners();
  }

  void selectAll() {
    _selectedTagIds = _tags.map((t) => t.id).toSet();
    _hasInitializedSelection = true;
    notifyListeners();
  }

  void deselectAll() {
    _selectedTagIds.clear();
    _hasInitializedSelection = true;
    notifyListeners();
  }
}
