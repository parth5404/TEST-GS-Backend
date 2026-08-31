# PRODUCTION image.
#
# Purani file mein do badi galtiyan thin:
#   1. `npm install --include=dev` — devDependencies (nodemon waghera) production
#      image mein chali jaati thin. Image bada, aur attack surface bada.
#   2. `CMD ["npm", "run", "dev"]` — production image nodemon chala rahi thi.
#      nodemon ek file-watcher hai; wo production ke liye bana hi nahi hai.
#
# Local development ke liye Dockerfile.dev use hota hai (compose usi ko build karta hai).

# ── Stage 1: dependencies ────────────────────────────────────────────────────
# "Multi-stage build" — kai FROM lines. Har FROM ek naya stage shuru karta hai.
# Aakhri stage hi final image banta hai; baaki stages sirf cheezein banane ke
# liye hote hain. Isse build tools final image mein nahi jaate.
FROM node:20-alpine AS deps

WORKDIR /app

# Sirf package files pehle copy kar rahe hain, pura source baad mein.
# Ye jaan-boojh ke hai — Docker har line ko ek "layer" mein cache karta hai,
# aur layer tabhi rebuild hoti hai jab uska input badle.
#
# Agar hum `COPY . .` pehle karte, toh kisi bhi controller file badalne pe
# npm ci dobara chalta (2-3 min). Sirf package files copy karne se npm ci ki
# layer tabhi rebuild hoti hai jab dependencies actually badlein.
COPY package.json package-lock.json ./

# `npm ci` vs `npm install`:
#   • ci  → lockfile ko exactly follow karta hai, reproducible, CI ke liye
#   • install → lockfile ko update kar sakta hai
# --omit=dev → devDependencies skip. Production ko nodemon ki zarurat nahi.
RUN npm ci --omit=dev

# ── Stage 2: runtime ─────────────────────────────────────────────────────────
FROM node:20-alpine

# `dumb-init` — ek chhota init process. Docker mein tera app PID 1 banta hai,
# aur PID 1 default signal handling nahi karta. Iske bina SIGTERM ignore ho
# jata hai aur `docker stop` 10 second baad zabardasti kill karta hai.
# Day 7 pe jab graceful shutdown likhega, ye line usko kaam karne degi.
RUN apk add --no-cache dumb-init

ENV NODE_ENV=production

WORKDIR /app

# node:alpine image mein ek non-root user `node` pehle se hota hai.
# Root ke bajaye usse chalana basic security hygiene hai — agar app compromise
# ho toh attacker ko root nahi milta.
COPY --chown=node:node --from=deps /app/node_modules ./node_modules
COPY --chown=node:node . .

USER node

EXPOSE 4000

# `node server.js` — nodemon nahi.
# Exec form (JSON array) use kar rahe hain, shell form nahi: isse node seedha
# PID 1 (dumb-init ke neeche) banta hai aur signals theek se milte hain.
ENTRYPOINT ["dumb-init", "--"]
CMD ["node", "server.js"]
