class OutfitGeneration {
  final String id;
  final String userId;
  final String ownedInventoryItemId;
  final int inventoryItemId;
  final String outfitName;
  final String selfieUrl;
  final String status;
  final String? error;
  final String? profilePictureUrl;

  const OutfitGeneration({
    required this.id,
    required this.userId,
    required this.ownedInventoryItemId,
    required this.inventoryItemId,
    required this.outfitName,
    required this.selfieUrl,
    required this.status,
    this.error,
    this.profilePictureUrl,
  });

  factory OutfitGeneration.fromJson(Map<String, dynamic> json) {
    return OutfitGeneration(
      id: json['id'] as String? ?? '',
      userId: json['userId'] as String? ?? '',
      ownedInventoryItemId: json['ownedInventoryItemId'] as String? ?? '',
      inventoryItemId: (json['inventoryItemId'] as num?)?.toInt() ?? 0,
      outfitName: json['outfitName'] as String? ?? '',
      selfieUrl: json['selfieUrl'] as String? ?? '',
      status: json['status'] as String? ?? '',
      error: json['error'] as String?,
      profilePictureUrl: json['profilePictureUrl'] as String?,
    );
  }

  bool get isPending => status == 'queued' || status == 'in_progress';
  bool get isComplete => status == 'complete';
  bool get isFailed => status == 'failed';
}
