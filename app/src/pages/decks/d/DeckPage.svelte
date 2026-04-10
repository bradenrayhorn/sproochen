<script lang="ts">
  import type { DocHandle } from "@automerge/vanillajs";
  import { setDeckCtx } from "./deck-context";
  import type { DeckProgressDoc } from "$lib/repo/repo.svelte";
  import { untrack } from "svelte";
  import type { Route } from "$lib/router/Router.svelte";
  import StudyPage from "./StudyPage.svelte";
  import NotFoundPage from "../../NotFoundPage.svelte";
  import Router from "$lib/router/Router.svelte";

  const { progressDoc }: { progressDoc: DocHandle<DeckProgressDoc> } = $props();

  setDeckCtx({ progress: untrack(() => progressDoc) });

  const routes: Route[] = [
    {
      path: "/study",
      component: StudyPage,
    },
    {
      path: "/.*",
      component: NotFoundPage,
    },
  ];
</script>

<div>Deck {progressDoc.url}</div>
<Router {routes} basePath="/d/([\w:]+)" />
