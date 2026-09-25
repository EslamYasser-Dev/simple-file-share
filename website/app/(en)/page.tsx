import type { Metadata } from "next";
import { Landing } from "../../components/landing";
import { getDict } from "../../lib/i18n";

const d = getDict("en");

export const metadata: Metadata = {
  title: d.meta.title,
  description: d.meta.description,
};

export default function HomePage() {
  return <Landing locale="en" />;
}
