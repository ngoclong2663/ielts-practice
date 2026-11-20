import { type RefObject, useCallback, useEffect, useRef, useState } from "react";
import type { TRunState } from "../lib/run-state";
import type { TListenMode, TPart } from "../types";

/** Seconds learners get to check their answers at the end of a full listening test in exam mode. */
export const CHECK_SECONDS = 120;
export const SPEEDS = [0.75, 1, 1.25] as const;

type Options = {
  audioRef: RefObject<HTMLAudioElement | null>;
  parts: TPart[];
  mode: TListenMode;
  fullTest: boolean;
  /** The part the learner is looking at (practice mode follows it). */
  currentPartNo: number;
  saved: TRunState["audio"];
  onSave: (audio: NonNullable<TRunState["audio"]>) => void;
  /** Exam mode moved on to the next part: show its questions. */
  onPartChange: (part: TPart) => void;
  /** Exam mode, full test: the check time is over. */
  onFinished: () => void;
};

export type TAudioPhase = "idle" | "playing" | "paused" | "cue" | "ended" | "check";

/**
 * The listening player's audio (Epic 10 §6). Practice mode: pause, −5 s, seek, speed, optional reading pauses;
 * the audio follows the part on screen. Exam mode: plays once with no pause or seek, stops at every cue for
 * its reading time, moves on to the next part by itself and, in a full test, ends with a 2-minute check.
 */
