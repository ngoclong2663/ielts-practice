#!/usr/bin/env bash
# One-time setup for the listening audio build (Epic 10 §8.3). Free, local, offline after the first run.
#   Kokoro-82M (Apache 2.0): American and British voices.  MeloTTS (MIT): the Australian voice.
# Needs Homebrew ffmpeg and espeak-ng (brew install ffmpeg espeak-ng) and uv.
set -euo pipefail
cd "$(dirname "$0")"
command -v ffmpeg >/dev/null || { echo "ffmpeg missing: brew install ffmpeg"; exit 1; }
[ -f /opt/homebrew/lib/libespeak-ng.dylib ] || { echo "espeak-ng missing: brew install espeak-ng"; exit 1; }
command -v uv >/dev/null || { echo "uv missing: https://docs.astral.sh/uv/"; exit 1; }

if [ ! -x .venv-kokoro/bin/python ]; then
  uv venv -q --python 3.12 .venv-kokoro
  uv pip install -q --python .venv-kokoro/bin/python "kokoro>=0.9.4" soundfile numpy
fi
# misaki (Kokoro's G2P) needs spaCy's small English model; installed here because its own download step
# expects pip inside the venv.
if ! .venv-kokoro/bin/python -I -c "import en_core_web_sm" 2>/dev/null; then
  uv pip install -q --python .venv-kokoro/bin/python \
    "en_core_web_sm @ https://github.com/explosion/spacy-models/releases/download/en_core_web_sm-3.8.0/en_core_web_sm-3.8.0-py3-none-any.whl"
fi
if [ ! -x .venv-melo/bin/python ]; then
  uv venv -q --python 3.11 .venv-melo
  uv pip install -q --python .venv-melo/bin/python "git+https://github.com/myshell-ai/MeloTTS.git"
  # MeloTTS imports old librosa (needs pkg_resources) and a Japanese dictionary even for English:
  # the small unidic-lite replaces the 500 MB unidic download.
  uv pip install -q --python .venv-melo/bin/python "setuptools<70" unidic-lite
  uv pip uninstall -q --python .venv-melo/bin/python unidic || true
fi
echo "Ready. Build a part with: python3 scripts/exam-audio/build.py be/seeds/exam/<set>/test<N>/listening-<P>.json"
