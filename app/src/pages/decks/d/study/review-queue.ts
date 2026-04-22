import type { CardProgressState } from "$lib/repo/repo.svelte";
import { State, type FSRS } from "ts-fsrs";

// How many seen cards to review before interleaving a new card.
const new_card_interleave_rate = 4;

const new_card_session_limit = 4;

const new_card_daily_limit = 20;

export function generateReviewQueue(
  size: number,
  scheduler: FSRS,
  cards: CardProgressState[],
): CardProgressState[] {
  const now = new Date();
  const oneDayAgo = new Date(now.getTime() - 24 * 60 * 60 * 1000);

  const newCardsToday = cards.filter(
    (card) =>
      card.state.state === State.Learning &&
      card.log.every((log) => log.review > oneDayAgo),
  ).length;

  // First, only consider due cards
  const dueCards = cards.filter((card) => card.state.due <= now);

  // P1 - Learning cards, oldest first
  const learningCards = dueCards
    .filter(
      (card) =>
        card.state.state === State.Learning ||
        card.state.state === State.Relearning,
    )
    .sort((a, b) => a.state.due.getTime() - b.state.due.getTime());

  // P2 - Reviewing cards, ordered by retrievability (least likely to remember first)
  const seenCards = dueCards
    .filter((card) => card.state.state === State.Review)
    .sort(
      (a, b) =>
        scheduler.get_retrievability(a.state, now, false) -
        scheduler.get_retrievability(b.state, now, false),
    );

  // P3 - New cards, random order, limit applied
  const newCards = dueCards
    .filter((card) => card.state.state === State.New)
    .sort(() => Math.random() - 0.5)
    .slice(
      0,
      Math.min(
        new_card_session_limit,
        Math.max(0, new_card_daily_limit - newCardsToday),
      ),
    );

  // pull the requested number of cards
  const queue = [];
  let cardsWithoutNew = 0;
  for (let i = 0; i < size; i++) {
    if (cardsWithoutNew > new_card_interleave_rate && newCards.length > 0) {
      queue.push(...newCards.splice(0, 1));
      cardsWithoutNew = 0;
    } else if (learningCards.length > 0) {
      queue.push(...learningCards.splice(0, 1));
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
