class AppLocation {
  final double latitude;
  final double longitude;
  final double accuracy;

  const AppLocation({
    required this.latitude,
    required this.longitude,
    this.accuracy = 0,
  });

  String get headerValue => '$latitude,$longitude,$accuracy';
}
