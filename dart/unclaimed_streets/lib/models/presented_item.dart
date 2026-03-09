/// Minimal model for NewItemModal / UsedItemModal.
class PresentedItem {
  final String name;
  final String imageUrl;
  final String? flavorText;
  final String? effectText;

  const PresentedItem({
    required this.name,
    required this.imageUrl,
    this.flavorText,
    this.effectText,
  });
}
