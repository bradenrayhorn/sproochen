<script lang="ts">
  import { onMount, type Component } from "svelte";
  import { proxyHistory } from "./history-proxy";
  import { setRouteContext, type RouteContext } from "./router-context";

  let RouteComponent = $state<Component | null>(null);

  let lastUri: string | undefined = undefined;

  export type Route = {
    path: string;
    component: Component;
    orRedirect?: () => string | undefined;
  };
  const props: { routes: Route[]; basePath?: string } = $props();

  const routeContext: RouteContext = $state({ params: {} });

  const routes = $derived.by(() => {
    return props.routes.map(({ path, ...route }) => {
      const basePath = (props.basePath ?? "").replace(/^\/+|\/+$/g, "");
      path = path.replace(/^\/+|\/+$/g, "");

      const regex = new RegExp(`${basePath}/${path.replaceAll("/", "\\/")}$`);

      return { path: regex, ...route };
    });
  });

  function updateRoute(uri: string) {
    uri = uri.replace("#", "");
    if (!uri.startsWith("/")) {
      uri = "/" + uri;
    }

    if (uri === lastUri) {
      return;
    }
    lastUri = uri;

    let route: Component | null = null;
    for (const candidate of routes) {
      const result = candidate.path.exec(uri);

      if (result) {
        if (candidate.orRedirect) {
          const goTo = candidate.orRedirect();
          if (goTo) {
            window.history.replaceState(undefined, "", goTo);
            updateRoute(goTo);
            return;
          }
        }

        routeContext.params = result.groups ?? {};
        route = candidate.component;
        break;
      }
    }

    RouteComponent = route;
  }

  setRouteContext(routeContext);

  onMount(() => {
    const doUpdate = () => updateRoute(location.hash);

    // pushstate and replacestate are custom events created by proxyHistory()
    window.addEventListener("popstate", doUpdate);
    window.addEventListener("pushstate", doUpdate);
    window.addEventListener("replacestate", doUpdate);

    return () => {
      window.removeEventListener("popstate", doUpdate);
      window.removeEventListener("pushstate", doUpdate);
      window.removeEventListener("replacestate", doUpdate);
    };
  });

  proxyHistory();
  updateRoute(location.hash);
</script>

<RouteComponent />