export function useListeningAudio({ audioRef, parts, mode, fullTest, currentPartNo, saved, onSave, onPartChange, onFinished }: Options) {
  const exam = mode === "exam";
  const startIdx = Math.max(0, parts.findIndex((p) => p.part_no === (saved?.part ?? currentPartNo)));
  const [partIdx, setPartIdx] = useState(startIdx);
  const [phase, setPhase] = useState<TAudioPhase>("idle");
  const [time, setTime] = useState(saved && parts[startIdx]?.part_no === saved.part ? saved.t : 0);
  const [duration, setDuration] = useState(parts[startIdx]?.audio?.duration ?? 0);
  const [speed, setSpeedState] = useState(1);
  const [keepPauses, setKeepPauses] = useState(false);
  const [cue, setCue] = useState<{ label: string; left: number } | null>(null);
  const [checkLeft, setCheckLeft] = useState<number | null>(null);
  const [error, setError] = useState(false);
  const [started, setStarted] = useState(saved?.started ?? false);
  const cuesDone = useRef(new Set<string>(saved?.cuesDone ?? []));
  const pendingSeek = useRef<number | null>(time > 0 ? time : null);
  const autoPlay = useRef(false);
  const lastSave = useRef(0);
  const part = parts[partIdx];

  const save = useCallback(
    (t: number, force = false) => {
      const now = Date.now();
      if (!force && now - lastSave.current < 3000) return;
      lastSave.current = now;
      onSave({ part: parts[partIdx].part_no, t, cuesDone: [...cuesDone.current], started });
    },
    [onSave, parts, partIdx, started],
  );

  // Practice mode: the audio follows the part the learner opens.
  useEffect(() => {
    if (exam) return;
    const idx = parts.findIndex((p) => p.part_no === currentPartNo);
    if (idx >= 0 && idx !== partIdx) {
      audioRef.current?.pause();
      // eslint-disable-next-line react-hooks/set-state-in-effect -- switching the loaded track follows the learner's part choice
      setPartIdx(idx);
      setTime(0);
      setPhase("idle");
      setCue(null);
      pendingSeek.current = null;
    }
  }, [exam, currentPartNo, parts, partIdx, audioRef]);

  // Reading pauses count down, then the audio continues.
  useEffect(() => {
    if (!cue) return;
    if (cue.left <= 0) {
      // eslint-disable-next-line react-hooks/set-state-in-effect -- the countdown reached zero: leave the pause
      setCue(null);
      setPhase("playing");
      void audioRef.current?.play().catch(() => setPhase("paused"));
      return;
    }
    const id = window.setTimeout(() => setCue((c) => (c ? { ...c, left: c.left - 1 } : c)), 1000);
    return () => window.clearTimeout(id);
  }, [cue, audioRef]);

  // The final check of a full test in exam mode, then submit.
  useEffect(() => {
    if (checkLeft === null) return;
    if (checkLeft <= 0) {
      onFinished();
      return;
    }
    const id = window.setTimeout(() => setCheckLeft((c) => (c === null ? c : c - 1)), 1000);
    return () => window.clearTimeout(id);
  }, [checkLeft, onFinished]);

  const handlers = {
    onLoadedMetadata: () => {
      const a = audioRef.current;
      if (!a) return;
      setDuration(Number.isFinite(a.duration) ? a.duration : (part.audio?.duration ?? 0));
      a.playbackRate = exam ? 1 : speed;
      if (pendingSeek.current !== null) {
        a.currentTime = pendingSeek.current;
        pendingSeek.current = null;
      }
      if (autoPlay.current) {
        autoPlay.current = false;
        void a.play().then(() => setPhase("playing")).catch(() => setPhase("paused"));
      }
    },
    onTimeUpdate: () => {
      const a = audioRef.current;
      if (!a) return;
      const t = a.currentTime;
      setTime(t);
      save(t);
      if (!(exam || keepPauses) || a.paused) return;
      const cues = part.audio?.cues ?? [];
      const i = cues.findIndex((c, idx) => c.at <= t + 0.1 && c.at > t - 2 && !cuesDone.current.has(`${part.part_no}:${idx}`));
      if (i >= 0) {
        cuesDone.current.add(`${part.part_no}:${i}`);
        a.pause();
        setPhase("cue");
        setCue({ label: cues[i].label, left: cues[i].seconds });
        save(t, true);
      }
    },
    onEnded: () => {
      if (exam && partIdx < parts.length - 1) {
        const next = partIdx + 1;
        setPartIdx(next);
        setTime(0);
        setDuration(parts[next].audio?.duration ?? 0);
        autoPlay.current = true;
        onPartChange(parts[next]);
        onSave({ part: parts[next].part_no, t: 0, cuesDone: [...cuesDone.current], started: true });
        return;
      }
      setPhase("ended");
      save(audioRef.current?.currentTime ?? 0, true);
      if (exam && fullTest) setCheckLeft(CHECK_SECONDS);
    },
    onError: () => setError(true),
    onPlay: () => setPhase((p) => (p === "cue" ? p : "playing")),
    onPause: () => setPhase((p) => (p === "playing" ? "paused" : p)),
  };

  const play = () => {
    const a = audioRef.current;
    if (!a || (exam && started && phase !== "idle")) return;
    setError(false);
    if (exam && !started) setStarted(true);
    void a.play().catch(() => setError(true));
    save(a.currentTime, true);
  };
  const pause = () => {
    if (!exam) audioRef.current?.pause();
  };
  const seek = (t: number) => {
    const a = audioRef.current;
    if (exam || !a) return;
    a.currentTime = Math.max(0, Math.min(t, duration || a.duration || 0));
    setTime(a.currentTime);
  };
  const setSpeed = (s: number) => {
    setSpeedState(s);
    if (audioRef.current && !exam) audioRef.current.playbackRate = s;
  };
  const skipCue = () => {
    if (!exam) setCue((c) => (c ? { ...c, left: 0 } : c));
  };
  const retry = () => {
    setError(false);
    audioRef.current?.load();
  };

  return {
    part,
    partIdx,
    phase,
    time,
    duration,
    speed,
    setSpeed,
    keepPauses,
    setKeepPauses,
    cue,
    skipCue,
    checkLeft,
    error,
    retry,
    started,
    // An exam-mode run that was interrupted (reload) resumes with one press; it cannot start over.
    canPlay: !exam || phase === "idle" || phase === "paused",
    play,
    pause,
    seek,
    handlers,
  };
}
