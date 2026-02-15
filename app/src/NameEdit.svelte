<script lang="ts">
  import { rootDoc } from "./lib/repo/repo.svelte";
  import { watch } from "./lib/repo/watch.svelte";

  let value = $state(rootDoc.doc().name);
  let curName = watch(rootDoc, "name");

  function saveName() {
    rootDoc.change((curDoc) => {
      curDoc.name = value;
    });
  }
</script>

<div>Name: `{curName.value}`</div>

<label>
  Name
  <input bind:value />
</label>

<button onclick={saveName}>Save</button>
