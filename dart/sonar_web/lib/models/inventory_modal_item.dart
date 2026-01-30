/// Item shown in "You got a X!" modal.
class PresentedItem {
  final String name;
  final String imageUrl;
  final String flavorText;
  final String effectText;

  const PresentedItem({
    required this.name,
    required this.imageUrl,
    this.flavorText = '',
    this.effectText = '',
  });
}

/// Item shown in "You used a X!" modal.
class UsedItem {
  final String name;
  final String imageUrl;

  const UsedItem({required this.name, required this.imageUrl});
}
