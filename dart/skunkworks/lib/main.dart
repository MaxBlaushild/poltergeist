import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:skunkworks/constants/api_constants.dart';
import 'package:skunkworks/providers/auth_provider.dart';
import 'package:skunkworks/providers/friend_provider.dart';
import 'package:skunkworks/providers/post_provider.dart';
import 'package:skunkworks/screens/feed_screen.dart';
import 'package:skunkworks/screens/login_screen.dart';
import 'package:skunkworks/screens/profile_screen.dart';
import 'package:skunkworks/screens/search_screen.dart';
import 'package:skunkworks/screens/upload_post_screen.dart';
import 'package:skunkworks/services/api_client.dart';
import 'package:skunkworks/services/auth_service.dart';
import 'package:skunkworks/services/friend_service.dart';
import 'package:skunkworks/services/post_service.dart';
import 'package:skunkworks/widgets/bottom_nav.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    // Initialize services
    final apiClient = APIClient(ApiConstants.baseUrl);
    final authService = AuthService(apiClient);
    final authProvider = AuthProvider(authService);
    final postService = PostService(apiClient);
    final postProvider = PostProvider(postService);
    final friendService = FriendService(apiClient);
    final friendProvider = FriendProvider(friendService);

    // Set up auth error callback to log out user when 401/403 occurs
    apiClient.setOnAuthError(() {
      authProvider.logout();
    });

    return MultiProvider(
      providers: [
        ChangeNotifierProvider.value(value: authProvider),
        ChangeNotifierProvider.value(value: postProvider),
        ChangeNotifierProvider.value(value: friendProvider),
      ],
      child: MaterialApp(
        title: 'Verifiable SN',
        theme: ThemeData(
          colorScheme: ColorScheme.fromSeed(seedColor: Colors.blue),
          useMaterial3: true,
        ),
        home: const HomeWidget(),
      ),
    );
  }
}

class HomeWidget extends StatefulWidget {
  const HomeWidget({super.key});

  @override
  State<HomeWidget> createState() => _HomeWidgetState();
}

class _HomeWidgetState extends State<HomeWidget> {
  NavTab _currentTab = NavTab.home;

  void _onTabChanged(NavTab tab) {
    setState(() {
      _currentTab = tab;
    });
  }

  Widget _getCurrentScreen() {
    switch (_currentTab) {
      case NavTab.home:
        return FeedScreen(onNavigate: _onTabChanged);
      case NavTab.search:
        return SearchScreen(onNavigate: _onTabChanged);
      case NavTab.upload:
        return UploadPostScreen(onNavigate: _onTabChanged);
      case NavTab.profile:
        return ProfileScreen(onNavigate: _onTabChanged);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Consumer<AuthProvider>(
      builder: (context, authProvider, child) {
        // Show loading screen while checking authentication
        if (authProvider.loading) {
          return const Scaffold(
            body: Center(
              child: CircularProgressIndicator(),
            ),
          );
        }

        // Show login screen if not authenticated
        if (!authProvider.isAuthenticated) {
          return const LoginScreen();
        }

        // Show main app with navigation
        return _getCurrentScreen();
      },
    );
  }
}