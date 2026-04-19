<script lang="ts">
  import { link } from "$lib/router/use-link";
  import { State } from "ts-fsrs";
  import { getDeckCtx } from "./deck-context";
  import { onMount } from "svelte";
  import type { CardProgressState } from "$lib/repo/repo.svelte";

  const { progress, baseURI } = getDeckCtx();
  const deckCardStates = progress.doc().deckCardStates;
  let now = $state(new Date());

  function aggregateCards(cards?: Record<string, CardProgressState>) {
    let t = 0,
      n = 0,
      d = 0;

    for (const card of Object.values(cards ?? {})) {
      t++;
      if (card.state.state === State.New) {
        n++;
      } else if (card.state.due <= now) {
        d++;
      }
    }

    return [t, n, d];
  }

  const [
    targetTotalCards,
    targetNewCards,
    targetDueCards,
    nativeTotalCards,
    nativeNewCards,
    nativeDueCards,
  ] = $derived.by(() => {
    const [tt, tn, td] = aggregateCards(deckCardStates.targetFront);
    const [nt, nn, nd] = aggregateCards(deckCardStates.nativeFront);
    return [tt, tn, td, nt, nn, nd];
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

<div>
  <h2>Study target language</h2>

  <p>You will see the target language.</p>

  <p>{targetTotalCards} cards</p>
  <p>{targetNewCards} new cards to learn</p>
  <p>{targetDueCards} cards ready to review</p>

  <a href={`${baseURI}/study/target`} use:link>Begin session</a>
</div>

<div>
  <h2>Study native language</h2>

  <p>You will see the native language.</p>

  <p>{nativeTotalCards} cards</p>
  <p>{nativeNewCards} new cards to learn</p>
  <p>{nativeDueCards} cards ready to review</p>

  <a href={`${baseURI}/study/native`} use:link>Begin session</a>
</div>
