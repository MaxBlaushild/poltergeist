import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:skunkworks/providers/post_provider.dart';
import 'package:skunkworks/screens/post_detail_screen.dart';
import 'package:skunkworks/widgets/bottom_nav.dart';

class PictureShowScreen extends StatefulWidget {
  final Function(NavTab) onNavigate;

  const PictureShowScreen({
    super.key,
    required this.onNavigate,
  });

  @override
  State<PictureShowScreen> createState() => _PictureShowScreenState();
}

class _PictureShowScreenState extends State<PictureShowScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<PostProvider>().loadFeed();
    });
  }

  Future<void> _refreshFeed() async {
    await context.read<PostProvider>().loadFeed();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.white,
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        title: const Text(
          'Picture Show',
          style: TextStyle(
            color: Colors.black,
            fontWeight: FontWeight.w400,
            letterSpacing: -1.0,
            fontSize: 24,
          ),
        ),
        centerTitle: false,
      ),
      body: Consumer<PostProvider>(
        builder: (context, postProvider, child) {
          if (postProvider.loading && postProvider.feedPosts.isEmpty) {
            return const Center(
              child: CircularProgressIndicator(),
            );
          }

          if (postProvider.error != null && postProvider.feedPosts.isEmpty) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(
                    'Error loading feed: ${postProvider.error}',
                    style: const TextStyle(color: Colors.red),
                  ),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: _refreshFeed,
                    child: const Text('Retry'),
                  ),
                ],
              ),
            );
          }

          if (postProvider.feedPosts.isEmpty) {
            return RefreshIndicator(
              onRefresh: _refreshFeed,
              child: SingleChildScrollView(
                physics: const AlwaysScrollableScrollPhysics(),
                child: SizedBox(
                  height: MediaQuery.of(context).size.height - 200,
                  child: const Center(
                    child: Text(
                      'No posts yet.\nFollow friends to see their posts!',
                      textAlign: TextAlign.center,
                      style: TextStyle(color: Colors.grey),
                    ),
                  ),
                ),
              ),
            );
          }

          return RefreshIndicator(
            onRefresh: _refreshFeed,
            child: GridView.builder(
              padding: const EdgeInsets.all(2),
              gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                crossAxisCount: 3,
                crossAxisSpacing: 2,
                mainAxisSpacing: 2,
              ),
              itemCount: postProvider.feedPosts.length,
              itemBuilder: (context, index) {
                final post = postProvider.feedPosts[index];
                return GestureDetector(
                  onTap: () {
                    if (post.id != null) {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) => PostDetailScreen(
                            postId: post.id!,
                            onNavigate: widget.onNavigate,
                          ),
                        ),
                      );
                    }
                  },
                  child: post.imageUrl != null
                      ? Stack(
                          fit: StackFit.expand,
                          children: [
                            Image.network(
                              post.imageUrl!,
                              fit: BoxFit.cover,
                              loadingBuilder: (context, child, loadingProgress) {
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
                              errorBuilder: (context, error, stackTrace) {
                                return Container(
                                  color: Colors.grey.shade200,
                                  child: const Icon(Icons.error),
                                );
                              },
                            ),
                            if (post.isVideo)
                              Container(
                                color: Colors.black.withOpacity(0.3),
                                child: const Center(
                                  child: Icon(
                                    Icons.play_circle_filled,
                                    color: Colors.white,
                                    size: 40,
                                  ),
                                ),
                              ),
                          ],
                        )
                      : Container(
                          color: Colors.grey.shade200,
                          child: const Icon(Icons.image),
                        ),
                );
              },
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
