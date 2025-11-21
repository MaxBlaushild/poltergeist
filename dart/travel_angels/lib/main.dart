import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/providers/user_level_provider.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/auth_service.dart';
import 'package:travel_angels/services/user_level_service.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/home_widget.dart';

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
    final userLevelService = UserLevelService(apiClient);
    final userLevelProvider = UserLevelProvider(userLevelService);

    return MultiProvider(
      providers: [
        ChangeNotifierProvider.value(value: authProvider),
        ChangeNotifierProvider.value(value: userLevelProvider),
      ],
      child: MaterialApp(
        title: 'Travel Angels',
        theme: AppTheme.lightTheme,
        home: const HomeWidget(),
      ),
    );
  }
}
