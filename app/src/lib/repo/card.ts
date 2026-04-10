import type { Card } from "ts-fsrs";

export function serializeCard(cardState: Card): Card {
  const newState = { ...cardState };
  if (cardState.last_review === undefined) {
    delete newState.last_review;
  }
  return newState;
}
