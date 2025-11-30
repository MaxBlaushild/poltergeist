import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/providers/auth_provider.dart';
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
      await authProvider.updateProfile(
        dateOfBirth: _dateOfBirth,
        gender: _selectedGender,
        latitude: _selectedLatitude,
        longitude: _selectedLongitude,
        locationAddress: _selectedLocationAddress,
        bio: _bioController.text.trim().isNotEmpty ? _bioController.text.trim() : null,
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
                // Wrap LocationPicker to prevent scroll conflicts with GoogleMap
                SizedBox(
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

