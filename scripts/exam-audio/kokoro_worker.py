"""Renders lines with Kokoro-82M. Input: a JSON job file [{"text", "voice", "speed", "out"}]; writes 24 kHz WAVs."""
import json
import sys

import espeakng_loader

# The wheel's bundled espeak-ng has a broken data path on macOS; use Homebrew's (brew install espeak-ng).
espeakng_loader.get_library_path = lambda: "/opt/homebrew/lib/libespeak-ng.dylib"
espeakng_loader.get_data_path = lambda: "/opt/homebrew/share/espeak-ng-data"

import numpy as np  # noqa: E402
import soundfile as sf  # noqa: E402
from kokoro import KPipeline  # noqa: E402

jobs = json.load(open(sys.argv[1]))
pipes = {}
for job in jobs:
    voice = job["voice"]  # e.g. bf_emma: first letter = accent (a American, b British)
    lang = voice[0]
    pipe = pipes.get(lang) or pipes.setdefault(lang, KPipeline(lang_code=lang, repo_id="hexgrad/Kokoro-82M"))
    chunks = [a.numpy() if hasattr(a, "numpy") else a for _, _, a in pipe(job["text"], voice=voice, speed=job.get("speed", 1.0))]
    sf.write(job["out"], np.concatenate(chunks), 24000, subtype="PCM_16")
    print("kokoro", voice, job["out"], flush=True)
