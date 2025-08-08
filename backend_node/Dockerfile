# Stage 1: Install dependencies (include devDeps)
FROM node:20-alpine AS deps
WORKDIR /app
COPY package*.json ./
RUN npm install --include=dev

# Stage 2: Copy source & run
FROM node:20-alpine
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

EXPOSE 4000
CMD ["npm", "run", "dev"]
