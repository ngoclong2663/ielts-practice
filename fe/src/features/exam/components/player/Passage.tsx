import type { TPart } from "../../types";

type PassageProps = {
  part: TPart;
  /** Reading text size in px (15 / 17 / 20). */
  size: number;
};

/** The reading passage. Each paragraph is a highlight block (data-hl="p-<id>") with its letter in the margin. */
export function Passage({ part, size }: PassageProps) {
  const lettered = (part.content.paragraphs ?? []).some((p) => p.label);
  return (
    <article className="mx-auto max-w-[70ch]" style={{ fontSize: size, lineHeight: 1.65 }}>
      <h1 className="mb-1.5 text-2xl leading-tight font-semibold tracking-tight">{part.title}</h1>
      {part.content.subtitle && <p className="mb-6 text-[15px] text-muted-foreground">{part.content.subtitle}</p>}
      <div className={part.content.subtitle ? "" : "mt-5"}>
        {(part.content.paragraphs ?? []).map((p) => (
          <div key={p.id} id={`para-${p.id}`} className={lettered ? "mb-4 grid grid-cols-[1.75rem_minmax(0,1fr)] gap-2" : "mb-4"}>
            {lettered && <b className="pt-0.5 text-[15px] text-muted-foreground">{p.label}</b>}
            <p data-hl={`p-${p.id}`} className="select-text">
              {p.text}
            </p>
          </div>
        ))}
      </div>
    </article>
  );
}
