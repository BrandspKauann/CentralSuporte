# The backend source is unchanged from v2.8.0; frontend customizations build separately.
FROM vikunja/vikunja:2.8.0 AS upstream
FROM node:24-alpine
RUN apk add --no-cache ca-certificates tzdata
COPY --from=upstream /app/vikunja/vikunja /usr/local/bin/vikunja
WORKDIR /opt/hirayama
COPY config.mjs start.mjs ./
COPY LICENSE ./LICENSE
LABEL org.opencontainers.image.licenses="AGPL-3.0-or-later"
ENV NODE_ENV=production
USER node
EXPOSE 80
CMD ["node", "start.mjs"]
