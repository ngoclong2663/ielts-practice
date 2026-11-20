"use client";

import { Flag } from "lucide-react";
import { useTranslations } from "next-intl";
import type { ReactNode } from "react";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { cn } from "@/lib/utils";
import { answerMode, overLimit, splitGaps } from "../../lib/questions";
import type { TAnswers, TGroup, TOption, TResult } from "../../types";

export type QuestionHandlers = {
  onAnswer: (n: number, value: string) => void;
  onToggleLetter: (groupFrom: number, letter: string, pick: number) => void;
  onFlag: (n: number) => void;
  onFocus: (n: number) => void;
};

type QuestionGroupProps = {
  group: TGroup;
  answers: TAnswers;
  current?: number;
  flags?: number[];
  /** Questions that cannot be changed in this run (a redo shows only the questions being redone as open). */
  locked?: (n: number) => boolean;
  /** Review: answers are shown, not edited; results colour them. */
  results?: Record<number, TResult>;
  handlers?: QuestionHandlers;
};

/** One block of questions sharing an instruction (Epic 10 §5.3). */
export function QuestionGroup({ group, answers, current, flags = [], locked, results, handlers }: QuestionGroupProps) {
  const t = useTranslations("exam");
  const readOnly = !handlers;
  const mode = answerMode(group.type);
  const ctx: Ctx = { group, answers, current, flags, locked: locked ?? (() => false), results, handlers, readOnly };

  return (
    <section className="mb-8" aria-labelledby={`g-${group.from}`}>
      <h3 id={`g-${group.from}`} className="text-[15px] font-semibold">
        {group.from === group.to ? t("question", { n: group.from }) : t("questions", { from: group.from, to: group.to })}
      </h3>
      <p className="mt-1 mb-4 text-sm text-muted-foreground">{group.instructions}</p>
      {group.figure && (
        // A map, plan or diagram (SVG, drawn on white so it reads in dark mode too).
        // eslint-disable-next-line @next/next/no-img-element -- our own SVG from the API, already sized by its viewBox
        <img
          src={group.figure}
          alt={group.title ?? t("figureAlt")}
          className="mb-4 w-full max-w-2xl rounded-xl border border-border bg-white"
        />
      )}
      {mode === "choice" && <ChoiceItems ctx={ctx} />}
      {mode === "choose_n" && <ChooseN ctx={ctx} />}
      {mode === "letter" && group.type !== "summary_bank" && <LetterItems ctx={ctx} />}
      {(mode === "gap" || group.type === "summary_bank") && <GapLayout ctx={ctx} />}
    </section>
  );
}

type Ctx = {
  group: TGroup;
  answers: TAnswers;
  current?: number;
  flags: number[];
  locked: (n: number) => boolean;
  results?: Record<number, TResult>;
  handlers?: QuestionHandlers;
  readOnly: boolean;
};

const str = (answers: TAnswers, n: number) => {
  const v = answers[String(n)];
  return typeof v === "string" ? v : "";
};

function Row({ ctx, n, children, stem }: { ctx: Ctx; n: number; children?: ReactNode; stem: string }) {
  const t = useTranslations("exam");
  const r = ctx.results?.[n];
  const flagged = ctx.flags.includes(n);
  return (
    <div
      data-q={n}
      className={cn("grid scroll-mt-4 grid-cols-[2.25rem_minmax(0,1fr)_2rem] items-start gap-2 border-t border-border py-3", ctx.locked(n) && "opacity-50")}
    >
      <QNum n={n} current={ctx.current === n} result={r} />
      <div className="min-w-0">
        <p className="pt-0.5 leading-relaxed" data-hl={`q${n}`}>
          {stem}
        </p>
        {children}
      </div>
      {!ctx.readOnly && !ctx.locked(n) ? (
        <button
          type="button"
          onClick={() => ctx.handlers?.onFlag(n)}
          aria-pressed={flagged}
          aria-label={t("flag", { n })}
          title={t("flagHint")}
          className={cn(
            "grid size-8 place-items-center rounded-md text-muted-foreground transition-colors hover:bg-muted",
            flagged && "bg-amber-100 text-amber-800 dark:bg-amber-500/20 dark:text-amber-300",
          )}
        >
          <Flag className="size-4" aria-hidden />
        </button>
      ) : (
        <span />
      )}
    </div>
  );
}

