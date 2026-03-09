import 'spell.dart';

class ScenarioItemReward {
  final int inventoryItemId;
  final int quantity;

  const ScenarioItemReward({
    required this.inventoryItemId,
    required this.quantity,
  });

  factory ScenarioItemReward.fromJson(Map<String, dynamic> json) {
    return ScenarioItemReward(
      inventoryItemId: (json['inventoryItemId'] as num?)?.toInt() ?? 0,
      quantity: (json['quantity'] as num?)?.toInt() ?? 0,
    );
  }
}

class ScenarioSpellReward {
  final String spellId;
  final Spell? spell;

  const ScenarioSpellReward({required this.spellId, this.spell});

  factory ScenarioSpellReward.fromJson(Map<String, dynamic> json) {
    Spell? spell;
    final rawSpell = json['spell'];
    if (rawSpell is Map<String, dynamic>) {
      spell = Spell.fromJson(rawSpell);
    } else if (rawSpell is Map) {
      spell = Spell.fromJson(Map<String, dynamic>.from(rawSpell));
    }
    return ScenarioSpellReward(
      spellId: json['spellId']?.toString() ?? '',
      spell: spell,
    );
  }
}

class ScenarioOption {
  final String id;
  final String optionText;
  final String successText;
  final String failureText;
  final String statTag;
  final List<String> proficiencies;
  final int? difficulty;
  final int rewardExperience;
  final int rewardGold;
  final List<ScenarioItemReward> itemRewards;
  final List<ScenarioItemReward> itemChoiceRewards;
  final List<ScenarioSpellReward> spellRewards;

  const ScenarioOption({
    required this.id,
    required this.optionText,
    this.successText = '',
    this.failureText = '',
    required this.statTag,
    this.proficiencies = const [],
    this.difficulty,
    this.rewardExperience = 0,
    this.rewardGold = 0,
    this.itemRewards = const [],
    this.itemChoiceRewards = const [],
    this.spellRewards = const [],
  });

  factory ScenarioOption.fromJson(Map<String, dynamic> json) {
    final proficiencies = <String>[];
    final rawProficiencies = json['proficiencies'];
    if (rawProficiencies is List) {
      for (final p in rawProficiencies) {
        final value = p?.toString().trim() ?? '';
        if (value.isNotEmpty) proficiencies.add(value);
      }
    }

    final rewards = <ScenarioItemReward>[];
    final rawRewards = json['itemRewards'];
    if (rawRewards is List) {
      for (final reward in rawRewards) {
        if (reward is Map<String, dynamic>) {
          rewards.add(ScenarioItemReward.fromJson(reward));
        }
      }
    }
    final choiceRewards = <ScenarioItemReward>[];
    final rawChoiceRewards = json['itemChoiceRewards'];
    if (rawChoiceRewards is List) {
      for (final reward in rawChoiceRewards) {
        if (reward is Map<String, dynamic>) {
          choiceRewards.add(ScenarioItemReward.fromJson(reward));
        }
      }
    }

    final spellRewards = <ScenarioSpellReward>[];
    final rawSpellRewards = json['spellRewards'];
    if (rawSpellRewards is List) {
      for (final reward in rawSpellRewards) {
        if (reward is Map<String, dynamic>) {
          spellRewards.add(ScenarioSpellReward.fromJson(reward));
        } else if (reward is Map) {
          spellRewards.add(
            ScenarioSpellReward.fromJson(Map<String, dynamic>.from(reward)),
          );
        }
      }
    }

    return ScenarioOption(
      id: json['id']?.toString() ?? '',
      optionText: json['optionText']?.toString() ?? '',
      successText: json['successText']?.toString() ?? '',
      failureText: json['failureText']?.toString() ?? '',
      statTag: json['statTag']?.toString() ?? '',
      proficiencies: proficiencies,
      difficulty: (json['difficulty'] as num?)?.toInt(),
      rewardExperience: (json['rewardExperience'] as num?)?.toInt() ?? 0,
      rewardGold: (json['rewardGold'] as num?)?.toInt() ?? 0,
      itemRewards: rewards,
      itemChoiceRewards: choiceRewards,
      spellRewards: spellRewards,
    );
  }
}

class Scenario {
  final String id;
  final String zoneId;
  final String? pointOfInterestId;
  final double latitude;
  final double longitude;
  final String prompt;
  final String imageUrl;
  final String thumbnailUrl;
  final int difficulty;
  final bool scaleWithUserLevel;
  final int rewardExperience;
  final int rewardGold;
  final bool openEnded;
  final List<ScenarioOption> options;
  final List<ScenarioItemReward> itemRewards;
  final List<ScenarioItemReward> itemChoiceRewards;
  final List<ScenarioSpellReward> spellRewards;
  final bool attemptedByUser;

