import 'dart:async';
import 'package:flutter/material.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/document_location.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/location_service.dart';

/// Widget for selecting multiple locations (cities, countries, continents)
class LocationSelector extends StatefulWidget {
  final List<DocumentLocation>? initialLocations;
  final Function(List<DocumentLocation>) onLocationsChanged;

  const LocationSelector({
    super.key,
    this.initialLocations,
    required this.onLocationsChanged,
  });

  @override
  State<LocationSelector> createState() => _LocationSelectorState();
}

class _LocationSelectorState extends State<LocationSelector> {
  final TextEditingController _searchController = TextEditingController();
  final LocationService _locationService = LocationService(
    APIClient(ApiConstants.baseUrl),
  );
  Timer? _debounceTimer;

  List<DocumentLocation> _selectedLocations = [];
  List<LocationCandidate> _searchResults = [];
  bool _isSearching = false;
  bool _showResults = false;

  @override
  void initState() {
    super.initState();
    _selectedLocations = List<DocumentLocation>.from(
      widget.initialLocations ?? [],
    );
    _searchController.addListener(_onSearchChanged);
  }

  @override
  void dispose() {
    _searchController.removeListener(_onSearchChanged);
    _searchController.dispose();
    _debounceTimer?.cancel();
    super.dispose();
  }

  void _onSearchChanged() {
    _debounceTimer?.cancel();

    final query = _searchController.text.trim();
    if (query.isEmpty) {
      setState(() {
        _searchResults = [];
        _showResults = false;
      });
      return;
    }

    _debounceTimer = Timer(const Duration(milliseconds: 500), () {
      _performSearch(query);
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
          SnackBar(
            content: Text('Error searching locations: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  void _selectLocation(LocationCandidate candidate) {
    // Determine location type from the candidate
    // For now, we'll try to infer from the name/address
    // In a more sophisticated implementation, we could check address components
    LocationType locationType = _inferLocationType(candidate);

    final location = DocumentLocation(
      id: '', // Will be set by backend
      documentId: '', // Will be set by backend
      placeId: candidate.placeId,
      name: candidate.name,
      formattedAddress: candidate.formattedAddress,
      latitude: candidate.latitude,
      longitude: candidate.longitude,
      locationType: locationType,
    );

    // Check if already selected
    if (_selectedLocations.any((loc) => loc.placeId == candidate.placeId)) {
      return;
    }

    setState(() {
      _selectedLocations.add(location);
      _searchController.clear();
      _searchResults = [];
      _showResults = false;
    });

    widget.onLocationsChanged(_selectedLocations);
  }

  LocationType _inferLocationType(LocationCandidate candidate) {
    // Simple heuristic: check if name contains country-like patterns
    // This is a basic implementation - could be improved with address component parsing
    final name = candidate.name.toLowerCase();
    final address = candidate.formattedAddress.toLowerCase();

    // Check for continent names (very basic)
    final continents = ['africa', 'asia', 'europe', 'north america', 'south america', 'oceania', 'antarctica'];
    if (continents.any((continent) => name.contains(continent) || address.contains(continent))) {
      return LocationType.continent;
    }

    // Check for country indicators
    // This is a simplified check - in production, you'd parse address components
    if (address.contains('country') || name.length < 20) {
      // Likely a country if it's a short name or explicitly mentioned
      return LocationType.country;
    }

    // Default to city
    return LocationType.city;
  }

  void _removeLocation(DocumentLocation location) {
    setState(() {
      _selectedLocations.removeWhere((loc) => loc.placeId == location.placeId);
    });
    widget.onLocationsChanged(_selectedLocations);
  }

  String _getLocationTypeLabel(LocationType type) {
    switch (type) {
      case LocationType.city:
        return 'City';
      case LocationType.country:
        return 'Country';
      case LocationType.continent:
        return 'Continent';
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      mainAxisSize: MainAxisSize.min,
      children: [
        // Search field
        TextField(
          controller: _searchController,
          decoration: InputDecoration(
            labelText: 'Search locations',
            hintText: 'Enter a city, country, or continent',
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
              color: theme.scaffoldBackgroundColor,
              border: Border.all(color: theme.colorScheme.outline),
              borderRadius: BorderRadius.circular(4),
            ),
            child: ListView.builder(
              shrinkWrap: true,
              itemCount: _searchResults.length,
              itemBuilder: (context, index) {
                final candidate = _searchResults[index];
                final isSelected = _selectedLocations.any(
                  (loc) => loc.placeId == candidate.placeId,
                );

                return ListTile(
                  title: Text(candidate.name),
                  subtitle: Text(
                    candidate.formattedAddress,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                  enabled: !isSelected,
                  trailing: isSelected
                      ? Icon(
                          Icons.check_circle,
                          color: theme.colorScheme.primary,
                        )
                      : const Icon(Icons.add_circle_outline),
                  onTap: isSelected ? null : () => _selectLocation(candidate),
                );
              },
            ),
          ),

        // Selected locations
        if (_selectedLocations.isNotEmpty) ...[
          const SizedBox(height: 16),
          Text(
            'Selected Locations',
            style: theme.textTheme.titleSmall?.copyWith(
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: _selectedLocations.map((location) {
              return Chip(
                label: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(
                      Icons.location_on,
                      size: 16,
                      color: theme.colorScheme.onPrimary,
                    ),
                    const SizedBox(width: 4),
                    Flexible(
                      child: Text(
                        location.name,
                        overflow: TextOverflow.ellipsis,
                        style: TextStyle(
                          color: theme.colorScheme.onPrimary,
                          fontSize: 12,
                        ),
                      ),
                    ),
                    const SizedBox(width: 4),
                    Text(
                      '(${_getLocationTypeLabel(location.locationType)})',
                      style: TextStyle(
                        color: theme.colorScheme.onPrimary.withOpacity(0.8),
                        fontSize: 11,
                      ),
                    ),
                  ],
                ),
                deleteIcon: Icon(
                  Icons.close,
                  size: 18,
                  color: theme.colorScheme.onPrimary,
                ),
                onDeleted: () => _removeLocation(location),
                backgroundColor: theme.colorScheme.primary,
              );
            }).toList(),
          ),
        ],
      ],
    );
  }
}

