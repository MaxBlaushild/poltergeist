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
    if (query.trim().isEmpty) {
      setState(() {
        _searchResults = [];
        _showResults = false;
        _isSearching = false;
      });
      return;
    }

    setState(() {
      _isSearching = true;
      _showResults = true;
    });

    print('LocationPicker: Performing search for "$query"');

    try {
      final results = await _locationService.searchLocations(query.trim());
      print('LocationPicker: Search returned ${results.length} results');
      
      setState(() {
        _searchResults = results;
        _isSearching = false;
        // Keep showResults true if we have results, false if empty
        _showResults = results.isNotEmpty;
      });
      
      if (results.isEmpty && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('No locations found. Try a different search term.'),
            duration: Duration(seconds: 2),
          ),
        );
      }
    } catch (e, stackTrace) {
      print('LocationPicker: Search error: $e');
      print('LocationPicker: Stack trace: $stackTrace');
      setState(() {
        _searchResults = [];
        _isSearching = false;
        _showResults = false;
      });
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Error searching locations: ${e.toString()}'),
            duration: const Duration(seconds: 5),
            action: SnackBarAction(
              label: 'Details',
              onPressed: () {
                showDialog(
                  context: context,
                  builder: (context) => AlertDialog(
                    title: const Text('Search Error'),
                    content: SingleChildScrollView(
                      child: Text('Error: $e\n\nStack trace: $stackTrace'),
                    ),
                    actions: [
                      TextButton(
                        onPressed: () => Navigator.of(context).pop(),
                        child: const Text('Close'),
                      ),
                    ],
                  ),
                );
              },
            ),
          ),
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
                : _searchController.text.isNotEmpty && _searchResults.isEmpty && !_isSearching
                    ? IconButton(
                        icon: const Icon(Icons.clear),
                        onPressed: () {
                          _searchController.clear();
                          setState(() {
                            _searchResults = [];
                            _showResults = false;
                          });
                        },
                      )
                    : null,
          ),
          onChanged: (value) {
            print('LocationPicker: TextField onChanged: "$value"');
            _onSearchChanged(value);
          },
          onTap: () {
            print('LocationPicker: TextField tapped, showing results: ${_searchResults.length}');
            if (_searchResults.isNotEmpty) {
              setState(() {
                _showResults = true;
              });
            }
          },
          onSubmitted: (value) {
            print('LocationPicker: TextField submitted: "$value"');
            if (value.trim().isNotEmpty) {
              _performSearch(value.trim());
            }
          },
        ),
        
        // Search results dropdown - using Stack to ensure it appears above other content
        if (_showResults && _searchResults.isNotEmpty)
          Container(
            constraints: const BoxConstraints(maxHeight: 200),
            margin: const EdgeInsets.only(top: 4),
            decoration: BoxDecoration(
              color: Theme.of(context).scaffoldBackgroundColor,
              border: Border.all(color: Colors.grey.shade300),
              borderRadius: BorderRadius.circular(4),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withOpacity(0.1),
                  blurRadius: 4,
                  offset: const Offset(0, 2),
                ),
              ],
            ),
            child: ListView.builder(
              shrinkWrap: true,
              physics: const ClampingScrollPhysics(),
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
        // GoogleMap needs explicit constraints and should not be in a scrollable context
        SizedBox(
          height: 300,
          width: double.infinity,
          child: ClipRRect(
            borderRadius: BorderRadius.circular(4),
            child: GoogleMap(
              initialCameraPosition: _initialCameraPosition,
              onMapCreated: (GoogleMapController controller) {
                _mapController = controller;
                print('GoogleMap created successfully');
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

