<script lang="ts">
  import { newDeck, type Deck } from "$lib/repo/deck";
  import { repo, rootDoc, type DeckProgressDoc } from "$lib/repo/repo.svelte";
  import { goto } from "$lib/router/goto";

  import { createEmptyCard, fsrs, Grades, Rating } from "ts-fsrs";
  import { getRouteContext } from "$lib/router/router-context";
  import type { AnyDocumentId } from "@automerge/vanillajs";
  import { getDeckCtx } from "./deck-context";
  import { cardMap } from "$lib/cards/card-set";

  const ctx = getDeckCtx();
  console.log(ctx.progress.doc());

  /*
  const scheduler = fsrs({});
  const card1 = createEmptyCard(new Date());
  console.log(card1.due);
  const x = scheduler.next(card1, new Date(), Rating.Again);
  console.log("next", x.log);
  console.log("now due", x.card.due);
  */
</script>

Study

{#each ctx.progress.doc().cardStates as card}
  <p>{cardMap[card.id].target_language}</p>
{/each}
