# Stage 1: Build
FROM node:22-alpine AS builder
 
WORKDIR /app
 
RUN npm install -g pnpm
 
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
 
COPY . .
RUN pnpm run build
 
# Stage 2: Serve
FROM nginx:1.30-alpine-slim
 
COPY --from=builder /app/dist /usr/share/nginx/html
 COPY nginx.conf /etc/nginx/conf.d/default.conf
 
EXPOSE 80
 
