"use client";

import { Search } from "lucide-react";
import { useFormatter, useTranslations } from "next-intl";
import { useState } from "react";
import { PageHeader } from "@/components/shared/PageHeader";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { useIsClient } from "@/hooks/use-is-client";
import { useAdminUsers } from "../hooks/use-admin";
import { useDebounced } from "../hooks/use-debounced";
import { ADMIN_USER_SORTS } from "../types";
import type { TAdminUser, TAdminUserSort } from "../types";
import { Pager } from "./Pager";
import { QueryError } from "./QueryError";

function UserCell({ user }: { user: TAdminUser }) {
  const name = user.display_name || user.email;
  return (
    <div className="flex min-w-0 items-center gap-3">
      <Avatar className="size-8">
        {user.avatar_url && <AvatarImage src={user.avatar_url} alt="" referrerPolicy="no-referrer" />}
        <AvatarFallback>{name.slice(0, 2).toUpperCase()}</AvatarFallback>
      </Avatar>
      <div className="min-w-0">
        <p className="truncate font-medium">{name}</p>
        <p className="truncate text-xs text-muted-foreground">
          {user.email} · {user.ui_locale}
        </p>
      </div>
    </div>
  );
}

export function UsersView() {
  const t = useTranslations("admin.users");
  const format = useFormatter();
  const isClient = useIsClient();
  const [search, setSearch] = useState("");
  const [sort, setSort] = useState<TAdminUserSort>("last_active");
  const [page, setPage] = useState(1);
  const q = useDebounced(search);
  const query = useAdminUsers({ q, sort, page });
  const data = isClient ? query.data : undefined;

  const n = (v: number) => format.number(v);
  const when = (iso: string | null) => (iso ? format.relativeTime(new Date(iso)) : "-");
  const sortItems = ADMIN_USER_SORTS.map((s) => ({ value: s, label: t(`sort.${s}`) }));

  return (
    <div className="mx-auto max-w-6xl">
      <PageHeader title={t("title")} description={data ? t("subtitle", { count: data.total }) : undefined} />

      <div className="mb-4 flex flex-col gap-3 sm:flex-row">
        <div className="relative flex-1">
          <Search className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" aria-hidden />
          <Input
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
            placeholder={t("search")}
            aria-label={t("search")}
            className="pl-9"
          />
        </div>
        <Select
          items={sortItems}
          value={sort}
          onValueChange={(v) => {
            setSort((v ?? "last_active") as TAdminUserSort);
            setPage(1);
          }}
        >
          <SelectTrigger className="w-full sm:w-56" aria-label={t("sortLabel")}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {sortItems.map((s) => (
              <SelectItem key={s.value} value={s.value}>
                {s.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {query.isError && <QueryError onRetry={() => void query.refetch()} />}

      {!data && !query.isError ? (
        <Skeleton className="h-96 rounded-2xl" />
      ) : data && data.items.length === 0 ? (
        <p className="rounded-2xl border border-dashed border-border p-8 text-center text-sm text-muted-foreground">{t("empty")}</p>
      ) : data ? (
        <div className="rounded-2xl border border-border bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="min-w-56">{t("col.user")}</TableHead>
                <TableHead>{t("col.joined")}</TableHead>
                <TableHead>{t("col.lastActive")}</TableHead>
                <TableHead className="text-right">{t("col.decksCards")}</TableHead>
                <TableHead className="text-right">{t("col.reviews")}</TableHead>
                <TableHead className="text-right">{t("col.aiWords")}</TableHead>
                <TableHead className="text-right">{t("col.topicLists")}</TableHead>
                <TableHead className="text-right">{t("col.writing")}</TableHead>
                <TableHead className="text-right">{t("col.writingAi")}</TableHead>
                <TableHead className="text-right">{t("col.feedback")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.items.map((u) => (
                <TableRow key={u.id}>
                  <TableCell className="max-w-72">
                    <UserCell user={u} />
                  </TableCell>
                  <TableCell className="whitespace-nowrap text-muted-foreground">
                    {format.dateTime(new Date(u.created_at), { day: "numeric", month: "short", year: "numeric" })}
                  </TableCell>
                  <TableCell className="whitespace-nowrap text-muted-foreground">{when(u.last_active_at)}</TableCell>
                  <TableCell className="text-right tabular-nums">
                    {n(u.decks)} / {n(u.cards)}
                  </TableCell>
                  <TableCell className="text-right tabular-nums">
                    {n(u.reviews_total)}
                    <span className="block text-xs text-muted-foreground">{t("week", { count: n(u.reviews_7d) })}</span>
                  </TableCell>
                  <TableCell className="text-right tabular-nums">{n(u.ai_words)}</TableCell>
                  <TableCell className="text-right tabular-nums">{n(u.topic_lists)}</TableCell>
                  <TableCell className="text-right tabular-nums">
                    {n(u.writing_attempts)}
                    <span className="block text-xs text-muted-foreground">{t("submitted", { count: n(u.writing_submitted) })}</span>
                  </TableCell>
                  <TableCell className="text-right tabular-nums">{n(u.writing_ai_calls)}</TableCell>
                  <TableCell className="text-right tabular-nums">{n(u.feedback)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      ) : null}

      {data && <Pager page={data.page} pageSize={data.page_size} total={data.total} onPage={setPage} />}
    </div>
  );
}
