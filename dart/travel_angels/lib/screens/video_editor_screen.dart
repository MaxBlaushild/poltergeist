import 'dart:io';
import 'dart:typed_data';

import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:path_provider/path_provider.dart';
import 'package:provider/provider.dart';
import 'package:pro_image_editor/pro_image_editor.dart';
import 'package:pro_video_editor/pro_video_editor.dart';
import 'package:record/record.dart';
import 'package:travel_angels/providers/auth_provider.dart';

/// Screen for editing a video (trim, speed, music, voiceover, export) using
/// [pro_video_editor]. Supports full editing on Android, iOS, and macOS only.
class VideoEditorScreen extends StatefulWidget {
  final File videoFile;
  final VoidCallback? onComplete;

  const VideoEditorScreen({
    super.key,
    required this.videoFile,
    this.onComplete,
  });

  @override
  State<VideoEditorScreen> createState() => _VideoEditorScreenState();
}

class _VideoEditorScreenState extends State<VideoEditorScreen> {
  final AudioRecorder _audioRecorder = AudioRecorder();

  VideoMetadata? _metadata;
  bool _metadataLoading = true;
  String? _metadataError;

  Duration _startTime = Duration.zero;
  Duration _endTime = Duration.zero;
  double _playbackSpeed = 1.0;
  bool _enableAudio = true;
  String? _customAudioPath;
  double _originalAudioVolume = 1.0;
  double _customAudioVolume = 0.5;

  bool _isRecording = false;
  bool _isExporting = false;
  double _exportProgress = 0.0;
  String? _renderTaskId;
  Uint8List? _overlayImageBytes;

  static const List<double> _speedOptions = [0.5, 1.0, 1.5, 2.0];

  @override
  void initState() {
    super.initState();
    _loadMetadata();
  }

  @override
  void dispose() {
    _audioRecorder.dispose();
    super.dispose();
  }

  Future<void> _loadMetadata() async {
    setState(() {
      _metadataLoading = true;
      _metadataError = null;
    });
    try {
      final video = EditorVideo.file(widget.videoFile);
      final meta = await ProVideoEditor.instance.getMetadata(video);
      if (!mounted) return;
      setState(() {
        _metadata = meta;
        _endTime = meta.duration;
        _metadataLoading = false;
        _metadataError = null;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _metadataLoading = false;
        _metadataError = e.toString();
      });
    }
  }

  Future<void> _pickMusic() async {
    final result = await FilePicker.platform.pickFiles(
      type: FileType.custom,
      allowedExtensions: ['mp3', 'm4a', 'aac', 'wav'],
      allowMultiple: false,
    );
    if (result == null || result.files.isEmpty) return;
    final path = result.files.single.path;
    if (path == null) return;
    setState(() => _customAudioPath = path);
  }

