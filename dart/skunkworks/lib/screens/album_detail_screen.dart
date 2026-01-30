import 'package:flutter/material.dart';
import 'package:skunkworks/constants/api_constants.dart';
import 'package:skunkworks/services/api_client.dart';
import 'package:skunkworks/constants/app_colors.dart';
import 'package:skunkworks/models/post.dart';
import 'package:skunkworks/services/album_service.dart';
import 'package:skunkworks/widgets/bottom_nav.dart';
import 'package:skunkworks/widgets/post_card.dart';

class AlbumDetailScreen extends StatefulWidget {
  final String albumId;
  final String albumName;
  final Function(NavTab) onNavigate;

  const AlbumDetailScreen({
    super.key,
    required this.albumId,
    required this.albumName,
    required this.onNavigate,
  });

  @override
  State<AlbumDetailScreen> createState() => _AlbumDetailScreenState();
}

class _AlbumDetailScreenState extends State<AlbumDetailScreen> {
  List<Post> _posts = [];
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadAlbum();
  }

  Future<void> _loadAlbum() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final apiClient = APIClient(ApiConstants.baseUrl);
      final albumService = AlbumService(apiClient);
      final data = await albumService.getAlbum(widget.albumId);
      if (mounted) {
        setState(() {
          _posts = data['posts'] as List<Post>;
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

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.warmWhite,
      appBar: AppBar(
        backgroundColor: AppColors.warmWhite,
        elevation: 0,
        title: Text(
          widget.albumName,
          style: const TextStyle(
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
                        onPressed: _loadAlbum,
                        child: const Text('Retry'),
                      ),
                    ],
                  ),
                )
              : _posts.isEmpty
                  ? Center(
                      child: Text(
                        'No posts in this album yet.\nPosts with the album\'s tags will appear here.',
                        textAlign: TextAlign.center,
                        style: TextStyle(fontSize: 16, color: Colors.grey.shade600),
                      ),
                    )
                  : RefreshIndicator(
                      onRefresh: _loadAlbum,
                      child: ListView.builder(
                        padding: const EdgeInsets.only(bottom: 80),
                        itemCount: _posts.length,
                        itemBuilder: (context, index) {
                          final post = _posts[index];
                          return PostCard(
                            post: post,
                            onNavigate: widget.onNavigate,
                          );
                        },
                      ),
                    ),
      bottomNavigationBar: BottomNav(
        currentTab: NavTab.profile,
        onTabChanged: widget.onNavigate,
      ),
    );
  }
}
