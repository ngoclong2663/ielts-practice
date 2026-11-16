import { Suspense } from "react";
import { AdminShell, AdminTabs, AdminTabsLive } from "@/features/admin/components/AdminShell";

export default function AdminBoardLayout({ children }: LayoutProps<"/[locale]/admin">) {
  return (
    <AdminShell
      tabs={
        <Suspense fallback={<AdminTabs pathname="" />}>
          <AdminTabsLive />
        </Suspense>
      }
    >
      {children}
    </AdminShell>
  );
}
