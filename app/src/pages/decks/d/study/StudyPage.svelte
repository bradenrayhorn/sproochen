<script lang="ts">
  import { cardMap } from "$lib/cards/card-set";
  import type { CardProgressState } from "$lib/repo/repo.svelte";
  import { fsrs, generatorParameters, type Grade } from "ts-fsrs";
  import { generateReviewQueue } from "./review-queue";
  import { serializeCard } from "$lib/repo/card";
  import { getDeckCtx } from "../deck-context";
  import CardReview from "./CardReview.svelte";

  const ctx = getDeckCtx();

  const scheduler = fsrs(generatorParameters({ enable_fuzz: true }));
  const queue = generateReviewQueue(
    20,
    scheduler,
    Object.values(ctx.progress.doc().cardStates),
  );

  let reviewing: CardProgressState | undefined = $state(queue.shift());

  function onRespond(grade: Grade) {
    if (reviewing === undefined) return;

    const newState = scheduler.next(reviewing.state, new Date(), grade);
    const reviewedId = reviewing.id;
    ctx.progress.change((doc) => {
      doc.cardStates[reviewedId].state = serializeCard(newState.card);
      doc.cardStates[reviewedId].log.push(newState.log);
    });

    reviewing = queue.shift();
  }
</script>

{#if reviewing !== undefined}
  {#key reviewing.id}
    <CardReview card={cardMap[reviewing.id]} {onRespond} />
  {/key}
{:else}
  Nothing to review!
{/if}
