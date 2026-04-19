<script lang="ts">
  import type { DocHandle } from "@automerge/vanillajs";
  import { setDeckCtx } from "./deck-context";
  import type { DeckProgressDoc } from "$lib/repo/repo.svelte";
  import { untrack } from "svelte";
  import type { Route } from "$lib/router/Router.svelte";
  import NotFoundPage from "../../NotFoundPage.svelte";
  import Router from "$lib/router/Router.svelte";
  import DeckLanding from "./DeckLanding.svelte";
  import StudyPage from "./study/StudyPage.svelte";
  import { AudioPlayer } from "$lib/audio/loader.svelte";

  const props: { progressDoc: DocHandle<DeckProgressDoc> } = $props();
  const progressDoc = untrack(() => props.progressDoc);

  if (progressDoc.doc().deckCardStates === undefined) {
    progressDoc.change((doc) => (doc.deckCardStates = {}));
  }

  const audioPlayer = new AudioPlayer();

  setDeckCtx({
    baseURI: `/d/${progressDoc.url}`,
    progress: progressDoc,
    audioPlayer,
  });

  const routes: Route[] = [
    {
      path: "/?",
      component: DeckLanding,
    },
    {
      path: "/study/(?<mode>[\\w]+)",
      component: StudyPage,
    },
    {
      path: "/.*",
      component: NotFoundPage,
    },
  ];
</script>

<Router {routes} basePath="/d/([\w:]+)" />
