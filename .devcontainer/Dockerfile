# https://github.com/devcontainers/features/issues/1543

FROM mcr.microsoft.com/devcontainers/go:2-1.25-bookworm
USER root
RUN rm -f /etc/apt/sources.list.d/yarn.list \
    && rm -f /usr/share/keyrings/yarn.gpg \
    && apt-get update
