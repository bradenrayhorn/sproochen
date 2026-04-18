import { cardSet } from "$lib/cards/card-set";
import {
  BroadcastChannelNetworkAdapter,
  DocHandle,
  IndexedDBStorageAdapter,
  Repo,
  type AnyDocumentId,
  type AutomergeUrl,
} from "@automerge/vanillajs";
import { createEmptyCard, type Card, type ReviewLog } from "ts-fsrs";
import { serializeCard } from "./card";

type RootDoc = {
  decks: Array<DeckIndex>;
};

type DeckIndex = {
  name: string;
  url: AutomergeUrl;
  progressUrl: AutomergeUrl;
};

export type CardProgressState = {
  id: string;
  state: Card;
  log: Array<ReviewLog>;
};

export type DeckProgressDoc = {
  cardStates: Record<string, CardProgressState>;
};

const repo = new Repo({
  network: [new BroadcastChannelNetworkAdapter()],
  storage: new IndexedDBStorageAdapter(),
});

function initializeDefaultDeckProgress(): DeckProgressDoc {
  const cardStates = cardSet.reduce(
    (cardData: Record<string, CardProgressState>, card) => {
      cardData[card.id] = {
        id: card.id,
        state: serializeCard(createEmptyCard()),
        log: [],
      };
      return cardData;
    },
    {},
  );
  return { cardStates };
}

async function initRootDoc(): Promise<DocHandle<RootDoc>> {
  const key = "sproochen-root";
  const docId = localStorage.getItem(key);
  if (!docId) {
    const deckDoc = repo.create({}).url;
    const progressDoc = repo.create<DeckProgressDoc>(
      initializeDefaultDeckProgress(),
    );

    const rootDoc = repo.create<RootDoc>({
      decks: [
        {
          name: "Default Deck",
          url: deckDoc,
          progressUrl: progressDoc.url,
        },
      ],
    });
    localStorage.setItem(key, rootDoc.url);
    return rootDoc;
  }

  const rootDoc = await repo.find<RootDoc>(docId as AnyDocumentId);

  return rootDoc;
}

const rootDoc = await initRootDoc();

export { repo, rootDoc };
