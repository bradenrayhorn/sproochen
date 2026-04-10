import { cardSet } from "$lib/cards/card-set";
import {
  Repo,
  BroadcastChannelNetworkAdapter,
  IndexedDBStorageAdapter,
  type AutomergeUrl,
  type AnyDocumentId,
  DocHandle,
} from "@automerge/vanillajs";

import { createEmptyCard, type Card } from "ts-fsrs";
import { serializeCard } from "./card";
type RootDoc = {
  decks: Array<DeckIndex>;
};

type DeckIndex = {
  name: string;
  url: AutomergeUrl;
  progressUrl: AutomergeUrl;
};

export type DeckProgressDoc = {
  cardStates: Array<ProgressCard>;
};

type ProgressCard = {
  id: string;
  state: Card;
};

const repo = new Repo({
  network: [new BroadcastChannelNetworkAdapter()],
  storage: new IndexedDBStorageAdapter(),
});

function initializeDefaultDeckProgress(): DeckProgressDoc {
  const cardStates = cardSet.map((cardData) => ({
    id: cardData.id,
    state: serializeCard(createEmptyCard()),
  }));
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
