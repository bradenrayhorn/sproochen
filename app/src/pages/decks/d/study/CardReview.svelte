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

<section class="card">
  <div class="front">
    <div class="word">{card.native_language}</div>
    {#if card.english_clarifier}
      <div class="subtext">({card.english_clarifier})</div>
    {/if}
  </div>

  {#if revealed}
    <div class="back">
      <div class="word">
        {card.target_language}
      </div>
      <div class="subtext">
        {card.part_of_speech}
      </div>
    </div>
  {/if}
</section>

<section>
  {#if revealed}
    <div class="actions">
      <button onclick={() => onRespond(Rating.Again)}>[1] Again</button>
      <button onclick={() => onRespond(Rating.Hard)}>[2] Hard</button>
      <button onclick={() => onRespond(Rating.Good)}>[3] Good</button>
      <button onclick={() => onRespond(Rating.Easy)}>[4] Easy</button>
    </div>
  {:else}
    <div class="actions">
      <button onclick={onReveal}>Reveal</button>
    </div>
  {/if}
</section>

<style>
  .card {
    display: grid;
    gap: 1rem;
    justify-items: center;

    & .word {
      font-size: 1.5rem;
      font-weight: 700;
    }

    & .front,
    .back {
      display: grid;
      justify-items: center;
    }
  }

  .actions {
    margin-block-start: 1rem;
    width: 100%;
    display: flex;
    justify-content: center;
    gap: 1rem;
  }
</style>
