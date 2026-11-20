#!/usr/bin/env python3
"""Builds the audio of IELTS listening parts from their seed files (Epic 10 §8.3).

    python3 scripts/exam-audio/build.py be/seeds/exam/s1/test1/listening-1.json [more files...]

For each part:
  1. every transcript line is rendered by its speaker's voice (the part's "voices" map), with Kokoro-82M
     ("kokoro:bf_emma") or MeloTTS ("melo:EN-AU"); clips are cached by voice and text in cache/, so editing
     one line re-renders only that line;
  2. the clips are joined with natural pauses into one 24 kHz mono track, encoded as 64 kbps MP3 and saved
     as out/<sha256>.mp3;
  3. the seed file is rewritten with each line's start time, the audio key (exam/media/<sha256>.mp3), its
     duration and the time of every cue (cues name the line they follow with "after").

Then upload with `cd be && go run ./cmd exam-media ../scripts/exam-audio/out/<sha256>.mp3` and run the seed
checks (`go test ./app/modules/exam/...`). Needs ./setup.sh once. Standard library only (plus ffmpeg).
"""

import hashlib
import json
import os
import shutil
import subprocess
import sys
import tempfile
import wave
from pathlib import Path

HERE = Path(__file__).resolve().parent
CACHE = HERE / "cache"
OUT = HERE / "out"
RATE = 24000

# Pauses in seconds: a change of speaker, the same speaker going on, after narration, around a cue.
GAP_SPEAKER = 0.55
GAP_SAME = 0.3
GAP_NARRATOR = 0.9
GAP_CUE = 1.5
LEAD_IN = 0.6
TAIL = 1.0

ENGINES = {
    "kokoro": (HERE / ".venv-kokoro/bin/python", HERE / "kokoro_worker.py"),
    "melo": (HERE / ".venv-melo/bin/python", HERE / "melo_worker.py"),
}


def parse_voice(spec: str):
    """"kokoro:bf_emma@0.95" → ("kokoro", "bf_emma", 0.95)."""
    engine, _, rest = spec.partition(":")
    voice, _, speed = rest.partition("@")
    if engine not in ENGINES or not voice:
        raise SystemExit(f"bad voice {spec!r}: use kokoro:<voice> or melo:<EN-US|EN-BR|EN-AU|EN_INDIA>")
    return engine, voice, float(speed or 1.0)


def clip_path(engine: str, voice: str, speed: float, text: str) -> Path:
    h = hashlib.sha256(f"{engine}|{voice}|{speed}|{text}".encode()).hexdigest()[:24]
    return CACHE / f"{h}.wav"


