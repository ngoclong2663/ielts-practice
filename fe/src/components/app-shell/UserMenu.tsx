"use client";

import { LogOut, MessageSquareText, Settings } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Skeleton } from "@/components/ui/skeleton";
import { useSignOut } from "@/features/auth/hooks/use-sign-out";
import { FeedbackDialog } from "@/features/feedback/components/FeedbackDialog";
import { useMe } from "@/features/profile/hooks/use-me";
import { useIsClient } from "@/hooks/use-is-client";
import { useRouter } from "@/i18n/navigation";

function initials(name: string, email: string): string {
  const source = name.trim() || email;
  const parts = source.split(/\s+/).filter(Boolean);
  return (parts.length > 1 ? parts[0][0] + parts[parts.length - 1][0] : source.slice(0, 2)).toUpperCase();
}

export function UserMenu() {
  const t = useTranslations("userMenu");
  const router = useRouter();
  const me = useMe();
  const isClient = useIsClient();
  const isPending = !isClient || me.isPending;
  const data = isClient ? me.data : undefined;
  const { signOut, isPending: isSigningOut } = useSignOut();
  const [feedbackOpen, setFeedbackOpen] = useState(false);

  if (isPending || !data) {
    return <Skeleton className="size-8 rounded-full" />;
  }
  const { user } = data;

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger
          aria-label={t("open")}
          className="rounded-full outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          <Avatar className="size-8">
            {user.avatar_url && <AvatarImage src={user.avatar_url} alt="" referrerPolicy="no-referrer" />}
            <AvatarFallback>{initials(user.display_name, user.email)}</AvatarFallback>
          </Avatar>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-60">
          <DropdownMenuGroup>
            <DropdownMenuLabel className="flex flex-col gap-0.5">
              <span className="truncate text-sm font-medium text-foreground">{user.display_name}</span>
              <span className="truncate text-xs font-normal text-muted-foreground">{user.email}</span>
            </DropdownMenuLabel>
          </DropdownMenuGroup>
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={() => router.push("/settings")}>
            <Settings className="size-4" strokeWidth={1.75} aria-hidden />
            {t("settings")}
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => setFeedbackOpen(true)}>
            <MessageSquareText className="size-4" strokeWidth={1.75} aria-hidden />
            {t("feedback")}
          </DropdownMenuItem>
          <DropdownMenuItem onClick={signOut} disabled={isSigningOut}>
            <LogOut className="size-4" strokeWidth={1.75} aria-hidden />
            {t("signOut")}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <FeedbackDialog open={feedbackOpen} onOpenChange={setFeedbackOpen} />
    </>
  );
}
