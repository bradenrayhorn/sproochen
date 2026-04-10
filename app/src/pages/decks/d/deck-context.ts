import type { DeckProgressDoc } from "$lib/repo/repo.svelte";
import type { DocHandle } from "@automerge/vanillajs";
import { getContext, setContext } from "svelte";

const key = "deck-context";

interface DeckContext {
  progress: DocHandle<DeckProgressDoc>;
}

export function setDeckCtx(ctx: DeckContext) {
  setContext(key, ctx);
}

export function getDeckCtx(): DeckContext {
  return getContext(key);
}