def render(missing: dict):
    """missing: engine → [job]. Runs each engine's worker once over all its lines."""
    for engine, jobs in missing.items():
        if not jobs:
            continue
        py, worker = ENGINES[engine]
        if not py.exists():
            raise SystemExit(f"{engine} is not set up: run scripts/exam-audio/setup.sh")
        with tempfile.NamedTemporaryFile("w", suffix=".json", delete=False) as f:
            json.dump(jobs, f)
            job_file = f.name
        try:
            subprocess.run([str(py), "-I", str(worker), job_file], check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        except subprocess.CalledProcessError:
            # Re-run with output shown so the error is visible.
            subprocess.run([str(py), "-I", str(worker), job_file], check=True)
        finally:
            os.unlink(job_file)


def to_pcm(path: Path) -> bytes:
    """Reads a clip as 24 kHz mono 16-bit frames, resampling with ffmpeg when needed (MeloTTS is 44.1 kHz)."""
    with wave.open(str(path)) as w:
        if w.getframerate() == RATE and w.getnchannels() == 1 and w.getsampwidth() == 2:
            return w.readframes(w.getnframes())
    fixed = path.with_suffix(".24k.wav")
    if not fixed.exists():
        subprocess.run(["ffmpeg", "-loglevel", "error", "-y", "-i", str(path), "-ac", "1", "-ar", str(RATE), "-sample_fmt", "s16", str(fixed)], check=True)
    with wave.open(str(fixed)) as w:
        return w.readframes(w.getnframes())


def silence(seconds: float) -> bytes:
    return b"\x00\x00" * int(RATE * seconds)


def build(seed_path: Path):
    data = json.loads(seed_path.read_text())
    lines = data.get("lines") or []
    voices = data.get("voices") or {}
    if not lines:
        raise SystemExit(f"{seed_path}: no transcript lines")
    audio = data.get("audio") or {}
    cues = audio.get("cues") or []
    cue_after = {c["after"] for c in cues if c.get("after")}

    plan, missing = [], {e: [] for e in ENGINES}
    for line in lines:
        spec = voices.get(line.get("speaker") or "NARRATOR")
        if not spec:
            raise SystemExit(f"{seed_path}: no voice for speaker {line.get('speaker')!r} (add it to \"voices\")")
        engine, voice, speed = parse_voice(spec)
        text = (line.get("say") or line["text"]).strip()
        path = clip_path(engine, voice, speed, text)
        if not path.exists() and not any(j["out"] == str(path) for j in missing[engine]):
            missing[engine].append({"text": text, "voice": voice, "speed": speed, "out": str(path)})
        plan.append((line, path))
    CACHE.mkdir(exist_ok=True)
    todo = sum(len(j) for j in missing.values())
    if todo:
        print(f"  rendering {todo} new line(s)…", flush=True)
    render(missing)

    pcm = bytearray(silence(LEAD_IN))
    ends = {}
    for i, (line, path) in enumerate(plan):
        line["at"] = round(len(pcm) / 2 / RATE, 2)
        pcm += to_pcm(path)
        ends[line["id"]] = round(len(pcm) / 2 / RATE, 2)
        nxt = plan[i + 1][0] if i + 1 < len(plan) else None
        if line["id"] in cue_after:
            gap = GAP_CUE
        elif not line.get("speaker") or line.get("speaker") == "NARRATOR":
            gap = GAP_NARRATOR
        elif nxt is not None and nxt.get("speaker") == line.get("speaker"):
            gap = GAP_SAME
        else:
            gap = GAP_SPEAKER
        pcm += silence(gap if nxt is not None else TAIL)
    duration = round(len(pcm) / 2 / RATE, 2)

    with tempfile.TemporaryDirectory() as tmp:
        wav_path, mp3_path = Path(tmp) / "part.wav", Path(tmp) / "part.mp3"
        with wave.open(str(wav_path), "wb") as w:
            w.setnchannels(1)
            w.setsampwidth(2)
            w.setframerate(RATE)
            w.writeframes(bytes(pcm))
        subprocess.run(["ffmpeg", "-loglevel", "error", "-y", "-i", str(wav_path), "-ac", "1", "-b:a", "64k", "-map_metadata", "-1", str(mp3_path)], check=True)
        sha = hashlib.sha256(mp3_path.read_bytes()).hexdigest()
        OUT.mkdir(exist_ok=True)
        final = OUT / f"{sha}.mp3"
        shutil.copyfile(mp3_path, final)

    for c in cues:
        if c.get("after"):
            if c["after"] not in ends:
                raise SystemExit(f"{seed_path}: cue follows unknown line {c['after']!r}")
            c["at"] = ends[c["after"]]
    data["audio"] = {"key": f"exam/media/{sha}.mp3", "duration": duration, "cues": cues}
    seed_path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n")
    size = final.stat().st_size
    print(f"  {seed_path.name}: {duration:.0f}s, {size / 1e6:.1f} MB → {final.relative_to(HERE.parent.parent)}")
    return final


def main():
    if len(sys.argv) < 2:
        raise SystemExit(__doc__)
    built = []
    for arg in sys.argv[1:]:
        print(arg, flush=True)
        built.append(build(Path(arg)))
    print("\nUpload:\n  cd be && go run ./cmd exam-media " + " ".join(f"../{p.relative_to(HERE.parent.parent)}" for p in built))


if __name__ == "__main__":
    main()
