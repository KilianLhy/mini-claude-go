#!/bin/sh
# Waits for the Ollama container to be ready, then pulls the model(s) listed in
# OLLAMA_MODELS (comma-separated). Run once by the ollama-init compose service.
set -e

OLLAMA_HOST="http://ollama:11434"
MODELS="${OLLAMA_MODELS:-llama3.2:3b}"

echo "Waiting for Ollama at ${OLLAMA_HOST}..."
until curl -sf "${OLLAMA_HOST}/api/tags" >/dev/null 2>&1; do
  sleep 2
done
echo "Ollama is up."

# OLLAMA_MODELS may contain several models separated by commas.
echo "${MODELS}" | tr ',' '\n' | while read -r model; do
  model="$(echo "${model}" | xargs)" # trim whitespace
  [ -z "${model}" ] && continue
  echo "Pulling model: ${model}"
  curl -sf "${OLLAMA_HOST}/api/pull" -d "{\"name\":\"${model}\"}" >/dev/null
  echo "  done: ${model}"
done

echo "All models ready."
