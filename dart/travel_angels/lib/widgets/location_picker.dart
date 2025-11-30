import 'dart:async';
import 'package:flutter/material.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/location_service.dart';

class LocationPicker extends StatefulWidget {
  final Function(double latitude, double longitude, String address)? onLocationSelected;
  final double? initialLatitude;
  final double? initialLongitude;
  final String? initialAddress;

  const LocationPicker({
    super.key,
    this.onLocationSelected,
    this.initialLatitude,
    this.initialLongitude,
    this.initialAddress,
  });

  @override
  State<LocationPicker> createState() => _LocationPickerState();
}

class _LocationPickerState extends State<LocationPicker> {
  final TextEditingController _searchController = TextEditingController();
  final LocationService _locationService = LocationService(APIClient(ApiConstants.baseUrl));
  GoogleMapController? _mapController;
  Timer? _debounceTimer;
  
  LatLng? _selectedLocation;
  String? _selectedAddress;
  List<LocationCandidate> _searchResults = [];
  bool _isSearching = false;
  bool _showResults = false;
  
  @override
  void initState() {
    super.initState();
    // Initialize with provided values if available
    if (widget.initialLatitude != null && widget.initialLongitude != null) {
      _selectedLocation = LatLng(widget.initialLatitude!, widget.initialLongitude!);
      _selectedAddress = widget.initialAddress ?? '${widget.initialLatitude}, ${widget.initialLongitude}';
      _searchController.text = widget.initialAddress ?? '';
    }
  }

  CameraPosition get _initialCameraPosition {
    if (_selectedLocation != null) {
      return CameraPosition(
        target: _selectedLocation!,
        zoom: 15,
      );
    }
    return const CameraPosition(
      target: LatLng(40.7128, -74.0060), // New York City
      zoom: 10,
    );
  }

  @override
  void dispose() {
    _searchController.dispose();
    _debounceTimer?.cancel();
    _mapController?.dispose();
    super.dispose();
  }

  void _onSearchChanged(String value) {
    _debounceTimer?.cancel();
    
    if (value.isEmpty) {
      setState(() {
        _searchResults = [];
        _showResults = false;
      });
      return;
    }

    _debounceTimer = Timer(const Duration(milliseconds: 500), () {
      _performSearch(value);
    });
  }

  Future<void> _performSearch(String query) async {
    setState(() {
      _isSearching = true;
      _showResults = true;
    });

    try {
      final results = await _locationService.searchLocations(query);
      setState(() {
        _searchResults = results;
        _isSearching = false;
      });
    } catch (e) {
      setState(() {
        _isSearching = false;
      });
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error searching locations: $e')),
        );
      }
    }
  }

  void _selectLocation(LocationCandidate candidate) {
    setState(() {
      _selectedLocation = LatLng(candidate.latitude, candidate.longitude);
      _selectedAddress = candidate.formattedAddress;
      _showResults = false;
      _searchController.text = candidate.formattedAddress;
    });

    _mapController?.animateCamera(
      CameraUpdate.newLatLngZoom(_selectedLocation!, 15),
    );

    widget.onLocationSelected?.call(
      candidate.latitude,
      candidate.longitude,
      candidate.formattedAddress,
    );
  }

  void _onMapTap(LatLng location) {
    setState(() {
      _selectedLocation = location;
      // When tapping on map, we don't have an address, so we'll use coordinates
      _selectedAddress = '${location.latitude}, ${location.longitude}';
    });

    widget.onLocationSelected?.call(
      location.latitude,
      location.longitude,
      _selectedAddress!,
    );
  }

  void _onCameraMove(CameraPosition position) {
    // Update marker position when camera moves (user drags map)
    setState(() {
      _selectedLocation = position.target;
      _selectedAddress = '${position.target.latitude}, ${position.target.longitude}';
    });
  }

  void _onCameraIdle() {
    // When camera stops moving, update the callback with final position
    if (_selectedLocation != null) {
      widget.onLocationSelected?.call(
        _selectedLocation!.latitude,
        _selectedLocation!.longitude,
        _selectedAddress ?? '${_selectedLocation!.latitude}, ${_selectedLocation!.longitude}',
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      mainAxisSize: MainAxisSize.min,
      children: [
        // Search field
        TextField(
          controller: _searchController,
          decoration: InputDecoration(
            labelText: 'Search location',
            hintText: 'Enter a location',
            border: const OutlineInputBorder(),
            prefixIcon: const Icon(Icons.search),
            suffixIcon: _isSearching
                ? const Padding(
                    padding: EdgeInsets.all(12.0),
                    child: SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    ),
                  )
                : null,
          ),
          onChanged: _onSearchChanged,
          onTap: () {
            if (_searchResults.isNotEmpty) {
              setState(() {
                _showResults = true;
              });
            }
          },
        ),
        
        // Search results dropdown
        if (_showResults && _searchResults.isNotEmpty)
          Container(
            constraints: const BoxConstraints(maxHeight: 200),
            decoration: BoxDecoration(
              color: Colors.white,
              border: Border.all(color: Colors.grey.shade300),
              borderRadius: const BorderRadius.vertical(bottom: Radius.circular(4)),
            ),
            child: ListView.builder(
              shrinkWrap: true,
              itemCount: _searchResults.length,
              itemBuilder: (context, index) {
                final candidate = _searchResults[index];
                return ListTile(
                  title: Text(candidate.name),
                  subtitle: Text(
                    candidate.formattedAddress,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                  onTap: () => _selectLocation(candidate),
                );
              },
            ),
          ),
        
        const SizedBox(height: 16),
        
        // Map - wrapped in a fixed-size container to prevent scroll issues
        SizedBox(
          height: 300,
          width: double.infinity,
          child: ClipRRect(
            borderRadius: BorderRadius.circular(4),
            child: GoogleMap(
              initialCameraPosition: _initialCameraPosition,
              onMapCreated: (GoogleMapController controller) {
                _mapController = controller;
                // If we have an initial location, center the map on it
                if (_selectedLocation != null && mounted) {
                  controller.animateCamera(
                    CameraUpdate.newLatLngZoom(_selectedLocation!, 15),
                  );
                }
              },
              onTap: _onMapTap,
              onCameraMove: _onCameraMove,
              onCameraIdle: _onCameraIdle,
              markers: _selectedLocation != null
                  ? {
                      Marker(
                        markerId: const MarkerId('selected_location'),
                        position: _selectedLocation!,
                        draggable: true,
                        onDragEnd: (LatLng newPosition) {
                          if (mounted) {
                            setState(() {
                              _selectedLocation = newPosition;
                              _selectedAddress = '${newPosition.latitude}, ${newPosition.longitude}';
                            });
                            widget.onLocationSelected?.call(
                              newPosition.latitude,
                              newPosition.longitude,
                              _selectedAddress!,
                            );
                          }
                        },
                      ),
                    }
                  : {},
              myLocationButtonEnabled: false,
              zoomControlsEnabled: true,
            ),
          ),
        ),
        
        // Selected address display
        if (_selectedAddress != null) ...[
          const SizedBox(height: 8),
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: Colors.grey.shade100,
              borderRadius: BorderRadius.circular(4),
            ),
            child: Row(
              children: [
                const Icon(Icons.location_on, color: Colors.blue),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    _selectedAddress!,
                    style: const TextStyle(fontSize: 14),
                  ),
                ),
              ],
            ),
          ),
        ],
      ],
    );
  }
}

