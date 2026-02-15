<script lang="ts">
  import { rootDoc } from "./lib/repo/repo.svelte";
  import NameEdit from "./NameEdit.svelte";

  let hide = $state(false);
</script>

<main>Sproochen</main>

<div>{rootDoc.url}</div>

{#await rootDoc.whenReady()}
  Loading...
{:then _}
  {#if hide}
    <button onclick={() => (hide = false)}>show</button>
  {:else}
    <NameEdit />
    <button onclick={() => (hide = true)}>hide</button>
  {/if}
{/await}

<style>
</style>
