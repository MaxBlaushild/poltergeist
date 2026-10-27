import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/models/user_level.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/providers/user_level_provider.dart';
import 'package:travel_angels/screens/library_screen.dart';
import 'package:travel_angels/screens/my_network_screen.dart';
import 'package:travel_angels/screens/travel_calendar_screen.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/credits_service.dart';
import 'package:travel_angels/services/document_service.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/credits_purchase_dialog.dart';
import 'package:travel_angels/widgets/edit_profile_dialog.dart';
import 'package:travel_angels/widgets/permissions_panel.dart';

class ProfileScreen extends StatefulWidget {
  const ProfileScreen({
    super.key,
    this.documentService,
    this.creditsService,
    this.onOpenCalendar,
    this.onOpenCircle,
  });

  final DocumentService? documentService;
  final CreditsService? creditsService;
  final VoidCallback? onOpenCalendar;
  final VoidCallback? onOpenCircle;

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  late final _documents =
      widget.documentService ??
      DocumentService(APIClient(ApiConstants.baseUrl));
  late final _credits =
      widget.creditsService ?? CreditsService(APIClient(ApiConstants.baseUrl));
  AuthProvider? _auth;
  String? _identity;
  int _generation = 0;
  int? _guideCount;
  int? _savedCount;
  bool _loadingLibrary = false;
  bool _libraryError = false;
  bool _refreshing = false;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final auth = context.read<AuthProvider>();
    if (_auth != auth) {
      _auth?.removeListener(_authChanged);
      _auth = auth;
      auth.addListener(_authChanged);
      _authChanged();
    }
  }

  void _authChanged() {
    final identity = _auth?.user?.id;
    if (identity == _identity) return;
    _identity = identity;
    _generation++;
    setState(() {
      _guideCount = null;
      _savedCount = null;
      _libraryError = false;
      _loadingLibrary = identity != null;
    });
    if (identity != null) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted && _identity == identity) _loadLibrary();
      });
    }
  }

  @override
  void dispose() {
    _generation++;
    _auth?.removeListener(_authChanged);
    super.dispose();
  }

  Future<void> _loadLibrary() async {
    if (_identity == null) return;
    final generation = ++_generation;
    setState(() {
      _loadingLibrary = true;
      _libraryError = false;
    });
    try {
      final results = await Future.wait([
        _documents.getLibraryDocuments('yours'),
        _documents.getLibraryDocuments('saved'),
      ]);
      if (!mounted || generation != _generation) return;
      setState(() {
        _guideCount = results[0].length;
        _savedCount = results[1].length;
      });
    } catch (_) {
      if (!mounted || generation != _generation) return;
      setState(() {
        _guideCount = null;
        _savedCount = null;
        _libraryError = true;
      });
    } finally {
      if (mounted && generation == _generation) {
        setState(() => _loadingLibrary = false);
      }
    }
  }

  Future<void> _refresh() async {
    if (_refreshing) return;
    setState(() => _refreshing = true);
    final auth = context.read<AuthProvider>();
    final level = context.read<UserLevelProvider>();
    try {
      await Future.wait([
        auth.verifyToken(),
        _loadLibrary(),
        level.fetchUserLevel(),
      ]);
      if (mounted && auth.error != null && auth.isAuthenticated) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Your profile could not be refreshed. Try again.'),
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _refreshing = false);
    }
  }

  Future<void> _edit() async {
    await showDialog<bool>(
      context: context,
      builder: (_) => EditProfileDialog(user: _auth?.user),
    );
  }

  Future<void> _openLibrary(String scope) async {
    await Navigator.push<void>(
      context,
      MaterialPageRoute(
        builder: (_) => Scaffold(
          appBar: AppBar(title: const Text('Library')),
          body: LibraryScreen(initialScope: scope, service: _documents),
        ),
      ),
    );
    if (mounted) await _loadLibrary();
  }

  void _openCalendar() {
    if (widget.onOpenCalendar != null) {
      widget.onOpenCalendar!();
    } else {
      Navigator.push<void>(
        context,
        MaterialPageRoute(builder: (_) => const TravelCalendarScreen()),
      );
    }
  }

  void _openCircle() {
    if (widget.onOpenCircle != null) {
      widget.onOpenCircle!();
    } else {
      Navigator.push<void>(
        context,
        MaterialPageRoute(
          builder: (_) => Scaffold(
            appBar: AppBar(title: const Text('My circle')),
            body: const MyNetworkScreen(),
          ),
        ),
      );
    }
  }

  void _connections() => Navigator.push<void>(
    context,
    MaterialPageRoute(
      builder: (_) => Scaffold(
        appBar: AppBar(title: const Text('Connected accounts')),
        body: SingleChildScrollView(
          padding: const EdgeInsets.all(16),
          child: Center(
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 640),
              child: const PermissionsPanel(),
            ),
          ),
        ),
      ),
    ),
  );

  void _buyCredits() {
    showDialog<void>(
      context: context,
      builder: (_) => CreditsPurchaseDialog(
        creditsService: _credits,
        onPurchaseComplete: () {
          if (mounted) _auth?.verifyToken();
        },
      ),
    );
  }

  Future<void> _logout() async {
    final auth = _auth;
    final identity = auth?.user?.id;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Log out?'),
        content: const Text('You can sign back in with your phone number.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext, false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(dialogContext, true),
            style: TextButton.styleFrom(foregroundColor: AppTheme.error),
            child: const Text('Log out'),
          ),
        ],
      ),
    );
    if (confirmed == true &&
        mounted &&
        identity != null &&
        auth?.user?.id == identity) {
      await auth?.logout();
    }
  }

  String _name(User user) {
    for (final value in [user.name, user.username]) {
      if (value?.trim().isNotEmpty == true) return value!.trim();
    }
    return 'Traveler';
  }

  String _initials(User user) => _name(user)
      .split(RegExp(r'\s+'))
      .take(2)
      .map((word) => word.characters.first)
      .join()
      .toUpperCase();

  Widget _avatar(User user, double size) {
    final fallback = Center(
      child: Padding(
        padding: const EdgeInsets.all(8),
        child: FittedBox(
          fit: BoxFit.scaleDown,
          child: Text(
            _initials(user),
            style: Theme.of(context).textTheme.headlineSmall,
          ),
        ),
      ),
    );
    return Container(
      width: size,
      height: size,
      decoration: const BoxDecoration(
        shape: BoxShape.circle,
        color: AppTheme.peach,
      ),
      clipBehavior: Clip.antiAlias,
      child: user.profilePictureUrl?.trim().isNotEmpty == true
          ? Image.network(
              user.profilePictureUrl!,
              fit: BoxFit.cover,
              errorBuilder: (_, _, _) => fallback,
            )
          : fallback,
    );
  }

  Widget _identityHeader(User user, bool largeText) {
    final theme = Theme.of(context).textTheme;
    final details = Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(_name(user), style: theme.headlineMedium),
        if (user.username?.trim().isNotEmpty == true) ...[
          const SizedBox(height: 4),
          Text(
            '@${user.username!.trim()}',
            style: theme.bodyMedium?.copyWith(color: AppTheme.muted),
          ),
        ],
        if (user.locationAddress?.trim().isNotEmpty == true) ...[
          const SizedBox(height: 8),
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Padding(
                padding: EdgeInsets.only(top: 2),
                child: Icon(Icons.place_outlined, size: 16),
              ),
              const SizedBox(width: 4),
              Expanded(
                child: Text(
                  user.locationAddress!.trim(),
                  style: theme.bodySmall,
                ),
              ),
            ],
          ),
        ],
      ],
    );
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (largeText)
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [_avatar(user, 56), const SizedBox(height: 12), details],
          )
        else
          Row(
            children: [
              _avatar(user, 72),
              const SizedBox(width: 18),
              Expanded(child: details),
            ],
          ),
        if (user.bio?.trim().isNotEmpty == true) ...[
          const SizedBox(height: 16),
          Text(user.bio!.trim(), style: theme.bodyMedium),
        ],
        if (user.createdAt != null) ...[
          const SizedBox(height: 10),
          Text(
            'Joined ${DateFormat.yMMM().format(user.createdAt!.toLocal())}',
            style: theme.labelSmall,
          ),
        ],
      ],
    );
  }

  Widget _libraryTile(
    String scope,
    String label,
    int? count,
    IconData icon,
    Color color,
  ) {
    return Material(
      color: color,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(4),
        side: const BorderSide(color: AppTheme.outline),
      ),
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        key: ValueKey('profile-library-$scope'),
        onTap: () => _openLibrary(scope),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Expanded(
                    child: Text(
                      count?.toString() ?? '—',
                      style: Theme.of(context).textTheme.headlineMedium,
                    ),
                  ),
                  Icon(icon, size: 20),
                ],
              ),
              const SizedBox(height: 6),
              Text(label, style: Theme.of(context).textTheme.bodyMedium),
            ],
          ),
        ),
      ),
    );
  }

  Widget _row({
    required String label,
    required IconData icon,
    required VoidCallback onTap,
    String? detail,
    Key? key,
  }) {
    return InkWell(
      key: key,
      onTap: onTap,
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 14),
        child: Row(
          children: [
            Icon(icon, size: 20),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(label, style: Theme.of(context).textTheme.bodyMedium),
                  if (detail != null)
                    Text(detail, style: Theme.of(context).textTheme.bodySmall),
                ],
              ),
            ),
            const SizedBox(width: 8),
            const Icon(Icons.chevron_right, size: 18),
          ],
        ),
      ),
    );
  }

  Widget _sectionTitle(String text) => Padding(
    padding: const EdgeInsets.only(bottom: 12),
    child: Text(text, style: Theme.of(context).textTheme.titleLarge),
  );

  Widget _travelSpace(bool largeText) => Column(
    crossAxisAlignment: CrossAxisAlignment.stretch,
    children: [
      _sectionTitle('Your travel space'),
      LayoutBuilder(
        builder: (context, constraints) {
          final width = (constraints.maxWidth - 12) / 2;
          return Wrap(
            spacing: 12,
            runSpacing: 12,
            children: [
              SizedBox(
                width: largeText && constraints.maxWidth < 450
                    ? constraints.maxWidth
                    : width,
                child: _libraryTile(
                  'yours',
                  'Your guides',
                  _guideCount,
                  Icons.auto_stories_outlined,
                  const Color(0xFFDCE8F2),
                ),
              ),
              SizedBox(
                width: largeText && constraints.maxWidth < 450
                    ? constraints.maxWidth
                    : width,
                child: _libraryTile(
                  'saved',
                  'Saved guides',
                  _savedCount,
                  Icons.bookmark_border,
                  const Color(0xFFE1E8D8),
                ),
              ),
            ],
          );
        },
      ),
      if (_loadingLibrary)
        const Padding(
          padding: EdgeInsets.only(top: 8),
          child: LinearProgressIndicator(minHeight: 2),
        ),
      if (_libraryError)
        Padding(
          padding: const EdgeInsets.only(top: 8),
          child: Wrap(
            crossAxisAlignment: WrapCrossAlignment.center,
            spacing: 8,
            children: [
              Text(
                'Guide counts are unavailable.',
                style: Theme.of(context).textTheme.bodySmall,
              ),
              TextButton(
                key: const ValueKey('profile-library-retry'),
                onPressed: _loadLibrary,
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
      const SizedBox(height: 12),
      _row(
        key: const ValueKey('profile-calendar'),
        label: 'Your calendar',
        icon: Icons.calendar_month_outlined,
        onTap: _openCalendar,
      ),
      const Divider(height: 1),
      _row(
        key: const ValueKey('profile-circle'),
        label: 'My circle',
        icon: Icons.people_outline,
        onTap: _openCircle,
      ),
    ],
  );

  Widget _account(User user, UserLevelProvider levels) {
    final level = levels.userLevel;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        _sectionTitle('Account'),
        _row(
          key: const ValueKey('profile-connections'),
          label: 'Connected accounts',
          detail: 'Google Drive',
          icon: Icons.cloud_outlined,
          onTap: _connections,
        ),
        const Divider(height: 1),
        _row(
          key: const ValueKey('profile-credits'),
          label: '${user.credits ?? 0} credits',
          detail: 'Add credits',
          icon: Icons.account_balance_wallet_outlined,
          onTap: _buyCredits,
        ),
        const Divider(height: 1),
        if (level != null)
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Wrap(
                  alignment: WrapAlignment.spaceBetween,
                  spacing: 12,
                  runSpacing: 4,
                  children: [
                    Text(
                      'Level ${level.level}',
                      style: Theme.of(context).textTheme.bodyMedium,
                    ),
                    Text(
                      '${level.experiencePointsOnLevel} / ${level.experienceToNextLevel} XP',
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                LinearProgressIndicator(
                  value: _progress(level),
                  minHeight: 4,
                  backgroundColor: AppTheme.peach,
                ),
              ],
            ),
          )
        else if (levels.loading)
          const Padding(
            padding: EdgeInsets.symmetric(vertical: 16),
            child: LinearProgressIndicator(minHeight: 2),
          ),
        const SizedBox(height: 8),
        Align(
          alignment: Alignment.centerLeft,
          child: TextButton.icon(
            key: const ValueKey('profile-logout'),
            style: TextButton.styleFrom(foregroundColor: AppTheme.error),
            onPressed: _logout,
            icon: const Icon(Icons.logout, size: 18),
            label: const Text('Log out'),
          ),
        ),
      ],
    );
  }

  double _progress(UserLevel level) => level.experienceToNextLevel <= 0
      ? 1
      : (level.experiencePointsOnLevel / level.experienceToNextLevel).clamp(
          0.0,
          1.0,
        );

  @override
  Widget build(BuildContext context) {
    final user = context.watch<AuthProvider>().user;
    final levels = context.watch<UserLevelProvider>();
    if (user == null) {
      return const Center(child: Text('Sign in to view your profile.'));
    }
    final largeText = MediaQuery.textScalerOf(context).scale(14) > 21;
    return Scaffold(
      key: const ValueKey('profile-screen'),
      body: SafeArea(
        child: RefreshIndicator(
          onRefresh: _refresh,
          child: SingleChildScrollView(
            physics: const AlwaysScrollableScrollPhysics(),
            child: Center(
              child: ConstrainedBox(
                constraints: const BoxConstraints(maxWidth: 1040),
                child: Padding(
                  padding: EdgeInsets.all(
                    MediaQuery.sizeOf(context).width < 600 ? 16 : 24,
                  ),
                  child: LayoutBuilder(
                    builder: (context, constraints) {
                      final wide = constraints.maxWidth >= 760 && !largeText;
                      return Column(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          Row(
                            children: [
                              Expanded(
                                child: Text(
                                  'Profile',
                                  style: Theme.of(
                                    context,
                                  ).textTheme.headlineMedium,
                                ),
                              ),
                              IconButton(
                                key: const ValueKey('profile-refresh'),
                                tooltip: 'Refresh profile',
                                onPressed: _refreshing ? null : _refresh,
                                icon: const Icon(Icons.refresh, size: 20),
                              ),
                              if (constraints.maxWidth < 400 || largeText)
                                IconButton(
                                  key: const ValueKey('profile-edit'),
                                  tooltip: 'Edit profile',
                                  onPressed: _edit,
                                  icon: const Icon(
                                    Icons.edit_outlined,
                                    size: 20,
                                  ),
                                )
                              else
                                TextButton.icon(
                                  key: const ValueKey('profile-edit'),
                                  onPressed: _edit,
                                  icon: const Icon(
                                    Icons.edit_outlined,
                                    size: 18,
                                  ),
                                  label: const Text('Edit profile'),
                                ),
                            ],
                          ),
                          const SizedBox(height: 20),
                          _identityHeader(
                            user,
                            largeText && constraints.maxWidth < 600,
                          ),
                          const SizedBox(height: 28),
                          const Divider(height: 1),
                          const SizedBox(height: 24),
                          if (wide)
                            Row(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Expanded(
                                  flex: 6,
                                  child: _travelSpace(largeText),
                                ),
                                const SizedBox(width: 40),
                                Expanded(
                                  flex: 4,
                                  child: _account(user, levels),
                                ),
                              ],
                            )
                          else ...[
                            _travelSpace(largeText),
                            const SizedBox(height: 24),
                            _account(user, levels),
                          ],
                          const SizedBox(height: 24),
                        ],
                      );
                    },
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
