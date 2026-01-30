import 'package:flutter/material.dart';
import 'package:skunkworks/constants/api_constants.dart';
import 'package:skunkworks/services/api_client.dart';
import 'package:skunkworks/constants/app_colors.dart';
import 'package:skunkworks/models/album.dart';
import 'package:skunkworks/services/album_service.dart';
import 'package:skunkworks/widgets/bottom_nav.dart';
import 'package:skunkworks/screens/album_detail_screen.dart';

class AlbumsScreen extends StatefulWidget {
  final Function(NavTab) onNavigate;

  const AlbumsScreen({super.key, required this.onNavigate});

  @override
  State<AlbumsScreen> createState() => _AlbumsScreenState();
}

class _AlbumsScreenState extends State<AlbumsScreen> {
  List<Album> _albums = [];
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadAlbums();
  }

  Future<void> _loadAlbums() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final apiClient = APIClient(ApiConstants.baseUrl);
      final albumService = AlbumService(apiClient);
      final albums = await albumService.getAlbums();
      if (mounted) {
        setState(() {
          _albums = albums;
          _loading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = e.toString();
          _loading = false;
        });
      }
    }
  }

  Future<void> _showCreateAlbumDialog() async {
    final nameController = TextEditingController();
    final tagController = TextEditingController();
    List<String> tags = [];

    final result = await showDialog<bool>(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setDialogState) => AlertDialog(
          title: const Text('New Album'),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                TextField(
                  controller: nameController,
                  decoration: const InputDecoration(
                    labelText: 'Album name',
                    border: OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: 16),
                Text(
                  'Tags',
                  style: TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                    color: Colors.grey.shade700,
                  ),
                ),
                const SizedBox(height: 8),
                Row(
                  children: [
                    Expanded(
                      child: TextField(
                        controller: tagController,
                        decoration: const InputDecoration(
                          hintText: 'Add a tag',
                          border: OutlineInputBorder(),
                          isDense: true,
                        ),
                        onSubmitted: (_) {
                          final t = tagController.text.trim();
                          if (t.isNotEmpty && !tags.contains(t)) {
                            setDialogState(() {
                              tags = [...tags, t];
                              tagController.clear();
                            });
                          }
                        },
                      ),
                    ),
                    const SizedBox(width: 8),
                    IconButton.filled(
                      onPressed: () {
                        final t = tagController.text.trim();
                        if (t.isNotEmpty && !tags.contains(t)) {
                          setDialogState(() {
                            tags = [...tags, t];
                            tagController.clear();
                          });
                        }
                      },
                      icon: const Icon(Icons.add),
                    ),
                  ],
                ),
                if (tags.isNotEmpty) ...[
                  const SizedBox(height: 8),
                  Wrap(
                    spacing: 6,
                    runSpacing: 6,
                    children: tags.map((tag) => Chip(
                      label: Text(tag),
                      deleteIcon: const Icon(Icons.close, size: 18),
                      onDeleted: () => setDialogState(() => tags = tags.where((t) => t != tag).toList()),
                    )).toList(),
                  ),
                ],
              ],
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () {
                if (nameController.text.trim().isEmpty) return;
                if (tags.isEmpty) return;
                Navigator.pop(context, true);
              },
              child: const Text('Create'),
            ),
          ],
        ),
      ),
    );

    if (result == true && mounted) {
      final name = nameController.text.trim();
      if (name.isEmpty || tags.isEmpty) return;
      try {
        final apiClient = APIClient(ApiConstants.baseUrl);
        final albumService = AlbumService(apiClient);
        await albumService.createAlbum(name, tags);
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Album created')),
          );
          _loadAlbums();
        }
      } catch (e) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Failed to create album: $e')),
          );
        }
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.warmWhite,
      appBar: AppBar(
        backgroundColor: AppColors.warmWhite,
        elevation: 0,
        title: const Text(
          'Albums',
          style: TextStyle(
            color: AppColors.graphiteInk,
            fontWeight: FontWeight.w600,
            fontSize: 18,
          ),
        ),
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(_error!, textAlign: TextAlign.center, style: TextStyle(color: Colors.grey.shade700)),
                      const SizedBox(height: 16),
                      TextButton(
                        onPressed: _loadAlbums,
                        child: const Text('Retry'),
                      ),
                    ],
                  ),
                )
              : _albums.isEmpty
                  ? Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(Icons.photo_album_outlined, size: 64, color: Colors.grey.shade400),
                          const SizedBox(height: 16),
                          Text(
                            'No albums yet',
                            style: TextStyle(fontSize: 16, color: Colors.grey.shade600),
                          ),
                          const SizedBox(height: 8),
                          Text(
                            'Create an album and link it to tags.\nPosts with those tags will appear in the album.',
                            textAlign: TextAlign.center,
                            style: TextStyle(fontSize: 14, color: Colors.grey.shade500),
                          ),
                        ],
                      ),
                    )
                  : ListView.builder(
                      padding: const EdgeInsets.all(16),
                      itemCount: _albums.length,
                      itemBuilder: (context, index) {
                        final album = _albums[index];
                        return Card(
                          margin: const EdgeInsets.only(bottom: 12),
                          child: ListTile(
                            leading: CircleAvatar(
                              backgroundColor: AppColors.softRealBlue.withOpacity(0.2),
                              child: Icon(Icons.photo_album, color: AppColors.softRealBlue),
                            ),
                            title: Text(album.name),
                            subtitle: album.tags.isEmpty
                                ? null
                                : Text(
                                    album.tags.map((t) => '#$t').join(' '),
                                    style: TextStyle(fontSize: 12, color: Colors.grey.shade600),
                                  ),
                            trailing: const Icon(Icons.chevron_right),
                            onTap: () {
                              if (album.id != null) {
                                Navigator.push(
                                  context,
                                  MaterialPageRoute(
                                    builder: (context) => AlbumDetailScreen(
                                      albumId: album.id!,
                                      albumName: album.name,
                                      onNavigate: widget.onNavigate,
                                    ),
                                  ),
                                );
                              }
                            },
                          ),
                        );
                      },
                    ),
      floatingActionButton: FloatingActionButton(
        onPressed: _showCreateAlbumDialog,
        child: const Icon(Icons.add),
      ),
      bottomNavigationBar: BottomNav(
        currentTab: NavTab.profile,
        onTabChanged: widget.onNavigate,
      ),
    );
  }
}
