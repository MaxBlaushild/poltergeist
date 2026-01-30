import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:intl/intl.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:skunkworks/constants/app_colors.dart';
import 'package:skunkworks/models/post.dart';
import 'package:skunkworks/models/certificate.dart';
import 'package:skunkworks/models/user.dart';
import 'package:skunkworks/providers/auth_provider.dart';
import 'package:skunkworks/services/post_service.dart';
import 'package:skunkworks/services/certificate_service.dart';
import 'package:skunkworks/services/api_client.dart';
import 'package:skunkworks/constants/api_constants.dart';
import 'package:skunkworks/widgets/bottom_nav.dart';
import 'package:skunkworks/screens/post_detail_screen.dart';

class ProfileScreen extends StatefulWidget {
  final Function(NavTab) onNavigate;
  final String? userId; // Optional: if provided, show this user's profile
  final User? user; // Optional: user data if available

  const ProfileScreen({
    super.key,
    required this.onNavigate,
    this.userId,
    this.user,
  });

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  List<Post> _userPosts = [];
  bool _loading = false;
  Certificate? _certificate;
  bool _loadingCertificate = false;
  User? _displayUser;
  bool _hasLoadedForUser = false;

  @override
  void initState() {
    super.initState();
    _loadUserPosts();
    _loadCertificate();
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    // Reload when viewing own profile and auth user just became available (e.g. right after login)
    if (!_isViewingOwnProfile()) return;
    final userId = context.read<AuthProvider>().user?.id;
    if (userId != null && !_hasLoadedForUser) {
      _loadUserPosts();
    }
  }

  bool _isViewingOwnProfile() {
    final authProvider = context.read<AuthProvider>();
    final currentUserId = authProvider.user?.id;
    return widget.userId == null || widget.userId == currentUserId;
  }

  Future<void> _loadUserPosts() async {
    final authProvider = context.read<AuthProvider>();
    final currentUser = authProvider.user;

    // Set display user immediately so profile details show before posts load
    if (widget.user != null) {
      _displayUser = widget.user;
    } else if (widget.userId == null || widget.userId == currentUser?.id) {
      _displayUser = currentUser;
    } else {
      _displayUser = null;
    }
    if (mounted) setState(() {});

    // Determine which user's posts to load
    final targetUserId = widget.userId ?? currentUser?.id;
    if (targetUserId == null) return;

    _hasLoadedForUser = true;
    setState(() {
      _loading = true;
    });

    try {
      final apiClient = APIClient(ApiConstants.baseUrl);
      final postService = PostService(apiClient);
      _userPosts = await postService.getUserPosts(targetUserId);

      // If we don't have user info and posts exist, get it from first post
      if (_displayUser == null &&
          _userPosts.isNotEmpty &&
          _userPosts[0].user != null) {
        _displayUser = _userPosts[0].user;
      }
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

  Future<void> _loadCertificate() async {
    setState(() {
      _loadingCertificate = true;
    });

    try {
      final apiClient = APIClient(ApiConstants.baseUrl);
      final certificateService = CertificateService(apiClient);
      
      Certificate? certificate;
      if (_isViewingOwnProfile()) {
        // Get own certificate
        certificate = await certificateService.getCertificate();
      } else {
        // Get other user's certificate
        final targetUserId = widget.userId;
        if (targetUserId != null) {
          certificate = await certificateService.getUserCertificate(targetUserId);
        }
      }
      
      if (mounted) {
        setState(() {
          _certificate = certificate;
          _loadingCertificate = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _loadingCertificate = false;
        });
        // Don't show error snackbar for 404 (no certificate) - that's expected
        if (e.toString().contains('404') == false) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Failed to load certificate: $e')),
          );
        }
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Consumer<AuthProvider>(
      builder: (context, authProvider, child) {
        final currentUser = authProvider.user;
        final displayUser = _displayUser ?? currentUser;
        final username = displayUser?.username ?? displayUser?.phoneNumber ?? 'Unknown';
        final profilePictureUrl = displayUser?.profilePictureUrl;
        final isOwnProfile = _isViewingOwnProfile();

        return Scaffold(
          backgroundColor: AppColors.warmWhite,
          appBar: AppBar(
            backgroundColor: AppColors.warmWhite,
            elevation: 0,
            title: Text(
              username,
              style: TextStyle(
                color: AppColors.graphiteInk,
                fontWeight: FontWeight.w600,
                fontSize: 18,
              ),
            ),
            actions: isOwnProfile
                ? [
                    IconButton(
                      icon: Icon(Icons.logout, color: AppColors.graphiteInk),
                      onPressed: () async {
                        await authProvider.logout();
                      },
                    ),
                  ]
                : null,
          ),
          body: SingleChildScrollView(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // Profile header – always show immediately (from AuthProvider)
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
                      
                      // Certificate section
                      _buildCertificateSection(),
                      
                      // Posts grid – show loading only for this section
                      if (_loading)
                        const Padding(
                          padding: EdgeInsets.all(32.0),
                          child: Center(
                            child: CircularProgressIndicator(),
                          ),
                        )
                      else if (_userPosts.isEmpty)
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
                    ],
                  ),
                ),
          bottomNavigationBar: isOwnProfile
              ? BottomNav(
                  currentTab: NavTab.profile,
                  onTabChanged: widget.onNavigate,
                )
              : null,
        );
      },
    );
  }

