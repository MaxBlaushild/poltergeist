class Certificate {
  final String certificatePem;
  final String fingerprint;
  final String publicKey;
  final DateTime? createdAt;

  Certificate({
    required this.certificatePem,
    required this.fingerprint,
    required this.publicKey,
    this.createdAt,
  });

  factory Certificate.fromJson(Map<String, dynamic> json) {
    return Certificate(
      certificatePem: json['certificatePem'] as String,
      fingerprint: json['fingerprint'] as String,
      publicKey: json['publicKey'] as String,
      createdAt: json['createdAt'] != null
          ? DateTime.parse(json['createdAt'] as String)
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'certificatePem': certificatePem,
      'fingerprint': fingerprint,
      'publicKey': publicKey,
      'createdAt': createdAt?.toIso8601String(),
    };
  }
}
