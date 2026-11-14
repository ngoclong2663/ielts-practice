"use client";

import { useCallback, useSyncExternalStore } from "react";

const noopSubscribe = () => () => {};

// Browser text-to-speech fallback (Web Speech API). Server-generated audio replaces it in Epic 3.
export function useSpeak(lang: "en-GB" | "en-US" = "en-GB") {
  // false on the server, real support on the client (no hydration mismatch).
  const supported = useSyncExternalStore(
    noopSubscribe,
    () => "speechSynthesis" in window,
    () => false,
  );

  const speak = useCallback(
    (text: string) => {
      if (!("speechSynthesis" in window)) {
        return;
      }
      window.speechSynthesis.cancel();
      const utterance = new SpeechSynthesisUtterance(text);
      utterance.lang = lang;
      utterance.rate = 0.9;
      const voice = window.speechSynthesis.getVoices().find((v) => v.lang === lang);
      if (voice) {
        utterance.voice = voice;
      }
      window.speechSynthesis.speak(utterance);
    },
    [lang],
  );

  return { speak, supported };
}
