import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';

import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/screens/login_screen.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/auth_service.dart';
import 'package:travel_angels/services/location_service.dart';
import 'package:travel_angels/services/media_service.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/location_picker.dart';

LocationCandidate _place(String name) => LocationCandidate(
  placeId: name,
  name: name,
  formattedAddress: '$name, Portugal',
  latitude: 38.7223,
  longitude: -9.1393,
);

class _Locations extends LocationService {
  _Locations() : super(APIClient('https://example.test'));
  final pending = <String, Completer<List<LocationCandidate>>>{};
  bool immediate = false;
  @override
  Future<List<LocationCandidate>> searchLocations(String query) {
    if (immediate) return Future.value([_place('Lisbon')]);
    return (pending[query] = Completer<List<LocationCandidate>>()).future;
  }
}

class _Auth extends AuthProvider {
  _Auth() : super(AuthService(APIClient('https://example.test')));
  final saves = <Map<String, dynamic>>[];
  @override
  Future<void> verifyToken() async {}
  @override
  bool get isWaitingForVerificationCode => true;
  @override
  bool get isRegister => true;
  @override
  bool get loading => false;
  @override
  User? get user => User(id: 'new-user');
  @override
  Future<void> logister(String phoneNumber, String code) async {}
  @override
  Future<bool> validateUsername(String username) async => true;
  @override
  Future<void> updateProfile({
    String? username,
    String? profilePictureUrl,
    DateTime? dateOfBirth,
    String? gender,
    double? latitude,
    double? longitude,
    String? locationAddress,
    String? bio,
  }) async {
    saves.add({
      'profilePictureUrl': profilePictureUrl,
      'latitude': latitude,
      'longitude': longitude,
      'locationAddress': locationAddress,
    });
  }
}

class _Media extends MediaService {
  _Media() : super(APIClient('https://example.test'));
  Uint8List? uploaded;
  String? filename;
  @override
  Future<String?> uploadProfilePictureBytes(
    Uint8List bytes,
    String filename,
    String userId,
  ) async {
    uploaded = bytes;
    this.filename = filename;
    return 'https://example.test/portrait.png';
  }
}

class _Picker extends FilePicker {
  Uint8List photo = base64Decode(
    'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRZkAAAAASUVORK5CYII=',
  );
  bool requestedBytes = false;
  @override
  Future<FilePickerResult?> pickFiles({
    String? dialogTitle,
    String? initialDirectory,
    FileType type = FileType.any,
    List<String>? allowedExtensions,
    Function(FilePickerStatus)? onFileLoading,
    bool allowCompression = true,
    int compressionQuality = 30,
    bool allowMultiple = false,
    bool withData = false,
    bool withReadStream = false,
    bool lockParentWindow = false,
    bool readSequential = false,
  }) async {
    requestedBytes = withData;
    return FilePickerResult([
      PlatformFile(name: 'portrait.png', size: photo.length, bytes: photo),
    ]);
  }
}

Finder _field(String label) => find.byWidgetPredicate(
  (widget) => widget is TextField && widget.decoration?.labelText == label,
);

Future<void> _tap(WidgetTester tester, Finder finder) async {
  await tester.ensureVisible(finder);
  await tester.pumpAndSettle();
  await tester.tap(finder);
  await tester.pumpAndSettle();
}

Future<void> _openSignup(WidgetTester tester, _Auth auth, _Media media) async {
  addTearDown(auth.dispose);
  await tester.pumpWidget(
    ChangeNotifierProvider<AuthProvider>.value(
      value: auth,
      child: MaterialApp(
        theme: AppTheme.lightTheme,
        home: LoginScreen(
          locationService: _Locations()..immediate = true,
          mediaService: media,
          showLocationMap: false,
        ),
      ),
    ),
  );
  await tester.enterText(_field('Verification code'), '123456');
  await _tap(tester, find.text('Register'));
  expect(find.text('Set up your profile'), findsOneWidget);
}

