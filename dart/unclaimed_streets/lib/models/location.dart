class AppLocation {
  final double latitude;
  final double longitude;
  final double accuracy;
  final double? heading;
  final double? speed;

  const AppLocation({
    required this.latitude,
    required this.longitude,
    this.accuracy = 0,
    this.heading,
    this.speed,
  });

  String get headerValue => '$latitude,$longitude,$accuracy';
}
