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
import 'services/character_stats_service.dart';
import 'services/inventory_service.dart';
import 'services/party_service.dart';
import 'services/poi_service.dart';
import 'services/quest_log_service.dart';
import 'services/tags_service.dart';
import 'services/user_level_service.dart';
import 'services/user_character_service.dart';
import 'providers/activity_feed_provider.dart';
import 'providers/completed_task_provider.dart';
import 'providers/discoveries_provider.dart';
import 'providers/inventory_modal_provider.dart';
import 'providers/log_provider.dart';
import 'providers/tags_provider.dart';
import 'providers/quest_log_provider.dart';
import 'providers/quest_filter_provider.dart';
import 'providers/zone_provider.dart';
import 'providers/map_focus_provider.dart';
import 'providers/character_stats_provider.dart';
import 'providers/user_level_provider.dart';

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
    final characterStatsService = CharacterStatsService(apiClient);
    final tagsService = TagsService(apiClient);
    final questLogService = QuestLogService(apiClient);
    final inventoryService = InventoryService(apiClient);
    final userLevelService = UserLevelService(apiClient);
    final userCharacterService = UserCharacterService(apiClient);
    final partyProvider = PartyProvider(partyService);
    final friendProvider = FriendProvider(friendService);
    final activityFeedProvider = ActivityFeedProvider(activityService);
    final logProvider = LogProvider(chatService);
    final tagsProvider = TagsProvider(tagsService);
    final questFilterProvider = QuestFilterProvider();
    final inventoryModalProvider = InventoryModalProvider();
    final completedTaskProvider = CompletedTaskProvider();
    final zoneProvider = ZoneProvider();
    final discoveriesProvider = DiscoveriesProvider(poiService, authProvider);
    final mapFocusProvider = MapFocusProvider();
    final userLevelProvider = UserLevelProvider(userLevelService, authProvider);

    apiClient.setOnAuthError(() {
      authProvider.logout();
    });

    AppLocation? getLocation() => locationProvider.location;
    apiClient.setGetLocation(getLocation);

    final router = createRouter(refreshListenable: authProvider);

    return MultiProvider(
      providers: [
        ChangeNotifierProvider<AuthProvider>.value(value: authProvider),
        ChangeNotifierProvider<LocationProvider>.value(value: locationProvider),
        Provider<MediaService>.value(value: mediaService),
        Provider<PoiService>.value(value: poiService),
        Provider<AdminService>.value(value: adminService),
        Provider<InventoryService>.value(value: inventoryService),
        Provider<UserCharacterService>.value(value: userCharacterService),
        ChangeNotifierProvider<PartyProvider>.value(value: partyProvider),
        ChangeNotifierProvider<FriendProvider>.value(value: friendProvider),
        ChangeNotifierProvider<ActivityFeedProvider>.value(value: activityFeedProvider),
        ChangeNotifierProxyProvider2<AuthProvider, ActivityFeedProvider,
            CharacterStatsProvider>(
          create: (_) => CharacterStatsProvider(characterStatsService),
          update: (_, auth, feed, stats) {
            stats ??= CharacterStatsProvider(characterStatsService);
            stats.updateAuth(auth);
            stats.updateActivityFeed(feed);
            return stats;
          },
        ),
        ChangeNotifierProvider<LogProvider>.value(value: logProvider),
        ChangeNotifierProvider<TagsProvider>.value(value: tagsProvider),
        ChangeNotifierProvider<QuestFilterProvider>.value(value: questFilterProvider),
        ChangeNotifierProvider<InventoryModalProvider>.value(value: inventoryModalProvider),
        ChangeNotifierProvider<CompletedTaskProvider>.value(value: completedTaskProvider),
        ChangeNotifierProvider<ZoneProvider>.value(value: zoneProvider),
        ChangeNotifierProvider<DiscoveriesProvider>.value(value: discoveriesProvider),
        ChangeNotifierProvider<MapFocusProvider>.value(value: mapFocusProvider),
        ChangeNotifierProvider<UserLevelProvider>.value(value: userLevelProvider),
        ChangeNotifierProvider<QuestLogProvider>.value(
          value: QuestLogProvider(
            questLogService,
            zoneProvider,
            tagsProvider,
            questFilterProvider,
          ),
        ),
      ],
      child: MaterialApp.router(
        title: 'Find your crew',
        theme: ThemeData(
          useMaterial3: true,
          fontFamily: GoogleFonts.crimsonPro().fontFamily,
          textTheme: () {
            final base = GoogleFonts.crimsonProTextTheme();
            final display = GoogleFonts.cinzelTextTheme();
            return base.copyWith(
              displayLarge: display.displayLarge,
              displayMedium: display.displayMedium,
              displaySmall: display.displaySmall,
              headlineLarge: display.headlineLarge,
              headlineMedium: display.headlineMedium,
              headlineSmall: display.headlineSmall,
              titleLarge: display.titleLarge,
              titleMedium: display.titleMedium,
              titleSmall: display.titleSmall,
            );
          }(),
          colorScheme: const ColorScheme.light(
            primary: Color(0xFF355C7D),
            onPrimary: Color(0xFFFDF6E3),
            secondary: Color(0xFF6B8E23),
            onSecondary: Color(0xFFFDF6E3),
            tertiary: Color(0xFFB87333),
            onTertiary: Color(0xFFFDF6E3),
            surface: Color(0xFFF4E9D6),
            onSurface: Color(0xFF2D2416),
            surfaceVariant: Color(0xFFE7D6B3),
            onSurfaceVariant: Color(0xFF3B2F1C),
            background: Color(0xFFF1E6CF),
            onBackground: Color(0xFF2D2416),
            error: Color(0xFFB00020),
            onError: Color(0xFFFFF7F0),
            outline: Color(0xFFB8A37E),
            outlineVariant: Color(0xFFD7C39F),
          ),
          scaffoldBackgroundColor: const Color(0xFFF1E6CF),
          appBarTheme: const AppBarTheme(
            backgroundColor: Color(0xFFF4E9D6),
            foregroundColor: Color(0xFF2D2416),
            elevation: 0,
            scrolledUnderElevation: 2,
            shadowColor: Color(0x332D2416),
            centerTitle: false,
            iconTheme: IconThemeData(color: Color(0xFF2D2416)),
            titleTextStyle: TextStyle(
              fontFamily: 'Cinzel',
              fontSize: 20,
              fontWeight: FontWeight.w600,
              color: Color(0xFF2D2416),
            ),
            shape: Border(
              bottom: BorderSide(
                color: Color(0xFFD7C39F),
                width: 1,
              ),
            ),
          ),
          bottomSheetTheme: const BottomSheetThemeData(
            backgroundColor: Color(0xFFF4E9D6),
            elevation: 0,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
            ),
          ),
          cardTheme: CardThemeData(
            color: const Color(0xFFE7D6B3),
            elevation: 2,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(16),
            ),
          ),
          iconTheme: const IconThemeData(
            color: Color(0xFF3B2F1C),
          ),
          iconButtonTheme: IconButtonThemeData(
            style: IconButton.styleFrom(
              foregroundColor: const Color(0xFF3B2F1C),
            ),
          ),
          listTileTheme: const ListTileThemeData(
            iconColor: Color(0xFF3B2F1C),
            textColor: Color(0xFF2D2416),
          ),
          dividerTheme: const DividerThemeData(
            color: Color(0xFFD7C39F),
            thickness: 1,
            space: 1,
          ),
          chipTheme: ChipThemeData(
            backgroundColor: const Color(0xFFE7D6B3),
            disabledColor: const Color(0xFFE7D6B3),
            selectedColor: const Color(0xFF355C7D),
            secondarySelectedColor: const Color(0xFF355C7D),
            labelStyle: const TextStyle(color: Color(0xFF2D2416)),
            secondaryLabelStyle: const TextStyle(color: Color(0xFFFDF6E3)),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(999),
            ),
          ),
          filledButtonTheme: FilledButtonThemeData(
            style: FilledButton.styleFrom(
              backgroundColor: const Color(0xFF355C7D),
              foregroundColor: const Color(0xFFFDF6E3),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(14),
              ),
              padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 12),
              textStyle: GoogleFonts.cinzel(fontWeight: FontWeight.w600),
            ),
          ),
          outlinedButtonTheme: OutlinedButtonThemeData(
            style: OutlinedButton.styleFrom(
              foregroundColor: const Color(0xFF355C7D),
              side: const BorderSide(color: Color(0xFF355C7D)),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(14),
              ),
              padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 12),
              textStyle: GoogleFonts.cinzel(fontWeight: FontWeight.w600),
            ),
          ),
          textButtonTheme: TextButtonThemeData(
            style: TextButton.styleFrom(
              foregroundColor: const Color(0xFF355C7D),
              textStyle: GoogleFonts.cinzel(fontWeight: FontWeight.w600),
            ),
          ),
          inputDecorationTheme: InputDecorationTheme(
            filled: true,
            fillColor: const Color(0xFFE7D6B3),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: Color(0xFFB8A37E)),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: Color(0xFFB8A37E)),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: Color(0xFF355C7D), width: 1.5),
            ),
            hintStyle: const TextStyle(color: Color(0xFF6E5B3B)),
          ),
        ),
        routerConfig: router,
      ),
    );
  }
}
