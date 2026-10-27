import 'dart:async';
import 'dart:typed_data';
import 'dart:ui' show ImageDescriptor, ImmutableBuffer;

import 'package:dio/dio.dart';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/location_service.dart';
import 'package:travel_angels/services/media_service.dart';
import 'package:travel_angels/theme/app_theme.dart';

class EditProfileDialog extends StatefulWidget {
  const EditProfileDialog({
    super.key,
    this.user,
    this.onSave,
    this.locationService,
    this.mediaService,
  });

  final User? user;
  final VoidCallback? onSave;
  final LocationService? locationService;
  final MediaService? mediaService;

  @override
  State<EditProfileDialog> createState() => _EditProfileDialogState();
}

class _EditProfileDialogState extends State<EditProfileDialog> {
  final _formKey = GlobalKey<FormState>();
  final _username = TextEditingController();
  final _bio = TextEditingController();
  final _homeBase = TextEditingController();
  late final LocationService _locations;
  late final MediaService _media;
  AuthProvider? _auth;
  User? _initial;
  Route<void>? _route;
  bool _initialized = false;
  bool _identityChanged = false;
  bool _saving = false;
  bool _picking = false;
  bool _searching = false;
  String? _error;
  String? _locationError;
  String? _selectedAddress;
  double? _latitude;
  double? _longitude;
  Uint8List? _photo;
  String? _photoName;
  List<LocationCandidate> _results = [];
  Timer? _searchTimer;
  int _searchVersion = 0;

  @override
  void initState() {
    super.initState();
    final client = APIClient(ApiConstants.baseUrl);
    _locations = widget.locationService ?? LocationService(client);
    _media = widget.mediaService ?? MediaService(client);
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final auth = context.read<AuthProvider>();
    if (_auth != auth) {
      _auth?.removeListener(_authChanged);
      _auth = auth;
      auth.addListener(_authChanged);
    }
    _route = ModalRoute.of(context);
    if (!_initialized) {
      _initialized = true;
      _initial = widget.user ?? auth.user;
      _username.text = _initial?.username ?? '';
      _bio.text = _initial?.bio ?? '';
      _selectedAddress = _initial?.locationAddress;
      _homeBase.text = _selectedAddress ?? '';
      _latitude = _initial?.latitude;
      _longitude = _initial?.longitude;
    }
    _checkIdentity();
  }

  bool get _sameIdentity =>
      !_identityChanged &&
      _initial?.id?.isNotEmpty == true &&
      _auth?.user?.id == _initial!.id;

  void _authChanged() {
    if (!mounted) return;
    _checkIdentity();
  }

