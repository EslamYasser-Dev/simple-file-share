import { ar } from "./ar";
import { en, type Dict } from "./en";

export type { Dict, MockRow, ChipTone } from "./en";

export type Locale = "en" | "ar";

export const LOCALES: readonly Locale[] = ["en", "ar"] as const;

export function getDict(locale: Locale): Dict {
  return locale === "ar" ? ar : en;
}

/** Prefix a locale-less site path (e.g. "/getting-started/") for a locale. */
export function lp(locale: Locale, path: string): string {
  if (locale === "en") return path;
  return path === "/" ? "/ar" : `/ar${path}`;
}
