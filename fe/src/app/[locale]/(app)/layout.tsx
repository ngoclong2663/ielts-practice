import { AppShell } from "@/components/app-shell/AppShell";

export default function AuthenticatedLayout({ children }: LayoutProps<"/[locale]">) {
  return <AppShell>{children}</AppShell>;
}
