class MovementPattern {
  final double startingLatitude;
  final double startingLongitude;

  const MovementPattern({
    required this.startingLatitude,
    required this.startingLongitude,
  });

  factory MovementPattern.fromJson(Map<String, dynamic> json) {
    return MovementPattern(
      startingLatitude: (json['startingLatitude'] as num).toDouble(),
      startingLongitude: (json['startingLongitude'] as num).toDouble(),
    );
  }
}
