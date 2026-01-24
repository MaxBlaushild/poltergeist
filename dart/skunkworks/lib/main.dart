import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:skunkworks/constants/api_constants.dart';
import 'package:skunkworks/constants/app_colors.dart';
import 'package:skunkworks/providers/auth_provider.dart';
import 'package:skunkworks/providers/friend_provider.dart';
import 'package:skunkworks/providers/post_provider.dart';
import 'package:skunkworks/screens/certificate_registration_screen.dart';
import 'package:skunkworks/screens/feed_screen.dart';
import 'package:skunkworks/screens/login_screen.dart';
import 'package:skunkworks/screens/profile_screen.dart';
import 'package:skunkworks/screens/search_screen.dart';
import 'package:skunkworks/screens/upload_post_screen.dart';
import 'package:skunkworks/services/api_client.dart';
import 'package:skunkworks/services/auth_service.dart';
import 'package:skunkworks/services/certificate_service.dart';
import 'package:skunkworks/services/friend_service.dart';
import 'package:skunkworks/services/post_service.dart';
import 'package:skunkworks/widgets/bottom_nav.dart';
import 'package:skunkworks/providers/certificate_provider.dart';

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
    final certificateService = CertificateService(apiClient);
    final certificateProvider = CertificateProvider(certificateService);

    // Set up auth error callback to log out user when 401/403 occurs
    apiClient.setOnAuthError(() {
      authProvider.logout();
    });

    return MultiProvider(
      providers: [
        ChangeNotifierProvider.value(value: authProvider),
        ChangeNotifierProvider.value(value: postProvider),
        ChangeNotifierProvider.value(value: friendProvider),
        ChangeNotifierProvider.value(value: certificateProvider),
      ],
      child: MaterialApp(
        title: 'Vera',
        theme: ThemeData(
          useMaterial3: true,
          fontFamily: GoogleFonts.inter().fontFamily,
          textTheme: GoogleFonts.interTextTheme(),
          colorScheme: ColorScheme.light(
            primary: AppColors.softRealBlue,
            secondary: AppColors.freshMint,
            background: AppColors.warmWhite,
            surface: AppColors.warmWhite,
            onPrimary: AppColors.warmWhite,
            onSecondary: AppColors.graphiteInk,
            onBackground: AppColors.graphiteInk,
            onSurface: AppColors.graphiteInk,
            error: AppColors.coralPop,
            onError: AppColors.warmWhite,
          ),
          scaffoldBackgroundColor: AppColors.warmWhite,
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
  bool _hasCheckedCertificate = false;

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
    return Consumer2<AuthProvider, CertificateProvider>(
      builder: (context, authProvider, certProvider, child) {
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
          _hasCheckedCertificate = false; // Reset when logged out
          return const LoginScreen();
        }

        // Check certificate on app startup (if authenticated and not already checked)
        if (!_hasCheckedCertificate && !certProvider.loading) {
          _hasCheckedCertificate = true;
          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (mounted && authProvider.isAuthenticated) {
              certProvider.checkCertificate();
            }
          });
        }
        
        // Show loading while checking certificate
        if (certProvider.loading) {
          return const Scaffold(
            body: Center(
              child: CircularProgressIndicator(),
            ),
          );
        }
        
        // Show certificate registration screen if no certificate
        if (!certProvider.hasCertificate) {
          return const CertificateRegistrationScreen();
        }

        // Show main app with navigation if certificate exists
        return _getCurrentScreen();
      },
    );
  }
}