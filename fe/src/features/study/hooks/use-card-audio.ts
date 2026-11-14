import { useCallback, useEffect, useRef } from "react";
import { useSpeak } from "@/hooks/use-speak";
import type { TCardContent } from "@/types/card-content";

/**
 * Reads words and sentences aloud with the browser's speech (Web Speech API) in the user's accent.
 * Cards carry no audio files since Epic 3; an audio URL is still played if old content has one.
 */
export function useCardAudio(accent: "uk" | "us") {
  const { speak } = useSpeak(accent === "us" ? "en-US" : "en-GB");
  const audioRef = useRef<HTMLAudioElement | null>(null);

  useEffect(() => () => audioRef.current?.pause(), []);

  const play = useCallback(
    (text: string, url?: string) => {
      audioRef.current?.pause();
      if (url) {
        const audio = new Audio(url);
        audioRef.current = audio;
        audio.play().catch(() => speak(text));
        return;
      }
      speak(text);
    },
    [speak],
  );

  const playHeadword = useCallback(
    (content: TCardContent) => play(content.headword, content.audio?.headword?.[accent]),
    [accent, play],
  );

  return { play, playHeadword };
}
