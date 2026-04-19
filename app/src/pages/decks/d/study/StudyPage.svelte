<script lang="ts">
  import { cardMap, cardSet } from "$lib/cards/card-set";
  import type { CardProgressState } from "$lib/repo/repo.svelte";
  import {
    createEmptyCard,
    fsrs,
    generatorParameters,
    State,
    type Grade,
  } from "ts-fsrs";
  import { generateReviewQueue } from "./review-queue";
  import { serializeCard } from "$lib/repo/card";
  import { getDeckCtx } from "../deck-context";
  import CardReview from "./CardReview.svelte";
  import { link } from "$lib/router/use-link";
  import { getRouteContext } from "$lib/router/router-context";

  const ctx = getDeckCtx();
  const mode = getRouteContext().params.mode as "native" | "target";
  const deckKey = mode === "target" ? "targetFront" : "nativeFront";

  const initialCardStates = ctx.progress.doc().deckCardStates[deckKey];

  // sync card state with card set
  if (
    initialCardStates === undefined ||
    Object.keys(initialCardStates).length !== cardSet.length
  ) {
    ctx.progress.change((doc) => {
      if (doc.deckCardStates[deckKey] === undefined) {
        doc.deckCardStates[deckKey] = {};
      }

      for (const card of cardSet) {
        if (doc.deckCardStates[deckKey][card.id] === undefined) {
          doc.deckCardStates[deckKey][card.id] = {
            id: card.id,
            state: serializeCard(createEmptyCard()),
            log: [],
          };
        }
      }
    });
  }

  const scheduler = fsrs(generatorParameters({ enable_fuzz: true }));
  const queue = generateReviewQueue(
    20,
    scheduler,
    Object.values(ctx.progress.doc().deckCardStates[deckKey] ?? {}),
  );

  let reviewing: CardProgressState | undefined = $state(queue.shift());

  function onRespond(grade: Grade) {
    if (reviewing === undefined) return;

    const newState = scheduler.next(reviewing.state, new Date(), grade);
    const reviewedId = reviewing.id;
    ctx.progress.change((doc) => {
      if (doc.deckCardStates[deckKey] === undefined) return;
      doc.deckCardStates[deckKey][reviewedId].state = serializeCard(
        newState.card,
      );
      doc.deckCardStates[deckKey][reviewedId].log.push(newState.log);
    });

    reviewing = queue.shift();
  }
</script>

{#if reviewing !== undefined}
  {#key reviewing.id}
    <CardReview
      card={cardMap[reviewing.id]}
      {mode}
      {onRespond}
      isNew={reviewing.state.state === State.New}
    />
  {/key}
{:else}
  Done!
  <a href={`${ctx.baseURI}`} use:link>Go back</a>
{/if}
