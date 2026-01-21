import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:skunkworks/models/post.dart';
import 'package:skunkworks/providers/auth_provider.dart';
import 'package:skunkworks/services/post_service.dart';
import 'package:skunkworks/services/api_client.dart';
import 'package:skunkworks/constants/api_constants.dart';
import 'package:skunkworks/widgets/bottom_nav.dart';

class ProfileScreen extends StatefulWidget {
  final Function(NavTab) onNavigate;

  const ProfileScreen({
    super.key,
    required this.onNavigate,
  });

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  List<Post> _userPosts = [];
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _loadUserPosts();
  }

  Future<void> _loadUserPosts() async {
    final authProvider = context.read<AuthProvider>();
    final user = authProvider.user;

    if (user?.id == null) return;

    setState(() {
      _loading = true;
    });

    try {
      final apiClient = APIClient(ApiConstants.baseUrl);
      final postService = PostService(apiClient);
      _userPosts = await postService.getUserPosts(user!.id!);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to load posts: $e')),
        );
      }
    } finally {
      if (mounted) {
        setState(() {
          _loading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Consumer<AuthProvider>(
      builder: (context, authProvider, child) {
        final user = authProvider.user;
        final username = user?.username ?? user?.phoneNumber ?? 'Unknown';
        final profilePictureUrl = user?.profilePictureUrl;

        return Scaffold(
          backgroundColor: Colors.white,
          appBar: AppBar(
            backgroundColor: Colors.white,
            elevation: 0,
            title: Text(
              username,
              style: const TextStyle(
                color: Colors.black,
                fontWeight: FontWeight.w600,
                fontSize: 18,
              ),
            ),
            actions: [
              IconButton(
                icon: const Icon(Icons.logout, color: Colors.black),
                onPressed: () async {
                  await authProvider.logout();
                },
              ),
            ],
          ),
          body: _loading
              ? const Center(child: CircularProgressIndicator())
              : SingleChildScrollView(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // Profile header
                      Padding(
                        padding: const EdgeInsets.all(16.0),
                        child: Row(
                          children: [
                            CircleAvatar(
                              radius: 40,
                              backgroundColor: Colors.grey.shade300,
                              backgroundImage: profilePictureUrl != null
                                  ? NetworkImage(profilePictureUrl)
                                  : null,
                              child: profilePictureUrl == null
                                  ? Text(
                                      username.isNotEmpty
                                          ? username[0].toUpperCase()
                                          : 'U',
                                      style: const TextStyle(
                                        fontSize: 32,
                                        color: Colors.grey,
                                      ),
                                    )
                                  : null,
                            ),
                            const SizedBox(width: 20),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    username,
                                    style: const TextStyle(
                                      fontWeight: FontWeight.w600,
                                      fontSize: 16,
                                    ),
                                  ),
                                  const SizedBox(height: 4),
                                  Text(
                                    '${_userPosts.length} posts',
                                    style: TextStyle(
                                      color: Colors.grey.shade600,
                                      fontSize: 14,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
                      ),
                      const Divider(),
                      
                      // Posts grid
                      if (_userPosts.isEmpty)
                        const Padding(
                          padding: EdgeInsets.all(32.0),
                          child: Center(
                            child: Text(
                              'No posts yet.\nShare your first photo!',
                              textAlign: TextAlign.center,
                              style: TextStyle(color: Colors.grey),
                            ),
                          ),
                        )
                      else
                        GridView.builder(
                          shrinkWrap: true,
                          physics: const NeverScrollableScrollPhysics(),
                          gridDelegate:
                              const SliverGridDelegateWithFixedCrossAxisCount(
                            crossAxisCount: 3,
                            crossAxisSpacing: 2,
                            mainAxisSpacing: 2,
                          ),
                          itemCount: _userPosts.length,
                          itemBuilder: (context, index) {
                            final post = _userPosts[index];
                            return GestureDetector(
                              onTap: () {
                                // Could show post details in a dialog
                              },
                              child: post.imageUrl != null
                                  ? Image.network(
                                      post.imageUrl!,
                                      fit: BoxFit.cover,
                                      loadingBuilder:
                                          (context, child, loadingProgress) {
                                        if (loadingProgress == null) {
                                          return child;
                                        }
                                        return Container(
                                          color: Colors.grey.shade200,
                                          child: const Center(
                                            child: CircularProgressIndicator(),
                                          ),
                                        );
                                      },
                                      errorBuilder:
                                          (context, error, stackTrace) {
                                        return Container(
                                          color: Colors.grey.shade200,
                                          child: const Icon(Icons.error),
                                        );
                                      },
                                    )
                                  : Container(
                                      color: Colors.grey.shade200,
                                      child: const Icon(Icons.image),
                                    ),
                            );
                          },
                        ),
                    ],
                  ),
                ),
          bottomNavigationBar: BottomNav(
            currentTab: NavTab.profile,
            onTabChanged: widget.onNavigate,
          ),
        );
      },
    );
  }
}

