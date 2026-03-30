import type { DocHandle } from "@automerge/vanillajs";
import { onDestroy } from "svelte";

export function watch<T, K extends keyof T>(
  doc: DocHandle<T>,
  key: K,
): { value: T[K] } {
  const state = $state({ value: doc.doc()[key] });

  const handler = (event: { doc: T }) => {
    state.value = event.doc[key];
  };

  doc.on("change", handler);
  onDestroy(() => doc.off("change", handler));

  return state;
}
