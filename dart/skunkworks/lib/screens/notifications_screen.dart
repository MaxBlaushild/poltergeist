import 'package:flutter/material.dart';
import 'package:skunkworks/constants/app_colors.dart';
import 'package:skunkworks/models/notification.dart';
import 'package:skunkworks/widgets/bottom_nav.dart';
import 'package:skunkworks/screens/album_invites_screen.dart';
import 'package:skunkworks/screens/album_detail_screen.dart';
import 'package:provider/provider.dart';
import 'package:skunkworks/providers/notification_provider.dart';

class NotificationsScreen extends StatefulWidget {
  final Function(NavTab) onNavigate;

  const NotificationsScreen({super.key, required this.onNavigate});

  @override
  State<NotificationsScreen> createState() => _NotificationsScreenState();
}

class _NotificationsScreenState extends State<NotificationsScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<NotificationProvider>().loadNotifications();
    });
  }

  String _notificationBody(Notification n) {
    final actorName = _actorDisplayName(n.actor);
    final albumName = n.album?['name'] as String? ?? 'an album';
    switch (n.type) {
      case 'album_invite':
        return '$actorName invited you to album $albumName';
      case 'album_invite_accepted':
        return '$actorName accepted your invite to $albumName';
      case 'album_photo_added':
        return '$actorName added a photo to $albumName';
      default:
        return '';
    }
  }

  String _actorDisplayName(Map<String, dynamic>? actor) {
    if (actor == null) return 'Someone';
    final username = actor['username'];
    final phoneNumber = actor['phoneNumber'];
    if (username != null && username.toString().isNotEmpty) return username.toString();
    if (phoneNumber != null && phoneNumber.toString().isNotEmpty) return phoneNumber.toString();
    return 'Someone';
  }

  void _onNotificationTap(Notification n) async {
    final provider = context.read<NotificationProvider>();
    if (n.id != null) {
      await provider.markAsRead(n.id!);
    }
    if (!mounted) return;
    if (n.type == 'album_invite') {
      Navigator.pushReplacement(
        context,
        MaterialPageRoute(
          builder: (context) => AlbumInvitesScreen(onNavigate: widget.onNavigate),
        ),
      );
    } else if (n.albumId != null) {
      final albumName = n.album?['name'] as String? ?? 'Album';
      Navigator.push(
        context,
        MaterialPageRoute(
          builder: (context) => AlbumDetailScreen(
            albumId: n.albumId!,
            albumName: albumName,
            onNavigate: widget.onNavigate,
          ),
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.warmWhite,
      appBar: AppBar(
        backgroundColor: AppColors.warmWhite,
        elevation: 0,
        title: const Text(
          'Notifications',
          style: TextStyle(
            color: AppColors.graphiteInk,
            fontWeight: FontWeight.w600,
            fontSize: 18,
          ),
        ),
        actions: [
          Consumer<NotificationProvider>(
            builder: (context, provider, _) {
              if (provider.unreadCount > 0) {
                return TextButton(
                  onPressed: () async {
                    await provider.markAllAsRead();
                  },
                  child: const Text('Mark all read'),
                );
              }
              return const SizedBox.shrink();
            },
          ),
        ],
      ),
      body: Consumer<NotificationProvider>(
        builder: (context, provider, _) {
          if (provider.loading && provider.notifications.isEmpty) {
            return const Center(child: CircularProgressIndicator());
          }
          if (provider.error != null && provider.notifications.isEmpty) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(provider.error!, textAlign: TextAlign.center, style: TextStyle(color: Colors.grey.shade700)),
                  const SizedBox(height: 16),
                  TextButton(onPressed: provider.loadNotifications, child: const Text('Retry')),
                ],
              ),
            );
          }
          if (provider.notifications.isEmpty) {
            return Center(
              child: Text(
                'No notifications yet',
                style: TextStyle(fontSize: 16, color: Colors.grey.shade600),
              ),
            );
          }
          return RefreshIndicator(
            onRefresh: provider.loadNotifications,
            child: ListView.builder(
              padding: const EdgeInsets.all(16),
              itemCount: provider.notifications.length,
              itemBuilder: (context, index) {
                final n = provider.notifications[index];
                return Card(
                  margin: const EdgeInsets.only(bottom: 12),
                  color: n.isRead ? null : AppColors.softRealBlue.withValues(alpha: 0.05),
                  child: ListTile(
                    leading: CircleAvatar(
                      backgroundColor: AppColors.softRealBlue.withValues(alpha: 0.2),
                      child: Icon(Icons.notifications_outlined, color: AppColors.softRealBlue, size: 24),
                    ),
                    title: Text(
                      _notificationBody(n),
                      style: TextStyle(
                        fontWeight: n.isRead ? FontWeight.normal : FontWeight.w600,
                        fontSize: 14,
                      ),
                    ),
                    subtitle: n.createdAt != null
                        ? Text(
                            _formatDate(n.createdAt!),
                            style: TextStyle(fontSize: 12, color: Colors.grey.shade600),
                          )
                        : null,
                    onTap: () => _onNotificationTap(n),
                  ),
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

  String _formatDate(DateTime d) {
    final now = DateTime.now();
    final diff = now.difference(d);
    if (diff.inMinutes < 1) return 'Just now';
    if (diff.inMinutes < 60) return '${diff.inMinutes}m ago';
    if (diff.inHours < 24) return '${diff.inHours}h ago';
    if (diff.inDays < 7) return '${diff.inDays}d ago';
    return '${d.month}/${d.day}/${d.year}';
  }
}
