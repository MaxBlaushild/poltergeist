import 'dart:io';
import 'dart:typed_data';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:provider/provider.dart';
import 'package:skunkworks/constants/api_constants.dart';
import 'package:skunkworks/constants/app_colors.dart';
import 'package:skunkworks/providers/auth_provider.dart';
import 'package:skunkworks/providers/post_provider.dart';
import 'package:skunkworks/models/draft.dart';
import 'package:skunkworks/screens/drafts_screen.dart';
import 'package:skunkworks/screens/image_editor_screen.dart';
// import 'package:skunkworks/screens/video_editor_screen.dart'; // Video editing disabled
import 'package:skunkworks/services/api_client.dart';
import 'package:skunkworks/services/draft_service.dart';
import 'package:skunkworks/services/c2pa_service.dart';
import 'package:skunkworks/services/certificate_service.dart';
import 'package:skunkworks/services/media_service.dart';
import 'package:skunkworks/utils/video_platform_utils.dart';
import 'package:skunkworks/widgets/bottom_nav.dart';
// import 'package:skunkworks/widgets/video_preview_dialog.dart'; // Video editing disabled
import 'package:chewie/chewie.dart';
import 'package:video_player/video_player.dart';

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
  final _tagInputController = TextEditingController();
  List<String> _selectedTags = [];
  List<String> _albumTagSuggestions = [];
  List<String> _recentTagSuggestions = [];
  bool _loadingSuggestions = false;
  File? _selectedMedia;
  File? _editedImage;
  Draft? _editingDraft;
  bool _uploading = false;
  bool _isVideoMode = false;
  final ImagePicker _picker = ImagePicker();
  final DraftService _draftService = DraftService();

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _loadTagSuggestions());
  }

  Future<void> _loadTagSuggestions() async {
    if (!mounted) return;
    setState(() => _loadingSuggestions = true);
    try {
      final resp = await context.read<PostProvider>().getPostTagSuggestions();
      if (!mounted) return;
      final albumRaw = resp['albumTags'];
      final recentRaw = resp['recentTags'];
      final album = albumRaw is List
          ? (albumRaw).map((e) => (e is Map ? e['tag']?.toString() : e?.toString()) ?? '').where((t) => t.isNotEmpty).toList()
          : <String>[];
      final recent = recentRaw is List
          ? (recentRaw).map((e) => (e is Map ? e['tag']?.toString() : e?.toString()) ?? '').where((t) => t.isNotEmpty).toList()
          : <String>[];
      setState(() {
        _albumTagSuggestions = album;
        _recentTagSuggestions = recent;
      });
    } catch (_) {
      if (mounted) setState(() {});
    } finally {
      if (mounted) setState(() => _loadingSuggestions = false);
    }
  }

  void _addTagFromSuggestion(String tag) {
    if (tag.isEmpty) return;
    final t = tag.length > 64 ? tag.substring(0, 64) : tag;
    if (!_selectedTags.contains(t)) {
      setState(() => _selectedTags = [..._selectedTags, t]);
    }
  }

  @override
  void dispose() {
    _captionController.dispose();
    _tagInputController.dispose();
    super.dispose();
  }

  void _addTag() {
    final t = _tagInputController.text.trim();
    if (t.isEmpty) return;
    final tag = t.length > 64 ? t.substring(0, 64) : t;
    if (!_selectedTags.contains(tag)) {
      setState(() {
        _selectedTags = [..._selectedTags, tag];
        _tagInputController.clear();
      });
    }
  }

  void _removeTag(String tag) {
    setState(() => _selectedTags = _selectedTags.where((t) => t != tag).toList());
  }

  Future<void> _pickMedia(ImageSource source) async {
    try {
      if (_isVideoMode) {
        final XFile? video = await _picker.pickVideo(source: source);
        if (video != null) {
          setState(() {
            _selectedMedia = File(video.path);
            _editedImage = null;
            _editingDraft = null;
          });
        }
      } else {
        final XFile? image = await _picker.pickImage(source: source);
        if (image != null) {
          setState(() {
            _selectedMedia = File(image.path);
            _editedImage = null;
            _editingDraft = null;
          });
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to capture ${_isVideoMode ? 'video' : 'photo'}: $e')),
        );
      }
    }
  }

  Future<void> _previewVideo(File videoFile) async {
    showDialog(
      context: context,
      builder: (context) => _VideoFilePreviewDialog(videoFile: videoFile),
    );
  }

  Future<void> _navigateToEditor() async {
    if (_selectedMedia == null) return;

    // Video editing is disabled/hidden
    if (_isVideoMode) {
      return;
    }

    // Handle image editing
    debugPrint('[UploadPost] _navigateToEditor: pushing ImageEditorScreen');
    try {
      final editedFile = await Navigator.push<File>(
        context,
        MaterialPageRoute(
          builder: (context) => ImageEditorScreen(
            imageFile: _selectedMedia!,
          ),
        ),
      );
      debugPrint('[UploadPost] _navigateToEditor: Navigator.push returned');
      debugPrint('[UploadPost] _navigateToEditor: editedFile=${editedFile?.path ?? "null"}');
      debugPrint('[UploadPost] _navigateToEditor: mounted=$mounted');

      if (editedFile != null && mounted) {
        final exists = await editedFile.exists();
        debugPrint('[UploadPost] _navigateToEditor: file exists=$exists');
        if (exists) {
          setState(() {
            _editedImage = editedFile;
          });
          debugPrint('[UploadPost] _navigateToEditor: set _editedImage');
        }
      } else {
        debugPrint('[UploadPost] _navigateToEditor: no file or not mounted, skipping setState');
      }
    } catch (e, st) {
      debugPrint('[UploadPost] _navigateToEditor: error $e');
      debugPrint('[UploadPost] _navigateToEditor: stack $st');
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Error editing image: $e'),
            backgroundColor: AppColors.coralPop,
          ),
        );
      }
    }
    debugPrint('[UploadPost] _navigateToEditor: done');
  }

  Future<void> _saveDraft() async {
    if (_selectedMedia == null || _isVideoMode) return;
    final image = _editedImage ?? _selectedMedia!;
    final caption = _captionController.text.trim();
    final c = caption.isEmpty ? null : caption;

    try {
      if (_editingDraft != null) {
        await _draftService.updateDraft(_editingDraft!.id, image, c);
      } else {
        await _draftService.saveDraft(image, c);
      }
      _captionController.clear();
      setState(() {
        _selectedMedia = null;
        _editedImage = null;
        _editingDraft = null;
      });
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Saved to draft')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to save draft: $e'),
            backgroundColor: AppColors.coralPop,
          ),
        );
      }
    }
  }

  Future<void> _navigateToDrafts() async {
    final draft = await Navigator.push<Draft?>(
      context,
      MaterialPageRoute(
        builder: (context) => const DraftsScreen(),
      ),
    );
    if (draft == null || !mounted) return;
    final file = File(draft.imagePath);
    if (!await file.exists()) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Draft image no longer available'),
            backgroundColor: AppColors.coralPop,
          ),
        );
      }
      return;
    }
    _captionController.text = draft.caption ?? '';
    setState(() {
      _selectedMedia = file;
      _editedImage = null;
      _editingDraft = draft;
      _isVideoMode = false;
    });
  }

  Future<void> _uploadPost() async {
    if (_selectedMedia == null) return;

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

      // Use edited image if available, otherwise use original
      final imageToUpload = _editedImage ?? _selectedMedia!;

      // Upload media to S3 first
      final mediaUrl = await mediaService.uploadPostImage(imageToUpload, user!.id!);
      if (mediaUrl == null) {
        throw Exception('Failed to upload ${_isVideoMode ? 'video' : 'image'}');
      }

      // Determine media type
      final mediaType = _isVideoMode ? 'video' : 'image';

      // Create C2PA manifest (skip for videos)
      String? manifestUrl;
      String? manifestHash;
      String? certFingerprint;
      String? assetId;

      if (!_isVideoMode) {
        try {
          final manifestData = await c2paService.createManifest(
            mediaUrl,
            assetId: mediaUrl, // Use media URL as asset ID
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
            assetId = mediaUrl;
          }
        } catch (e) {
          // If manifest creation fails, log but don't fail the post
          // User might not have a certificate enrolled yet
          print('Warning: Failed to create C2PA manifest: $e');
        }
      }

      // Create post with manifest data
      await postProvider.createPost(
        mediaUrl,
        caption: _captionController.text.trim().isNotEmpty
            ? _captionController.text.trim()
            : null,
        manifestUrl: manifestUrl,
        manifestHash: manifestHash,
        certFingerprint: certFingerprint,
        assetId: assetId,
        mediaType: mediaType,
        tags: _selectedTags.isNotEmpty ? _selectedTags : null,
      );

      final editingDraft = _editingDraft;
      if (editingDraft != null) {
        await _draftService.deleteDraft(editingDraft.id);
      }

      _captionController.clear();
      _tagInputController.clear();
      setState(() {
        _selectedMedia = null;
        _editedImage = null;
        _editingDraft = null;
        _selectedTags = [];
        _isVideoMode = false;
      });

      if (mounted) {
        widget.onNavigate(NavTab.home);
      }
    } catch (e) {
      // Log detailed error information
      print('Upload post error: $e');
      if (e is DioException) {
        print('DioException - Status: ${e.response?.statusCode}');
        print('DioException - Response body: ${e.response?.data}');
        print('DioException - Request URL: ${e.requestOptions.uri}');
        print('DioException - Request data: ${e.requestOptions.data}');
      }
      
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
      backgroundColor: AppColors.warmWhite,
      appBar: AppBar(
        backgroundColor: AppColors.warmWhite,
        elevation: 0,
        title: const Text(
          'New Post',
          style: TextStyle(
            color: AppColors.graphiteInk,
            fontWeight: FontWeight.w600,
            fontSize: 18,
          ),
        ),
        actions: [
          if (!_uploading)
            IconButton(
              icon: Icon(Icons.drafts_outlined, color: AppColors.graphiteInk),
              onPressed: _navigateToDrafts,
              tooltip: 'Drafts',
            ),
          if (_selectedMedia != null && !_isVideoMode && !_uploading)
            TextButton(
              onPressed: _saveDraft,
              child: Text(
                'Save draft',
                style: TextStyle(
                  color: AppColors.graphiteInk,
                  fontWeight: FontWeight.w500,
                  fontSize: 16,
                ),
              ),
            ),
          if (_selectedMedia != null && !_uploading)
            TextButton(
              onPressed: _uploadPost,
              child: Text(
                'Share',
                style: TextStyle(
                  color: AppColors.softRealBlue,
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
                  // Media picker/preview
                  Container(
                    height: 400,
                    color: Colors.grey.shade100,
                    child: _selectedMedia != null
                        ? _isVideoMode
                            ? Stack(
                                fit: StackFit.expand,
                                children: [
                                  GestureDetector(
                                    onTap: () => _previewVideo(_selectedMedia!),
                                    child: Image.file(
                                      _selectedMedia!,
                                      fit: BoxFit.cover,
                                      errorBuilder: (context, error, stackTrace) {
                                        return Center(
                                          child: Column(
                                            mainAxisAlignment: MainAxisAlignment.center,
                                            children: [
                                              Icon(
                                                Icons.videocam,
                                                size: 64,
                                                color: Colors.grey.shade400,
                                              ),
                                              const SizedBox(height: 16),
                                              Text(
                                                'Video selected',
                                                style: TextStyle(
                                                  color: Colors.grey.shade600,
                                                  fontSize: 16,
                                                ),
                                              ),
                                            ],
                                          ),
                                        );
                                      },
                                    ),
                                  ),
                                  Center(
                                    child: GestureDetector(
                                      onTap: () => _previewVideo(_selectedMedia!),
                                      child: Icon(
                                        Icons.play_circle_filled,
                                        size: 64,
                                        color: Colors.white.withValues(alpha: 0.8),
                                      ),
                                    ),
                                  ),
                                  // Edit button overlay - HIDDEN for videos
                                  // Positioned(
                                  //   top: 12,
                                  //   right: 12,
                                  //   child: Material(
                                  //     color: AppColors.softRealBlue,
                                  //     borderRadius: BorderRadius.circular(20),
                                  //     child: InkWell(
                                  //       onTap: _navigateToEditor,
                                  //       borderRadius: BorderRadius.circular(20),
                                  //       child: Container(
                                  //         padding: const EdgeInsets.symmetric(
                                  //           horizontal: 16,
                                  //           vertical: 8,
                                  //         ),
                                  //         child: Row(
                                  //           mainAxisSize: MainAxisSize.min,
                                  //           children: [
                                  //             const Icon(
                                  //               Icons.edit,
                                  //               color: Colors.white,
                                  //               size: 18,
                                  //             ),
                                  //             const SizedBox(width: 6),
                                  //             Text(
                                  //               'Edit',
                                  //               style: TextStyle(
                                  //                 color: Colors.white,
                                  //                 fontWeight: FontWeight.w600,
                                  //                 fontSize: 14,
                                  //               ),
                                  //             ),
                                  //           ],
                                  //         ),
                                  //       ),
                                  //     ),
                                  //   ),
                                  // ),
                                ],
                              )
                            : Stack(
                                fit: StackFit.expand,
                                children: [
                                  Image.file(
                                    _editedImage ?? _selectedMedia!,
                                    fit: BoxFit.cover,
                                  ),
                                  // Edit button overlay
                                  if (!_isVideoMode || supportsFullVideoEditing)
                                    Positioned(
                                      top: 12,
                                      right: 12,
                                      child: Material(
                                        color: AppColors.softRealBlue,
                                        borderRadius: BorderRadius.circular(20),
                                        child: InkWell(
                                          onTap: _navigateToEditor,
                                          borderRadius: BorderRadius.circular(20),
                                          child: Container(
                                            padding: const EdgeInsets.symmetric(
                                              horizontal: 16,
                                              vertical: 8,
                                            ),
                                            child: Row(
                                              mainAxisSize: MainAxisSize.min,
                                              children: [
                                                const Icon(
                                                  Icons.edit,
                                                  color: Colors.white,
                                                  size: 18,
                                                ),
                                                const SizedBox(width: 6),
                                                Text(
                                                  'Edit',
                                                  style: TextStyle(
                                                    color: Colors.white,
                                                    fontWeight: FontWeight.w600,
                                                    fontSize: 14,
                                                  ),
                                                ),
                                              ],
                                            ),
                                          ),
                                        ),
                                      ),
                                    ),
                                ],
                              )
                        : Center(
                            child: Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Icon(
                                  _isVideoMode ? Icons.videocam_outlined : Icons.image_outlined,
                                  size: 64,
                                  color: Colors.grey.shade400,
                                ),
                                const SizedBox(height: 16),
                                Text(
                                  _isVideoMode ? 'Record a video' : 'Take a photo',
                                  style: TextStyle(
                                    color: Colors.grey.shade600,
                                    fontSize: 16,
                                  ),
                                ),
                                const SizedBox(height: 24),
                                Row(
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  children: [
                                    SegmentedButton<bool>(
                                      segments: const [
                                        ButtonSegment(
                                          value: false,
                                          label: Text('Photo'),
                                          icon: Icon(Icons.camera_alt),
                                        ),
                                        ButtonSegment(
                                          value: true,
                                          label: Text('Video'),
                                          icon: Icon(Icons.videocam),
                                        ),
                                      ],
                                      selected: {_isVideoMode},
                                      onSelectionChanged: (Set<bool> newSelection) {
                                        setState(() {
                                          _isVideoMode = newSelection.first;
                                          _selectedMedia = null;
                                          _editedImage = null;
                                          _editingDraft = null;
                                        });
                                      },
                                    ),
                                  ],
                                ),
                                const SizedBox(height: 16),
                                ElevatedButton.icon(
                                  onPressed: () => _pickMedia(ImageSource.camera),
                                  icon: Icon(_isVideoMode ? Icons.videocam : Icons.camera_alt),
                                  label: Text(_isVideoMode ? 'Record video' : 'Take photo'),
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
                  // Tags
                  Padding(
                    padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Tags',
                          style: TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w600,
                            color: Colors.grey.shade700,
                          ),
                        ),
                        const SizedBox(height: 8),
                        if (_loadingSuggestions)
                          const Padding(
                            padding: EdgeInsets.only(bottom: 8),
                            child: SizedBox(height: 24, width: 24, child: CircularProgressIndicator(strokeWidth: 2)),
                          )
                        else ...[
                          if (_albumTagSuggestions.isNotEmpty) ...[
                            Text(
                              'From your albums',
                              style: TextStyle(fontSize: 12, color: Colors.grey.shade600),
                            ),
                            const SizedBox(height: 4),
                            Wrap(
                              spacing: 6,
                              runSpacing: 6,
                              children: _albumTagSuggestions
                                  .where((t) => !_selectedTags.contains(t))
                                  .map((tag) => ActionChip(
                                        label: Text(tag, style: const TextStyle(fontSize: 13)),
                                        onPressed: () => _addTagFromSuggestion(tag),
                                        materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                                        backgroundColor: Colors.grey.shade200,
                                      ))
                                  .toList(),
                            ),
                            const SizedBox(height: 12),
                          ],
                          if (_recentTagSuggestions.isNotEmpty) ...[
                            Text(
                              'Recently used',
                              style: TextStyle(fontSize: 12, color: Colors.grey.shade600),
                            ),
                            const SizedBox(height: 4),
                            Wrap(
                              spacing: 6,
                              runSpacing: 6,
                              children: _recentTagSuggestions
                                  .where((t) => !_selectedTags.contains(t))
                                  .map((tag) => ActionChip(
                                        label: Text(tag, style: const TextStyle(fontSize: 13)),
                                        onPressed: () => _addTagFromSuggestion(tag),
                                        materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                                        backgroundColor: Colors.grey.shade200,
                                      ))
                                  .toList(),
                            ),
                            const SizedBox(height: 12),
                          ],
                        ],
                        Row(
                          children: [
                            Expanded(
                              child: TextField(
                                controller: _tagInputController,
                                decoration: const InputDecoration(
                                  hintText: 'Add a tag',
                                  border: OutlineInputBorder(),
                                  isDense: true,
                                  contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                                ),
                                onSubmitted: (_) => _addTag(),
                              ),
                            ),
                            const SizedBox(width: 8),
                            IconButton.filled(
                              onPressed: _addTag,
                              icon: const Icon(Icons.add, size: 20),
                              tooltip: 'Add tag',
                            ),
                          ],
                        ),
                        if (_selectedTags.isNotEmpty) ...[
                          const SizedBox(height: 8),
                          Wrap(
                            spacing: 6,
                            runSpacing: 6,
                            children: _selectedTags.map((tag) => Chip(
                              label: Text(tag, style: const TextStyle(fontSize: 13)),
                              deleteIcon: const Icon(Icons.close, size: 18),
                              onDeleted: () => _removeTag(tag),
                              materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                            )).toList(),
                          ),
                        ],
                      ],
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

/// Dialog for previewing a video from a File
class _VideoFilePreviewDialog extends StatefulWidget {
  final File videoFile;

  const _VideoFilePreviewDialog({required this.videoFile});

  @override
  State<_VideoFilePreviewDialog> createState() => _VideoFilePreviewDialogState();
}

class _VideoFilePreviewDialogState extends State<_VideoFilePreviewDialog> {
  VideoPlayerController? _videoPlayerController;
  ChewieController? _chewieController;
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _initializePlayer();
  }

  Future<void> _initializePlayer() async {
    try {
      _videoPlayerController = VideoPlayerController.file(widget.videoFile);
      await _videoPlayerController!.initialize();
      
      if (!mounted) return;
      
      _chewieController = ChewieController(
        videoPlayerController: _videoPlayerController!,
        autoPlay: true,
        looping: false,
        aspectRatio: _videoPlayerController!.value.aspectRatio,
        errorBuilder: (context, errorMessage) {
          return Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(Icons.error_outline, size: 48, color: Colors.red),
                const SizedBox(height: 16),
                Text(
                  'Error loading video',
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                const SizedBox(height: 8),
                Text(
                  errorMessage,
                  style: Theme.of(context).textTheme.bodySmall,
                  textAlign: TextAlign.center,
                ),
              ],
            ),
          );
        },
      );
      
      setState(() {
        _isLoading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _isLoading = false;
        _error = e.toString();
      });
    }
  }

  @override
  void dispose() {
    _chewieController?.dispose();
    _videoPlayerController?.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      child: Container(
        width: MediaQuery.of(context).size.width * 0.9,
        height: MediaQuery.of(context).size.height * 0.7,
        padding: const EdgeInsets.all(16),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text(
                  'Video Preview',
                  style: TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.close),
                  onPressed: () => Navigator.of(context).pop(),
                ),
              ],
            ),
            const SizedBox(height: 16),
            Expanded(
              child: _isLoading
                  ? const Center(child: CircularProgressIndicator())
                  : _error != null
                      ? Center(
                          child: Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              const Icon(Icons.error_outline,
                                  size: 48, color: Colors.red),
                              const SizedBox(height: 16),
                              Text(
                                'Error loading video',
                                style: Theme.of(context).textTheme.titleMedium,
                              ),
                              const SizedBox(height: 8),
                              Text(
                                _error!,
                                style: Theme.of(context).textTheme.bodySmall,
                                textAlign: TextAlign.center,
                              ),
                            ],
                          ),
                        )
                      : _chewieController != null
                          ? Chewie(controller: _chewieController!)
                          : const Center(child: CircularProgressIndicator()),
            ),
          ],
        ),
      ),
    );
  }
}