  Widget _buildProfileInfoSection(User? user, AuthProvider authProvider) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 8.0),
      child: Card(
        elevation: 0,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
          side: BorderSide(color: Colors.grey.shade300),
        ),
        child: Column(
          children: [
            ListTile(
              leading: Icon(Icons.info_outline, color: AppColors.graphiteInk),
              title: const Text(
                'Profile Information',
                style: TextStyle(
                  fontWeight: FontWeight.w600,
                  fontSize: 16,
                ),
              ),
              trailing: IconButton(
                icon: Icon(Icons.edit, color: AppColors.softRealBlue),
                onPressed: () => _showEditProfileDialog(context, user, authProvider),
                tooltip: 'Edit profile',
              ),
            ),
            const Divider(height: 1),
            Padding(
              padding: const EdgeInsets.all(16.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _buildInfoRow(
                    Icons.category,
                    'Category',
                    user?.category ?? 'Not set',
                  ),
                  const SizedBox(height: 12),
                  _buildInfoRow(
                    Icons.calendar_today,
                    'Age Range',
                    user?.ageRange ?? 'Not set',
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildInfoRow(IconData icon, String label, String value) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(icon, size: 20, color: Colors.grey.shade600),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                label,
                style: TextStyle(
                  fontSize: 12,
                  color: Colors.grey.shade600,
                ),
              ),
              const SizedBox(height: 2),
              Text(
                value,
                style: const TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  void _showEditProfileDialog(BuildContext context, User? user, AuthProvider authProvider) {
    final bioController = TextEditingController(text: user?.bio ?? '');

    final categoryOptions = [
      'Travel',
      'Food',
      'Art',
      'Music',
      'Sports',
      'Technology',
      'Fashion',
      'Fitness',
      'Photography',
      'Other',
    ];

    final ageRangeOptions = [
      '18-25',
      '26-35',
      '36-45',
      '46-55',
      '56+',
    ];

    String? selectedCategory = user?.category;
    String? selectedAgeRange = user?.ageRange;

    showDialog(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setDialogState) => AlertDialog(
          title: const Text('Edit Profile'),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Category dropdown
                const Text(
                  'Category',
                  style: TextStyle(fontWeight: FontWeight.w600, fontSize: 14),
                ),
                const SizedBox(height: 8),
                DropdownButtonFormField<String>(
                  value: selectedCategory,
                  decoration: InputDecoration(
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(4),
                    ),
                    filled: true,
                    fillColor: Colors.grey.shade50,
                  ),
                  hint: const Text('Select category'),
                  items: categoryOptions.map((cat) {
                    return DropdownMenuItem(
                      value: cat,
                      child: Text(cat),
                    );
                  }).toList(),
                  onChanged: (value) {
                    setDialogState(() {
                      selectedCategory = value;
                    });
                  },
                ),
                const SizedBox(height: 16),
                // Age range dropdown
                const Text(
                  'Age Range',
                  style: TextStyle(fontWeight: FontWeight.w600, fontSize: 14),
                ),
                const SizedBox(height: 8),
                DropdownButtonFormField<String>(
                  value: selectedAgeRange,
                  decoration: InputDecoration(
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(4),
                    ),
                    filled: true,
                    fillColor: Colors.grey.shade50,
                  ),
                  hint: const Text('Select age range'),
                  items: ageRangeOptions.map((range) {
                    return DropdownMenuItem(
                      value: range,
                      child: Text(range),
                    );
                  }).toList(),
                  onChanged: (value) {
                    setDialogState(() {
                      selectedAgeRange = value;
                    });
                  },
                ),
                const SizedBox(height: 16),
                // Bio text field
                const Text(
                  'Bio',
                  style: TextStyle(fontWeight: FontWeight.w600, fontSize: 14),
                ),
                const SizedBox(height: 8),
                TextField(
                  controller: bioController,
                  decoration: InputDecoration(
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(4),
                    ),
                    filled: true,
                    fillColor: Colors.grey.shade50,
                    hintText: 'Tell us about yourself',
                  ),
                  maxLines: 3,
                ),
              ],
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(context).pop(),
              child: const Text('Cancel'),
            ),
            TextButton(
              onPressed: () async {
                try {
                  await authProvider.updateProfile(
                    category: selectedCategory,
                    ageRange: selectedAgeRange,
                    bio: bioController.text.trim().isEmpty
                        ? null
                        : bioController.text.trim(),
                  );
                  if (context.mounted) {
                    Navigator.of(context).pop();
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(content: Text('Profile updated successfully')),
                    );
                  }
                } catch (e) {
                  if (context.mounted) {
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(content: Text('Failed to update profile: $e')),
                    );
                  }
                }
              },
              child: const Text('Save'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildCertificateSection() {
    if (_loadingCertificate) {
      return const Padding(
        padding: EdgeInsets.all(16.0),
        child: Center(
          child: CircularProgressIndicator(),
        ),
      );
    }

    if (_certificate == null) {
      return Padding(
        padding: const EdgeInsets.all(16.0),
        child: Card(
          elevation: 0,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(8),
            side: BorderSide(color: Colors.grey.shade300),
          ),
          child: Padding(
            padding: const EdgeInsets.all(16.0),
            child: Row(
              children: [
                Icon(Icons.info_outline, color: Colors.grey.shade600),
                const SizedBox(width: 12),
                Expanded(
                  child: Text(
                    'No certificate enrolled',
                    style: TextStyle(
                      color: Colors.grey.shade600,
                      fontSize: 14,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      );
    }

    final cert = _certificate!;
    final dateFormat = DateFormat('MMM dd, yyyy');
    final createdDate = cert.createdAt != null
        ? dateFormat.format(cert.createdAt!)
        : 'Unknown';

    return Padding(
      padding: const EdgeInsets.all(16.0),
      child: Card(
        elevation: 0,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
          side: BorderSide(color: Colors.grey.shade300),
        ),
        child: ExpansionTile(
          initiallyExpanded: false,
          leading: Icon(Icons.verified, color: AppColors.softRealBlue),
          title: const Text(
            'Certificate',
            style: TextStyle(
              fontWeight: FontWeight.w600,
              fontSize: 16,
            ),
          ),
          children: [
            const Divider(height: 1),
            
            // Status
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 12.0),
              child: Row(
                children: [
                  const Text(
                    'Status: ',
                    style: TextStyle(
                      fontSize: 14,
                      color: Colors.grey,
                    ),
                  ),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                    decoration: BoxDecoration(
                      color: (cert.active == true)
                          ? Colors.green.shade100
                          : Colors.grey.shade200,
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Text(
                      (cert.active == true) ? 'Active' : 'Inactive',
                      style: TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w600,
                        color: (cert.active == true)
                            ? Colors.green.shade800
                            : Colors.grey.shade700,
                      ),
                    ),
                  ),
                ],
              ),
            ),
            
            // Fingerprint
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 8.0),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Icon(Icons.fingerprint, size: 18, color: Colors.grey.shade600),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Fingerprint',
                          style: TextStyle(
                            fontSize: 12,
                            color: Colors.grey.shade600,
                          ),
                        ),
                        const SizedBox(height: 4),
                        SelectableText(
                          cert.fingerprint,
                          style: const TextStyle(
                            fontSize: 12,
                            fontFamily: 'monospace',
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            
            // Created date
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 8.0),
              child: Row(
                children: [
                  Icon(Icons.calendar_today, size: 18, color: Colors.grey.shade600),
                  const SizedBox(width: 8),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Created',
                        style: TextStyle(
                          fontSize: 12,
                          color: Colors.grey.shade600,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        createdDate,
                        style: const TextStyle(
                          fontSize: 14,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
            
            // Block Explorer Link
            if (cert.transactionHash != null)
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 8.0),
                child: InkWell(
                  onTap: () async {
                    final url = _getBlockExplorerUrl(cert.transactionHash!, cert.chainId);
                    final uri = Uri.parse(url);
                    if (await canLaunchUrl(uri)) {
                      await launchUrl(uri, mode: LaunchMode.externalApplication);
                    } else {
                      if (mounted) {
                        ScaffoldMessenger.of(context).showSnackBar(
                          SnackBar(
                            content: Text('Could not open block explorer: $url'),
                            backgroundColor: Theme.of(context).colorScheme.error,
                          ),
                        );
                      }
                    }
                  },
                  child: Row(
                    children: [
                      Icon(Icons.open_in_new, size: 18, color: AppColors.softRealBlue),
                      const SizedBox(width: 8),
                      Text(
                        'View on Block Explorer',
                        style: TextStyle(
                          fontSize: 14,
                          color: AppColors.softRealBlue,
                          decoration: TextDecoration.underline,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            
            const SizedBox(height: 8),
            
            // Public Key (expandable)
            ExpansionTile(
              title: const Text(
                'Public Key',
                style: TextStyle(fontSize: 14),
              ),
              leading: Icon(Icons.key, size: 20, color: Colors.grey.shade600),
              children: [
                Padding(
                  padding: const EdgeInsets.all(16.0),
                  child: SelectableText(
                    cert.publicKey,
                    style: const TextStyle(
                      fontSize: 11,
                      fontFamily: 'monospace',
                    ),
                  ),
                ),
              ],
            ),
            
            // Certificate PEM (expandable)
            ExpansionTile(
              title: const Text(
                'Certificate',
                style: TextStyle(fontSize: 14),
              ),
              leading: Icon(Icons.description, size: 20, color: Colors.grey.shade600),
              children: [
                Padding(
                  padding: const EdgeInsets.all(16.0),
                  child: SelectableText(
                    cert.certificatePem,
                    style: const TextStyle(
                      fontSize: 11,
                      fontFamily: 'monospace',
                    ),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  String _getBlockExplorerUrl(String txHash, int? chainId) {
    switch (chainId) {
      case 84532: // Base Sepolia
        return 'https://sepolia.basescan.org/tx/$txHash';
      case 1: // Ethereum Mainnet
        return 'https://etherscan.io/tx/$txHash';
      case 11155111: // Sepolia
        return 'https://sepolia.etherscan.io/tx/$txHash';
      default:
        // Default to Base Sepolia if unknown
        return 'https://sepolia.basescan.org/tx/$txHash';
    }
  }
}

