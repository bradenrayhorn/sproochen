<script lang="ts">
  import { rootDoc } from "./lib/repo/repo.svelte";
  import type { Route } from "./lib/router/Router.svelte";
  import Router from "./lib/router/Router.svelte";
  import NewDeckPage from "./pages/decks/NewDeckPage.svelte";
  import LandingPage from "./pages/LandingPage.svelte";
  import NotFoundPage from "./pages/NotFoundPage.svelte";

  const routes: Route[] = [
    { path: "/", component: LandingPage },
    { path: "/d/new", component: NewDeckPage },
    { path: ".*", component: NotFoundPage },
  ];
</script>

<main>
  {#await rootDoc.whenReady()}
    Loading...
  {:then}
    <Router {routes} />
  {:catch}
    An error occurred.
  {/await}
</main>
