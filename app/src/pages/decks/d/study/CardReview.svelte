<script lang="ts">
  import type { CardData } from "$lib/cards/card-set";
  import { Rating, type Grade } from "ts-fsrs";

  const {
    card,
    onRespond,
  }: { card: CardData; onRespond: (rating: Grade) => void } = $props();

  let revealed = $state(false);

  function onReveal() {
    revealed = true;
  }

  function onKeydown(event: KeyboardEvent) {
    if (revealed) {
      switch (event.key) {
        case "1":
          onRespond(Rating.Again);
          break;
        case "2":
          onRespond(Rating.Hard);
          break;
        case "3":
          onRespond(Rating.Good);
          break;
        case "4":
          onRespond(Rating.Easy);
          break;
      }
    } else {
      switch (event.key) {
        case " ":
        case "Enter":
          onReveal();
          break;
      }
    }
  }
</script>

<svelte:window on:keydown={onKeydown} />

<section>
  <div>
    {card.native_language}
  </div>

  {#if revealed}
    <div>
      {card.target_language}
    </div>
    <div>
      {card.part_of_speech}
    </div>
  {/if}
</section>

<section>
  {#if revealed}
    <div class="ratings">
      <button onclick={() => onRespond(Rating.Again)}>[1] Again</button>
      <button onclick={() => onRespond(Rating.Hard)}>[2] Hard</button>
      <button onclick={() => onRespond(Rating.Good)}>[3] Good</button>
      <button onclick={() => onRespond(Rating.Easy)}>[4] Easy</button>
    </div>
  {:else}
    <div class="ratings">
      <button onclick={onReveal}>Reveal</button>
    </div>
  {/if}
</section>

<style>
  .ratings {
    width: 100%;
    display: flex;
    justify-content: center;
    gap: 1rem;
  }
</style>
