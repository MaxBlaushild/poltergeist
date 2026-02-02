import 'package:flutter/material.dart';
import 'package:skunkworks/constants/api_constants.dart';
import 'package:skunkworks/services/api_client.dart';
import 'package:skunkworks/constants/app_colors.dart';
import 'package:skunkworks/services/album_service.dart';
import 'package:skunkworks/widgets/bottom_nav.dart';
import 'package:skunkworks/screens/album_detail_screen.dart';

class AlbumInvitesScreen extends StatefulWidget {
  final Function(NavTab) onNavigate;

  const AlbumInvitesScreen({super.key, required this.onNavigate});

  @override
  State<AlbumInvitesScreen> createState() => _AlbumInvitesScreenState();
}

class _AlbumInvitesScreenState extends State<AlbumInvitesScreen> {
  List<dynamic> _invites = [];
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final apiClient = APIClient(ApiConstants.baseUrl);
      final albumService = AlbumService(apiClient);
      final invites = await albumService.getMyAlbumInvites();
      if (mounted) {
        setState(() {
          _invites = invites;
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

  Future<void> _accept(String inviteId) async {
    try {
      await AlbumService(APIClient(ApiConstants.baseUrl)).acceptAlbumInvite(inviteId);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Invite accepted')));
        _load();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Failed: $e')));
      }
    }
  }

  Future<void> _reject(String inviteId) async {
    try {
      await AlbumService(APIClient(ApiConstants.baseUrl)).rejectAlbumInvite(inviteId);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Invite rejected')));
        _load();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Failed: $e')));
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
          'Album Invites',
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
                      TextButton(onPressed: _load, child: const Text('Retry')),
                    ],
                  ),
                )
              : _invites.isEmpty
                  ? Center(
                      child: Text(
                        'No pending invites',
                        style: TextStyle(fontSize: 16, color: Colors.grey.shade600),
                      ),
                    )
                  : ListView.builder(
                      padding: const EdgeInsets.all(16),
                      itemCount: _invites.length,
                      itemBuilder: (context, index) {
                        final inv = _invites[index] as Map<String, dynamic>;
                        final album = inv['album'] as Map<String, dynamic>?;
                        final inviter = inv['inviter'] as Map<String, dynamic>?;
                        final albumName = album?['name'] as String? ?? 'Unknown album';
                        final inviterName = inviter?['username'] ?? inviter?['phoneNumber'] ?? 'Someone';
                        final inviteId = inv['id']?.toString();
                        if (inviteId == null) return const SizedBox.shrink();
                        return Card(
                          margin: const EdgeInsets.only(bottom: 12),
                          child: ListTile(
                            title: Text(albumName, style: const TextStyle(fontWeight: FontWeight.w600)),
                            subtitle: Text('$inviterName invited you'),
                            trailing: Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                TextButton(
                                  onPressed: () => _reject(inviteId),
                                  child: const Text('Reject'),
                                ),
                                FilledButton(
                                  onPressed: () async {
                                    await _accept(inviteId);
                                    if (mounted && album?['id'] != null) {
                                      Navigator.pushReplacement(
                                        context,
                                        MaterialPageRoute(
                                          builder: (context) => AlbumDetailScreen(
                                            albumId: album!['id'].toString(),
                                            albumName: albumName,
                                            onNavigate: widget.onNavigate,
                                          ),
                                        ),
                                      );
                                    }
                                  },
                                  child: const Text('Accept'),
                                ),
                              ],
                            ),
                          ),
                        );
                      },
                    ),
      bottomNavigationBar: BottomNav(
        currentTab: NavTab.home,
        onTabChanged: widget.onNavigate,
      ),
    );
  }
}
