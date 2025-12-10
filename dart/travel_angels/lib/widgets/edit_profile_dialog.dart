import 'dart:io';
import 'package:flutter/material.dart';
import 'package:file_picker/file_picker.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/media_service.dart';
import 'package:travel_angels/widgets/location_picker.dart';

class EditProfileDialog extends StatefulWidget {
  final User? user;
  final VoidCallback? onSave;

  const EditProfileDialog({
    super.key,
    this.user,
    this.onSave,
  });

  @override
  State<EditProfileDialog> createState() => _EditProfileDialogState();
}

class _EditProfileDialogState extends State<EditProfileDialog> {
  DateTime? _dateOfBirth;
  String? _selectedGender;
  double? _selectedLatitude;
  double? _selectedLongitude;
  String? _selectedLocationAddress;
  final _bioController = TextEditingController();
  bool _isSaving = false;
  File? _profilePicture;

  @override
  void initState() {
    super.initState();
    // Initialize with current user values
    _dateOfBirth = widget.user?.dateOfBirth;
    _selectedGender = widget.user?.gender;
    _selectedLatitude = widget.user?.latitude;
    _selectedLongitude = widget.user?.longitude;
    _selectedLocationAddress = widget.user?.locationAddress;
    _bioController.text = widget.user?.bio ?? '';
  }

  @override
  void dispose() {
    _bioController.dispose();
    super.dispose();
  }

