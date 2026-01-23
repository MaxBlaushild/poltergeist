class Certificate {
  final String certificatePem;
  final String fingerprint;
  final String publicKey;
  final DateTime? createdAt;
  final bool? active;
  final String? transactionHash;
  final int? chainId;

  Certificate({
    required this.certificatePem,
    required this.fingerprint,
    required this.publicKey,
    this.createdAt,
    this.active,
    this.transactionHash,
    this.chainId,
  });

  factory Certificate.fromJson(Map<String, dynamic> json) {
    return Certificate(
      certificatePem: json['certificatePem'] as String,
      fingerprint: json['fingerprint'] as String,
      publicKey: json['publicKey'] as String,
      createdAt: json['createdAt'] != null
          ? DateTime.parse(json['createdAt'] as String)
          : null,
      active: json['active'] as bool?,
      transactionHash: json['transactionHash'] as String?,
      chainId: json['chainId'] as int?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'certificatePem': certificatePem,
      'fingerprint': fingerprint,
      'publicKey': publicKey,
      'createdAt': createdAt?.toIso8601String(),
      'active': active,
      'transactionHash': transactionHash,
      'chainId': chainId,
    };
  }
}
