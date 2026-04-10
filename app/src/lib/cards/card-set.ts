import cards from "./flashcards.json";

export const cardSet = cards
  .slice(0, 8)
  .map((card) => ({ id: `${card.lod_id}-${card.meaning_id}`, ...card }));

export const cardMap: Record<string, (typeof cardSet)[0]> = cardSet.reduce(
  (prev, cur) => ({ ...prev, [cur.id]: cur }),
  {},
);
