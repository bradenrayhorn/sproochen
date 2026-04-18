<script lang="ts">
  import { rootDoc } from "$lib/repo/repo.svelte";
  import { watch } from "$lib/repo/watch.svelte";
  import { link } from "$lib/router/use-link";

  const decks = watch(rootDoc, "decks");
</script>

Your ID: <code>{rootDoc.url}</code>

<main>
  <div>
    {#each decks.value as deck (deck.url)}
      <div>
        <a href={`/d/${deck.progressUrl}`} use:link>{deck.name}</a>
      </div>
    {/each}
  </div>
</main>
