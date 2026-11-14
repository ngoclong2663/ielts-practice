import { cn } from "@/lib/utils";
import type { TSitePhoto } from "@/assets/images";

const UTM = "utm_source=ielts_practice&utm_medium=referral";

type PhotoCreditProps = {
  photo: TSitePhoto;
  className?: string;
};

// Attribution required by the Unsplash API guidelines: photographer + Unsplash, with UTM links.
export function PhotoCredit({ photo, className }: PhotoCreditProps) {
  return (
    <p className={cn("text-[11px] leading-tight", className)}>
      Photo by{" "}
      <a href={`${photo.authorUrl}?${UTM}`} target="_blank" rel="noopener noreferrer" className="underline underline-offset-2">
        {photo.author}
      </a>{" "}
      on{" "}
      <a href={`https://unsplash.com/?${UTM}`} target="_blank" rel="noopener noreferrer" className="underline underline-offset-2">
        Unsplash
      </a>
    </p>
  );
}