export function QNum({ n, current, result, label }: { n: number; current?: boolean; result?: TResult; label?: string }) {
  return (
    <span
      className={cn(
        "grid h-7 min-w-7 place-items-center rounded-md bg-muted px-1 text-[13px] font-semibold tabular-nums",
        current && "bg-accent text-accent-foreground ring-[1.5px] ring-primary",
        result && (result.correct ? "bg-accent text-accent-foreground" : "bg-destructive/10 text-destructive"),
      )}
    >
      {label ?? n}
    </span>
  );
}

function ChoiceButtons({ ctx, n, options }: { ctx: Ctx; n: number; options: string[] }) {
  const value = str(ctx.answers, n);
  const disabled = ctx.readOnly || ctx.locked(n);
  return (
    <div role="radiogroup" aria-label={String(n)} className="mt-2 flex flex-wrap gap-1.5">
      {options.map((o) => (
        <button
          key={o}
          type="button"
          role="radio"
          aria-checked={value === o}
          disabled={disabled}
          onClick={() => ctx.handlers?.onAnswer(n, value === o ? "" : o)}
          onFocus={() => ctx.handlers?.onFocus(n)}
          className={cn(
            "h-9 min-w-10 rounded-lg border border-input bg-background px-3 text-[13px] font-medium transition-colors hover:bg-muted disabled:cursor-default disabled:hover:bg-background",
            value === o && "border-primary bg-accent text-accent-foreground hover:bg-accent",
          )}
        >
          {o}
        </button>
      ))}
    </div>
  );
}

function OptionRows({ ctx, n, name, options, checked, multi, onPick }: {
  ctx: Ctx;
  n: number;
  name: string;
  options: TOption[];
  checked: (key: string) => boolean;
  multi?: boolean;
  onPick: (key: string) => void;
}) {
  const disabled = ctx.readOnly || ctx.locked(n);
  return (
    <div className="mt-2 grid gap-1">
      {options.map((o) => (
        <label
          key={o.key}
          className={cn(
            "flex cursor-pointer items-start gap-2.5 rounded-lg border border-transparent px-2.5 py-2 transition-colors hover:bg-muted",
            checked(o.key) && "border-primary bg-accent hover:bg-accent",
            disabled && "cursor-default hover:bg-transparent",
          )}
        >
          <input
            type={multi ? "checkbox" : "radio"}
            name={name}
            className="mt-1 accent-primary"
            checked={checked(o.key)}
            disabled={disabled}
            onChange={() => onPick(o.key)}
            onFocus={() => ctx.handlers?.onFocus(n)}
          />
          <span className="w-4 shrink-0 font-semibold">{o.key}</span>
          <span>{o.text}</span>
        </label>
      ))}
    </div>
  );
}

function ChoiceItems({ ctx }: { ctx: Ctx }) {
  const { group } = ctx;
  const fixed = group.type === "tfng" ? ["TRUE", "FALSE", "NOT GIVEN"] : group.type === "ynng" ? ["YES", "NO", "NOT GIVEN"] : null;
  return (
    <div>
      {(group.items ?? []).map((it) => (
        <Row key={it.n} ctx={ctx} n={it.n} stem={it.text}>
          {fixed ? (
            <ChoiceButtons ctx={ctx} n={it.n} options={fixed} />
          ) : (
            <OptionRows
              ctx={ctx}
              n={it.n}
              name={`mcq-${it.n}`}
              options={it.options ?? []}
              checked={(k) => str(ctx.answers, it.n) === k}
              onPick={(k) => ctx.handlers?.onAnswer(it.n, k)}
            />
          )}
        </Row>
      ))}
    </div>
  );
}

function ChooseN({ ctx }: { ctx: Ctx }) {
  const t = useTranslations("exam");
  const { group } = ctx;
  const v = ctx.answers[String(group.from)];
  const chosen = Array.isArray(v) ? v : [];
  const pick = group.pick ?? 2;
  return (
    <div data-q={group.from} className="grid scroll-mt-4 grid-cols-[2.25rem_minmax(0,1fr)] items-start gap-2 border-t border-border py-3">
      <QNum n={group.from} label={`${group.from}–${group.to}`} current={ctx.current !== undefined && ctx.current >= group.from && ctx.current <= group.to} />
      <div className="min-w-0">
        <p className="pt-0.5 leading-relaxed" data-hl={`q${group.from}`}>
          {group.stem}
        </p>
        <OptionRows
          ctx={ctx}
          n={group.from}
          name={`cn-${group.from}`}
          multi
          options={group.options ?? []}
          checked={(k) => chosen.includes(k)}
          onPick={(k) => ctx.handlers?.onToggleLetter(group.from, k, pick)}
        />
        <p className="mt-1.5 text-[13px] text-muted-foreground">{t("chosen", { count: chosen.length, pick })}</p>
      </div>
    </div>
  );
}

