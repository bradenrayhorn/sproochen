<script lang="ts">
  import { newDeck, type Deck } from "$lib/repo/deck";
  import { repo, rootDoc } from "$lib/repo/repo.svelte";
  import { goto } from "$lib/router/goto";

  let deckName = $state("");

  function onSubmit() {
    const deck = repo.create<Deck>(newDeck(deckName));
    rootDoc.change((curRoot) => {
      curRoot.decks.push({ name: deck.doc().name, url: deck.url });
    });

    goto("/");
  }
</script>

<form onsubmit={onSubmit}>
  <label>
    Name
    <input bind:value={deckName} />
  </label>
  <button type="submit">Create deck</button>
</form>
