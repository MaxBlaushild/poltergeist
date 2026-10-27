import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';
import 'dart:ui' as ui;

import 'package:dio/dio.dart';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/auth_service.dart';
import 'package:travel_angels/services/location_service.dart';
import 'package:travel_angels/services/media_service.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/edit_profile_dialog.dart';

import 'support/app_fonts.dart';

Finder _key(String key) => find.byKey(ValueKey(key));

class _Auth extends AuthProvider {
  _Auth(this.current) : super(AuthService(APIClient('https://example.test')));
  User? current;
  Object? failure;
  final saves = <Map<String, dynamic>>[];
  @override
  User? get user => current;
  @override
  Future<void> verifyToken() async {}
  void change(User? user) {
    current = user;
    notifyListeners();
  }

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
      'username': username,
      'profilePictureUrl': profilePictureUrl,
      'dateOfBirth': dateOfBirth,
      'gender': gender,
      'latitude': latitude,
      'longitude': longitude,
      'locationAddress': locationAddress,
      'bio': bio,
    });
    if (failure != null) throw failure!;
    notifyListeners();
  }
}

class _Locations extends LocationService {
  _Locations() : super(APIClient('https://example.test'));
  @override
  Future<List<LocationCandidate>> searchLocations(String query) async => [
    LocationCandidate(
      placeId: 'lisbon',
      name: 'Lisbon',
      formattedAddress: 'Lisbon, Portugal',
      latitude: 38.7223,
      longitude: -9.1393,
    ),
  ];
}

class _Media extends MediaService {
  _Media() : super(APIClient('https://example.test'));
  Uint8List? uploaded;
  final pending = Completer<String?>();
  bool wait = false;
  @override
  Future<String?> uploadProfilePictureBytes(
    Uint8List bytes,
    String filename,
    String userId,
  ) async {
    uploaded = bytes;
    return wait ? pending.future : 'https://example.test/avatar.png';
  }
}

