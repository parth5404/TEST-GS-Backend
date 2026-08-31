# Docker — is project ke through samajhna

Ye generic Docker tutorial nahi hai. Ye batata hai ki **is repo mein** Docker kya kar
raha hai aur kyun. Har concept ke saath actual file ka reference hai.

---

## 1. Teen cheezein: image, container, volume

Sabse pehle ye teen alag-alag samajh le. Log yahi gadbad karte hain.

| Cheez | Kya hai | Analogy |
|---|---|---|
| **Image** | Read-only template. Build hoti hai, badalti nahi. | Class |
| **Container** | Chalta hua instance. Delete hone pe iska data gaya. | Object |
| **Volume** | Alag se store, container ke bahar. Container marne pe bhi zinda. | Database file disk pe |

Isliye compose mein Mongo ke liye ye likha hai:

```yaml
volumes:
  - mongo-data:/data/db
```

Bina iske `docker compose down` karte hi tera saara data uda jaata. Volume ke saath
container marta hai par data bacha rehta hai.

Data jaan-boojh ke mitana ho:

```bash
docker compose down -v      # -v matlab volumes bhi hata do
```

---

## 2. Layers aur caching — `COPY package*.json` pehle kyun?

`backend_node/Dockerfile` khol aur dekh:

```dockerfile
COPY package.json package-lock.json ./
RUN npm ci --omit=dev
COPY . .
```

Aisa kyun na kiya:

```dockerfile
COPY . .              # ← sab kuch ek saath
RUN npm ci --omit=dev
```

**Kyunki Dockerfile ki har line ek "layer" banati hai, aur Docker layers cache karta hai.**

Docker har layer ke liye dekhta hai: "iska input pichli baar se badla?" Nahi badla toh
cache se utha leta hai, dobara nahi chalata.

Ab do situations:

**Sahi tareeka (abhi wala):**
```
COPY package.json      → badla? nahi (tune sirf controller edit kiya)  → CACHE ✅
RUN npm ci             → upar wali layer same hai                      → CACHE ✅  (2 min bache)
COPY . .               → badla? haan                                    → rebuild
```

**Galat tareeka:**
```
COPY . .               → badla? haan (controller edit kiya)  → rebuild
RUN npm ci             → upar wali layer badli               → rebuild ❌ (2 min barbaad)
```

Ek line ka farq, par har build pe 2-3 minute ka. Isliye rule ye hai:
**jo cheez kam badalti hai, use pehle rakho.**

---

## 3. Multi-stage build

`backend_node/Dockerfile` mein do `FROM` lines hain. Har `FROM` naya stage shuru karta hai,
aur **sirf aakhri stage final image banta hai.**

```dockerfile
FROM node:20-alpine AS deps     # stage 1 — yahan dependencies install hoti hain
...
FROM node:20-alpine             # stage 2 — asli image
COPY --from=deps /app/node_modules ./node_modules
```

Fayda: build ke waqt jo cheezein chahiye (compilers, dev tools, build cache) wo final
image mein nahi jaatin. Chhoti image = tez deploy, kam attack surface.

Go wale `Dockerfile` mein ye aur saaf dikhta hai — stage 1 mein pura Go toolchain
(~300MB) hai, aur stage 2 sirf alpine + compiled binary (~15MB).

---

## 4. `depends_on` sirf order deta hai, readiness nahi

Ye Docker ka sabse common confusion hai.

```yaml
depends_on:
  - mongo
```

Ye sirf kehta hai: **"mongo container pehle START karo."** Ye ye NAHI kehta ki mongo
connections lene ke liye taiyaar hai. Container 1 second mein start ho jaata hai, par
MongoDB ko taiyaar hone mein 5-10 second lagte hain.

Result: Node start hota hai, Mongo se connect karta hai, `ECONNREFUSED` khaata hai,
`process.exit(1)` kar deta hai. Classic race condition.

Isliye humne healthcheck lagaya hai:

```yaml
depends_on:
  mongo:
    condition: service_healthy    # ← healthcheck pass hone ka intezaar
```

Ab compose tab tak Node start nahi karega jab tak `mongo` ka healthcheck green na ho.

Hamara healthcheck do kaam ek saath karta hai — batata hai Mongo zinda hai, **aur**
pehli baar replica set initiate bhi kar deta hai:

```javascript
try { rs.status().ok } catch (e) { rs.initiate({...}) }
```

`rs.status()` error deta hai jab tak RS initiate na ho — isliye `catch` mein initiate.
Ek baar ho gaya toh aage har check simply pass hota hai.

---

## 5. Networking — service ka naam hi hostname hai

Compose ek private network banata hai jisme **har service ka naam ek hostname ban jaata hai.**

Isliye Node ka config `http://go-mail:8080` hai, `http://localhost:8080` nahi.

**`localhost` kyun nahi chalega:** har container ka apna alag network namespace hota hai.
Node container ke andar `localhost` ka matlab "Node container khud" hai — Go nahi.

```
┌─ app-network ──────────────────────────────────┐
│                                                │
│  node-app  ──http://go-mail:8080──▶  go-mail   │
│      │                                  │      │
│      └──mongodb://mongo:27017──▶  mongo ◀┘     │
│                                                │
└────────────────────────────────────────────────┘
         │
         │  ports: "4000:4000"
         ▼
   tera laptop (localhost:4000)
```

`ports:` sirf tab chahiye jab **bahar se** (tere laptop se) access karna ho. Container
aapas mein bina `ports:` ke baat kar sakte hain.

