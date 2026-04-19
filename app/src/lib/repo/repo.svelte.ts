import {
  BroadcastChannelNetworkAdapter,
  DocHandle,
  IndexedDBStorageAdapter,
  Repo,
  type AnyDocumentId,
  type AutomergeUrl,
} from "@automerge/vanillajs";
import { type Card, type ReviewLog } from "ts-fsrs";

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

export type DeckCardStates = {
  nativeFront?: Record<string, CardProgressState>;
  targetFront?: Record<string, CardProgressState>;
};

export type DeckProgressDoc = {
  deckCardStates: DeckCardStates;
};

const repo = new Repo({
  network: [new BroadcastChannelNetworkAdapter()],
  storage: new IndexedDBStorageAdapter(),
});

function initializeDefaultDeckProgress(): DeckProgressDoc {
  return { deckCardStates: {} };
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
