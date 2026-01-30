import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';

import 'config/router.dart';
import 'constants/api_constants.dart';
import 'models/location.dart';
import 'providers/auth_provider.dart';
import 'providers/friend_provider.dart';
import 'providers/location_provider.dart';
import 'providers/party_provider.dart';
import 'services/api_client.dart';
import 'services/auth_service.dart';
import 'services/friend_service.dart';
import 'services/location_service.dart';
import 'services/media_service.dart';
import 'services/activity_service.dart';
import 'services/admin_service.dart';
import 'services/chat_service.dart';
import 'services/inventory_service.dart';
import 'services/party_service.dart';
import 'services/poi_service.dart';
import 'services/quest_log_service.dart';
import 'services/tags_service.dart';
import 'providers/activity_feed_provider.dart';
import 'providers/completed_task_provider.dart';
import 'providers/discoveries_provider.dart';
import 'providers/inventory_modal_provider.dart';
import 'providers/log_provider.dart';
import 'providers/tags_provider.dart';
import 'providers/quest_log_provider.dart';
import 'providers/zone_provider.dart';

void main() {
  runApp(const SonarApp());
}

class SonarApp extends StatelessWidget {
  const SonarApp({super.key});

  @override
  Widget build(BuildContext context) {
    final apiClient = ApiClient(
      ApiConstants.baseUrl,
      onAuthError: () {
        // Router redirect will send to / on 401/403; auth provider logout
        // is triggered elsewhere when we detect auth error
      },
    );
    final authService = AuthService(apiClient);
    final authProvider = AuthProvider(authService);
    final locationService = LocationService();
    final locationProvider = LocationProvider(locationService);
    final mediaService = MediaService(apiClient);
    final poiService = PoiService(apiClient);
    final adminService = AdminService(apiClient);
    final partyService = PartyService(apiClient);
    final friendService = FriendService(apiClient);
    final activityService = ActivityService(apiClient);
    final chatService = ChatService(apiClient);
    final tagsService = TagsService(apiClient);
    final questLogService = QuestLogService(apiClient);
    final inventoryService = InventoryService(apiClient);
    final partyProvider = PartyProvider(partyService);
    final friendProvider = FriendProvider(friendService);
    final activityFeedProvider = ActivityFeedProvider(activityService);
    final logProvider = LogProvider(chatService);
    final tagsProvider = TagsProvider(tagsService);
    final inventoryModalProvider = InventoryModalProvider();
    final completedTaskProvider = CompletedTaskProvider();
    final zoneProvider = ZoneProvider();
    final discoveriesProvider = DiscoveriesProvider(poiService, authProvider);

    apiClient.setOnAuthError(() {
      authProvider.logout();
    });

    AppLocation? getLocation() => locationProvider.location;
    apiClient.setGetLocation(getLocation);

    final rootKey = GlobalKey<NavigatorState>();
    final router = createRouter(rootKey);

    return MultiProvider(
      providers: [
        ChangeNotifierProvider<AuthProvider>.value(value: authProvider),
        ChangeNotifierProvider<LocationProvider>.value(value: locationProvider),
        Provider<MediaService>.value(value: mediaService),
        Provider<PoiService>.value(value: poiService),
        Provider<AdminService>.value(value: adminService),
        Provider<InventoryService>.value(value: inventoryService),
        ChangeNotifierProvider<PartyProvider>.value(value: partyProvider),
        ChangeNotifierProvider<FriendProvider>.value(value: friendProvider),
        ChangeNotifierProvider<ActivityFeedProvider>.value(value: activityFeedProvider),
        ChangeNotifierProvider<LogProvider>.value(value: logProvider),
        ChangeNotifierProvider<TagsProvider>.value(value: tagsProvider),
        ChangeNotifierProvider<InventoryModalProvider>.value(value: inventoryModalProvider),
        ChangeNotifierProvider<CompletedTaskProvider>.value(value: completedTaskProvider),
        ChangeNotifierProvider<ZoneProvider>.value(value: zoneProvider),
        ChangeNotifierProvider<DiscoveriesProvider>.value(value: discoveriesProvider),
        ChangeNotifierProvider<QuestLogProvider>.value(
          value: QuestLogProvider(questLogService, zoneProvider, tagsProvider),
        ),
      ],
      child: MaterialApp.router(
        title: 'Find your crew',
        theme: ThemeData(
          useMaterial3: true,
          fontFamily: GoogleFonts.inter().fontFamily,
          textTheme: GoogleFonts.interTextTheme(),
          colorScheme: ColorScheme.fromSeed(seedColor: Colors.blue),
        ),
        routerConfig: router,
      ),
    );
  }
}