  Future<void> _pickProfilePicture() async {
    try {
      final result = await FilePicker.platform.pickFiles(
        type: FileType.image,
        allowMultiple: false,
      );

      if (result != null && result.files.single.path != null) {
        setState(() {
          _profilePicture = File(result.files.single.path!);
        });
      }
    } catch (e) {
      // Handle error
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to pick image: $e')),
        );
      }
    }
  }

  Future<void> _handleSave() async {
    if (_dateOfBirth == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please select your date of birth')),
      );
      return;
    }

    if (_selectedGender == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please select your gender')),
      );
      return;
    }

    if (_selectedLatitude == null || _selectedLongitude == null || _selectedLocationAddress == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please select your location')),
      );
      return;
    }

    setState(() {
      _isSaving = true;
    });

    try {
      final authProvider = context.read<AuthProvider>();
      String? profilePictureUrl;

      // Upload profile picture if selected
      if (_profilePicture != null) {
        try {
          final apiClient = APIClient(ApiConstants.baseUrl);
          final mediaService = MediaService(apiClient);
          final user = authProvider.user;
          
          if (user != null && user.id != null) {
            profilePictureUrl = await mediaService.uploadProfilePicture(
              _profilePicture!,
              user.id!,
            );
            
            if (profilePictureUrl == null) {
              if (mounted) {
                setState(() {
                  _isSaving = false;
                });
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('Failed to upload profile picture')),
                );
              }
              return;
            }
          }
        } catch (e) {
          if (mounted) {
            setState(() {
              _isSaving = false;
            });
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(content: Text('Failed to upload profile picture: $e')),
            );
          }
          return;
        }
      }

      await authProvider.updateProfile(
        dateOfBirth: _dateOfBirth,
        gender: _selectedGender,
        latitude: _selectedLatitude,
        longitude: _selectedLongitude,
        locationAddress: _selectedLocationAddress,
        bio: _bioController.text.trim().isNotEmpty ? _bioController.text.trim() : null,
        profilePictureUrl: profilePictureUrl,
      );

      if (mounted) {
        Navigator.of(context).pop();
        widget.onSave?.call();
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Profile updated successfully'),
            backgroundColor: Colors.green,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _isSaving = false;
        });
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to update profile: $e')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    
    return Container(
      decoration: BoxDecoration(
        color: theme.scaffoldBackgroundColor,
        borderRadius: const BorderRadius.vertical(top: Radius.circular(20)),
      ),
      child: DraggableScrollableSheet(
        initialChildSize: 0.9,
        minChildSize: 0.5,
        maxChildSize: 0.95,
        builder: (context, scrollController) {
          return ListView(
            controller: scrollController,
            padding: const EdgeInsets.all(24.0),
            children: [
                // Handle bar
                Center(
                  child: Container(
                    width: 40,
                    height: 4,
                    margin: const EdgeInsets.only(bottom: 16),
                    decoration: BoxDecoration(
                      color: Colors.grey.shade300,
                      borderRadius: BorderRadius.circular(2),
                    ),
                  ),
                ),
                // Title
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      'Edit Profile',
                      style: theme.textTheme.headlineSmall?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    IconButton(
                      icon: const Icon(Icons.close),
                      onPressed: () => Navigator.of(context).pop(),
                    ),
                  ],
                ),
                const SizedBox(height: 24),
                // Date of birth picker
                const Text(
                  'Date of Birth *',
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w500,
                  ),
                ),
                const SizedBox(height: 8),
                InkWell(
                  onTap: () async {
                    final DateTime? picked = await showDatePicker(
                      context: context,
                      initialDate: _dateOfBirth ?? DateTime.now().subtract(const Duration(days: 365 * 18)),
                      firstDate: DateTime(1900),
                      lastDate: DateTime.now(),
                    );
                    if (picked != null) {
                      setState(() {
                        _dateOfBirth = picked;
                      });
                    }
                  },
                  child: Container(
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      border: Border.all(color: Colors.grey),
                      borderRadius: BorderRadius.circular(4),
                    ),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          _dateOfBirth != null
                              ? '${_dateOfBirth!.year}-${_dateOfBirth!.month.toString().padLeft(2, '0')}-${_dateOfBirth!.day.toString().padLeft(2, '0')}'
                              : 'Select date of birth',
                          style: TextStyle(
                            color: _dateOfBirth != null ? Colors.black : Colors.grey,
                          ),
                        ),
                        const Icon(Icons.calendar_today),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 24),
                // Gender dropdown
                const Text(
                  'Gender *',
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w500,
                  ),
                ),
                const SizedBox(height: 8),
                DropdownButtonFormField<String>(
                  value: _selectedGender,
                  decoration: const InputDecoration(
                    border: OutlineInputBorder(),
                    hintText: 'Select gender',
                  ),
                  items: const [
                    DropdownMenuItem(value: 'Male', child: Text('Male')),
                    DropdownMenuItem(value: 'Female', child: Text('Female')),
                    DropdownMenuItem(value: 'Non-binary', child: Text('Non-binary')),
                    DropdownMenuItem(value: 'Prefer not to say', child: Text('Prefer not to say')),
                  ],
                  onChanged: (value) {
                    setState(() {
                      _selectedGender = value;
                    });
                  },
                ),
                const SizedBox(height: 24),
                // Location picker
                const Text(
                  'Location *',
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w500,
                  ),
                ),
                const SizedBox(height: 8),
                // LocationPicker - wrapped to prevent scroll conflicts with GoogleMap
                // The map needs explicit width constraints from MediaQuery to render properly
                SizedBox(
                  width: MediaQuery.of(context).size.width - 48, // Account for padding
                  child: LocationPicker(
                    initialLatitude: _selectedLatitude,
                    initialLongitude: _selectedLongitude,
                    initialAddress: _selectedLocationAddress,
                    onLocationSelected: (latitude, longitude, address) {
                      setState(() {
                        _selectedLatitude = latitude;
                        _selectedLongitude = longitude;
                        _selectedLocationAddress = address;
                      });
                    },
                  ),
                ),
                const SizedBox(height: 24),
                // Profile picture picker
                const Text(
                  'Profile Picture (Optional)',
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w500,
                  ),
                ),
                const SizedBox(height: 8),
                Row(
                  children: [
                    if (_profilePicture != null) ...[
                      ClipRRect(
                        borderRadius: BorderRadius.circular(8),
                        child: Image.file(
                          _profilePicture!,
                          width: 100,
                          height: 100,
                          fit: BoxFit.cover,
                        ),
                      ),
                      const SizedBox(width: 16),
                    ] else if (widget.user?.profilePictureUrl != null && widget.user!.profilePictureUrl!.isNotEmpty) ...[
                      ClipRRect(
                        borderRadius: BorderRadius.circular(8),
                        child: Image.network(
                          widget.user!.profilePictureUrl!,
                          width: 100,
                          height: 100,
                          fit: BoxFit.cover,
                          errorBuilder: (context, error, stackTrace) {
                            return Container(
                              width: 100,
                              height: 100,
                              color: Colors.grey[300],
                              child: const Icon(Icons.person, size: 50),
                            );
                          },
                        ),
                      ),
                      const SizedBox(width: 16),
                    ],
                    Expanded(
                      child: OutlinedButton.icon(
                        onPressed: _pickProfilePicture,
                        icon: const Icon(Icons.image),
                        label: Text(_profilePicture == null ? 'Pick Image' : 'Change Image'),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 24),
                // Bio field
                TextField(
                  controller: _bioController,
                  decoration: const InputDecoration(
                    labelText: 'Bio (Optional)',
                    hintText: 'Tell us about yourself',
                    border: OutlineInputBorder(),
                  ),
                  maxLines: 3,
                ),
                const SizedBox(height: 32),
                // Save button
                ElevatedButton(
                  onPressed: _isSaving ? null : _handleSave,
                  style: ElevatedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(vertical: 16),
                  ),
                  child: _isSaving
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Text('Save Changes'),
                ),
                const SizedBox(height: 16),
            ],
          );
        },
      ),
    );
  }
}