Isi wajah se maine `go-mail` se `ports:` **hata diya** hai. Pehle `"8080:8080"` tha,
matlab `/send-email` poori duniya ko khula tha — bina auth ke. Ab Node usse network ke
andar se access karta hai, par bahar se koi nahi pahunch sakta.

---

## 6. Replica set ka gotcha — `directConnection=true`

Ye wo jagah hai jahan log ghante barbaad karte hain, toh dhyan se padh.

Humne replica set ko is host naam se initiate kiya hai:

```javascript
rs.initiate({ _id: 'rs0', members: [{ _id: 0, host: 'mongo:27017' }] })
```

`mongo` compose network **ke andar** resolve hota hai. Container se connect karna ho:

```
mongodb://mongo:27017/gsacademia?replicaSet=rs0        ✅ chalega
```

Par **tere laptop se** `mongo` hostname exist hi nahi karta. Aur MongoDB driver
`replicaSet=rs0` dekhte hi "topology discovery" karta hai — seed se poochta hai
"members kaun hain?", jawab milta hai `mongo:27017`, aur wo resolve nahi hota → timeout.

Solution:

```
mongodb://localhost:27017/gsacademia?directConnection=true   ✅ laptop se
```

`directConnection=true` discovery skip karke seedha usi address pe connect karta hai.

**Yaad rakhne wala rule:** container ke andar se `replicaSet=`, laptop se `directConnection=true`.

---

## 7. Roz ke commands

```bash
# Sab shuru karo (background mein)
docker compose up -d

# Shuru karo aur logs dekho (Ctrl+C se band)
docker compose up

# Kaun chal raha hai, kaun healthy hai
docker compose ps

# Logs — ek service ke, live
docker compose logs -f go-mail

# Aakhri 50 lines
docker compose logs --tail=50 node-app

# Container ke andar shell
docker compose exec node-app sh
docker compose exec mongo mongosh

# Ek service restart
docker compose restart node-app

# Dockerfile ya dependencies badalne pe rebuild
docker compose up -d --build

# Band karo (data bacha rahega)
docker compose down

# Band karo AUR database wipe karo
docker compose down -v
```

---

## 8. Jab kuch toote

**"port is already allocated"**
Us port pe pehle se kuch chal raha hai. Dhundh: `lsof -i :27017` — ya local Mongo band kar.

**Node start hote hi crash, `ECONNREFUSED`**
`docker compose ps` chala. Mongo `healthy` hai ya `starting`? `condition: service_healthy`
lagi hui hai ya nahi check kar.

**Code change kiya par kuch nahi hua**
`node-app` mein bind mount hai, toh JS changes turant dikhne chahiye. Agar nahi dikh rahe:
`docker compose logs -f node-app` mein dekh nodemon restart hua ya nahi. `package.json`
badla ho toh rebuild chahiye: `docker compose up -d --build`.

**Go ka change nahi dikh raha**
Go compiled hai, bind mount nahi hai. Har change pe rebuild: `docker compose up -d --build go-mail`.

**"mongo hostname not found" laptop se**
Section 6 dobara padh. `directConnection=true` lagana bhool gaya.

---

## 9. Exercises — khud haath laga

Har ek 5-10 minute ka. Asli seekh cheezein **todne** se aati hai, padhne se nahi.

**1. Port badal ke dekh**
`docker-compose.yml` mein mongo ka `"27017:27017"` → `"27018:27017"` kar de.
`docker compose up -d` chala. Ab tere laptop se `mongosh mongodb://localhost:27017`
kaam karega ya nahi? Aur container ke andar se Node ka connection? Sochne ke baad
test kar — phir wapas theek kar de.
*(Seekh: `ports:` sirf bahar wali duniya ke liye hai, andar wali baat cheet pe koi asar nahi.)*

**2. Go ke startup logs padh**
```bash
docker compose logs -f go-mail
```
Cron scheduler start hua? Mongo se connect hua? Kaunse endpoints register hue?
Ab `docker compose restart go-mail` kar aur poora startup sequence dobara dekh.

**3. Healthcheck hata ke race condition khud dekh**
`node-app` ke `depends_on` mein `condition: service_healthy` ko `condition: service_started`
kar de. Phir:
```bash
docker compose down -v && docker compose up
```
Node ke logs mein dekh — Mongo taiyaar hone se pehle connect karne ki koshish mein
crash hoga. Ye wahi bug hai jo dher saare tutorials mein chhupa hota hai. Wapas theek kar de.

**4. Layer caching apni aankhon se dekh**
```bash
docker compose build node-app          # pehli baar — pura chalega
docker compose build node-app          # dobara — sab "CACHED" dikhega
```
Ab `backend_node/server.js` mein ek comment add kar aur phir build kar. Kaunsi layers
dobara chalin? Ab `package.json` mein koi dependency add kar aur build kar — ab kya farq hai?

**5. `.dockerignore` ka asar**
`backend_node/.dockerignore` se `.env` line hata de. Build kar, phir andar jhaank:
```bash
docker compose build node-app
docker compose run --rm node-app sh -c "ls -la .env"
```
Teri secrets file image ke andar aa gayi. **Wapas add kar de.** Ye asli production
incident ka pattern hai — log aise hi image registry pe secrets push kar dete hain.

---

## Aage kya

> Poora day-wise roadmap: [`PLAN-14-DAYS.md`](./PLAN-14-DAYS.md)

Day 6 pe do Go worker ek saath chalane honge, tab ye kaam aayega:

```bash
docker compose up -d --scale go-mail=2
```

Tab tu verify karega ki dono worker ek hi outbox row kabhi process na karein.
Note: `--scale` tabhi kaam karta hai jab service pe fixed `ports:` na ho — ek aur
wajah ki `go-mail` se port mapping hata diya gaya.
