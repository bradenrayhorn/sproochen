export type Deck = {
  name: string;
  cards: Card[];
  cardSets: CardSet[];
};

export type Card = {
  id: string;
  front: string;
  back: string;
};

export type CardSet = {
  id: string;
  name: string;
  parentSetID: string | null;
  cardIDs: Array<Card["id"]>;
};

export function newDeck(name: string): Deck {
  return {
    name,
    cards: [],
    cardSets: [],
  };
}
