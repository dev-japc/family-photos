# ---dependencies ---
FROM node:24-alpine AS deps
WORKDIR /app
COPY package*.json ./
RUN npm ci

# --- build ---
FROM node:24-alpine AS builder
WORKDIR /app
ENV NODE_ENV=production
COPY --from=deps /app/node_modules ./node_modules
COPY . .

RUN npm run build

# --- production / development ---
FROM node:24-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production
COPY --from=builder /app/dist ./dist
COPY package*.json ./
RUN npm ci --only=production --ignore-scripts

ENV HOST=0.0.0.0
ENV PORT=4321
EXPOSE ${PORT}

CMD [ "node", "dist/server/entry.mjs" ]
