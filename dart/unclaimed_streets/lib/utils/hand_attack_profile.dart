import 'dart:math' as math;

import '../models/equipment_item.dart';
import '../models/inventory_item.dart';

class HandAttackContribution {
  const HandAttackContribution({
    required this.name,
    required this.damageMin,
    required this.damageMax,
    required this.swipesPerAttack,
    required this.isOffHandWeapon,
  });

  final String name;
  final int damageMin;
  final int damageMax;
  final int swipesPerAttack;
  final bool isOffHandWeapon;

  String get sourceLabel => isOffHandWeapon ? '$name (off-hand)' : name;
}

class HandAttackProfile {
  const HandAttackProfile({required this.contributions});

  final List<HandAttackContribution> contributions;

  bool get hasWeapon => contributions.isNotEmpty;

  int get damageMin =>
      contributions.fold<int>(0, (sum, item) => sum + item.damageMin);

  int get damageMax =>
      contributions.fold<int>(0, (sum, item) => sum + item.damageMax);

  int get swipesPerAttack =>
      contributions.fold<int>(0, (sum, item) => sum + item.swipesPerAttack);

  String get source => hasWeapon
      ? contributions.map((item) => item.sourceLabel).join(' + ')
      : 'Unarmed';
}

HandAttackProfile buildHandAttackProfile(Iterable<EquippedItem> equippedItems) {
  final contributions = <HandAttackContribution>[];
  for (final equipped in equippedItems) {
    final item = equipped.inventoryItem;
    if (item == null) continue;
    final damageMin = item.damageMin;
    final damageMax = item.damageMax;
    if (damageMin == null ||
        damageMax == null ||
        damageMin <= 0 ||
        damageMax <= 0) {
      continue;
    }

    final slot = equipped.slot.trim().toLowerCase();
    final isOffHandWeapon = slot == 'off_hand' && _isOneHandedWeapon(item);
    if (slot != 'dominant_hand' && !isOffHandWeapon) {
      continue;
    }

    final effectiveMin = isOffHandWeapon ? _halfDamage(damageMin) : damageMin;
    final effectiveMax = isOffHandWeapon ? _halfDamage(damageMax) : damageMax;
    final swipes = math.max(1, item.swipesPerAttack ?? 1);
    contributions.add(
      HandAttackContribution(
        name: item.name,
        damageMin: effectiveMin,
        damageMax: math.max(effectiveMin, effectiveMax),
        swipesPerAttack: swipes,
        isOffHandWeapon: isOffHandWeapon,
      ),
    );
  }
  return HandAttackProfile(contributions: contributions);
}

bool _isOneHandedWeapon(InventoryItem item) {
  final category = item.handItemCategory?.trim().toLowerCase();
  final handedness = item.handedness?.trim().toLowerCase();
  return category == 'weapon' && handedness == 'one_handed';
}

int _halfDamage(int value) {
  return math.max(1, value ~/ 2);
}
