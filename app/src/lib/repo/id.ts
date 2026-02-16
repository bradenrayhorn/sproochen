export function id(): string {
  const array = new Uint8Array(32);
  self.crypto.getRandomValues(array);

  return Array.from(array)
    .map((byte) => byte.toString(16).padStart(2, "0"))
    .join("");
}
