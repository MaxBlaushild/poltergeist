import 'package:flutter/material.dart';
import 'package:travel_angels/models/trending_destination.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/trending_service.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:dio/dio.dart';

class TrendingDestinationsWidget extends StatefulWidget {
  const TrendingDestinationsWidget({super.key});

  @override
  State<TrendingDestinationsWidget> createState() => _TrendingDestinationsWidgetState();
}

class _TrendingDestinationsWidgetState extends State<TrendingDestinationsWidget> {
  final TrendingService _trendingService = TrendingService(
    APIClient(ApiConstants.baseUrl),
  );

  List<TrendingDestination> _cities = [];
  List<TrendingDestination> _countries = [];
  bool _isLoading = true;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _loadTrendingDestinations();
  }

  Future<void> _loadTrendingDestinations() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final result = await _trendingService.getTrendingDestinations();
      setState(() {
        _cities = result['cities'] ?? [];
        _countries = result['countries'] ?? [];
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _isLoading = false;
        if (e is DioException) {
          if (e.response != null && e.response?.data != null) {
            final errorData = e.response?.data as Map<String, dynamic>?;
            _errorMessage = errorData?['error']?.toString() ?? 'Failed to load trending destinations';
          } else {
            _errorMessage = 'Failed to load trending destinations: ${e.message ?? e.toString()}';
          }
        } else {
          _errorMessage = 'Failed to load trending destinations: ${e.toString()}';
        }
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    if (_isLoading) {
      return Card(
        margin: const EdgeInsets.only(bottom: 16.0),
        child: Padding(
          padding: const EdgeInsets.all(16.0),
          child: Center(
            child: CircularProgressIndicator(
              color: theme.colorScheme.primary,
            ),
          ),
        ),
      );
    }

    if (_errorMessage != null) {
      return Card(
        margin: const EdgeInsets.only(bottom: 16.0),
        child: Padding(
          padding: const EdgeInsets.all(16.0),
          child: Column(
            children: [
              Icon(
                Icons.error_outline,
                color: theme.colorScheme.error,
              ),
              const SizedBox(height: 8),
              Text(
                _errorMessage!,
                style: theme.textTheme.bodyMedium?.copyWith(
                  color: theme.colorScheme.error,
                ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 8),
              TextButton(
                onPressed: _loadTrendingDestinations,
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
      );
    }

    // Don't show widget if there's no data
    if (_cities.isEmpty && _countries.isEmpty) {
      return const SizedBox.shrink();
    }

    return Card(
      margin: const EdgeInsets.only(bottom: 16.0),
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Title
            Row(
              children: [
                Icon(
                  Icons.trending_up,
                  color: theme.colorScheme.primary,
                ),
                const SizedBox(width: 8),
                Text(
                  'Top 5 Trending Destinations',
                  style: theme.textTheme.titleLarge?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            // Cities section
            if (_cities.isNotEmpty) ...[
              Text(
                'Cities',
                style: theme.textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.w600,
                ),
              ),
              const SizedBox(height: 8),
              Wrap(
                spacing: 8.0,
                runSpacing: 8.0,
                children: _cities.map((city) => _buildDestinationChip(
                  context,
                  city,
                  theme,
                )).toList(),
              ),
              const SizedBox(height: 16),
            ],
            // Countries section
            if (_countries.isNotEmpty) ...[
              Text(
                'Countries',
                style: theme.textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.w600,
                ),
              ),
              const SizedBox(height: 8),
              Wrap(
                spacing: 8.0,
                runSpacing: 8.0,
                children: _countries.map((country) => _buildDestinationChip(
                  context,
                  country,
                  theme,
                )).toList(),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildDestinationChip(
    BuildContext context,
    TrendingDestination destination,
    ThemeData theme,
  ) {
    return Chip(
      avatar: CircleAvatar(
        backgroundColor: theme.colorScheme.primaryContainer,
        child: Text(
          '${destination.rank}',
          style: TextStyle(
            fontSize: 12,
            fontWeight: FontWeight.bold,
            color: theme.colorScheme.onPrimaryContainer,
          ),
        ),
      ),
      label: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            destination.name,
            style: theme.textTheme.bodyMedium?.copyWith(
              fontWeight: FontWeight.w500,
            ),
          ),
          Text(
            '${destination.documentCount} ${destination.documentCount == 1 ? 'doc' : 'docs'}',
            style: theme.textTheme.bodySmall?.copyWith(
              color: theme.colorScheme.onSurface.withOpacity(0.7),
            ),
          ),
        ],
      ),
      padding: const EdgeInsets.symmetric(horizontal: 8.0, vertical: 4.0),
      materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
      visualDensity: VisualDensity.compact,
    );
  }
}