  Future<void> _toggleVoiceover() async {
    print('[VideoEditor] _toggleVoiceover called, _isRecording: $_isRecording');
    if (_isRecording) {
      print('[VideoEditor] _toggleVoiceover: stopping recording');
      final path = await _audioRecorder.stop();
      print('[VideoEditor] _toggleVoiceover: stopped, path: $path');
      setState(() {
        _isRecording = false;
        if (path != null) {
          _customAudioPath = path;
          print('[VideoEditor] _toggleVoiceover: set _customAudioPath to $path');
        } else {
          print('[VideoEditor] _toggleVoiceover: path is null after stop');
        }
      });
    } else {
      print('[VideoEditor] _toggleVoiceover: starting recording');
      final hasPermission = await _audioRecorder.hasPermission();
      print('[VideoEditor] _toggleVoiceover: hasPermission: $hasPermission');
      if (!hasPermission) {
        if (!mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Microphone permission required')),
        );
        return;
      }
      final dir = await getTemporaryDirectory();
      final path = '${dir.path}/voiceover_${DateTime.now().millisecondsSinceEpoch}.m4a';
      print('[VideoEditor] _toggleVoiceover: starting recording to $path');
      try {
        await _audioRecorder.start(const RecordConfig(), path: path);
        print('[VideoEditor] _toggleVoiceover: recording started successfully');
        setState(() {
          _isRecording = true;
          _customAudioPath = path;
          print('[VideoEditor] _toggleVoiceover: set _isRecording=true, _customAudioPath=$path');
        });
      } catch (e) {
        print('[VideoEditor] _toggleVoiceover: error starting recording: $e');
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Failed to start recording: $e')),
          );
        }
      }
    }
  }

  void _clearCustomAudio() {
    setState(() {
      _customAudioPath = null;
      _isRecording = false;
    });
  }

  Future<void> _addTextOrStickers() async {
    print('[VideoEditor] _addTextOrStickers called');
    if (_metadata == null) {
      print('[VideoEditor] _addTextOrStickers: metadata is null');
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Video metadata not loaded yet')),
      );
      return;
    }
    try {
      print('[VideoEditor] _addTextOrStickers: getting thumbnail at ${_startTime.inMilliseconds}ms');
      final video = EditorVideo.file(widget.videoFile);
      final thumbnails = await ProVideoEditor.instance.getThumbnails(
        ThumbnailConfigs(
          video: video,
          outputSize: const Size(1920, 1080),
          timestamps: [_startTime],
        ),
      );
      print('[VideoEditor] _addTextOrStickers: got ${thumbnails.length} thumbnails');
      if (thumbnails.isEmpty || !mounted) {
        print('[VideoEditor] _addTextOrStickers: no thumbnails or not mounted');
        return;
      }

      final thumbnailBytes = thumbnails.first;
      print('[VideoEditor] _addTextOrStickers: thumbnail size: ${thumbnailBytes.length} bytes');
      print('[VideoEditor] _addTextOrStickers: opening ProImageEditor');
      final editedBytes = await Navigator.push<Uint8List>(
        context,
        MaterialPageRoute(
          builder: (context) => ProImageEditor.memory(
            thumbnailBytes,
            callbacks: ProImageEditorCallbacks(
              onImageEditingComplete: (Uint8List bytes) async {
                print('[VideoEditor] _addTextOrStickers: onImageEditingComplete, ${bytes.length} bytes');
                Navigator.of(context).pop(bytes);
              },
              onCloseEditor: (EditorMode mode) {
                print('[VideoEditor] _addTextOrStickers: onCloseEditor, mode: $mode');
                Navigator.of(context).pop();
              },
            ),
          ),
        ),
      );

      print('[VideoEditor] _addTextOrStickers: got editedBytes: ${editedBytes != null}, size: ${editedBytes?.length ?? 0}');
      if (editedBytes != null && mounted) {
        setState(() {
          _overlayImageBytes = editedBytes;
          print('[VideoEditor] _addTextOrStickers: set _overlayImageBytes, size: ${editedBytes.length}');
        });
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Overlay added successfully')),
        );
      } else {
        print('[VideoEditor] _addTextOrStickers: editedBytes is null or not mounted');
      }
    } catch (e, stackTrace) {
      print('[VideoEditor] _addTextOrStickers: error: $e');
      print('[VideoEditor] _addTextOrStickers: stackTrace: $stackTrace');
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Failed to add text/stickers: $e')),
      );
    }
  }

  void _clearOverlay() {
    setState(() => _overlayImageBytes = null);
  }

  Future<void> _export() async {
    print('[VideoEditor] _export called');
    if (_metadata == null) {
      print('[VideoEditor] _export: metadata is null, returning');
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Video metadata not loaded yet')),
      );
      return;
    }
    
    print('[VideoEditor] _export: metadata loaded, duration: ${_metadata!.duration}');
    final user = context.read<AuthProvider>().user;
    if (user?.id == null) {
      print('[VideoEditor] _export: user not authenticated');
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('You must be logged in to export')),
      );
      return;
    }

    print('[VideoEditor] _export: starting export process');
    setState(() {
      _isExporting = true;
      _exportProgress = 0.0;
    });

    final video = EditorVideo.file(widget.videoFile);
    print('[VideoEditor] _export: created EditorVideo from ${widget.videoFile.path}');
    print('[VideoEditor] _export: trim: ${_startTime.inMilliseconds}ms to ${_endTime.inMilliseconds}ms');
    print('[VideoEditor] _export: speed: $_playbackSpeed');
    print('[VideoEditor] _export: enableAudio: $_enableAudio');
    print('[VideoEditor] _export: customAudioPath: $_customAudioPath');
    print('[VideoEditor] _export: originalAudioVolume: $_originalAudioVolume');
    print('[VideoEditor] _export: customAudioVolume: $_customAudioVolume');
    print('[VideoEditor] _export: overlayImageBytes: ${_overlayImageBytes != null ? "${_overlayImageBytes!.length} bytes" : "null"}');
    
    final renderData = VideoRenderData.withQualityPreset(
      video: video,
      qualityPreset: VideoQualityPreset.p720,
      startTime: _startTime,
      endTime: _endTime,
      playbackSpeed: _playbackSpeed,
      enableAudio: _enableAudio,
      customAudioPath: _customAudioPath,
      originalAudioVolume: _originalAudioVolume,
      customAudioVolume: _customAudioVolume,
      imageBytes: _overlayImageBytes,
    );
    _renderTaskId = renderData.id;
    print('[VideoEditor] _export: created render data, id: ${renderData.id}');

    final dir = await getTemporaryDirectory();
    final outputPath = '${dir.path}/edited_${DateTime.now().millisecondsSinceEpoch}.mp4';
    print('[VideoEditor] _export: output path: $outputPath');

    try {
      final sub = renderData.progressStream.listen((p) {
        print('[VideoEditor] _export: progress: ${p.progress}');
        if (mounted) setState(() => _exportProgress = p.progress);
      });

      print('[VideoEditor] _export: calling renderVideoToFile');
      await ProVideoEditor.instance.renderVideoToFile(outputPath, renderData);
      await sub.cancel();
      print('[VideoEditor] _export: renderVideoToFile completed');

      if (!mounted) {
        print('[VideoEditor] _export: not mounted after render, returning');
        return;
      }
      
      final outputFile = File(outputPath);
      final exists = await outputFile.exists();
      print('[VideoEditor] _export: output file exists: $exists, path: ${outputFile.path}');
      
      if (!exists) {
        throw Exception('Output file was not created');
      }
      
      // Return the edited file to the caller
      setState(() => _isExporting = false);
      widget.onComplete?.call();
      print('[VideoEditor] _export: calling Navigator.pop with output file');
      if (mounted) {
        Navigator.of(context).pop(outputFile);
      }
    } on RenderCanceledException {
      print('[VideoEditor] _export: RenderCanceledException');
      if (mounted) setState(() => _isExporting = false);
    } catch (e, stackTrace) {
      print('[VideoEditor] _export: error: $e');
      print('[VideoEditor] _export: stackTrace: $stackTrace');
      if (mounted) {
        setState(() => _isExporting = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Export failed: $e')),
        );
      }
    } finally {
      _renderTaskId = null;
    }
  }

  Future<void> _cancelExport() async {
    final id = _renderTaskId;
    if (id == null) return;
    try {
      await ProVideoEditor.instance.cancel(id);
    } catch (_) {}
  }

  String _formatDuration(Duration d) {
    final m = d.inMinutes;
    final s = d.inSeconds % 60;
    return '${m}m ${s}s';
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      backgroundColor: theme.scaffoldBackgroundColor,
      appBar: AppBar(
        backgroundColor: theme.scaffoldBackgroundColor,
        elevation: 0,
        title: Text(
          'Edit Video',
          style: TextStyle(
            color: theme.colorScheme.onSurface,
            fontWeight: FontWeight.w600,
          ),
        ),
        leading: IconButton(
          icon: Icon(Icons.close, color: theme.colorScheme.onSurface),
          onPressed: () => Navigator.of(context).pop(),
        ),
        actions: [
          if (_isExporting)
            const Padding(
              padding: EdgeInsets.all(16.0),
              child: SizedBox(
                width: 20,
                height: 20,
                child: CircularProgressIndicator(strokeWidth: 2),
              ),
            )
          else
            TextButton(
              onPressed: () {
                print('[VideoEditor] Export button pressed, metadata: ${_metadata != null}, isExporting: $_isExporting');
                if (_metadata != null) {
                  _export();
                } else {
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('Please wait for video to load')),
                  );
                }
              },
              child: Text(
                'Export',
                style: TextStyle(
                  color: _metadata != null
                      ? theme.colorScheme.primary
                      : theme.disabledColor,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ),
        ],
      ),
      body: Stack(
        children: [
          _metadataLoading
              ? const Center(child: CircularProgressIndicator())
              : _metadataError != null
                  ? Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          const Icon(Icons.error_outline, size: 48, color: Colors.red),
                          const SizedBox(height: 16),
                          Text(
                            'Failed to load video',
                            style: Theme.of(context).textTheme.titleMedium,
                          ),
                          const SizedBox(height: 8),
                          Text(
                            _metadataError!,
                            style: Theme.of(context).textTheme.bodySmall,
                            textAlign: TextAlign.center,
                          ),
                        ],
                      ),
                    )
                  : SingleChildScrollView(
                      child: Padding(
                        padding: const EdgeInsets.all(16.0),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.stretch,
                          children: [
                            // Trim controls
                            Card(
                              child: Padding(
                                padding: const EdgeInsets.all(16.0),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    const Text(
                                      'Trim',
                                      style: TextStyle(
                                        fontSize: 18,
                                        fontWeight: FontWeight.bold,
                                      ),
                                    ),
                                    const SizedBox(height: 8),
                                    RangeSlider(
                                      values: RangeValues(
                                        _startTime.inMilliseconds.toDouble(),
                                        _endTime.inMilliseconds.toDouble(),
                                      ),
                                      min: 0,
                                      max: _metadata!.duration.inMilliseconds.toDouble(),
                                      onChanged: (values) {
                                        setState(() {
                                          _startTime = Duration(milliseconds: values.start.toInt());
                                          _endTime = Duration(milliseconds: values.end.toInt());
                                        });
                                      },
                                    ),
                                    Row(
                                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                      children: [
                                        Text('Start: ${_formatDuration(_startTime)}'),
                                        Text('End: ${_formatDuration(_endTime)}'),
                                      ],
                                    ),
                                  ],
                                ),
                              ),
                            ),
                            const SizedBox(height: 16),
                            // Speed controls
                            Card(
                              child: Padding(
                                padding: const EdgeInsets.all(16.0),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    const Text(
                                      'Playback Speed',
                                      style: TextStyle(
                                        fontSize: 18,
                                        fontWeight: FontWeight.bold,
                                      ),
                                    ),
                                    const SizedBox(height: 8),
                                    SegmentedButton<double>(
                                      segments: _speedOptions
                                          .map((speed) => ButtonSegment(
                                                value: speed,
                                                label: Text('${speed}x'),
                                              ))
                                          .toList(),
                                      selected: {_playbackSpeed},
                                      onSelectionChanged: (Set<double> newSelection) {
                                        setState(() => _playbackSpeed = newSelection.first);
                                      },
                                    ),
                                  ],
                                ),
                              ),
                            ),
                            const SizedBox(height: 16),
                            // Audio controls
                            Card(
                              child: Padding(
                                padding: const EdgeInsets.all(16.0),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Row(
                                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                      children: [
                                        const Text(
                                          'Audio',
                                          style: TextStyle(
                                            fontSize: 18,
                                            fontWeight: FontWeight.bold,
                                          ),
                                        ),
                                        Switch(
                                          value: _enableAudio,
                                          onChanged: (value) => setState(() => _enableAudio = value),
                                        ),
                                      ],
                                    ),
                                    if (_enableAudio) ...[
                                      const SizedBox(height: 16),
                                      Text('Original Volume: ${(_originalAudioVolume * 100).toInt()}%'),
                                      Slider(
                                        value: _originalAudioVolume,
                                        onChanged: (value) => setState(() => _originalAudioVolume = value),
                                      ),
                                      if (_customAudioPath != null) ...[
                                        const SizedBox(height: 8),
                                        Text('Custom Audio Volume: ${(_customAudioVolume * 100).toInt()}%'),
                                        Slider(
                                          value: _customAudioVolume,
                                          onChanged: (value) => setState(() => _customAudioVolume = value),
                                        ),
                                      ],
                                    ],
                                  ],
                                ),
                              ),
                            ),
                            const SizedBox(height: 16),
                            // Music and voiceover
                            Card(
                              child: Padding(
                                padding: const EdgeInsets.all(16.0),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.stretch,
                                  children: [
                                    const Text(
                                      'Audio Tracks',
                                      style: TextStyle(
                                        fontSize: 18,
                                        fontWeight: FontWeight.bold,
                                      ),
                                    ),
                                    const SizedBox(height: 16),
                                    Row(
                                      children: [
                                        Expanded(
                                          child: OutlinedButton.icon(
                                            icon: const Icon(Icons.music_note),
                                            label: const Text('Add Music'),
                                            onPressed: _pickMusic,
                                          ),
                                        ),
                                        const SizedBox(width: 8),
                                    Expanded(
                                      child: OutlinedButton.icon(
                                        icon: Icon(_isRecording ? Icons.stop : Icons.mic),
                                        label: Text(_isRecording ? 'Stop' : 'Voiceover'),
                                        onPressed: _toggleVoiceover,
                                        style: OutlinedButton.styleFrom(
                                          backgroundColor: _customAudioPath != null && !_isRecording
                                              ? theme.colorScheme.primary.withValues(alpha: 0.1)
                                              : null,
                                        ),
                                      ),
                                    ),
                                      ],
                                    ),
                                if (_customAudioPath != null) ...[
                                  const SizedBox(height: 8),
                                  Container(
                                    padding: const EdgeInsets.all(8),
                                    decoration: BoxDecoration(
                                      color: theme.colorScheme.primary.withValues(alpha: 0.1),
                                      borderRadius: BorderRadius.circular(8),
                                    ),
                                    child: Row(
                                      children: [
                                        Icon(Icons.check_circle, color: theme.colorScheme.primary, size: 20),
                                        const SizedBox(width: 8),
                                        Expanded(
                                          child: Text(
                                            _isRecording 
                                                ? 'Recording...' 
                                                : 'Custom audio: ${_customAudioPath!.split('/').last}',
                                            style: const TextStyle(fontSize: 12),
                                          ),
                                        ),
                                        TextButton(
                                          onPressed: _clearCustomAudio,
                                          child: const Text('Clear'),
                                        ),
                                      ],
                                    ),
                                  ),
                                ],
                                  ],
                                ),
                              ),
                            ),
                            const SizedBox(height: 16),
                            // Text and stickers
                            Card(
                              child: Padding(
                                padding: const EdgeInsets.all(16.0),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.stretch,
                                  children: [
                                    const Text(
                                      'Overlays',
                                      style: TextStyle(
                                        fontSize: 18,
                                        fontWeight: FontWeight.bold,
                                      ),
                                    ),
                                    const SizedBox(height: 16),
                                    Row(
                                      children: [
                                    Expanded(
                                      child: OutlinedButton.icon(
                                        icon: const Icon(Icons.text_fields),
                                        label: const Text('Text/Stickers'),
                                        onPressed: _addTextOrStickers,
                                        style: OutlinedButton.styleFrom(
                                          backgroundColor: _overlayImageBytes != null
                                              ? theme.colorScheme.primary.withValues(alpha: 0.1)
                                              : null,
                                        ),
                                      ),
                                    ),
                                    if (_overlayImageBytes != null) ...[
                                      const SizedBox(width: 8),
                                      Expanded(
                                        child: OutlinedButton.icon(
                                          icon: const Icon(Icons.clear),
                                          label: const Text('Clear'),
                                          onPressed: _clearOverlay,
                                        ),
                                      ),
                                    ],
                                  ],
                                ),
                                if (_overlayImageBytes != null) ...[
                                  const SizedBox(height: 8),
                                  Container(
                                    padding: const EdgeInsets.all(8),
                                    decoration: BoxDecoration(
                                      color: theme.colorScheme.primary.withValues(alpha: 0.1),
                                      borderRadius: BorderRadius.circular(8),
                                    ),
                                    child: Row(
                                      children: [
                                        Icon(Icons.check_circle, color: theme.colorScheme.primary, size: 20),
                                        const SizedBox(width: 8),
                                        const Expanded(
                                          child: Text(
                                            'Overlay added',
                                            style: TextStyle(fontSize: 12),
                                          ),
                                        ),
                                        TextButton(
                                          onPressed: _clearOverlay,
                                          child: const Text('Clear'),
                                        ),
                                      ],
                                    ),
                                  ),
                                ],
                              ],
                            ),
                          ),
                        ),
                          ],
                        ),
                      ),
                    ),
          // Export progress overlay
          if (_isExporting)
            ModalBarrier(
              color: Colors.black.withValues(alpha: 0.5),
              dismissible: false,
            ),
          if (_isExporting)
            Center(
              child: Card(
                child: Padding(
                  padding: const EdgeInsets.all(24.0),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const CircularProgressIndicator(),
                      const SizedBox(height: 16),
                      Text('Exporting... ${(_exportProgress * 100).toInt()}%'),
                      const SizedBox(height: 8),
                      TextButton(
                        onPressed: _cancelExport,
                        child: const Text('Cancel'),
                      ),
                    ],
                  ),
                ),
              ),
            ),
        ],
      ),
    );
  }
}
