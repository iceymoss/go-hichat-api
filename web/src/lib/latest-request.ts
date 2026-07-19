export function createLatestRequest() {
  let generation = 0;

  return {
    begin() {
      const current = ++generation;
      return () => current === generation;
    },
    invalidate() {
      generation += 1;
    },
  };
}