  const Scenario({
    required this.id,
    required this.zoneId,
    this.pointOfInterestId,
    required this.latitude,
    required this.longitude,
    required this.prompt,
    required this.imageUrl,
    required this.thumbnailUrl,
    required this.difficulty,
    this.scaleWithUserLevel = false,
    required this.rewardExperience,
    required this.rewardGold,
    required this.openEnded,
    this.options = const [],
    this.itemRewards = const [],
    this.itemChoiceRewards = const [],
    this.spellRewards = const [],
    this.attemptedByUser = false,
  });

  factory Scenario.fromJson(Map<String, dynamic> json) {
    final options = <ScenarioOption>[];
    final rawOptions = json['options'];
    if (rawOptions is List) {
      for (final option in rawOptions) {
        if (option is Map<String, dynamic>) {
          options.add(ScenarioOption.fromJson(option));
        }
      }
    }

    final itemRewards = <ScenarioItemReward>[];
    final rawRewards = json['itemRewards'];
    if (rawRewards is List) {
      for (final reward in rawRewards) {
        if (reward is Map<String, dynamic>) {
          itemRewards.add(ScenarioItemReward.fromJson(reward));
        }
      }
    }
    final itemChoiceRewards = <ScenarioItemReward>[];
    final rawChoiceRewards = json['itemChoiceRewards'];
    if (rawChoiceRewards is List) {
      for (final reward in rawChoiceRewards) {
        if (reward is Map<String, dynamic>) {
          itemChoiceRewards.add(ScenarioItemReward.fromJson(reward));
        }
      }
    }

    final spellRewards = <ScenarioSpellReward>[];
    final rawSpellRewards = json['spellRewards'];
    if (rawSpellRewards is List) {
      for (final reward in rawSpellRewards) {
        if (reward is Map<String, dynamic>) {
          spellRewards.add(ScenarioSpellReward.fromJson(reward));
        } else if (reward is Map) {
          spellRewards.add(
            ScenarioSpellReward.fromJson(Map<String, dynamic>.from(reward)),
          );
        }
      }
    }

    return Scenario(
      id: json['id']?.toString() ?? '',
      zoneId: json['zoneId']?.toString() ?? '',
      pointOfInterestId: json['pointOfInterestId']?.toString(),
      latitude: (json['latitude'] as num?)?.toDouble() ?? 0.0,
      longitude: (json['longitude'] as num?)?.toDouble() ?? 0.0,
      prompt: json['prompt']?.toString() ?? '',
      imageUrl: json['imageUrl']?.toString() ?? '',
      thumbnailUrl: json['thumbnailUrl']?.toString() ?? '',
      difficulty: (json['difficulty'] as num?)?.toInt() ?? 24,
      scaleWithUserLevel: json['scaleWithUserLevel'] as bool? ?? false,
      rewardExperience: (json['rewardExperience'] as num?)?.toInt() ?? 0,
      rewardGold: (json['rewardGold'] as num?)?.toInt() ?? 0,
      openEnded: json['openEnded'] as bool? ?? false,
      options: options,
      itemRewards: itemRewards,
      itemChoiceRewards: itemChoiceRewards,
      spellRewards: spellRewards,
      attemptedByUser: json['attemptedByUser'] as bool? ?? false,
    );
  }
}

class ScenarioPerformResult {
  final bool successful;
  final String reason;
  final String outcomeText;
  final String scenarioId;
  final String? scenarioOptionId;
  final int roll;
  final String statTag;
  final int statValue;
  final List<String> proficiencies;
  final int proficiencyBonus;
  final int creativityBonus;
  final int threshold;
  final int totalScore;
  final int failureHealthDrained;
  final int failureManaDrained;
  final List<ScenarioAppliedFailureStatus> failureStatusesApplied;
  final int successHealthRestored;
  final int successManaRestored;
  final List<ScenarioAppliedFailureStatus> successStatusesApplied;
  final int rewardExperience;
  final int rewardGold;
  final List<Map<String, dynamic>> itemsAwarded;
  final List<Map<String, dynamic>> itemChoiceRewards;
  final List<Spell> spellsAwarded;

  const ScenarioPerformResult({
    required this.successful,
    required this.reason,
    required this.outcomeText,
    required this.scenarioId,
    this.scenarioOptionId,
    required this.roll,
    required this.statTag,
    required this.statValue,
    this.proficiencies = const [],
    required this.proficiencyBonus,
    required this.creativityBonus,
    required this.threshold,
    required this.totalScore,
    this.failureHealthDrained = 0,
    this.failureManaDrained = 0,
    this.failureStatusesApplied = const [],
    this.successHealthRestored = 0,
    this.successManaRestored = 0,
    this.successStatusesApplied = const [],
    required this.rewardExperience,
    required this.rewardGold,
    this.itemsAwarded = const [],
    this.itemChoiceRewards = const [],
    this.spellsAwarded = const [],
  });

