import { createContext } from "svelte";

export interface RouteContext {
  params: Record<string, string>;
}

export const [getRouteContext, setRouteContext] = createContext<RouteContext>();
