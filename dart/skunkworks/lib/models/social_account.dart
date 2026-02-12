class SocialAccount {
  final String provider;
  final String? accountId;
  final String? username;
  final DateTime? expiresAt;

  SocialAccount({
    required this.provider,
    this.accountId,
    this.username,
    this.expiresAt,
  });

  factory SocialAccount.fromJson(Map<String, dynamic> json) {
    return SocialAccount(
      provider: json['provider'] as String,
      accountId: json['accountId'] as String?,
      username: json['username'] as String?,
      expiresAt: json['expiresAt'] != null
          ? DateTime.tryParse(json['expiresAt'] as String)
          : null,
    );
  }
}
