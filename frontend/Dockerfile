FROM node:22-alpine AS build
WORKDIR /app
RUN corepack enable
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile --config.package-import-method=copy
COPY . .
RUN pnpm build

FROM node:22-alpine
WORKDIR /app
RUN corepack enable
COPY --from=build /app/build ./build
COPY --from=build /app/package.json ./
COPY --from=build /app/pnpm-lock.yaml ./
RUN pnpm install --prod --frozen-lockfile --config.package-import-method=copy
ENV PORT=3000
EXPOSE 3000
CMD ["node", "build"]