  void _checkIdentity() {
    if (_sameIdentity || _identityChanged) return;
    _identityChanged = true;
    _searchTimer?.cancel();
    _searchVersion++;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      final route = _route;
      if (route != null && route.isActive) {
        Navigator.of(context).removeRoute(route);
      } else {
        setState(() {});
      }
    });
  }

  @override
  void dispose() {
    _auth?.removeListener(_authChanged);
    _searchTimer?.cancel();
    _username.dispose();
    _bio.dispose();
    _homeBase.dispose();
    super.dispose();
  }

  Future<void> _pickPhoto() async {
    if (!_sameIdentity || _picking || _saving) return;
    setState(() {
      _picking = true;
      _error = null;
    });
    try {
      final result = await FilePicker.platform.pickFiles(
        type: FileType.custom,
        allowedExtensions: const ['jpg', 'jpeg', 'png', 'webp'],
        allowMultiple: false,
        withData: true,
      );
      if (!mounted || !_sameIdentity || result == null) return;
      final file = result.files.single;
      final bytes = file.bytes;
      if (bytes == null || bytes.isEmpty) {
        throw const FormatException(
          'This photo could not be read. Try another image.',
        );
      }
      if (bytes.length > 8 * 1024 * 1024) {
        throw const FormatException('Choose a photo smaller than 8 MB.');
      }
      // Decode before saving so a mislabeled file never becomes the avatar.
      final buffer = await ImmutableBuffer.fromUint8List(bytes);
      try {
        final descriptor = await ImageDescriptor.encoded(buffer);
        try {
          if (descriptor.width * descriptor.height > 20000000) {
            throw const FormatException(
              'Choose a photo smaller than 20 megapixels.',
            );
          }
        } finally {
          descriptor.dispose();
        }
      } finally {
        buffer.dispose();
      }
      if (!mounted || !_sameIdentity) return;
      setState(() {
        _photo = bytes;
        _photoName = file.name;
      });
    } catch (error) {
      if (mounted && _sameIdentity) {
        setState(() {
          _error = error is FormatException
              ? error.message
              : 'We could not open that photo. Choose a JPG, PNG, or WebP image.';
        });
      }
    } finally {
      if (mounted && _sameIdentity) setState(() => _picking = false);
    }
  }

  void _searchHomeBase(String value) {
    _searchTimer?.cancel();
    final version = ++_searchVersion;
    setState(() {
      _results = [];
      _searching = false;
      _locationError = null;
    });
    if (value.trim().length < 2 || value.trim() == _selectedAddress) return;
    _searchTimer = Timer(const Duration(milliseconds: 400), () async {
      if (!mounted || !_sameIdentity) return;
      setState(() => _searching = true);
      try {
        final results = await _locations.searchLocations(value.trim());
        if (!mounted || !_sameIdentity || version != _searchVersion) return;
        setState(() {
          _results = results;
          _searching = false;
          _locationError = results.isEmpty
              ? 'No places found. Try a city or region.'
              : null;
        });
      } catch (_) {
        if (!mounted || !_sameIdentity || version != _searchVersion) return;
        setState(() {
          _searching = false;
          _locationError = 'Could not search places. Try again.';
        });
      }
    });
  }

  void _selectPlace(LocationCandidate place) {
    _searchTimer?.cancel();
    _searchVersion++;
    setState(() {
      _selectedAddress = place.formattedAddress;
      _latitude = place.latitude;
      _longitude = place.longitude;
      _homeBase.text = place.formattedAddress;
      _results = [];
      _searching = false;
      _locationError = null;
    });
    FocusScope.of(context).unfocus();
  }

  Future<void> _save() async {
    if (_saving || _picking || !_sameIdentity) return;
    if (!_formKey.currentState!.validate()) return;
    FocusScope.of(context).unfocus();
    setState(() {
      _saving = true;
      _error = null;
    });
    try {
      String? url;
      if (_photo != null) {
        url = await _media.uploadProfilePictureBytes(
          _photo!,
          _photoName!,
          _initial!.id!,
        );
        if (!mounted || !_sameIdentity) return;
        if (url == null) {
          throw const FormatException(
            'Your photo could not be uploaded. Please try saving again.',
          );
        }
      }
      if (!mounted || !_sameIdentity) return;
      final username = _username.text.trim();
      final locationChanged = _selectedAddress != _initial?.locationAddress;
      await _auth!.updateProfile(
        username: username == (_initial?.username ?? '') ? null : username,
        profilePictureUrl: url,
        bio: _bio.text.trim(),
        latitude: locationChanged ? _latitude : null,
        longitude: locationChanged ? _longitude : null,
        locationAddress: locationChanged ? _selectedAddress : null,
      );
      if (!mounted || !_sameIdentity) return;
      widget.onSave?.call();
      Navigator.of(context).pop(true);
    } catch (error) {
      if (!mounted || !_sameIdentity) return;
      setState(() {
        _saving = false;
        _error = _saveError(error);
      });
    }
  }

  String _saveError(Object error) {
    if (error is FormatException) return error.message;
    if (error is DioException) {
      final data = error.response?.data;
      final message = data is Map ? data['error'] : null;
      if (message is String && message.isNotEmpty) return message;
      if (error.type == DioExceptionType.connectionError ||
          error.type == DioExceptionType.connectionTimeout) {
        return 'Could not connect. Check your connection and try again.';
      }
    }
    return 'Your changes could not be saved. Please try again.';
  }

  Widget _photoControl() {
    final url = _initial?.profilePictureUrl;
    const fallback = ColoredBox(
      color: AppTheme.peach,
      child: Center(
        child: Icon(Icons.person_outline, size: 36, color: AppTheme.ink),
      ),
    );
    return Row(
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        Semantics(
          label: _photo == null
              ? 'Current profile photo'
              : 'Selected profile photo',
          image: true,
          child: ClipRRect(
            borderRadius: BorderRadius.circular(4),
            child: SizedBox(
              width: 64,
              height: 64,
              child: _photo != null
                  ? Image.memory(_photo!, fit: BoxFit.cover, cacheWidth: 192)
                  : url != null && url.isNotEmpty
                  ? Image.network(
                      url,
                      fit: BoxFit.cover,
                      cacheWidth: 192,
                      errorBuilder: (_, _, _) => fallback,
                    )
                  : fallback,
            ),
          ),
        ),
        const SizedBox(width: 16),
        Expanded(
          child: Align(
            alignment: Alignment.centerLeft,
            child: TextButton.icon(
              key: const ValueKey('profile-photo'),
              onPressed: _saving || _picking ? null : _pickPhoto,
              icon: const Icon(Icons.add_a_photo_outlined, size: 20),
              label: Text(_picking ? 'Opening photos…' : 'Change photo'),
            ),
          ),
        ),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    if (!_sameIdentity) return const SizedBox.shrink();
    final theme = Theme.of(context);
    return PopScope(
      canPop: !_saving,
      child: Dialog(
        insetPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 24),
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 520),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Padding(
                padding: const EdgeInsets.fromLTRB(20, 12, 8, 8),
                child: Row(
                  children: [
                    Expanded(
                      child: Text(
                        'Edit profile',
                        style: theme.textTheme.headlineSmall,
                      ),
                    ),
                    IconButton(
                      tooltip: 'Close profile editor',
                      onPressed: _saving
                          ? null
                          : () => Navigator.of(context).pop(),
                      icon: const Icon(Icons.close),
                    ),
                  ],
                ),
              ),
              const Divider(height: 1),
              Flexible(
                child: SingleChildScrollView(
                  key: const ValueKey('profile-edit-scroll'),
                  padding: const EdgeInsets.all(20),
                  child: Form(
                    key: _formKey,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        _photoControl(),
                        const SizedBox(height: 20),
                        TextFormField(
                          key: const ValueKey('profile-username'),
                          controller: _username,
                          enabled: !_saving,
                          autocorrect: false,
                          textInputAction: TextInputAction.next,
                          decoration: const InputDecoration(
                            labelText: 'Username',
                            prefixText: '@',
                          ),
                          validator: (value) {
                            final username = value?.trim() ?? '';
                            if (username == (_initial?.username ?? '')) {
                              return null;
                            }
                            if (username.isEmpty) return 'Enter a username.';
                            if (!RegExp(r'^[a-zA-Z0-9]+$').hasMatch(username)) {
                              return 'Use letters and numbers only.';
                            }
                            return null;
                          },
                        ),
                        const SizedBox(height: 16),
                        TextFormField(
                          key: const ValueKey('profile-home-base'),
                          controller: _homeBase,
                          enabled: !_saving,
                          onChanged: _searchHomeBase,
                          textInputAction: TextInputAction.next,
                          decoration: InputDecoration(
                            labelText: 'Home base',
                            hintText: 'Search a city or region',
                            suffixIcon: _searching
                                ? const Padding(
                                    padding: EdgeInsets.all(14),
                                    child: SizedBox.square(
                                      dimension: 18,
                                      child: CircularProgressIndicator(
                                        strokeWidth: 2,
                                      ),
                                    ),
                                  )
                                : _homeBase.text != (_selectedAddress ?? '') &&
                                      (_selectedAddress?.isNotEmpty ?? false)
                                ? IconButton(
                                    tooltip: 'Keep current home base',
                                    onPressed: () {
                                      _homeBase.text = _selectedAddress!;
                                      _searchHomeBase(_homeBase.text);
                                    },
                                    icon: const Icon(Icons.undo),
                                  )
                                : const Icon(Icons.search, size: 20),
                          ),
                          validator: (value) =>
                              (value?.trim() ?? '') == (_selectedAddress ?? '')
                              ? null
                              : 'Choose a matching place from the search results.',
                        ),
                        if (_locationError != null)
                          Padding(
                            padding: const EdgeInsets.only(top: 8),
                            child: Text(
                              _locationError!,
                              style: const TextStyle(color: AppTheme.error),
                            ),
                          ),
                        if (_results.isNotEmpty)
                          Container(
                            decoration: BoxDecoration(
                              color: AppTheme.peach,
                              border: Border.all(color: AppTheme.outline),
                            ),
                            child: Column(
                              children: [
                                for (final place in _results.take(5))
                                  ListTile(
                                    key: ValueKey(
                                      'profile-place-${place.placeId}',
                                    ),
                                    title: Text(place.name),
                                    subtitle: Text(place.formattedAddress),
                                    onTap: _saving
                                        ? null
                                        : () => _selectPlace(place),
                                  ),
                              ],
                            ),
                          ),
                        const SizedBox(height: 16),
                        TextFormField(
                          key: const ValueKey('profile-bio'),
                          controller: _bio,
                          enabled: !_saving,
                          minLines: 3,
                          maxLines: 5,
                          textCapitalization: TextCapitalization.sentences,
                          decoration: const InputDecoration(
                            labelText: 'A little about you',
                            hintText:
                                'Places you love, how you like to travel…',
                            alignLabelWithHint: true,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
              const Divider(height: 1),
              if (_error != null)
                Padding(
                  padding: const EdgeInsets.fromLTRB(20, 12, 20, 0),
                  child: Semantics(
                    liveRegion: true,
                    child: Text(
                      _error!,
                      key: const ValueKey('profile-edit-error'),
                      style: const TextStyle(color: AppTheme.error),
                    ),
                  ),
                ),
              Padding(
                padding: const EdgeInsets.all(12),
                child: OverflowBar(
                  alignment: MainAxisAlignment.end,
                  spacing: 8,
                  overflowSpacing: 8,
                  children: [
                    TextButton(
                      key: const ValueKey('profile-edit-cancel'),
                      onPressed: _saving
                          ? null
                          : () => Navigator.of(context).pop(),
                      child: const Text('Cancel'),
                    ),
                    FilledButton(
                      key: const ValueKey('profile-edit-save'),
                      onPressed: _saving || _picking ? null : _save,
                      child: Text(_saving ? 'Saving…' : 'Save changes'),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
