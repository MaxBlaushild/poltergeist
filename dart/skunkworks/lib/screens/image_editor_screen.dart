import 'dart:io';
import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:path_provider/path_provider.dart';
import 'package:pro_image_editor/pro_image_editor.dart';
import 'package:skunkworks/constants/app_colors.dart';

class ImageEditorScreen extends StatefulWidget {
  final File imageFile;

  const ImageEditorScreen({
    super.key,
    required this.imageFile,
  });

  @override
  State<ImageEditorScreen> createState() => _ImageEditorScreenState();
}

class _ImageEditorScreenState extends State<ImageEditorScreen> {
  Uint8List? _imageBytes;
  bool _hasPopped = false;

  @override
  void initState() {
    super.initState();
    _loadImage();
  }

  Future<void> _loadImage() async {
    debugPrint('[ImageEditor] _loadImage: start, path=${widget.imageFile.path}');
    final bytes = await widget.imageFile.readAsBytes();
    debugPrint('[ImageEditor] _loadImage: loaded ${bytes.length} bytes');
    setState(() {
      _imageBytes = bytes;
    });
  }

  Future<File?> _saveEditedImage(Uint8List editedBytes) async {
    debugPrint('[ImageEditor] _saveEditedImage: start, ${editedBytes.length} bytes');
    try {
      final tempDir = await getTemporaryDirectory();
      final timestamp = DateTime.now().millisecondsSinceEpoch;
      final editedFile = File('${tempDir.path}/edited_$timestamp.jpg');
      
      await editedFile.writeAsBytes(editedBytes);
      debugPrint('[ImageEditor] _saveEditedImage: wrote to ${editedFile.path}');
      
      final exists = await editedFile.exists();
      final len = exists ? await editedFile.length() : 0;
      debugPrint('[ImageEditor] _saveEditedImage: exists=$exists, length=$len');
      
      if (exists && len > 0) {
        return editedFile;
      }
      debugPrint('[ImageEditor] _saveEditedImage: returning null (empty or missing)');
      return null;
    } catch (e, st) {
      debugPrint('[ImageEditor] _saveEditedImage: error $e');
      debugPrint('[ImageEditor] _saveEditedImage: stack $st');
      return null;
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_imageBytes == null) {
      debugPrint('[ImageEditor] build: showing loading');
      return Scaffold(
        backgroundColor: AppColors.warmWhite,
        body: const Center(
          child: CircularProgressIndicator(),
        ),
      );
    }

    debugPrint('[ImageEditor] build: showing ProImageEditor');
    return ProImageEditor.memory(
      _imageBytes!,
      callbacks: ProImageEditorCallbacks(
        onImageEditingComplete: (Uint8List editedBytes) async {
          debugPrint('[ImageEditor] onImageEditingComplete: called, ${editedBytes.length} bytes');
          final editedFile = await _saveEditedImage(editedBytes);
          debugPrint('[ImageEditor] onImageEditingComplete: saved, editedFile=${editedFile?.path ?? "null"}');
          debugPrint('[ImageEditor] onImageEditingComplete: mounted=$mounted');
          if (mounted) {
            _hasPopped = true;
            debugPrint('[ImageEditor] onImageEditingComplete: calling Navigator.pop(editedFile)');
            Navigator.of(context).pop(editedFile);
            debugPrint('[ImageEditor] onImageEditingComplete: Navigator.pop returned');
          } else {
            debugPrint('[ImageEditor] onImageEditingComplete: skipped pop (not mounted)');
          }
        },
        onCloseEditor: (EditorMode mode) {
          debugPrint('[ImageEditor] onCloseEditor: called, mode=$mode, _hasPopped=$_hasPopped');
          if (_hasPopped) {
            debugPrint('[ImageEditor] onCloseEditor: skipping pop (already popped from onImageEditingComplete)');
            return;
          }
          if (mounted) {
            debugPrint('[ImageEditor] onCloseEditor: calling Navigator.pop()');
            Navigator.of(context).pop();
          }
        },
      ),
    );
  }
}