class _Picker extends FilePicker {
  final photo = base64Decode(
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

Future<GlobalKey> _open(
  WidgetTester tester,
  _Auth auth, {
  _Media? media,
  double width = 800,
  double scale = 1,
}) async {
  final size = Size(width, 844);
  await tester.binding.setSurfaceSize(size);
  addTearDown(() => tester.binding.setSurfaceSize(null));
  addTearDown(auth.dispose);
  final boundary = GlobalKey();
  await tester.pumpWidget(
    ChangeNotifierProvider<AuthProvider>.value(
      value: auth,
      child: RepaintBoundary(
        key: boundary,
        child: MaterialApp(
          debugShowCheckedModeBanner: false,
          theme: AppTheme.lightTheme,
          builder: (_, child) => MediaQuery(
            data: MediaQueryData(
              size: size,
              textScaler: TextScaler.linear(scale),
            ),
            child: child!,
          ),
          home: Builder(
            builder: (context) => Scaffold(
              body: TextButton(
                onPressed: () => showDialog<bool>(
                  context: context,
                  builder: (_) => EditProfileDialog(
                    user: auth.user,
                    locationService: _Locations(),
                    mediaService: media ?? _Media(),
                  ),
                ),
                child: const Text('Edit'),
              ),
            ),
          ),
        ),
      ),
    ),
  );
  await tester.tap(find.text('Edit'));
  await tester.pumpAndSettle();
  return boundary;
}

Future<void> _tap(WidgetTester tester, Finder finder) async {
  await tester.ensureVisible(finder);
  await tester.pumpAndSettle();
  await tester.tap(finder);
  await tester.pumpAndSettle();
}

Future<void> _pickPhoto(WidgetTester tester) async {
  await tester.runAsync(() async {
    await tester.tap(_key('profile-photo'));
    // ImageDescriptor uses the engine's asynchronous codec, outside fake time.
    await Future<void>.delayed(const Duration(milliseconds: 100));
  });
  await tester.pumpAndSettle();
}

void main() {
  setUpAll(loadAppFonts);
  setUp(() => FilePicker.platform = _Picker());

  testWidgets('saves basic profile without demographics and clears the bio', (
    tester,
  ) async {
    final auth = _Auth(User(id: 'me', username: 'old.name', bio: 'Old bio'));
    await _open(tester, auth);
    await tester.enterText(_key('profile-bio'), '');
    await _tap(tester, _key('profile-edit-save'));
    expect(auth.saves.single['bio'], '');
    expect(auth.saves.single['username'], isNull);
    expect(auth.saves.single['dateOfBirth'], isNull);
    expect(auth.saves.single['gender'], isNull);
    expect(auth.saves.single['locationAddress'], isNull);
    expect(find.byType(EditProfileDialog), findsNothing);
  });

  testWidgets(
    'validates changed username and preserves selected place coordinates',
    (tester) async {
      final auth = _Auth(User(id: 'me', username: 'traveler'));
      await _open(tester, auth);
      await tester.enterText(_key('profile-username'), 'bad name');
      await _tap(tester, _key('profile-edit-save'));
      expect(find.text('Use letters and numbers only.'), findsOneWidget);
      expect(auth.saves, isEmpty);
      await tester.enterText(_key('profile-username'), 'newtraveler');
      await tester.enterText(_key('profile-home-base'), 'Lis');
      await tester.pump(const Duration(milliseconds: 450));
      await tester.pumpAndSettle();
      await _tap(tester, _key('profile-edit-save'));
      expect(auth.saves, isEmpty);
      expect(
        find.text('Choose a matching place from the search results.'),
        findsOneWidget,
      );
      await _tap(tester, _key('profile-place-lisbon'));
      await _tap(tester, _key('profile-edit-save'));
      expect(auth.saves.single['username'], 'newtraveler');
      expect(auth.saves.single['locationAddress'], 'Lisbon, Portugal');
      expect(auth.saves.single['latitude'], 38.7223);
      expect(auth.saves.single['longitude'], -9.1393);
    },
  );

  testWidgets('server errors preserve changes for a retry', (tester) async {
    final auth = _Auth(User(id: 'me', username: 'traveler'));
    final request = RequestOptions(path: '/profile');
    auth.failure = DioException(
      requestOptions: request,
      response: Response(
        requestOptions: request,
        statusCode: 400,
        data: {'error': 'Username already taken'},
      ),
    );
    await _open(tester, auth);
    await tester.enterText(_key('profile-username'), 'newtraveler');
    await _tap(tester, _key('profile-edit-save'));
    expect(find.text('Username already taken'), findsOneWidget);
    expect(_key('profile-edit-error').hitTestable(), findsOneWidget);
    expect(
      tester.widget<TextFormField>(_key('profile-username')).controller!.text,
      'newtraveler',
    );
    auth.failure = null;
    await _tap(tester, _key('profile-edit-save'));
    expect(auth.saves, hasLength(2));
    expect(find.byType(EditProfileDialog), findsNothing);
  });

  testWidgets('uploads a photo that has bytes and no native file path', (
    tester,
  ) async {
    final picker = _Picker();
    final originalPicker = FilePicker.platform;
    FilePicker.platform = picker;
    addTearDown(() => FilePicker.platform = originalPicker);
    final auth = _Auth(User(id: 'me', username: 'traveler'));
    final media = _Media();
    await _open(tester, auth, media: media);
    await _pickPhoto(tester);
    expect(picker.requestedBytes, isTrue);
    await _tap(tester, _key('profile-edit-save'));
    expect(media.uploaded, picker.photo);
    expect(
      auth.saves.single['profilePictureUrl'],
      'https://example.test/avatar.png',
    );
  });

  testWidgets(
    'account changes discard editor and prevent an upload from saving to another account',
    (tester) async {
      final originalPicker = FilePicker.platform;
      FilePicker.platform = _Picker();
      addTearDown(() => FilePicker.platform = originalPicker);
      final auth = _Auth(User(id: 'me', username: 'traveler'));
      final media = _Media()..wait = true;
      await _open(tester, auth, media: media);
      await _pickPhoto(tester);
      await _tap(tester, _key('profile-edit-save'));
      expect(media.uploaded, isNotNull);
      auth.change(User(id: 'someone-else', username: 'other'));
      await tester.pumpAndSettle();
      expect(find.byType(EditProfileDialog), findsNothing);
      media.pending.complete('https://example.test/avatar.png');
      await tester.pumpAndSettle();
      expect(auth.saves, isEmpty);
    },
  );

  for (final (width, scale) in [(800.0, 1.0), (390.0, 1.0), (320.0, 2.0)]) {
    testWidgets('editor controls remain usable at $width with ${scale}x text', (
      tester,
    ) async {
      final auth = _Auth(
        User(
          id: 'me',
          username: 'traveler',
          locationAddress: 'Lisbon, Portugal',
          bio: 'Slow mornings and long walks.',
        ),
      );
      final boundary = await _open(tester, auth, width: width, scale: scale);
      expect(tester.takeException(), isNull);
      expect(_key('profile-edit-save').hitTestable(), findsOneWidget);
      if (const bool.fromEnvironment('TRIP_SCREENSHOTS')) {
        await tester.runAsync(() async {
          final image =
              await (boundary.currentContext!.findRenderObject()!
                      as RenderRepaintBoundary)
                  .toImage();
          final bytes = await image.toByteData(format: ui.ImageByteFormat.png);
          await File(
            '/private/tmp/travel-angels-edit-profile-${width.toInt()}-${scale}x.png',
          ).writeAsBytes(bytes!.buffer.asUint8List());
          image.dispose();
        });
      }
      await tester.ensureVisible(_key('profile-bio'));
      await tester.pumpAndSettle();
      await tester.enterText(_key('profile-bio'), 'New travel bio');
      await _tap(tester, _key('profile-edit-save'));
      expect(auth.saves.single['bio'], 'New travel bio');
      expect(tester.takeException(), isNull);
    });
  }
}
