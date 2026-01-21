import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:skunkworks/providers/post_provider.dart';
import 'package:skunkworks/widgets/bottom_nav.dart';
import 'package:skunkworks/widgets/post_card.dart';

class FeedScreen extends StatefulWidget {
  final Function(NavTab) onNavigate;

  const FeedScreen({
    super.key,
    required this.onNavigate,
  });

  @override
  State<FeedScreen> createState() => _FeedScreenState();
}

class _FeedScreenState extends State<FeedScreen> {
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
          'Verifiable SN',
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
            child: ListView.builder(
              itemCount: postProvider.feedPosts.length,
              itemBuilder: (context, index) {
                return PostCard(post: postProvider.feedPosts[index]);
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

