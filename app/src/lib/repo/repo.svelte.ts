import {
  Repo,
  BroadcastChannelNetworkAdapter,
  IndexedDBStorageAdapter,
  type AutomergeUrl,
  type AnyDocumentId,
  DocHandle,
} from "@automerge/vanillajs";

type RootDoc = {
  name: string;
  progress: AutomergeUrl;
  decks: Array<AutomergeUrl>;
};

const repo = new Repo({
  network: [new BroadcastChannelNetworkAdapter()],
  storage: new IndexedDBStorageAdapter(),
});

async function initRootDoc(): Promise<DocHandle<RootDoc>> {
  const key = "sproochen-root";
  const docId = localStorage.getItem(key);
  if (!docId) {
    const progressDoc = repo.create({}).url;
    const rootDoc = repo.create<RootDoc>({
      progress: progressDoc,
      decks: [],
      name: "testname",
    });
    localStorage.setItem(key, rootDoc.url);
    return rootDoc;
  }

  const rootDoc = await repo.find<RootDoc>(docId as AnyDocumentId);

  return rootDoc;
}

const rootDoc = await initRootDoc();

export { repo, rootDoc };
