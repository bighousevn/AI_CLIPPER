# Stage 1: Build a production-ready application
FROM node:20-alpine AS builder
WORKDIR /app

COPY FE/package*.json ./
RUN npm install

COPY FE/ .

RUN npm run build

# Stage 2: Create the final, optimized image
FROM node:20-alpine
WORKDIR /app

COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/package.json .

RUN npm install --only=production

EXPOSE 3000

CMD ["npm", "start"]