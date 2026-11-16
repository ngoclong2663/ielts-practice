import type { Metadata } from "next";

// The admin board is private (Epic 6 §5): keep it out of search engines.
export const metadata: Metadata = { robots: { index: false, follow: false } };

export default function AdminLayout({ children }: LayoutProps<"/[locale]/admin">) {
  return children;
}
