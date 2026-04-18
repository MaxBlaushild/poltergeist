import 'dart:async';
import 'dart:collection';

import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:unclaimed_streets/models/character_stats.dart';
import 'package:unclaimed_streets/models/user.dart';
import 'package:unclaimed_streets/providers/auth_provider.dart';
import 'package:unclaimed_streets/providers/character_stats_provider.dart';
import 'package:unclaimed_streets/services/api_client.dart';
import 'package:unclaimed_streets/services/auth_service.dart';
import 'package:unclaimed_streets/services/character_stats_service.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() {
    SharedPreferences.setMockInitialValues({});
  });

  test(
    'refresh waits for an in-flight stats request before returning',
    () async {
      final service = _FakeCharacterStatsService();
      service.enqueue(Future.value(_stats(level: 1, health: 37, mana: 19)));

      final authProvider = AuthProvider(
        AuthService(ApiClient('http://localhost')),
      );
      await Future<void>.delayed(Duration.zero);
      authProvider.setUser(
        const User(
          id: 'user-1',
          phoneNumber: '+15555550123',
          name: 'Test User',
          username: 'tester',
          profilePictureUrl: '',
        ),
      );

      final provider = CharacterStatsProvider(service);
      provider.updateAuth(authProvider);
      await provider.refresh(silent: true);

      final nextResponse = Completer<CharacterStats?>();
      service.enqueue(nextResponse.future);
      final callCountBefore = service.getStatsCallCount;

      final firstRefresh = provider.refresh(silent: true);
      final secondRefresh = provider.refresh(silent: true);

      expect(service.getStatsCallCount, callCountBefore + 1);

      nextResponse.complete(_stats(level: 2, health: 100, mana: 100));
      await Future.wait([firstRefresh, secondRefresh]);

      expect(provider.level, 2);
      expect(provider.health, 100);
      expect(provider.mana, 100);
    },
  );
}

class _FakeCharacterStatsService extends CharacterStatsService {
  _FakeCharacterStatsService() : super(ApiClient('http://localhost'));

  final Queue<Future<CharacterStats?>> _responses =
      Queue<Future<CharacterStats?>>();
  int getStatsCallCount = 0;

  void enqueue(Future<CharacterStats?> response) {
    _responses.add(response);
  }

  @override
  Future<CharacterStats?> getStats() {
    getStatsCallCount += 1;
    if (_responses.isEmpty) {
      return Future<CharacterStats?>.value(null);
    }
    return _responses.removeFirst();
  }
}

CharacterStats _stats({
  required int level,
  required int health,
  required int mana,
}) {
  return CharacterStats(
    strength: 10,
    dexterity: 10,
    constitution: 10,
    intelligence: 10,
    wisdom: 10,
    charisma: 10,
    health: health,
    maxHealth: 100,
    mana: mana,
    maxMana: 100,
    unspentPoints: 0,
    level: level,
  );
}
