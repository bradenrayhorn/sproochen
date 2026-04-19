<script lang="ts">
  import type { CardData } from "$lib/cards/card-set";
  import { Rating, type Grade } from "ts-fsrs";
  import { getDeckCtx } from "../deck-context";
  import { onMount } from "svelte";

  const { audioPlayer } = getDeckCtx();

  const {
    card,
    mode,
    onRespond,
    isNew,
  }: {
    card: CardData;
    mode: "target" | "native";
    isNew: boolean;
    onRespond: (rating: Grade) => void;
  } = $props();

  let revealed = $state(false);
  let isPlaying = $state(false);

  function onReveal() {
    revealed = true;

    if (mode === "native") {
      onPlayAudio();
    }
  }

  function onPlayAudio() {
    if (isPlaying) return;
    isPlaying = true;
    audioPlayer
      .playAudio(card.lod_id)
      .catch((error) => console.error(error))
      .finally(() => {
        isPlaying = false;
      });
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
        case "r":
          onPlayAudio();
          break;
      }
    } else {
      switch (event.key) {
        case "r":
          if (mode === "target") onPlayAudio();
          break;
        case " ":
        case "Enter":
          onReveal();
          break;
      }
    }
  }

  onMount(() => {
    if (mode === "target") {
      onPlayAudio();
    }
  });
</script>

<svelte:window on:keydown={onKeydown} />

<section class="card">
  <div class="front">
    {#if mode === "native"}
      <div class="word">{card.native_language}</div>
      {#if card.english_clarifier}
        <div class="subtext">({card.english_clarifier})</div>
      {/if}
    {:else}
      <div class="word">
        {card.target_language}
      </div>
      {#if card.plural}
        <div class="subtext">
          plural: {card.plural?.join(" / ")}
        </div>
      {/if}

      {#if revealed}
        <div class="subtext">
          {card.part_of_speech}
        </div>
      {/if}

      <button class="player" onclick={onPlayAudio} disabled={isPlaying}>
        {#if isPlaying}
          Playing...
        {:else}
          Hear 🔊
        {/if}
      </button>
    {/if}
  </div>

  {#if revealed}
    <div class="back">
      {#if mode === "native"}
        <div class="word">
          {card.target_language}
        </div>
        {#if card.plural}
          <div class="subtext">
            plural: {card.plural?.join(" / ")}
          </div>
        {/if}
        <div class="subtext">
          {card.part_of_speech}
        </div>

        <button class="player" onclick={onPlayAudio} disabled={isPlaying}>
          {#if isPlaying}
            Playing...
          {:else}
            Hear 🔊
          {/if}
        </button>
      {:else}
        <div class="word">{card.native_language}</div>
        {#if card.english_clarifier}
          <div class="subtext">({card.english_clarifier})</div>
        {/if}
      {/if}
    </div>
  {/if}
</section>

<section>
  {#if revealed}
    <div class="actions">
      {#if isNew}
        <button onclick={() => onRespond(Rating.Again)}>Next</button>
      {:else}
        <button onclick={() => onRespond(Rating.Again)}>[1] Again</button>
        <button onclick={() => onRespond(Rating.Hard)}>[2] Hard</button>
        <button onclick={() => onRespond(Rating.Good)}>[3] Good</button>
        <button onclick={() => onRespond(Rating.Easy)}>[4] Easy</button>
      {/if}
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
    gap: 3rem;
    justify-items: center;

    & .word {
      font-size: 1.5rem;
      font-weight: 700;
    }

    & .front,
    .back {
      display: grid;
      justify-items: center;

      & .player {
        margin-block-start: 1rem;
      }
    }
  }

  .actions {
    margin-block-start: 3rem;
    width: 100%;
    display: flex;
    justify-content: center;
    gap: 1rem;
  }
</style>
