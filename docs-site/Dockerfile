# Build the Hugo docs site with a pinned Hugo version, then serve it with nginx.
FROM debian:bookworm-slim AS build
ARG HUGO_VERSION=0.164.0
RUN apt-get update \
 && apt-get install -y --no-install-recommends curl ca-certificates \
 && arch="$(dpkg --print-architecture)" \
 && curl -sL "https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/hugo_extended_${HUGO_VERSION}_linux-${arch}.tar.gz" -o /tmp/hugo.tgz \
 && tar -xzf /tmp/hugo.tgz -C /usr/local/bin hugo \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY . .
RUN hugo --minify --gc --destination /public --baseURL http://localhost:1313/

FROM nginx:alpine
COPY --from=build /public /usr/share/nginx/html
EXPOSE 80
