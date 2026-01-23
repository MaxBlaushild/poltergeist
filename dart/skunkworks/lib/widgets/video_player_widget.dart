import 'package:flutter/material.dart';
import 'package:video_player/video_player.dart';

class VideoPlayerWidget extends StatefulWidget {
  final String videoUrl;
  final bool autoPlay;
  final bool showControls;
  final BoxFit fit;

  const VideoPlayerWidget({
    super.key,
    required this.videoUrl,
    this.autoPlay = false,
    this.showControls = true,
    this.fit = BoxFit.cover,
  });

  @override
  State<VideoPlayerWidget> createState() => _VideoPlayerWidgetState();
}

class _VideoPlayerWidgetState extends State<VideoPlayerWidget> {
  VideoPlayerController? _controller;
  bool _isInitialized = false;
  bool _hasError = false;

  @override
  void initState() {
    super.initState();
    _initializePlayer();
  }

  Future<void> _initializePlayer() async {
    try {
      _controller = VideoPlayerController.networkUrl(Uri.parse(widget.videoUrl));
      await _controller!.initialize();
      
      if (widget.autoPlay) {
        _controller!.play();
      }
      
      if (mounted) {
        setState(() {
          _isInitialized = true;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _hasError = true;
        });
      }
    }
  }

  @override
  void dispose() {
    _controller?.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    if (_hasError) {
      return Container(
        color: Colors.grey.shade200,
        child: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                Icons.error_outline,
                size: 48,
                color: Colors.grey.shade600,
              ),
              const SizedBox(height: 8),
              Text(
                'Failed to load video',
                style: TextStyle(
                  color: Colors.grey.shade600,
                  fontSize: 14,
                ),
              ),
            ],
          ),
        ),
      );
    }

    if (!_isInitialized || _controller == null) {
      return Container(
        color: Colors.grey.shade200,
        child: const Center(
          child: CircularProgressIndicator(),
        ),
      );
    }

    return Stack(
      fit: StackFit.expand,
      children: [
        Center(
          child: AspectRatio(
            aspectRatio: _controller!.value.aspectRatio,
            child: VideoPlayer(_controller!),
          ),
        ),
        if (widget.showControls)
          _VideoControlsOverlay(
            controller: _controller!,
            autoPlay: widget.autoPlay,
          ),
      ],
    );
  }
}

class _VideoControlsOverlay extends StatefulWidget {
  final VideoPlayerController controller;
  final bool autoPlay;

  const _VideoControlsOverlay({
    required this.controller,
    required this.autoPlay,
  });

  @override
  State<_VideoControlsOverlay> createState() => _VideoControlsOverlayState();
}

class _VideoControlsOverlayState extends State<_VideoControlsOverlay> {
  bool _isPlaying = false;
  bool _showControls = true;

  @override
  void initState() {
    super.initState();
    _isPlaying = widget.autoPlay;
    widget.controller.addListener(_videoListener);
    
    // Hide controls after 3 seconds if playing
    if (_isPlaying) {
      Future.delayed(const Duration(seconds: 3), () {
        if (mounted) {
          setState(() {
            _showControls = false;
          });
        }
      });
    }
  }

  @override
  void dispose() {
    widget.controller.removeListener(_videoListener);
    super.dispose();
  }

  void _videoListener() {
    if (mounted) {
      setState(() {
        _isPlaying = widget.controller.value.isPlaying;
      });
    }
  }

  void _togglePlayPause() {
    setState(() {
      _showControls = true;
    });
    
    if (widget.controller.value.isPlaying) {
      widget.controller.pause();
    } else {
      widget.controller.play();
    }
    
    // Hide controls after 3 seconds
    Future.delayed(const Duration(seconds: 3), () {
      if (mounted && widget.controller.value.isPlaying) {
        setState(() {
          _showControls = false;
        });
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: () {
        setState(() {
          _showControls = !_showControls;
        });
      },
      child: AnimatedOpacity(
        opacity: _showControls ? 1.0 : 0.0,
        duration: const Duration(milliseconds: 300),
        child: Container(
          color: Colors.black.withOpacity(0.3),
          child: Center(
            child: IconButton(
              icon: Icon(
                _isPlaying ? Icons.pause_circle_filled : Icons.play_circle_filled,
                size: 64,
                color: Colors.white,
              ),
              onPressed: _togglePlayPause,
            ),
          ),
        ),
      ),
    );
  }
}
