import 'dart:async';
import 'dart:io';
import 'package:app_links/app_links.dart';
import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:skunkworks/constants/api_constants.dart';
import 'package:skunkworks/constants/app_colors.dart';
import 'package:skunkworks/providers/auth_provider.dart';
import 'package:skunkworks/providers/friend_provider.dart';
import 'package:skunkworks/providers/post_provider.dart';
import 'package:skunkworks/screens/certificate_registration_screen.dart';
import 'package:skunkworks/screens/albums_screen.dart';
import 'package:skunkworks/screens/login_screen.dart';
import 'package:skunkworks/screens/profile_screen.dart';
import 'package:skunkworks/screens/search_screen.dart';
import 'package:skunkworks/screens/post_detail_screen.dart';
import 'package:skunkworks/screens/upload_post_screen.dart';
import 'package:skunkworks/screens/notifications_screen.dart';
import 'package:skunkworks/screens/album_detail_screen.dart';
import 'package:skunkworks/screens/album_invites_screen.dart';
import 'package:skunkworks/screens/shared_album_screen.dart';
import 'package:skunkworks/services/api_client.dart';
import 'package:skunkworks/services/auth_service.dart';
import 'package:skunkworks/services/certificate_service.dart';
import 'package:skunkworks/services/friend_service.dart';
import 'package:skunkworks/services/post_service.dart';
import 'package:skunkworks/widgets/bottom_nav.dart';
import 'package:skunkworks/providers/certificate_provider.dart';
import 'package:skunkworks/providers/notification_provider.dart';
import 'package:skunkworks/services/notification_service.dart';

final GlobalKey<NavigatorState> navigatorKey = GlobalKey<NavigatorState>();

@pragma('vm:entry-point')
Future<void> _firebaseMessagingBackgroundHandler(RemoteMessage message) async {
  await Firebase.initializeApp();
}

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  try {
    await Firebase.initializeApp();
    FirebaseMessaging.onBackgroundMessage(_firebaseMessagingBackgroundHandler);
  } catch (_) {
    // Firebase not configured - app works without push
  }
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
    final notificationService = NotificationService(apiClient);
    final notificationProvider = NotificationProvider(notificationService);

    // Set up auth error callback to log out user when 401/403 occurs
    apiClient.setOnAuthError(() {
      authProvider.logout();
    });

    return MultiProvider(
      providers: [
        Provider.value(value: apiClient),
        ChangeNotifierProvider.value(value: authProvider),
        ChangeNotifierProvider.value(value: postProvider),
        ChangeNotifierProvider.value(value: friendProvider),
        ChangeNotifierProvider.value(value: certificateProvider),
        ChangeNotifierProvider.value(value: notificationProvider),
      ],
      child: MaterialApp(
        navigatorKey: navigatorKey,
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
  bool _hasInitializedFcm = false;
  StreamSubscription<Uri>? _linkSubscription;
  final AppLinks _appLinks = AppLinks();
  String? _pendingSharedAlbumToken;

  @override
  void initState() {
    super.initState();
    _initDeepLinks();
  }

  Future<void> _initFcm(BuildContext context) async {
    if (_hasInitializedFcm) return;
    _hasInitializedFcm = true;
    try {
      final notificationProvider = context.read<NotificationProvider>();
      final messaging = FirebaseMessaging.instance;
      if (Platform.isIOS) {
        await messaging.requestPermission(alert: true, badge: true, sound: true);
      }
      final token = await messaging.getToken();
      if (token != null && context.mounted) {
        final apiClient = APIClient(ApiConstants.baseUrl);
        final platform = Platform.isIOS ? 'ios' : 'android';
        await NotificationService(apiClient).registerDeviceToken(token, platform);
      }
      FirebaseMessaging.onMessage.listen((RemoteMessage message) {
        notificationProvider.loadNotifications();
      });
      FirebaseMessaging.onMessageOpenedApp.listen((RemoteMessage message) {
        _handleNotificationTap(message.data);
      });
      final initialMessage = await FirebaseMessaging.instance.getInitialMessage();
      if (initialMessage != null && mounted) {
        WidgetsBinding.instance.addPostFrameCallback((_) {
          _handleNotificationTap(initialMessage.data);
        });
      }
    } catch (_) {}
  }

  void _handleNotificationTap(Map<String, dynamic> data) {
    final type = data['type'] as String?;
    final albumId = data['albumId'] as String?;
    if (navigatorKey.currentState == null) return;
    if (type == 'album_invite') {
      navigatorKey.currentState!.push(
        MaterialPageRoute(
          builder: (context) => AlbumInvitesScreen(onNavigate: _onTabChanged),
        ),
      );
    } else if (albumId != null) {
      navigatorKey.currentState!.push(
        MaterialPageRoute(
          builder: (context) => AlbumDetailScreen(
            albumId: albumId,
            albumName: 'Album',
            onNavigate: _onTabChanged,
          ),
        ),
      );
    } else {
      navigatorKey.currentState!.push(
        MaterialPageRoute(
          builder: (context) => NotificationsScreen(onNavigate: _onTabChanged),
        ),
      );
    }
  }

  @override
  void dispose() {
    _linkSubscription?.cancel();
    super.dispose();
  }

  Future<void> _initDeepLinks() async {
    // Handle link when app is opened via deep link
    final initialUri = await _appLinks.getInitialLink();
    if (initialUri != null && mounted) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted) _handleDeepLink(initialUri);
      });
    }
    // Handle links when app is in background
    _linkSubscription = _appLinks.uriLinkStream.listen((uri) {
      if (mounted) _handleDeepLink(uri);
    });
  }

  void _handleDeepLink(Uri uri) {
    if (uri.scheme != 'vera') return;
    final pathSegments = uri.pathSegments;
    if (pathSegments.isEmpty) return;

    if (uri.host == 'post') {
      final postId = pathSegments.first;
      if (postId.isEmpty) return;
      navigatorKey.currentState?.push(
        MaterialPageRoute(
          builder: (context) => PostDetailScreen(
            postId: postId,
            onNavigate: _onTabChanged,
          ),
        ),
      );
      return;
    }

    if (uri.host == 'album') {
      final token = pathSegments.first;
      if (token.isEmpty) return;
      final navContext = navigatorKey.currentContext;
      if (navContext != null) {
        final authProvider = Provider.of<AuthProvider>(navContext, listen: false);
        if (authProvider.isAuthenticated) {
          navigatorKey.currentState?.push(
            MaterialPageRoute(
              builder: (context) => SharedAlbumScreen(
                shareToken: token,
                onNavigate: _onTabChanged,
              ),
            ),
          );
          return;
        }
      }
      setState(() {
        _pendingSharedAlbumToken = token;
      });
    }
  }

  void _onTabChanged(NavTab tab) {
    setState(() {
      _currentTab = tab;
    });
  }

  Widget _getCurrentScreen() {
    switch (_currentTab) {
      case NavTab.home:
        return AlbumsScreen(onNavigate: _onTabChanged);
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
        if (_pendingSharedAlbumToken != null) {
          return SharedAlbumScreen(
            shareToken: _pendingSharedAlbumToken!,
            onNavigate: (tab) {
              setState(() {
                _pendingSharedAlbumToken = null;
                _currentTab = tab;
              });
            },
          );
        }
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
        WidgetsBinding.instance.addPostFrameCallback((_) {
          if (mounted) _initFcm(context);
        });
        return _getCurrentScreen();
      },
    );
  }
}
