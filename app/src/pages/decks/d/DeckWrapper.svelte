<script lang="ts">
  import { repo, type DeckProgressDoc } from "$lib/repo/repo.svelte";
  import { getRouteContext } from "$lib/router/router-context";
  import type { AnyDocumentId } from "@automerge/vanillajs";
  import DeckRouter from "./DeckRouter.svelte";

  const deckProgressURL = getRouteContext().params.deckURL;
  const progressDoc = repo.find<DeckProgressDoc>(
    deckProgressURL as AnyDocumentId,
  );
</script>

{#await progressDoc}
  Loading doc...
{:then doc}
  <DeckRouter progressDoc={doc} />
{:catch error}
  Doc error
{/await}
