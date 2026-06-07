FROM oven/bun:1.2-alpine AS deps

WORKDIR /app

COPY package.json bun.lock* ./
RUN bun install --frozen-lockfile


FROM oven/bun:1.2-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production

COPY --from=deps /app/node_modules ./node_modules
COPY package.json bun.lock* ./
COPY tsconfig.json ./
COPY src ./src
COPY scripts ./scripts
COPY domains ./domains

EXPOSE 3000

CMD ["bun", "run", "src/main.ts"]
