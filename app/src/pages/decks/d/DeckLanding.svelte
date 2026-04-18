<script lang="ts">
  import { link } from "$lib/router/use-link";
  import { State } from "ts-fsrs";
  import { getDeckCtx } from "./deck-context";
  import { onMount } from "svelte";

  const { progress, baseURI } = getDeckCtx();
  const cards = Object.values(progress.doc().cardStates);
  let now = $state(new Date());

  const [totalCards, newCards, dueCards] = $derived.by(() => {
    let t = 0,
      n = 0,
      d = 0;

    for (const card of cards) {
      t++;
      if (card.state.state === State.New) {
        n++;
      } else if (card.state.due <= now) {
        d++;
      }
    }

    return [t, n, d];
  });

  function onRefresh() {
    now = new Date();
  }

  onMount(() => {
    const interval = setInterval(onRefresh, 5000);
    return () => clearInterval(interval);
  });
</script>

<svelte:document onvisibilitychange={onRefresh} />

<div>{totalCards} cards</div>

<div>{newCards} new cards to learn</div>

<div>{dueCards} cards ready to review</div>

<a href={`${baseURI}/study`} use:link>Begin session</a>
