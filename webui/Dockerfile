FROM node:slim AS builder
ARG TARGETPLATFORM

WORKDIR /app
COPY package.json yarn.lock .yarnrc ./
RUN yarn install
COPY . .
RUN yarn run build

FROM node:slim

WORKDIR /app

COPY package.json yarn.lock .yarnrc ./
RUN yarn install --production

COPY --from=builder /app/build /app/build

ENV BODY_SIZE_LIMIT 0
ENV HOST 0.0.0.0
ENV PORT 5173

CMD ["node", "build"]