function OptionList({ group }: { group: TGroup }) {
  const opts = (group.options ?? []).filter((o) => o.text);
  if (!opts.length) return null;
  return (
    <div className="mb-3 rounded-xl border border-border bg-card p-3 text-sm">
      {group.title && <p className="mb-2 font-semibold">{group.title}</p>}
      <dl className="grid grid-cols-[1.75rem_minmax(0,1fr)] gap-x-2 gap-y-1">
        {opts.map((o) => (
          <div key={o.key} className="contents">
            <dt className="font-semibold">{o.key}</dt>
            <dd>{o.text}</dd>
          </div>
        ))}
      </dl>
    </div>
  );
}

function LetterPicker({ ctx, n, compact }: { ctx: Ctx; n: number; compact?: boolean }) {
  const t = useTranslations("exam");
  const keys = (ctx.group.options ?? []).map((o) => o.key);
  const value = str(ctx.answers, n);
  const disabled = ctx.readOnly || ctx.locked(n);
  if (keys.length <= 8 && !compact) return <ChoiceButtons ctx={ctx} n={n} options={keys} />;
  return (
    <Select
      items={keys.map((k) => ({ value: k, label: k }))}
      value={value || null}
      onValueChange={(v) => ctx.handlers?.onAnswer(n, v ? String(v) : "")}
      disabled={disabled}
    >
      <SelectTrigger className={cn("h-8 min-w-16", compact ? "inline-flex" : "mt-2")} aria-label={t("question", { n })} onFocus={() => ctx.handlers?.onFocus(n)}>
        <SelectValue placeholder={String(n)} />
      </SelectTrigger>
      <SelectContent>
        {keys.map((k) => (
          <SelectItem key={k} value={k}>
            {k}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}

function LetterItems({ ctx }: { ctx: Ctx }) {
  return (
    <div>
      <OptionList group={ctx.group} />
      {(ctx.group.items ?? []).map((it) => (
        <Row key={it.n} ctx={ctx} n={it.n} stem={it.text}>
          <LetterPicker ctx={ctx} n={it.n} />
        </Row>
      ))}
    </div>
  );
}

function Gap({ ctx, n }: { ctx: Ctx; n: number }) {
  const t = useTranslations("exam");
  const { group } = ctx;
  const r = ctx.results?.[n];
  if (group.type === "summary_bank") {
    if (ctx.readOnly) return <ReadOnlyGap n={n} value={str(ctx.answers, n)} result={r} />;
    return (
      <span data-q={n} className="inline-flex scroll-mt-4 items-center gap-1 align-baseline">
        <b className="text-xs text-muted-foreground">{n}</b>
        <LetterPicker ctx={ctx} n={n} compact />
      </span>
    );
  }
  if (ctx.readOnly) return <ReadOnlyGap n={n} value={str(ctx.answers, n)} result={r} />;
  const value = str(ctx.answers, n);
  const over = overLimit(value, group.word_limit, group.allow_number);
  const words = value.trim() ? value.trim().split(/\s+/).length : 0;
  return (
    <span data-q={n} className="inline-flex scroll-mt-4 items-center gap-1 align-baseline">
      <b className="text-xs text-muted-foreground">{n}</b>
      <input
        id={`gap-${n}`}
        value={value}
        disabled={ctx.locked(n)}
        onChange={(e) => ctx.handlers?.onAnswer(n, e.target.value)}
        onFocus={() => ctx.handlers?.onFocus(n)}
        autoComplete="off"
        autoCapitalize="off"
        autoCorrect="off"
        spellCheck={false}
        aria-label={t("question", { n })}
        aria-invalid={over || undefined}
        className={cn(
          "h-8 w-36 rounded-md border border-input bg-background px-2 text-[15px] outline-none focus:border-primary focus:ring-2 focus:ring-ring/40 disabled:opacity-60 sm:w-40",
          ctx.current === n && "border-primary",
          over && "border-destructive",
        )}
      />
      {over && <span className="text-xs font-semibold text-destructive">{t("words", { count: words })}</span>}
    </span>
  );
}

function ReadOnlyGap({ n, value, result }: { n: number; value: string; result?: TResult }) {
  return (
    <span className="inline-flex items-baseline gap-1">
      <b className="text-xs text-muted-foreground">{n}</b>
      <span className={cn("rounded px-1 font-semibold", result ? (result.correct ? "text-primary" : "text-destructive") : "")}>
        {value || "—"}
      </span>
    </span>
  );
}

function WithGaps({ ctx, text }: { ctx: Ctx; text: string }) {
  return (
    <>
      {splitGaps(text).map((piece, i) => (typeof piece === "number" ? <Gap key={i} ctx={ctx} n={piece} /> : <span key={i}>{piece}</span>))}
    </>
  );
}

function GapLayout({ ctx }: { ctx: Ctx }) {
  const { group } = ctx;
  const lines = group.lines ?? [];
  let body: ReactNode = null;
  switch (group.type) {
    case "notes":
    case "diagram":
      // A line starting with "## " is a section heading inside the notes.
      body = (
        <ul className="grid list-disc gap-3 pl-5">
          {lines.map((l, i) =>
            l.startsWith("## ") ? (
              <li key={i} className="-ml-5 list-none pt-1 font-semibold">
                {l.slice(3)}
              </li>
            ) : (
              <li key={i} className="leading-loose">
                <WithGaps ctx={ctx} text={l} />
              </li>
            ),
          )}
        </ul>
      );
      break;
    case "flowchart":
      body = (
        <ol className="grid gap-2">
          {lines.map((l, i) => (
            <li key={i} className="flex flex-col items-center gap-2">
              <span className="w-full rounded-lg border border-border bg-background px-3 py-2 text-center leading-loose">
                <WithGaps ctx={ctx} text={l} />
              </span>
              {i < lines.length - 1 && <span aria-hidden className="text-muted-foreground">↓</span>}
            </li>
          ))}
        </ol>
      );
      break;
    case "table":
    case "form":
      body = (
        <div className="overflow-x-auto">
          <table className="w-full border-collapse text-sm">
            {group.header && (
              <thead>
                <tr>
                  {group.header.map((h, i) => (
                    <th key={i} className="border border-border px-2 py-1.5 text-left font-semibold">
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
            )}
            <tbody>
              {(group.rows ?? []).map((row, ri) =>
                // A row with one cell is a section heading across the table or form.
                row.length === 1 ? (
                  <tr key={ri}>
                    <th colSpan={Math.max(group.header?.length ?? 0, ...(group.rows ?? []).map((r) => r.length))} className="border border-border bg-muted/60 px-2 py-1.5 text-left font-semibold">
                      {row[0]}
                    </th>
                  </tr>
                ) : (
                <tr key={ri}>
                  {row.map((cell, ci) => (
                    <td key={ci} className={cn("border border-border px-2 py-1.5 align-top leading-loose", group.type === "form" && ci === 0 && "text-muted-foreground")}>
                      <WithGaps ctx={ctx} text={cell} />
                    </td>
                  ))}
                </tr>
                ),
              )}
            </tbody>
          </table>
        </div>
      );
      break;
    case "short_answer":
      body = (group.items ?? []).map((it) => (
        <Row key={it.n} ctx={ctx} n={it.n} stem={it.text}>
          <div className="mt-2">
            <Gap ctx={ctx} n={it.n} />
          </div>
        </Row>
      ));
      break;
    case "sentence":
      body = (
        <ul className="grid gap-3">
          {lines.map((l, i) => (
            <li key={i} className="leading-loose">
              <WithGaps ctx={ctx} text={l} />
            </li>
          ))}
        </ul>
      );
      break;
    default:
      // summary and summary_bank: running text
      body = lines.map((l, i) => (
        <p key={i} className="leading-[2.4]">
          <WithGaps ctx={ctx} text={l} />
        </p>
      ));
  }
  return (
    <div>
      {group.type === "summary_bank" && <OptionList group={{ ...group, title: undefined }} />}
      <div className="rounded-xl border border-border bg-card p-4">
        {group.title && <p className="mb-3 text-center font-semibold">{group.title}</p>}
        {body}
      </div>
    </div>
  );
}
