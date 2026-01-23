import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:intl/intl.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:skunkworks/models/post.dart';
import 'package:skunkworks/models/certificate.dart';
import 'package:skunkworks/providers/auth_provider.dart';
import 'package:skunkworks/services/post_service.dart';
import 'package:skunkworks/services/certificate_service.dart';
import 'package:skunkworks/services/api_client.dart';
import 'package:skunkworks/constants/api_constants.dart';
import 'package:skunkworks/widgets/bottom_nav.dart';
import 'package:skunkworks/screens/post_detail_screen.dart';

class ProfileScreen extends StatefulWidget {
  final Function(NavTab) onNavigate;

  const ProfileScreen({
    super.key,
    required this.onNavigate,
  });

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  List<Post> _userPosts = [];
  bool _loading = false;
  Certificate? _certificate;
  bool _loadingCertificate = false;

  @override
  void initState() {
    super.initState();
    _loadUserPosts();
    _loadCertificate();
  }

  Future<void> _loadUserPosts() async {
    final authProvider = context.read<AuthProvider>();
    final user = authProvider.user;

    if (user?.id == null) return;

    setState(() {
      _loading = true;
    });

    try {
      final apiClient = APIClient(ApiConstants.baseUrl);
      final postService = PostService(apiClient);
      _userPosts = await postService.getUserPosts(user!.id!);
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
      final certificate = await certificateService.getCertificate();
      
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
        final user = authProvider.user;
        final username = user?.username ?? user?.phoneNumber ?? 'Unknown';
        final profilePictureUrl = user?.profilePictureUrl;

        return Scaffold(
          backgroundColor: Colors.white,
          appBar: AppBar(
            backgroundColor: Colors.white,
            elevation: 0,
            title: Text(
              username,
              style: const TextStyle(
                color: Colors.black,
                fontWeight: FontWeight.w600,
                fontSize: 18,
              ),
            ),
            actions: [
              IconButton(
                icon: const Icon(Icons.logout, color: Colors.black),
                onPressed: () async {
                  await authProvider.logout();
                },
              ),
            ],
          ),
          body: _loading
              ? const Center(child: CircularProgressIndicator())
              : SingleChildScrollView(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // Profile header
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
                      
                      // Posts grid
                      if (_userPosts.isEmpty)
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
                                  ? Image.network(
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
          bottomNavigationBar: BottomNav(
            currentTab: NavTab.profile,
            onTabChanged: widget.onNavigate,
          ),
        );
      },
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
          leading: Icon(Icons.verified, color: Colors.blue.shade700),
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
                      Icon(Icons.open_in_new, size: 18, color: Colors.blue.shade700),
                      const SizedBox(width: 8),
                      Text(
                        'View on Block Explorer',
                        style: TextStyle(
                          fontSize: 14,
                          color: Colors.blue.shade700,
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

