import type { InventoryItem, InventoryRecipe } from '@poltergeist/types';

export type CraftingStationKey = 'alchemy' | 'workshop';

export type CraftingPartyReference = {
  itemId: number;
  itemName: string;
  archived: boolean;
};

export type CraftingIngredientReference = CraftingPartyReference & {
  quantity: number;
};

export type CraftingRecipeReference = {
  key: string;
  id: string;
  station: CraftingStationKey;
  stationLabel: string;
  resultItemId: number;
  resultItemName: string;
  resultItemArchived: boolean;
  resultItemLevel: number;
  tier: number;
  isPublic: boolean;
  ingredientEntries: CraftingIngredientReference[];
  ingredientItemIds: number[];
  teacherEntries: CraftingPartyReference[];
  teacherItemIds: number[];
  missingTeacher: boolean;
  archivedDependency: boolean;
};

export type CraftingItemRelationship = {
  producedRecipes: CraftingRecipeReference[];
  ingredientInRecipes: CraftingRecipeReference[];
  teachesRecipes: CraftingRecipeReference[];
  taughtByItemIds: number[];
  orphanedPrivateRecipes: CraftingRecipeReference[];
};

export type CraftingGraph = {
  recipes: CraftingRecipeReference[];
  relationshipsByItemId: Map<number, CraftingItemRelationship>;
  itemIdsByRole: {
    results: number[];
    ingredients: number[];
    teachers: number[];
    orphanedPrivateResults: number[];
  };
  recipeCounts: {
    total: number;
    byStation: Record<CraftingStationKey, number>;
    missingTeacher: number;
    archivedDependency: number;
  };
};

const compareText = (left: string, right: string) =>
  left.localeCompare(right, undefined, { sensitivity: 'base' });

export const craftingStationLabel = (station: CraftingStationKey) =>
  station === 'alchemy' ? 'Alchemy' : 'Workshop';

export const summarizeCraftingRecipe = (recipe: CraftingRecipeReference) =>
  `${recipe.stationLabel} T${recipe.tier} ${recipe.isPublic ? 'Public' : 'Private'}`;

export const compareCraftingRecipes = (
  left: CraftingRecipeReference,
  right: CraftingRecipeReference
) => {
  if (left.station !== right.station) {
    return compareText(left.station, right.station);
  }
  if (left.tier !== right.tier) {
    return left.tier - right.tier;
  }
  return compareText(left.resultItemName, right.resultItemName);
};

export const matchesCraftingRecipeQuery = (
  recipe: CraftingRecipeReference,
  query: string
) => {
  const normalizedQuery = query.trim().toLowerCase();
  if (!normalizedQuery) return true;
  const haystack = [
    recipe.id,
    recipe.stationLabel,
    recipe.resultItemName,
    recipe.tier.toString(),
    recipe.isPublic ? 'public' : 'private',
    ...recipe.ingredientEntries.map(
      (entry) => `${entry.itemName} ${entry.quantity}`
    ),
    ...recipe.teacherEntries.map((entry) => entry.itemName),
  ]
    .join(' ')
    .toLowerCase();
  return haystack.includes(normalizedQuery);
};

const createEmptyRelationship = (): CraftingItemRelationship => ({
  producedRecipes: [],
  ingredientInRecipes: [],
  teachesRecipes: [],
  taughtByItemIds: [],
  orphanedPrivateRecipes: [],
});

const normalizeRecipeIngredients = (recipe?: InventoryRecipe | null) =>
  (recipe?.ingredients ?? []).filter(
    (ingredient) => ingredient.itemId > 0 && ingredient.quantity > 0
  );

const sortAndDedupeRecipes = (recipes: CraftingRecipeReference[]) =>
  Array.from(
    new Map(recipes.map((recipe) => [recipe.key, recipe])).values()
  ).sort(compareCraftingRecipes);

