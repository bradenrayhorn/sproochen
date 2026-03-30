<script lang="ts">
  import { rootDoc } from "$lib/repo/repo.svelte";
  import { watch } from "$lib/repo/watch.svelte";
  import { link } from "$lib/router/use-link";

  const decks = watch(rootDoc, "decks");
</script>

Landing page

<a href="/d/new" use:link>new deck</a>

<div>
  {#each decks.value as deck (deck.url)}
    <div>
      <div>{deck.name}</div>
      <a href={`/d/${deck.url}`}>Study</a>
    </div>
  {/each}
</div>
