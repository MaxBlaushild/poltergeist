import 'dart:io';
import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:provider/provider.dart';
import 'package:skunkworks/constants/api_constants.dart';
import 'package:skunkworks/providers/auth_provider.dart';
import 'package:skunkworks/providers/post_provider.dart';
import 'package:skunkworks/services/api_client.dart';
import 'package:skunkworks/services/c2pa_service.dart';
import 'package:skunkworks/services/certificate_service.dart';
import 'package:skunkworks/services/media_service.dart';
import 'package:skunkworks/widgets/bottom_nav.dart';

class UploadPostScreen extends StatefulWidget {
  final Function(NavTab) onNavigate;

  const UploadPostScreen({
    super.key,
    required this.onNavigate,
  });

  @override
  State<UploadPostScreen> createState() => _UploadPostScreenState();
}

class _UploadPostScreenState extends State<UploadPostScreen> {
  final _captionController = TextEditingController();
  File? _selectedImage;
  bool _uploading = false;
  final ImagePicker _picker = ImagePicker();

  @override
  void dispose() {
    _captionController.dispose();
    super.dispose();
  }

  Future<void> _pickImage(ImageSource source) async {
    try {
      final XFile? image = await _picker.pickImage(source: source);
      if (image != null) {
        setState(() {
          _selectedImage = File(image.path);
        });
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to pick image: $e')),
        );
      }
    }
  }

  Future<void> _uploadPost() async {
    if (_selectedImage == null) return;

    final authProvider = context.read<AuthProvider>();
    final postProvider = context.read<PostProvider>();
    final user = authProvider.user;

    if (user?.id == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('User not authenticated')),
      );
      return;
    }

    setState(() {
      _uploading = true;
    });

    try {
      final apiClient = APIClient(ApiConstants.baseUrl);
      final mediaService = MediaService(apiClient);
      final certificateService = CertificateService(apiClient);
      final c2paService = C2PAService(certificateService);

      // Upload image to S3 first
      final imageUrl = await mediaService.uploadPostImage(_selectedImage!, user!.id!);
      if (imageUrl == null) {
        throw Exception('Failed to upload image');
      }

      // Create C2PA manifest
      String? manifestUrl;
      String? manifestHash;
      String? certFingerprint;
      String? assetId;

      try {
        final manifestData = await c2paService.createManifest(
          imageUrl,
          assetId: imageUrl, // Use image URL as asset ID
        );

        // Upload manifest to S3
        final manifestBytes = manifestData['manifestBytes'] as Uint8List;
        manifestUrl = await mediaService.uploadManifest(
          manifestBytes,
          user.id!,
        );

        if (manifestUrl != null) {
          // Convert bytes to hex strings for API
          manifestHash = _bytesToHex(manifestData['manifestHash'] as Uint8List);
          certFingerprint = _bytesToHex(manifestData['certFingerprint'] as Uint8List);
          assetId = imageUrl;
        }
      } catch (e) {
        // If manifest creation fails, log but don't fail the post
        // User might not have a certificate enrolled yet
        print('Warning: Failed to create C2PA manifest: $e');
      }

      // Create post with manifest data
      await postProvider.createPost(
        imageUrl,
        caption: _captionController.text.trim().isNotEmpty
            ? _captionController.text.trim()
            : null,
        manifestUrl: manifestUrl,
        manifestHash: manifestHash,
        certFingerprint: certFingerprint,
        assetId: assetId,
      );

      // Clear form and navigate to feed
      _captionController.clear();
      setState(() {
        _selectedImage = null;
      });

      if (mounted) {
        widget.onNavigate(NavTab.home);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to upload post: $e')),
        );
      }
    } finally {
      if (mounted) {
        setState(() {
          _uploading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.white,
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        title: const Text(
          'New Post',
          style: TextStyle(
            color: Colors.black,
            fontWeight: FontWeight.w600,
            fontSize: 18,
          ),
        ),
        actions: [
          if (_selectedImage != null && !_uploading)
            TextButton(
              onPressed: _uploadPost,
              child: const Text(
                'Share',
                style: TextStyle(
                  color: Colors.blue,
                  fontWeight: FontWeight.w600,
                  fontSize: 16,
                ),
              ),
            ),
          if (_uploading)
            const Padding(
              padding: EdgeInsets.all(16.0),
              child: SizedBox(
                width: 20,
                height: 20,
                child: CircularProgressIndicator(strokeWidth: 2),
              ),
            ),
        ],
      ),
      body: Column(
        children: [
          Expanded(
            child: SingleChildScrollView(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  // Image picker/preview
                  Container(
                    height: 400,
                    color: Colors.grey.shade100,
                    child: _selectedImage != null
                        ? Image.file(
                            _selectedImage!,
                            fit: BoxFit.cover,
                          )
                        : Center(
                            child: Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Icon(
                                  Icons.image_outlined,
                                  size: 64,
                                  color: Colors.grey.shade400,
                                ),
                                const SizedBox(height: 16),
                                Text(
                                  'Select an image',
                                  style: TextStyle(
                                    color: Colors.grey.shade600,
                                    fontSize: 16,
                                  ),
                                ),
                                const SizedBox(height: 24),
                                Row(
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  children: [
                                    ElevatedButton.icon(
                                      onPressed: () => _pickImage(ImageSource.gallery),
                                      icon: const Icon(Icons.photo_library),
                                      label: const Text('Gallery'),
                                    ),
                                    const SizedBox(width: 16),
                                    ElevatedButton.icon(
                                      onPressed: () => _pickImage(ImageSource.camera),
                                      icon: const Icon(Icons.camera_alt),
                                      label: const Text('Camera'),
                                    ),
                                  ],
                                ),
                              ],
                            ),
                          ),
                  ),
                  // Caption input
                  Padding(
                    padding: const EdgeInsets.all(16.0),
                    child: TextField(
                      controller: _captionController,
                      decoration: const InputDecoration(
                        hintText: 'Write a caption...',
                        border: InputBorder.none,
                      ),
                      maxLines: 5,
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
      bottomNavigationBar: BottomNav(
        currentTab: NavTab.upload,
        onTabChanged: widget.onNavigate,
      ),
    );
  }

  /// Converts bytes to hex string
  String _bytesToHex(Uint8List bytes) {
    return bytes.map((byte) => byte.toRadixString(16).padLeft(2, '0')).join('');
  }
}