void main() {
  testWidgets('map-free location search selects coordinates and clears edits', (
    tester,
  ) async {
    final service = _Locations()..immediate = true;
    LocationCandidate? selected;
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: LocationPicker(
            locationService: service,
            showMap: false,
            onLocationSelected: (latitude, longitude, address) =>
                selected = LocationCandidate(
                  placeId: '',
                  name: '',
                  formattedAddress: address,
                  latitude: latitude,
                  longitude: longitude,
                ),
            onLocationCleared: () => selected = null,
          ),
        ),
      ),
    );
    expect(find.byType(GoogleMap), findsNothing);
    await tester.enterText(_field('Search location'), 'Lis');
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Lisbon'));
    await tester.pumpAndSettle();
    expect(selected?.formattedAddress, 'Lisbon, Portugal');
    expect(selected?.latitude, 38.7223);
    await tester.enterText(_field('Search location'), '');
    await tester.pumpAndSettle();
    expect(selected, isNull);
    expect(find.text('Lisbon, Portugal'), findsNothing);
  });

  testWidgets(
    'location searches discard stale responses and survive disposal',
    (tester) async {
      final service = _Locations();
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: LocationPicker(locationService: service, showMap: false),
          ),
        ),
      );
      await tester.enterText(_field('Search location'), 'old');
      await tester.pump(const Duration(milliseconds: 500));
      await tester.enterText(_field('Search location'), 'new');
      await tester.pump(const Duration(milliseconds: 500));
      service.pending['new']!.complete([_place('New place')]);
      await tester.pumpAndSettle();
      service.pending['old']!.complete([_place('Old place')]);
      await tester.pumpAndSettle();
      expect(find.text('New place'), findsOneWidget);
      expect(find.text('Old place'), findsNothing);
      await tester.enterText(_field('Search location'), 'removed');
      await tester.pump(const Duration(milliseconds: 500));
      await tester.pumpWidget(const SizedBox());
      service.pending['removed']!.completeError(StateError('offline'));
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets('map-free location errors remain recoverable', (tester) async {
    final service = _Locations();
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: LocationPicker(locationService: service, showMap: false),
        ),
      ),
    );
    await tester.enterText(_field('Search location'), 'Lis');
    await tester.pump(const Duration(milliseconds: 500));
    service.pending['Lis']!.completeError(StateError('offline'));
    await tester.pumpAndSettle();
    expect(
      find.text('Locations could not be loaded. Try your search again.'),
      findsOneWidget,
    );
    service.immediate = true;
    await tester.enterText(_field('Search location'), 'Lisbon');
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pumpAndSettle();
    expect(find.text('Lisbon'), findsWidgets);
  });

  testWidgets(
    'signup previews and uploads a pathless image with selected place',
    (tester) async {
      final auth = _Auth();
      final media = _Media();
      final picker = _Picker();
      FilePicker.platform = picker;
      await _openSignup(tester, auth, media);
      await tester.enterText(_field('Username'), 'traveler');
      await tester.pumpAndSettle();
      await _tap(tester, find.text('Select date of birth'));
      await _tap(tester, find.text('OK'));
      await _tap(tester, find.byType(DropdownButtonFormField<String>));
      await _tap(tester, find.text('Prefer not to say').last);
      await tester.ensureVisible(_field('Search location'));
      await tester.enterText(_field('Search location'), 'Lis');
      await tester.pump(const Duration(milliseconds: 500));
      await tester.pumpAndSettle();
      await _tap(tester, find.text('Lisbon'));
      await tester.ensureVisible(find.text('Pick Image'));
      await tester.pumpAndSettle();
      await tester.runAsync(() async {
        await tester.tap(find.text('Pick Image'));
        await Future<void>.delayed(const Duration(milliseconds: 100));
      });
      await tester.pumpAndSettle();
      expect(picker.requestedBytes, isTrue);
      expect(
        find.byWidgetPredicate(
          (widget) => widget is Image && widget.image is MemoryImage,
        ),
        findsOneWidget,
      );
      await _tap(tester, find.text('Complete Profile'));
      expect(media.uploaded, picker.photo);
      expect(media.filename, 'portrait.png');
      expect(
        auth.saves.single['profilePictureUrl'],
        'https://example.test/portrait.png',
      );
      expect(auth.saves.single['locationAddress'], 'Lisbon, Portugal');
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets('signup rejects oversized image bytes before upload', (
    tester,
  ) async {
    final media = _Media();
    FilePicker.platform = _Picker()..photo = Uint8List(8 * 1024 * 1024 + 1);
    await _openSignup(tester, _Auth(), media);
    await _tap(tester, find.text('Pick Image'));
    expect(find.text('Choose a photo smaller than 8 MB.'), findsOneWidget);
    expect(media.uploaded, isNull);
  });
}
