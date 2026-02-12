import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:skunkworks/constants/app_colors.dart';
import 'package:skunkworks/providers/post_provider.dart';
import 'package:skunkworks/screens/drafts_screen.dart';
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

  Widget _buildDraftsSection(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
      child: Card(
        elevation: 0,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
          side: BorderSide(color: AppColors.graphiteInk.withOpacity(0.08)),
        ),
        child: InkWell(
          borderRadius: BorderRadius.circular(16),
          onTap: () {
            Navigator.of(context).push(
              MaterialPageRoute(builder: (_) => const DraftsScreen()),
            );
          },
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
            child: Row(
              children: [
                Container(
                  width: 40,
                  height: 40,
                  decoration: BoxDecoration(
                    color: AppColors.coralPop.withOpacity(0.12),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: const Icon(
                    Icons.drafts_outlined,
                    color: AppColors.coralPop,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        'Drafts',
                        style: TextStyle(
                          color: AppColors.graphiteInk,
                          fontWeight: FontWeight.w600,
                          fontSize: 16,
                        ),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        'View your image drafts',
                        style: TextStyle(
                          color: AppColors.graphiteInk.withOpacity(0.6),
                          fontSize: 13,
                        ),
                      ),
                    ],
                  ),
                ),
                const Icon(
                  Icons.chevron_right,
                  color: AppColors.graphiteInk,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.warmWhite,
      appBar: AppBar(
        backgroundColor: AppColors.warmWhite,
        elevation: 0,
        title: const Text(
          'Vera',
          style: TextStyle(
            color: AppColors.graphiteInk,
            fontWeight: FontWeight.w400,
            letterSpacing: -1.0,
            fontSize: 24,
          ),
        ),
        centerTitle: false,
      ),
      body: Consumer<PostProvider>(
        builder: (context, postProvider, child) {
          final hasPosts = postProvider.feedPosts.isNotEmpty;
          final showLoading = postProvider.loading && !hasPosts;
          final showError = postProvider.error != null && !hasPosts;

          final itemCount = hasPosts ? postProvider.feedPosts.length + 1 : 2;

          return RefreshIndicator(
            onRefresh: _refreshFeed,
            child: ListView.builder(
              physics: const AlwaysScrollableScrollPhysics(),
              itemCount: itemCount,
              itemBuilder: (context, index) {
                final draftsIndex = itemCount - 1;
                if (index == draftsIndex) {
                  return Padding(
                    padding: const EdgeInsets.only(bottom: 12),
                    child: _buildDraftsSection(context),
                  );
                }

                if (showLoading) {
                  return const Padding(
                    padding: EdgeInsets.only(top: 32),
                    child: Center(child: CircularProgressIndicator()),
                  );
                }

                if (showError) {
                  return Padding(
                    padding: const EdgeInsets.only(top: 24),
                    child: Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Text(
                            'Error loading feed: ${postProvider.error}',
                            style: TextStyle(color: AppColors.coralPop),
                          ),
                          const SizedBox(height: 16),
                          ElevatedButton(
                            onPressed: _refreshFeed,
                            child: const Text('Retry'),
                          ),
                        ],
                      ),
                    ),
                  );
                }

                if (!hasPosts) {
                  return Padding(
                    padding: const EdgeInsets.only(top: 16),
                    child: Center(
                      child: Text(
                        'No posts yet.\nFollow friends to see their posts!',
                        textAlign: TextAlign.center,
                        style: TextStyle(
                          color: AppColors.graphiteInk.withOpacity(0.6),
                        ),
                      ),
                    ),
                  );
                }

                final postIndex = index;
                return PostCard(
                  post: postProvider.feedPosts[postIndex],
                  onNavigate: widget.onNavigate,
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