  factory ScenarioPerformResult.fromJson(Map<String, dynamic> json) {
    final proficiencies = <String>[];
    final rawProficiencies = json['proficiencies'];
    if (rawProficiencies is List) {
      for (final p in rawProficiencies) {
        final value = p?.toString().trim() ?? '';
        if (value.isNotEmpty) proficiencies.add(value);
      }
    }
    final failureStatuses = <ScenarioAppliedFailureStatus>[];
    final rawFailureStatuses = json['failureStatusesApplied'];
    if (rawFailureStatuses is List) {
      for (final status in rawFailureStatuses) {
        if (status is Map<String, dynamic>) {
          failureStatuses.add(ScenarioAppliedFailureStatus.fromJson(status));
        } else if (status is Map) {
          failureStatuses.add(
            ScenarioAppliedFailureStatus.fromJson(
              Map<String, dynamic>.from(status),
            ),
          );
        }
      }
    }
    final successStatuses = <ScenarioAppliedFailureStatus>[];
    final rawSuccessStatuses = json['successStatusesApplied'];
    if (rawSuccessStatuses is List) {
      for (final status in rawSuccessStatuses) {
        if (status is Map<String, dynamic>) {
          successStatuses.add(ScenarioAppliedFailureStatus.fromJson(status));
        } else if (status is Map) {
          successStatuses.add(
            ScenarioAppliedFailureStatus.fromJson(
              Map<String, dynamic>.from(status),
            ),
          );
        }
      }
    }

    return ScenarioPerformResult(
      successful: json['successful'] as bool? ?? false,
      reason: json['reason']?.toString() ?? '',
      outcomeText: json['outcomeText']?.toString() ?? '',
      scenarioId: json['scenarioId']?.toString() ?? '',
      scenarioOptionId: json['scenarioOptionId']?.toString(),
      roll: (json['roll'] as num?)?.toInt() ?? 0,
      statTag: json['statTag']?.toString() ?? '',
      statValue: (json['statValue'] as num?)?.toInt() ?? 0,
      proficiencies: proficiencies,
      proficiencyBonus: (json['proficiencyBonus'] as num?)?.toInt() ?? 0,
      creativityBonus: (json['creativityBonus'] as num?)?.toInt() ?? 0,
      threshold: (json['threshold'] as num?)?.toInt() ?? 0,
      totalScore: (json['totalScore'] as num?)?.toInt() ?? 0,
      failureHealthDrained:
          (json['failureHealthDrained'] as num?)?.toInt() ?? 0,
      failureManaDrained: (json['failureManaDrained'] as num?)?.toInt() ?? 0,
      failureStatusesApplied: failureStatuses,
      successHealthRestored:
          (json['successHealthRestored'] as num?)?.toInt() ?? 0,
      successManaRestored: (json['successManaRestored'] as num?)?.toInt() ?? 0,
      successStatusesApplied: successStatuses,
      rewardExperience: (json['rewardExperience'] as num?)?.toInt() ?? 0,
      rewardGold: (json['rewardGold'] as num?)?.toInt() ?? 0,
      itemsAwarded:
          (json['itemsAwarded'] as List<dynamic>?)
              ?.whereType<Map>()
              .map((item) => Map<String, dynamic>.from(item))
              .toList() ??
          const [],
      itemChoiceRewards:
          (json['itemChoiceRewards'] as List<dynamic>?)
              ?.whereType<Map>()
              .map((item) => Map<String, dynamic>.from(item))
              .toList() ??
          const [],
      spellsAwarded:
          (json['spellsAwarded'] as List<dynamic>?)
              ?.whereType<Map>()
              .map((spell) => Spell.fromJson(Map<String, dynamic>.from(spell)))
              .toList() ??
          const [],
    );
  }
}

class ScenarioAppliedFailureStatus {
  final String name;
  final String description;
  final String effect;
  final bool positive;
  final int durationSeconds;

  const ScenarioAppliedFailureStatus({
    required this.name,
    required this.description,
    required this.effect,
    required this.positive,
    required this.durationSeconds,
  });

  factory ScenarioAppliedFailureStatus.fromJson(Map<String, dynamic> json) {
    return ScenarioAppliedFailureStatus(
      name: json['name']?.toString() ?? '',
      description: json['description']?.toString() ?? '',
      effect: json['effect']?.toString() ?? '',
      positive: json['positive'] as bool? ?? false,
      durationSeconds: (json['durationSeconds'] as num?)?.toInt() ?? 0,
    );
  }

  Map<String, dynamic> toJson() => {
    'name': name,
    'description': description,
    'effect': effect,
    'positive': positive,
    'durationSeconds': durationSeconds,
  };
}
