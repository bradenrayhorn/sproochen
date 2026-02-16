export function goto(route: string) {
  window.history.pushState(undefined, "", route);
}
