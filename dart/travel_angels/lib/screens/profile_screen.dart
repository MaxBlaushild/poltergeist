import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/models/user_level.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/providers/user_level_provider.dart';
import 'package:travel_angels/screens/documents_screen.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/credits_service.dart';
import 'package:travel_angels/services/document_service.dart';
import 'package:travel_angels/widgets/credits_purchase_dialog.dart';
import 'package:travel_angels/widgets/permissions_panel.dart';
import 'package:travel_angels/widgets/edit_profile_dialog.dart';
import 'package:intl/intl.dart';

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
    // Load initial data
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      _loadDocumentCount();
    });
  }

  Future<void> _loadDocumentCount() async {
    if (!mounted) return;
    
    final authProvider = context.read<AuthProvider>();
    final user = authProvider.user;

    if (!authProvider.isAuthenticated || user?.id == null) {
      return;
    }

    if (!mounted) return;
    setState(() {
      _isLoadingDocs = true;
    });

    try {
      final documentsJson = await _documentService.getDocumentsByUserId(user!.id!);
      if (mounted) {
        setState(() {
          _docsShared = documentsJson.length;
          _isLoadingDocs = false;
        });
      }
    } catch (e) {
      // Silently handle errors - don't show error state, just keep count at 0
      if (mounted) {
        setState(() {
          _isLoadingDocs = false;
        });
      }
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

  /// Calculate age from date of birth
  static int? _calculateAge(DateTime? dateOfBirth) {
    if (dateOfBirth == null) return null;
    final now = DateTime.now();
    int age = now.year - dateOfBirth.year;
    if (now.month < dateOfBirth.month ||
        (now.month == dateOfBirth.month && now.day < dateOfBirth.day)) {
      age--;
    }
    return age;
  }

  /// Format date of birth for display
  static String _formatDateOfBirth(DateTime? dateOfBirth) {
    if (dateOfBirth == null) return 'Not set';
    return DateFormat('MMMM d, yyyy').format(dateOfBirth);
  }

  /// Format location for display
  static String _formatLocation(String? locationAddress) {
    if (locationAddress == null || locationAddress.isEmpty) return 'Not set';
    return locationAddress;
  }

  /// Build demographic information section
  Widget _buildDemographicSection(User? user, ThemeData theme) {
    final age = _calculateAge(user?.dateOfBirth);
    final dateOfBirthText = user?.dateOfBirth != null
        ? '${_formatDateOfBirth(user!.dateOfBirth)} (${age} years old)'
        : 'Not set';
    final genderText = user?.gender ?? 'Not set';
    final locationText = _formatLocation(user?.locationAddress);
    final bioText = user?.bio ?? 'Not set';

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  'Profile Information',
                  style: theme.textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.edit),
                  onPressed: () => _openEditDialog(context, user),
                  tooltip: 'Edit profile',
                ),
              ],
            ),
            const SizedBox(height: 16),
            _buildInfoRow(
              theme,
              Icons.cake,
              'Date of Birth',
              dateOfBirthText,
            ),
            const SizedBox(height: 12),
            _buildInfoRow(
              theme,
              Icons.person,
              'Gender',
              genderText,
            ),
            const SizedBox(height: 12),
            _buildInfoRow(
              theme,
              Icons.location_on,
              'Location',
              locationText,
            ),
            const SizedBox(height: 12),
            _buildInfoRow(
              theme,
              Icons.description,
              'Bio',
              bioText,
              isMultiline: true,
            ),
          ],
        ),
      ),
    );
  }

  /// Build a single info row
  Widget _buildInfoRow(
    ThemeData theme,
    IconData icon,
    String label,
    String value, {
    bool isMultiline = false,
  }) {
    return Row(
      crossAxisAlignment: isMultiline ? CrossAxisAlignment.start : CrossAxisAlignment.center,
      children: [
        Icon(
          icon,
          size: 20,
          color: theme.colorScheme.primary,
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                label,
                style: theme.textTheme.bodySmall?.copyWith(
                  color: theme.colorScheme.onSurface.withOpacity(0.6),
                  fontWeight: FontWeight.w500,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                value,
                style: theme.textTheme.bodyMedium?.copyWith(
                  color: value == 'Not set'
                      ? theme.colorScheme.onSurface.withOpacity(0.5)
                      : theme.colorScheme.onSurface,
                ),
                maxLines: isMultiline ? null : 1,
                overflow: isMultiline ? null : TextOverflow.ellipsis,
              ),
            ],
          ),
        ),
      ],
    );
  }

  /// Open edit profile dialog
  void _openEditDialog(BuildContext context, User? user) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => EditProfileDialog(
        user: user,
        onSave: () {
          // Refresh user data after save
          context.read<AuthProvider>().verifyToken();
        },
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
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(16.0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.center,
            children: [
              // Profile header
              Align(
                alignment: Alignment.centerLeft,
                child: Text(
                  'Profile',
                  style: theme.textTheme.headlineMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
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
            // Demographic Information Section
            _buildDemographicSection(user, theme),
            const SizedBox(height: 16),
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
            // Permissions Panel
            const PermissionsPanel(),
            const SizedBox(height: 16),
            // Logout Section
            Card(
              child: InkWell(
                onTap: () => _showLogoutConfirmation(context),
                borderRadius: BorderRadius.circular(12),
                child: Padding(
                  padding: const EdgeInsets.all(16.0),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Row(
                        children: [
                          Icon(
                            Icons.logout,
                            size: 32,
                            color: theme.colorScheme.error,
                          ),
                          const SizedBox(width: 12),
                          Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                'Log Out',
                                style: theme.textTheme.titleMedium?.copyWith(
                                  fontWeight: FontWeight.bold,
                                  color: theme.colorScheme.error,
                                ),
                              ),
                              const SizedBox(height: 4),
                              Text(
                                'Sign out of your account',
                                style: theme.textTheme.bodyMedium?.copyWith(
                                  color: theme.colorScheme.onSurface.withOpacity(0.7),
                                ),
                              ),
                            ],
                          ),
                        ],
                      ),
                      Icon(
                        Icons.chevron_right,
                        color: theme.colorScheme.onSurface.withOpacity(0.5),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
        ),
    );
  }

  /// Show logout confirmation dialog
  void _showLogoutConfirmation(BuildContext context) {
    showDialog(
      context: context,
      builder: (BuildContext dialogContext) {
        final theme = Theme.of(context);
        return AlertDialog(
          title: const Text('Log Out'),
          content: const Text('Are you sure you want to log out?'),
          actions: [
            TextButton(
              onPressed: () {
                Navigator.of(dialogContext).pop();
              },
              child: const Text('Cancel'),
            ),
            TextButton(
              onPressed: () async {
                Navigator.of(dialogContext).pop();
                await context.read<AuthProvider>().logout();
              },
              style: TextButton.styleFrom(
                foregroundColor: theme.colorScheme.error,
              ),
              child: const Text('Log Out'),
            ),
          ],
        );
      },
    );
  }
}