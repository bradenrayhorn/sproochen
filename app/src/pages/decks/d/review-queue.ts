import type { CardProgressState } from "$lib/repo/repo.svelte";
import type { FSRS } from "ts-fsrs";

// How many seen cards to review before interleaving a new card.
const new_card_interleave_rate = 4;

export function generateReviewQueue(
  size: number,
  scheduler: FSRS,
  cards: CardProgressState[],
): CardProgressState[] {
  const dueCards = cards.filter((card) => card.state.due <= new Date());

  const newCards = dueCards.filter(
    (card) => card.state.last_review === undefined,
  );
  // shuffle new cards
  newCards.sort(() => Math.random() - 0.5);

  const seenCards = dueCards.filter(
    (card) => card.state.last_review !== undefined,
  );
  // sort seenCards by most likely to get wrong first
  seenCards.sort(
    (a, b) =>
      scheduler.get_retrievability(a.state, new Date(), false) -
      scheduler.get_retrievability(b.state, new Date(), false),
  );

  // pull the requested number of cards
  const queue = [];
  let cardsWithoutNew = 0;
  for (let i = 0; i < size; i++) {
    if (cardsWithoutNew > new_card_interleave_rate && newCards.length > 0) {
      queue.push(...newCards.splice(0, 1));
      cardsWithoutNew = 0;
    } else if (seenCards.length > 0) {
      queue.push(...seenCards.splice(0, 1));
      cardsWithoutNew++;
    } else if (newCards.length > 0) {
      queue.push(...newCards.splice(0, 1));
      cardsWithoutNew = 0;
    } else {
      break;
    }
  }
  return queue;
}
