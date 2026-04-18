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

  const { progressDoc }: { progressDoc: DocHandle<DeckProgressDoc> } = $props();

  setDeckCtx({
    baseURI: untrack(() => `/d/${progressDoc.url}`),
    progress: untrack(() => progressDoc),
  });

  const routes: Route[] = [
    {
      path: "/?",
      component: DeckLanding,
    },
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

<Router {routes} basePath="/d/([\w:]+)" />
