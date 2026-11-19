import type { TLibraryWordsParams } from "./types";

export const libraryKeys = {
  all: ["library"] as const,
  collections: () => [...libraryKeys.all, "collections"] as const,
  words: (slug: string) => [...libraryKeys.all, "words", slug] as const,
  wordList: (slug: string, params: TLibraryWordsParams) => [...libraryKeys.words(slug), params] as const,
};
