"""Renders lines with MeloTTS (English accents: EN-US, EN-BR, EN-AU, EN_INDIA). Same job format as kokoro_worker."""
import json
import sys

from melo.api import TTS

jobs = json.load(open(sys.argv[1]))
model = TTS(language="EN", device="cpu")
speakers = model.hps.data.spk2id
for job in jobs:
    model.tts_to_file(job["text"], speakers[job["voice"]], job["out"], speed=job.get("speed", 1.0), quiet=True)
    print("melo", job["voice"], job["out"], flush=True)