export const buildCraftingGraph = (items: InventoryItem[]): CraftingGraph => {
  const itemById = new Map(items.map((item) => [item.id, item]));
  const relationshipsByItemId = new Map<number, CraftingItemRelationship>(
    items.map((item) => [item.id, createEmptyRelationship()])
  );
  const teacherItemIdsByRecipeId = new Map<string, Set<number>>();

  items.forEach((item) => {
    (item.consumeTeachRecipeIds ?? []).forEach((rawRecipeId) => {
      const recipeId = rawRecipeId.trim();
      if (!recipeId) return;
      const current =
        teacherItemIdsByRecipeId.get(recipeId) ?? new Set<number>();
      current.add(item.id);
      teacherItemIdsByRecipeId.set(recipeId, current);
    });
  });

  const recipes: CraftingRecipeReference[] = [];
  const resultItemIds = new Set<number>();
  const ingredientItemIds = new Set<number>();
  const teacherItemIds = new Set<number>();
  let missingTeacherCount = 0;
  let archivedDependencyCount = 0;

  const registerRecipe = (
    station: CraftingStationKey,
    resultItem: InventoryItem,
    recipe: InventoryRecipe,
    index: number
  ) => {
    const recipeId = recipe.id.trim();
    const ingredientEntries = normalizeRecipeIngredients(recipe).map(
      (ingredient) => {
        const ingredientItem = itemById.get(ingredient.itemId);
        return {
          itemId: ingredient.itemId,
          itemName: ingredientItem?.name ?? `Item ${ingredient.itemId}`,
          archived: Boolean(ingredientItem?.archived),
          quantity: ingredient.quantity,
        };
      }
    );
    const teacherEntries = recipeId
      ? Array.from(teacherItemIdsByRecipeId.get(recipeId) ?? [])
          .map((teacherItemId) => {
            const teacherItem = itemById.get(teacherItemId);
            return {
              itemId: teacherItemId,
              itemName: teacherItem?.name ?? `Item ${teacherItemId}`,
              archived: Boolean(teacherItem?.archived),
            };
          })
          .sort((left, right) => compareText(left.itemName, right.itemName))
      : [];
    const missingTeacher = !recipe.isPublic && teacherEntries.length === 0;
    const archivedDependency =
      Boolean(resultItem.archived) ||
      ingredientEntries.some((entry) => entry.archived) ||
      teacherEntries.some((entry) => entry.archived);
    const reference: CraftingRecipeReference = {
      key: recipeId || `${station}:${resultItem.id}:${index}`,
      id: recipeId,
      station,
      stationLabel: craftingStationLabel(station),
      resultItemId: resultItem.id,
      resultItemName: resultItem.name,
      resultItemArchived: Boolean(resultItem.archived),
      resultItemLevel: resultItem.itemLevel ?? 1,
      tier: recipe.tier > 0 ? recipe.tier : 1,
      isPublic: recipe.isPublic ?? true,
      ingredientEntries,
      ingredientItemIds: ingredientEntries.map((entry) => entry.itemId),
      teacherEntries,
      teacherItemIds: teacherEntries.map((entry) => entry.itemId),
      missingTeacher,
      archivedDependency,
    };

    recipes.push(reference);
    resultItemIds.add(resultItem.id);

    const resultRelationship =
      relationshipsByItemId.get(resultItem.id) ?? createEmptyRelationship();
    resultRelationship.producedRecipes.push(reference);
    if (missingTeacher) {
      resultRelationship.orphanedPrivateRecipes.push(reference);
      missingTeacherCount += 1;
    }
    resultRelationship.taughtByItemIds = Array.from(
      new Set([
        ...resultRelationship.taughtByItemIds,
        ...reference.teacherItemIds,
      ])
    );
    relationshipsByItemId.set(resultItem.id, resultRelationship);

    reference.ingredientEntries.forEach((entry) => {
      ingredientItemIds.add(entry.itemId);
      const ingredientRelationship =
        relationshipsByItemId.get(entry.itemId) ?? createEmptyRelationship();
      ingredientRelationship.ingredientInRecipes.push(reference);
      relationshipsByItemId.set(entry.itemId, ingredientRelationship);
    });

    reference.teacherEntries.forEach((entry) => {
      teacherItemIds.add(entry.itemId);
      const teacherRelationship =
        relationshipsByItemId.get(entry.itemId) ?? createEmptyRelationship();
      teacherRelationship.teachesRecipes.push(reference);
      relationshipsByItemId.set(entry.itemId, teacherRelationship);
    });

    if (archivedDependency) {
      archivedDependencyCount += 1;
    }
  };

  items.forEach((item) => {
    (item.alchemyRecipes ?? []).forEach((recipe, index) =>
      registerRecipe('alchemy', item, recipe, index)
    );
    (item.workshopRecipes ?? []).forEach((recipe, index) =>
      registerRecipe('workshop', item, recipe, index)
    );
  });

  relationshipsByItemId.forEach((relationship, itemId) => {
    relationship.producedRecipes = sortAndDedupeRecipes(
      relationship.producedRecipes
    );
    relationship.ingredientInRecipes = sortAndDedupeRecipes(
      relationship.ingredientInRecipes
    );
    relationship.teachesRecipes = sortAndDedupeRecipes(
      relationship.teachesRecipes
    );
    relationship.orphanedPrivateRecipes = sortAndDedupeRecipes(
      relationship.orphanedPrivateRecipes
    );
    relationship.taughtByItemIds = Array.from(
      new Set(relationship.taughtByItemIds)
    ).sort((left, right) => {
      const leftName = itemById.get(left)?.name ?? `${left}`;
      const rightName = itemById.get(right)?.name ?? `${right}`;
      return compareText(leftName, rightName);
    });
    relationshipsByItemId.set(itemId, relationship);
  });

  const sortedRecipes = [...recipes].sort(compareCraftingRecipes);

  return {
    recipes: sortedRecipes,
    relationshipsByItemId,
    itemIdsByRole: {
      results: Array.from(resultItemIds).sort((left, right) => left - right),
      ingredients: Array.from(ingredientItemIds).sort(
        (left, right) => left - right
      ),
      teachers: Array.from(teacherItemIds).sort((left, right) => left - right),
      orphanedPrivateResults: Array.from(
        new Set(
          sortedRecipes
            .filter((recipe) => recipe.missingTeacher)
            .map((recipe) => recipe.resultItemId)
        )
      ).sort((left, right) => left - right),
    },
    recipeCounts: {
      total: sortedRecipes.length,
      byStation: {
        alchemy: sortedRecipes.filter((recipe) => recipe.station === 'alchemy')
          .length,
        workshop: sortedRecipes.filter(
          (recipe) => recipe.station === 'workshop'
        ).length,
      },
      missingTeacher: missingTeacherCount,
      archivedDependency: archivedDependencyCount,
    },
  };
};
