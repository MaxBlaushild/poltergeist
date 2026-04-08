import 'character_action.dart';
import 'point_of_interest.dart';
import 'spell.dart';

class Exposition {
  final String id;
  final String zoneId;
  final String? pointOfInterestId;
  final PointOfInterest? pointOfInterest;
  final double latitude;
  final double longitude;
  final String title;
  final String description;
  final List<DialogueMessage> dialogue;
  final String imageUrl;
  final String thumbnailUrl;
  final String rewardMode;
  final String randomRewardSize;
  final int rewardExperience;
  final int rewardGold;
  final List<Map<String, dynamic>> materialRewards;
  final List<Map<String, dynamic>> itemRewards;
  final List<Spell> spellRewards;

  const Exposition({
    required this.id,
    required this.zoneId,
    this.pointOfInterestId,
    this.pointOfInterest,
    required this.latitude,
    required this.longitude,
    required this.title,
    required this.description,
    this.dialogue = const [],
    this.imageUrl = '',
    this.thumbnailUrl = '',
    this.rewardMode = 'random',
    this.randomRewardSize = 'small',
    this.rewardExperience = 0,
    this.rewardGold = 0,
    this.materialRewards = const [],
    this.itemRewards = const [],
    this.spellRewards = const [],
  });

  factory Exposition.fromJson(Map<String, dynamic> json) {
    PointOfInterest? pointOfInterest;
    final rawPoi = json['pointOfInterest'];
    if (rawPoi is Map<String, dynamic>) {
      pointOfInterest = PointOfInterest.fromJson(rawPoi);
    } else if (rawPoi is Map) {
      pointOfInterest = PointOfInterest.fromJson(
        Map<String, dynamic>.from(rawPoi),
      );
    }

    final dialogue = <DialogueMessage>[];
    final rawDialogue = json['dialogue'];
    if (rawDialogue is List) {
      for (final entry in rawDialogue) {
        if (entry is Map<String, dynamic>) {
          dialogue.add(DialogueMessage.fromJson(entry));
        } else if (entry is Map) {
          dialogue.add(
            DialogueMessage.fromJson(Map<String, dynamic>.from(entry)),
          );
        }
      }
    }

    final spellRewards = <Spell>[];
    final rawSpellRewards = json['spellRewards'];
    if (rawSpellRewards is List) {
      for (final reward in rawSpellRewards) {
        if (reward is Map<String, dynamic>) {
          final rawSpell = reward['spell'];
          if (rawSpell is Map<String, dynamic>) {
            spellRewards.add(Spell.fromJson(rawSpell));
          } else if (rawSpell is Map) {
            spellRewards.add(
              Spell.fromJson(Map<String, dynamic>.from(rawSpell)),
            );
          }
        } else if (reward is Map) {
          final rawSpell = reward['spell'];
          if (rawSpell is Map<String, dynamic>) {
            spellRewards.add(Spell.fromJson(rawSpell));
          } else if (rawSpell is Map) {
            spellRewards.add(
              Spell.fromJson(Map<String, dynamic>.from(rawSpell)),
            );
          }
        }
      }
    }

    return Exposition(
      id: json['id']?.toString() ?? '',
      zoneId: json['zoneId']?.toString() ?? '',
      pointOfInterestId: json['pointOfInterestId']?.toString(),
      pointOfInterest: pointOfInterest,
      latitude: (json['latitude'] as num?)?.toDouble() ?? 0.0,
      longitude: (json['longitude'] as num?)?.toDouble() ?? 0.0,
      title: json['title']?.toString() ?? '',
      description: json['description']?.toString() ?? '',
      dialogue: dialogue,
      imageUrl: json['imageUrl']?.toString() ?? '',
      thumbnailUrl: json['thumbnailUrl']?.toString() ?? '',
      rewardMode: json['rewardMode']?.toString() ?? 'random',
      randomRewardSize: json['randomRewardSize']?.toString() ?? 'small',
      rewardExperience: (json['rewardExperience'] as num?)?.toInt() ?? 0,
      rewardGold: (json['rewardGold'] as num?)?.toInt() ?? 0,
      materialRewards:
          (json['materialRewards'] as List<dynamic>?)
              ?.whereType<Map>()
              .map((entry) => Map<String, dynamic>.from(entry))
              .toList() ??
          const [],
      itemRewards:
          (json['itemRewards'] as List<dynamic>?)
              ?.whereType<Map>()
              .map((entry) => Map<String, dynamic>.from(entry))
              .toList() ??
          const [],
      spellRewards: spellRewards,
    );
  }
}

class ExpositionPerformResult {
  final String expositionId;
  final bool successful;
  final String title;
  final int rewardExperience;
  final int rewardGold;
  final List<Map<String, dynamic>> baseResourcesAwarded;
  final List<Map<String, dynamic>> itemsAwarded;
  final List<Spell> spellsAwarded;
  final bool awardedRewards;

  const ExpositionPerformResult({
    required this.expositionId,
    required this.successful,
    required this.title,
    required this.rewardExperience,
    required this.rewardGold,
    this.baseResourcesAwarded = const [],
    this.itemsAwarded = const [],
    this.spellsAwarded = const [],
    this.awardedRewards = false,
  });

  factory ExpositionPerformResult.fromJson(Map<String, dynamic> json) {
    final spellsAwarded =
        (json['spellsAwarded'] as List<dynamic>?)
            ?.whereType<Map>()
            .map((spell) => Spell.fromJson(Map<String, dynamic>.from(spell)))
            .toList() ??
        const <Spell>[];

    return ExpositionPerformResult(
      expositionId: json['expositionId']?.toString() ?? '',
      successful: json['successful'] == true,
      title: json['title']?.toString() ?? '',
      rewardExperience: (json['rewardExperience'] as num?)?.toInt() ?? 0,
      rewardGold: (json['rewardGold'] as num?)?.toInt() ?? 0,
      baseResourcesAwarded:
          (json['baseResourcesAwarded'] as List<dynamic>?)
              ?.whereType<Map>()
              .map((entry) => Map<String, dynamic>.from(entry))
              .toList() ??
          const [],
      itemsAwarded:
          (json['itemsAwarded'] as List<dynamic>?)
              ?.whereType<Map>()
              .map((entry) => Map<String, dynamic>.from(entry))
              .toList() ??
          const [],
      spellsAwarded: spellsAwarded,
      awardedRewards: json['awardedRewards'] == true,
    );
  }
}
