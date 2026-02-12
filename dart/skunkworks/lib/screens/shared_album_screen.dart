import 'package:flutter/material.dart';
import 'package:skunkworks/constants/api_constants.dart';
import 'package:skunkworks/constants/app_colors.dart';
import 'package:skunkworks/models/album.dart';
import 'package:skunkworks/models/post.dart';
import 'package:skunkworks/services/api_client.dart';
import 'package:skunkworks/services/album_service.dart';
import 'package:skunkworks/widgets/bottom_nav.dart';
import 'package:skunkworks/widgets/post_card.dart';

class SharedAlbumScreen extends StatefulWidget {
  final String shareToken;
  final Function(NavTab) onNavigate;

  const SharedAlbumScreen({
    super.key,
    required this.shareToken,
    required this.onNavigate,
  });

  @override
  State<SharedAlbumScreen> createState() => _SharedAlbumScreenState();
}

class _SharedAlbumScreenState extends State<SharedAlbumScreen> {
  Album? _album;
  List<Post> _posts = [];
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadSharedAlbum();
  }

  Future<void> _loadSharedAlbum() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final apiClient = APIClient(ApiConstants.baseUrl);
      final albumService = AlbumService(apiClient);
      final data = await albumService.getSharedAlbum(widget.shareToken);
      if (mounted) {
        setState(() {
          _album = data['album'] as Album;
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
    final title = _album?.name ?? 'Shared Album';
    return Scaffold(
      backgroundColor: AppColors.warmWhite,
      appBar: AppBar(
        backgroundColor: AppColors.warmWhite,
        elevation: 0,
        title: Text(
          title,
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
                      TextButton(onPressed: _loadSharedAlbum, child: const Text('Retry')),
                    ],
                  ),
                )
              : RefreshIndicator(
                  onRefresh: _loadSharedAlbum,
                  child: SingleChildScrollView(
                    physics: const AlwaysScrollableScrollPhysics(),
                    padding: const EdgeInsets.only(bottom: 80),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        if (_posts.isEmpty)
                          Padding(
                            padding: const EdgeInsets.all(24),
                            child: Center(
                              child: Text(
                                'No posts in this album yet.',
                                textAlign: TextAlign.center,
                                style: TextStyle(fontSize: 16, color: Colors.grey.shade600),
                              ),
                            ),
                          )
                        else
                          ..._posts.map((post) => PostCard(post: post, onNavigate: widget.onNavigate)),
                      ],
                    ),
                  ),
                ),
      bottomNavigationBar: BottomNav(
        currentTab: NavTab.profile,
        onTabChanged: widget.onNavigate,
      ),
    );
  }
}
