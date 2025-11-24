import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/models/user_level.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/providers/user_level_provider.dart';
import 'package:travel_angels/screens/documents_screen.dart';
import 'package:travel_angels/screens/my_network_screen.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/credits_service.dart';
import 'package:travel_angels/services/document_service.dart';
import 'package:travel_angels/widgets/credits_purchase_dialog.dart';
import 'package:travel_angels/widgets/permissions_panel.dart';

/// Profile screen for user profile and settings
class ProfileScreen extends StatefulWidget {
  const ProfileScreen({super.key});

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  final DocumentService _documentService = DocumentService(
    APIClient(ApiConstants.baseUrl),
  );
  final CreditsService _creditsService = CreditsService(
    APIClient(ApiConstants.baseUrl),
  );

  int _docsShared = 0;
  bool _isLoadingDocs = false;
  String? _lastLoadedUserId;

  /// Calculate progress percentage towards next level
  /// Returns a value between 0.0 and 1.0
  static double _calculateProgress(UserLevel userLevel) {
    final experiencePointsOnLevel = userLevel.experiencePointsOnLevel;
    final experienceToNextLevel = userLevel.experienceToNextLevel;

    if (experienceToNextLevel <= 0) return 1.0;
    if (experiencePointsOnLevel <= 0) return 0.0;
    if (experiencePointsOnLevel >= experienceToNextLevel) return 1.0;

    return experiencePointsOnLevel / experienceToNextLevel;
  }

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _loadDocumentCount();
    });
  }

  Future<void> _loadDocumentCount() async {
    final authProvider = context.read<AuthProvider>();
    final user = authProvider.user;

    if (!authProvider.isAuthenticated || user?.id == null) {
      return;
    }

    // Don't reload if we already loaded for this user
    if (_lastLoadedUserId == user?.id && user?.id != null) {
      return;
    }

    setState(() {
      _isLoadingDocs = true;
    });

    try {
      final documentsJson = await _documentService.getDocumentsByUserId(user!.id!);
      setState(() {
        _docsShared = documentsJson.length;
        _isLoadingDocs = false;
        _lastLoadedUserId = user.id;
      });
    } catch (e) {
      // Silently handle errors - don't show error state, just keep count at 0
      setState(() {
        _isLoadingDocs = false;
      });
    }
  }

  /// Get user initials for fallback avatar
  static String _getInitials(User? user) {
    if (user == null) return 'T';
    final name = user.name ?? user.username ?? '';
    if (name.isEmpty) return 'T';
    final parts = name.trim().split(' ');
    if (parts.length >= 2) {
      return '${parts[0][0]}${parts[1][0]}'.toUpperCase();
    }
    return name[0].toUpperCase();
  }

  /// Build profile picture widget
  Widget _buildProfilePicture(User? user, ThemeData theme) {
    final profilePictureUrl = user?.profilePictureUrl;
    
    if (profilePictureUrl != null && profilePictureUrl.isNotEmpty) {
      return CircleAvatar(
        radius: 50,
        backgroundImage: NetworkImage(profilePictureUrl),
        onBackgroundImageError: (exception, stackTrace) {
          // Handle error silently - will show fallback
        },
        child: profilePictureUrl.isEmpty ? _buildFallbackAvatar(user, theme) : null,
      );
    }
    
    return _buildFallbackAvatar(user, theme);
  }

  Widget _buildFallbackAvatar(User? user, ThemeData theme) {
    return CircleAvatar(
      radius: 50,
      backgroundColor: theme.colorScheme.primaryContainer,
      child: Text(
        _getInitials(user),
        style: TextStyle(
          fontSize: 32,
          fontWeight: FontWeight.bold,
          color: theme.colorScheme.onPrimaryContainer,
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final authProvider = context.watch<AuthProvider>();
    final user = authProvider.user;
    // Use select to only rebuild when userLevel or loading changes
    final userLevel = context.select<UserLevelProvider, UserLevel?>((provider) => provider.userLevel);
    final isLoading = context.select<UserLevelProvider, bool>((provider) => provider.loading);

    final userName = user?.name ?? user?.username ?? 'Traveler';
    
    // Reload document count if user changes
    if (user?.id != null && user?.id != _lastLoadedUserId && !_isLoadingDocs) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        _loadDocumentCount();
      });
    }

    // Use default values if user level is not loaded yet
    final level = userLevel?.level ?? 1;
    final experiencePointsOnLevel = userLevel?.experiencePointsOnLevel ?? 0;
    final experienceToNextLevel = userLevel?.experienceToNextLevel ?? 100;

    // Calculate experience progress
    double progress = 0.0;
    if (userLevel != null) {
      progress = _calculateProgress(userLevel);
    }

    return Scaffold(
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: [
            const SizedBox(height: 24),
            // Profile Picture
            _buildProfilePicture(user, theme),
            const SizedBox(height: 16),
            // User Name
            Text(
              userName,
              style: theme.textTheme.headlineSmall?.copyWith(
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 8),
            // Level Badge
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              decoration: BoxDecoration(
                color: theme.colorScheme.primaryContainer,
                borderRadius: BorderRadius.circular(20),
              ),
              child: Text(
                'Level $level',
                style: theme.textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                  color: theme.colorScheme.onPrimaryContainer,
                ),
              ),
            ),
            const SizedBox(height: 32),
            // Experience Bar Section
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          'Experience',
                          style: theme.textTheme.titleMedium?.copyWith(
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        Text(
                          '$experiencePointsOnLevel / $experienceToNextLevel XP',
                          style: theme.textTheme.bodyMedium?.copyWith(
                            color: theme.colorScheme.onSurface.withOpacity(0.7),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    isLoading
                        ? const LinearProgressIndicator()
                        : TweenAnimationBuilder<double>(
                            tween: Tween<double>(begin: 0.0, end: progress),
                            duration: const Duration(milliseconds: 500),
                            curve: Curves.easeOut,
                            builder: (context, animatedProgress, child) {
                              return LinearProgressIndicator(
                                value: animatedProgress,
                                minHeight: 8,
                                borderRadius: BorderRadius.circular(4),
                                backgroundColor: theme.colorScheme.surfaceContainerHighest,
                                valueColor: AlwaysStoppedAnimation<Color>(
                                  theme.colorScheme.primary,
                                ),
                              );
                            },
                          ),
                    const SizedBox(height: 8),
                    Text(
                      '${experiencePointsOnLevel.toStringAsFixed(0)} / $experienceToNextLevel XP to Level ${level + 1}',
                      style: theme.textTheme.bodySmall?.copyWith(
                        color: theme.colorScheme.onSurface.withOpacity(0.6),
                      ),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),
            // Docs Shared Section
            Card(
              child: InkWell(
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => const DocumentsScreen(),
                    ),
                  );
                },
                borderRadius: BorderRadius.circular(12),
                child: Padding(
                  padding: const EdgeInsets.all(16.0),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Docs Shared',
                            style: theme.textTheme.titleMedium?.copyWith(
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          const SizedBox(height: 4),
                          _isLoadingDocs
                              ? SizedBox(
                                  width: 16,
                                  height: 16,
                                  child: CircularProgressIndicator(
                                    strokeWidth: 2,
                                    valueColor: AlwaysStoppedAnimation<Color>(
                                      theme.colorScheme.onSurface.withOpacity(0.7),
                                    ),
                                  ),
                                )
                              : Text(
                                  '$_docsShared docs shared',
                                  style: theme.textTheme.bodyMedium?.copyWith(
                                    color: theme.colorScheme.onSurface.withOpacity(0.7),
                                  ),
                                ),
                        ],
                      ),
                      Icon(
                        Icons.description,
                        size: 32,
                        color: theme.colorScheme.primary,
                      ),
                    ],
                  ),
                ),
              ),
            ),
            const SizedBox(height: 16),
            // Credits Section
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          'Credits',
                          style: theme.textTheme.titleMedium?.copyWith(
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        ElevatedButton.icon(
                          onPressed: () {
                            showDialog(
                              context: context,
                              builder: (context) => CreditsPurchaseDialog(
                                creditsService: _creditsService,
                                onPurchaseComplete: () {
                                  // Refresh user data
                                  context.read<AuthProvider>().verifyToken();
                                },
                              ),
                            );
                          },
                          icon: const Icon(Icons.add, size: 18),
                          label: const Text('Buy Credits'),
                          style: ElevatedButton.styleFrom(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 16,
                              vertical: 8,
                            ),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    Row(
                      children: [
                        Icon(
                          Icons.account_balance_wallet,
                          size: 32,
                          color: theme.colorScheme.primary,
                        ),
                        const SizedBox(width: 12),
                        Text(
                          '${user?.credits ?? 0} credits',
                          style: theme.textTheme.headlineSmall?.copyWith(
                            fontWeight: FontWeight.bold,
                            color: theme.colorScheme.primary,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),
            // Manage Friends Section
            Card(
              child: InkWell(
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => const MyNetworkScreen(),
                    ),
                  );
                },
                borderRadius: BorderRadius.circular(12),
                child: Padding(
                  padding: const EdgeInsets.all(16.0),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Manage Friends',
                            style: theme.textTheme.titleMedium?.copyWith(
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          const SizedBox(height: 4),
                          Text(
                            'View friends and manage invites',
                            style: theme.textTheme.bodyMedium?.copyWith(
                              color: theme.colorScheme.onSurface.withOpacity(0.7),
                            ),
                          ),
                        ],
                      ),
                      Icon(
                        Icons.people,
                        size: 32,
                        color: theme.colorScheme.primary,
                      ),
                    ],
                  ),
                ),
              ),
            ),
            const SizedBox(height: 16),
            // Permissions Panel
            const PermissionsPanel(),
          ],
        ),
      ),
    );
  }
}