# 14-Day Rebuild — day-wise plan

Roz **2 ghante, hard stop.** Din over ho jaye toh agle din carry kar — sequence
important hai, calendar nahi.

Ye plan asli architectural defects theek karke sikhata hai. Koi toy exercise nahi —
har din is repo mein ek asli problem band hoti hai.

---

## Kyun — 8 defects

| # | Defect | Kab band hoga |
|---|---|---|
| 1 | Email request path ke andar bhejta hai (sync SMTP) | Day 5–6 |
| 2 | OTP: DB loop, koi index nahi, plaintext, consume nahi hota | Day 2–3 |
| 3 | Go seedha `users` collection padhta hai (shared schema) | Day 15+ |
| 4 | Email service transport + scheduler dono hai → replicate nahi ho sakti | Day 15+ |
| 5 | Koi transaction nahi — paisa aur enrollment atomic nahi | Day 10–12 |
| 6 | Bidirectional arrays, unbounded growth, drift | Day 8–9 |
| 7 | Student dashboard pe N+1 queries | Day 13 |
| 8 | Koi layering nahi — Express business logic se juda hua | Day 4, 13 |

**Week 1** = kaam ko request se bahar nikalna.
**Week 2** = data model aur paisa theek karna.

---

## Day 0 — ✅ HO CHUKA HAI

Ye scaffolding push ho chuki hai (5 commits). Day 1 shuru karne se pehle ye chal jaana chahiye:

```bash
cp backend_node/.env.example backend_node/.env
cp go_email/.env.example go_email/.env
docker compose up -d --build
docker compose ps        # teeno healthy?
```

Kya-kya set hua:

- **Mongo service** `--replSet rs0` ke saath — pehle repo mein Mongo tha hi nahi.
  Replica set Day 12 (transactions) aur Day 15 (change streams) dono ke liye zaruri hai.
- **`image:` → `build:`** — pehle compose Docker Hub se prebuilt image kheenchta tha,
  matlab tera local code chalta hi nahi tha.
- **Production Dockerfile** ab `node server.js` chalati hai, nodemon nahi.
- **`go-mail` se published port hata diya** — pehle `/send-email` bina auth ke
  duniya ko khula tha.
- **ESLint + Prettier + husky** ratchet ke saath (neeche detail mein).
- **GitHub Actions CI** — Node lint/test + Go vet/build/fmt.
- **Guides** — [`DOCKER.md`](./DOCKER.md) aur [`CI.md`](./CI.md), exercises ke saath.

> **Abhi tak nahi hua:** live Render deployment band karna. Payment bypass aur mail
> relay dono abhi production mein reachable hain. Day 11 tak deployment band rakh,
> ya kam se kam Gmail ka app-password rotate kar de.

---

## Standing practices — roz, alag task nahi

Ye alag se time nahi maangte. Sirf dhyaan maangte hain.

- **Ek logical change = ek commit**, imperative mein.
  `feat(otp): replace collision loop with crypto.randomInt` — na ki `prod changes`.
  Purana log public hai; agle 30 commits alag kahani sunayenge.
- **`git add -p` se stage kar.** Har line padhne pe majboor karta hai. Ye ek habit
  kisi bhi linter se zyada pakadti hai.
- **Roz ek branch, PR kholo, apna diff khud review karo, phir merge.**
  Solo PR teen din tak bewakoofi lagti hai. Uske baad diff view aur "kyun" likhne
  ki jagah milti hai.
- **Delete karo, comment out mat karo.** Abhi ~250 line commented code padi hai —
  `PaymentC.js` mein pura dead Razorpay webhook. Git sab yaad rakhta hai.
- **Har fix ke saath ek test** jo fix ke bina fail ho (Day 7 ke baad, jab harness aa jayega).

---

# Week 1 — Day 1 se 7

## Day 1 — Pehle naapo

**Seekh:** Query plans, aur baseline ka discipline. Aage ka har claim performance ya
correctness ke baare mein hai — aaj tu wo instrument banayega jo un claims ko judge karega.

**Banao:**
- `scripts/seed.js` — 5,000 users, 2,000 OTPs, 200 courses, enrollments.
  Batched `insertMany` use kar.
- `mongoose.set('debug', true)` ek env flag ke peeche — isse per-request query count dikhega.
- Do timing script: `POST /api/v1/auth/sendotp` aur `GET /api/v1/users/getenrolledcourses`,
  20-20 baar, p50 aur p95 print karo.
- `mongosh` mein:
  ```javascript
  db.otps.find({ otp: "123456" }).explain("executionStats")
  ```
  `totalDocsExamined` aur winning stage note kar.

**Done jab:** `BASELINE.md` mein dono endpoints ke p50/p95 aur `COLLSCAN` ka proof paste ho.

> Ye din skip karne ka mann karega. Mat karna — iske bina Day 3 aur Day 13 ka har
> improvement sirf andaaza rahega.

---

## Day 2 — Indexes, theek se

**Seekh:** B-tree indexes, selectivity, compound prefix rule, TTL indexes, covered queries.

Abhi `models/OTP.js` pe **ek bhi index nahi** hai, sirf `createdAt` pe TTL. Isliye
kal ka explain COLLSCAN dikha raha tha.

**Banao:**
- Naya `models/Otp.js` (purane ko abhi chhod de): `email` (unique), `otpHash`,
  `expiresAt`, `attempts`.
- `expiresAt` pe TTL index, `expireAfterSeconds: 0`. Sochh ki ye `createdAt` pe
  `expires: '10m'` se kaise alag hai — aur kaunsa tujhe per-document lifetime deta hai.
- Baaki missing indexes: `CourseProgress` pe `(userId, courseId)` unique,
  `Review` pe `(user, course)` unique.

**Ratchet se ek line hatao:** `controllers/CourseProgressC.js` mein `ErrorResponse`
ka `require` add kar. Phir `eslint.config.js` ke legacy block se us file ki line
hata de. Ye asli bug hai — har error path pe `ReferenceError` aata hai.

> **Trap:** Tere `User` schema mein `unique: [true, 'Enter a valid email']` likha hai.
> **Wo message kabhi use nahi hota.** `unique` ek index directive hai, validator nahi.
> Duplicate email daal ke dekh asli error kya aata hai — phir decide kar usse kahan catch karna hai.

**Done jab:** Codebase ki har query ya toh index use karti ho, ya tune likh rakha ho ki kyun nahi chahiye.

---

## Day 3 — OTP flow dobara likho

**Seekh:** Entropy vs uniqueness, short-lived secrets ko hash karna, timing-safe
comparison, idempotent upsert.

Abhi ka loop DB se globally unique code maangta hai:

```javascript
do {
  otp = otpGenerator.generate(6, options);
  user = await OTP.findOne({ otp });
} while (user);
```

Par chahiye ye tha ki code **ek email ke liye unpredictable** ho — aur ye koi query
bata hi nahi sakti. Global uniqueness ulta entropy kam karti hai.

**Banao:**
- `crypto.randomInt(0, 1000000)`, zero-padded. `do/while` poora hata de.
- `sha256(code)` store kar; `crypto.timingSafeEqual` se compare kar.
- Issue: `findOneAndUpdate({ email }, ..., { upsert: true })` — ek email pe ek hi live code.
- Verify: ek indexed read, `$inc: { attempts: 1 }`, 5 ke baad reject, success pe delete.

**Ratchet se ek line hatao:** `models/User.js` ka email regex (`no-useless-escape`).

**Done jab:** Timing script Day 1 se saaf kam p95 dikhaye, aur successful signup ke
baad OTP row DB se gayab ho.

> `NOTES.md` mein likh: bcrypt yahan galat kyun hai par passwords ke liye sahi kyun hai?

---

## Day 4 — Pehla service extract karo

**Seekh:** Layering, side effects kahan hone chahiye, dependency direction.

Abhi OTP email ek Mongoose `post('save')` hook se jaati hai — data layer network I/O
kar raha hai. Koi bhi seed script ya test jo OTP banaye, chupke se asli mail bhej deta hai.

**Banao:**
- `services/otpService.js` — `issueOtp(email)` aur `verifyOtp(email, code)`.
  Plain async functions. `req`, `res`, `next` kahin nahi.
- `AppError` class ek `code` ke saath; controller code ko HTTP status pe map kare.
  Dhyaan de ki service ko ab HTTP ka pata hi nahi.
- `models/OTP.js` ka `post('save')` hook delete kar. Ab service khud mail bhejegi, saamne.
- Ek `config.js` jo `process.env` ko zod schema se boot pe parse kare aur missing pe
  throw kare — abhi 9 files mein alag-alag `dotenv.config()` bikhra hai.
  Missing `JWT_SECRET` startup pe process maar de, na ki pehle login pe 500 de.

**Done jab:** `issueOtp('x@y.com')` Node REPL se chal jaye bina Express chalaye,
aur `.env` se koi required variable hatane pe server boot hi na ho.

---

## Day 5 — Outbox, Node side

**Seekh:** Transactional outbox, at-least-once delivery, idempotency keys.

Aaj se notification transaction ka hissa nahi, uska **nateeja** ban jayega.

**Banao:**
- `models/OutboxEmail.js`: `to`, `template`, `payload`, `status`, `attempts`,
  `availableAt`, `dedupeKey` (unique, sparse), `lastError`.
  Index `(status, availableAt)` pe.
- `services/emailService.js` mein `enqueueEmail()`. Har `await emailSender(...)`
  call site ko isse replace kar.
- `utils/emailSender.js` se axios call poori tarah hata de.

**Done jab:** `POST /api/v1/auth/sendotp` 50ms se kam mein return kare aur outbox mein
`pending` row aa jaye.

> **Aaj koi mail nahi jayegi.** Wo Day 6 ka kaam hai. Ye ek din ka gap hi pattern
> ka asli point hai — ek din isko mehsoos kar.

---

## Day 6 — Outbox, Go side

**Seekh:** Bina message broker ke kaam atomically claim karna, exponential backoff,
dead-lettering, crashed worker ka orphaned kaam wapas lena.

**Banao (Go):**
- Ek worker goroutine, 2 second ka ticker.
- **Claim:** `FindOneAndUpdate` on `{status:"pending", availableAt:{$lte:now}}`,
  set `status:"sending"` + `lockedAt`, sort by `availableAt`.
  **Yahi ek operation pattern ko safe banata hai.**
- Success → `sent`. Fail → `attempts++`, `availableAt = now + 2^attempts` seconds,
  wapas `pending`. 5 attempts ke baad → `failed` with `lastError` — ek queryable
  dead-letter list, jo abhi bilkul nahi hai.
- Ek reaper jo 5 minute se `sending` mein atki rows wapas `pending` kare.

**Done jab:**
```bash
docker compose up -d --scale go-mail=2
```
Do worker ek outbox pe chalein aur koi row do baar na jaye. **Yahi test asli seekh hai.**

---

## Day 7 — Jaan-boojh ke todo, aur safety net banao

**Seekh:** Failure injection, asli DB ke against integration testing, structured
logging — kyunki jo failure tu dekh hi nahi sakta, usse inject bhi nahi kar sakta.

**Banao:**
- `console.log` aur `clgDev` ko `pino` se replace kar. Har request ko ek ID do aur
  usse Go service ko header mein bhejo — taaki ek signup dono process mein trace ho.
  Error handler ko **error log karna** sikhao.

  > Abhi wo kuch bhi log nahi karta, aur `npm start` `NODE_ENV=production` set karta hai
  > jo `morgan` aur `clgDev` dono band kar deta hai. **Production ka 500 kahin koi
  > nishaan nahi chhodta.**

- `mongodb-memory-server` + `supertest`. Do test se shuru: OTP issue karne pe exactly
  ek outbox row bane; 5 galat code pe address lock ho jaye.
- CI workflow mein test step already hai — ab wo asli tests chalayega.
- Node mein graceful shutdown: `SIGTERM` pe naye connection lena band, in-flight
  requests drain, Mongo close. **Tera Go service ye already sahi karta hai** —
  `go_email/main.go` padh aur wahi shape copy kar.
- Ab todo: Go service band kar → signup phir bhi 201 de. Wapas chalu kar → rows drain hon.
  `MAIL_HOST` ko kahin galat point kar → backoff curve aur `failed` dekh.

**Done jab:** Tu evidence ke saath bata sake ki Gmail down hone pe signup ka kya hota hai,
aur ek request ID Node ke log se Go ke log tak follow kar sake.

> Day 1 pe iska imaandaar jawab tha: *"500, aur account bana ya nahi pata nahi."*

---

# Week 2 — Day 8 se 14

## Day 8 — Relationship ko theek se model karo

**Seekh:** Join collection vs embedded array, 16MB document limit, unbounded array
growth ka har parent read pe asar.

**Banao:**
- `models/Enrollment.js` — `userId`, `courseId`, `enrolledAt`.
  Unique compound index `(userId, courseId)` pe, secondary `courseId` pe.
- Ek migration jo har `Course.studentsEnrolled[]` aur `User.courses[]` ko
  Enrollment rows mein badle.

  > **Dono match nahi karenge.** `controllers/CourseC.js` ka `deleteCourse` kabhi
  > cleanup karta hi nahi. Har discrepancy log kar — chupke se koi ek mat chun.

**Done jab:** Migration idempotent ho (do baar chalao, same row count), aur tere paas
likha hua count ho ki asli data mein dono arrays kitne drift kar chuke the.

---

## Day 9 — Reads migrate karo, bina downtime

**Seekh:** Expand aur contract (parallel change pattern). Backfill → dual-write →
verify → reads cut over → **phir** purana shape hatao.

> Zyadatar log ye kabhi nahi seekhte, kyunki tutorials hamesha khaali database maante hain.

**Banao:**
- Arrays **aur** Enrollment dono mein likhte raho.
- Reads shift karo: `getEnrolledCourses`, `createOrder` ka already-enrolled check,
  aur `getEnrolledCourseData`.
- Ek drift-check script jo dono representations compare kare.

**Done jab:** Har read path Enrollment use kare, writes abhi bhi dono jagah jayein,
aur drift zero report kare.

> **Arrays abhi delete mat karna.** Contract step ko aakhir mein rakhna hi poora discipline hai.

---

## Day 10 — Paise ko model karo

**Seekh:** State machines, server-authoritative state, aur amounts integer kyun hote hain.

Abhi tera payment state poori tarah Razorpay + client ke request body mein hai.

**Banao:**
- `models/Order.js`: `razorpayOrderId` (unique), `userId`, `courses[]`, `amountPaise`,
  `status` (`created | paid | fulfilled | failed`), timestamps.
- `createOrder` client ko kuch return karne se **pehle** Order persist kare.
- Paise integer mein store kar. Phir sochh ki `totalAmount * 100` ka kya hota hai
  jab course ka price 1999.99 ho.

**Done jab:** Har Razorpay order ka ek matching row ho, aur tu "is user ne kya kharida?"
ka jawab apne DB se de sake, bina Razorpay ka API call kiye.

---

## Day 11 — Payment bypass band karo

**Seekh:** Replay protection aur mutating endpoints pe idempotency.

Signature sirf ye proof karta hai ki **payment hua**. Ye nahi batata ki **kya khareeda gaya** —
jab tak tune khud record na kiya ho.

**Banao:**
- HMAC verify karo, `razorpay_order_id` se Order dhundo, phir
  **`req.body.courses` ko poori tarah ignore karo** aur Order ke courses enroll karo.
- Transition guard karo:
  ```javascript
  findOneAndUpdate({ _id, status: 'paid' }, { $set: { status: 'fulfilled' } })
  ```
  `null` return matlab kisi ne pehle hi fulfill kar diya — toh replay no-op ban jaata hai.
- Ab apne hi endpoint pe attack kar. Valid signature replay kar. Alag course list ke
  saath signature bhej. **Dono fail hone chahiye.**

**Done jab:** Tune wo exact request likh rakhi ho jo pehle free courses deti thi,
aur dikha sako ki ab wo error deti hai.

> Ab live deployment wapas chalu kar sakta hai.

---

## Day 12 — Atomicity

**Seekh:** Sessions, `withTransaction`, retryable writes — aur limits: transactions ko
replica set chahiye, wo locks hold karte hain, aur achhe modelling ka replacement nahi hain.

Abhi poore codebase mein **ek bhi `startSession` nahi hai.**

**Banao:**
- Fulfillment ko ek transaction mein wrap kar: Order status flip + Enrollment inserts +
  counter increments + outbox row.
- Phir prove kar: doosre Enrollment insert ke baad jaan-boojh ke throw kar aur confirm
  kar ki kuch bhi persist nahi hua.

**Done jab:** Forced mid-transaction failure DB ko bilkul waise chhode jaisa tha —
aur tu bata sake ki **outbox row transaction ke andar kyun honi chahiye.**

> Wo ek sentence hi wajah hai ki pattern ka naam *transactional* outbox hai.

---

## Day 13 — N+1 maaro, controller patla karo

**Seekh:** `$in` se batching, aggregation vs populate, `asyncHandler` wrapper.

`controllers/UserC.js` ka `getEnrolledCourses` har enrolled course pe ek query karta hai —
kisi bhi LMS ki sabse zyada hit hone wali screen pe N+1.

**Banao:**
- Ek Enrollment query + ek `CourseProgress.find({ userId, courseId: { $in: ids } })`,
  memory mein assemble. Phir wahi ek `$lookup` aggregation se likh aur dono compare kar.
- `utils/asyncHandler.js` likh aur poore ek controller file se try/catch hata —
  codebase mein ~30 ek jaise block hain.
- Day 1 wali dashboard timing dobara chala.

**Ratchet se do line hatao:** `controllers/CourseC.js` (`no-prototype-builtins` →
`Object.prototype.hasOwnProperty.call`) aur `models/Review.js` (`no-useless-catch` —
bekaar `try { } catch (e) { throw e }` hata).

> Ab legacy block khali ho jayega. **Block delete kar de.** Ratchet ka kaam khatam.

**Done jab:** Dashboard constant number of queries kare, chahe student ke kitne bhi
course hon — `mongoose.set('debug')` output se prove kar, maan ke mat chal.

---

## Day 14 — Consolidate karo aur likho

**Seekh:** Jo kaam documented nahi, wo invisible hai. Aaj ye pura fortnight ek aisi
cheez ban jaata hai jo dikhayi aur samjhaayi ja sake.

**Banao:**
- Test gaps bharo: payment path (replay aur course-swap dono fail hone chahiye) aur
  enrollment migration ki idempotency.
- `BASELINE.md` mein after-numbers ko before-numbers ke saath rakho.
- **Root `README.md` dobara likho** — abhi wo sirf links ki list hai. Chahiye:
  system kya karta hai, do service kyun hain, outbox ke through request path ka diagram,
  ek command mein local kaise chalao, aur ek imaandaar section ki kya nahi hua.

  > Wo aakhri section confidence dikhata hai, kamzori nahi.
  >
  > **Root README English mein likhna** — public-facing wahi hai. Ye docs Hinglish
  > mein theek hain, wo tere seekhne ke liye hain.

- **Chaar chhote ADR** `docs/adr/` mein, ek-ek page: Redis queue ke bajaye outbox kyun;
  embedded arrays ke bajaye join collection kyun; Order server-authoritative kyun;
  OTP hash kyun. Context, decision, consequences.

  > Tere level pe koi ye nahi likhta. Aur interview jo reasoning dhundta hai, ye
  > exactly wahi dikhate hain.

- CI badge README mein:
  ```markdown
  ![CI](https://github.com/parth5404/TEST-GS-Backend/actions/workflows/ci.yml/badge.svg)
  ```

**Done jab:** Tu kisi aur engineer ko 5 minute mein, bina notes ke samjha sake ki
outbox pattern kyun exist karta hai aur uski keemat kya hai — aur koi banda jisne repo
kabhi nahi dekha, sirf README se local chala le.

---

## ESLint ratchet — kaunsi file kis din

`backend_node/eslint.config.js` ke `LEGACY RATCHET` block se ye lines hatani hain.
**Ye khud ek checklist hai.**

| File | Rule | Din | Kya karna |
|---|---|---|---|
| `controllers/CourseProgressC.js` | `no-undef` | Day 2 | `ErrorResponse` ka `require` add karo — **asli bug** |
| `models/User.js` | `no-useless-escape` | Day 3 | Email regex ka `\.` character class ke andar |
| `controllers/CourseC.js` | `no-prototype-builtins` | Day 13 | `Object.prototype.hasOwnProperty.call(updates, key)` |
| `models/Review.js` | `no-useless-catch` | Day 13 | `try { } catch (e) { throw e }` hatao |

Chaaron hat gaye = poora block delete kar do.

Abhi ka status: **0 errors, 159 warnings.** Warnings CI fail nahi karti — wo tera backlog hai.

---

## Day 15+ — jo abhi bacha hai

28 ghante mein ya toh kuch cheezein gehrai se hoti hain, ya sab kuch upar-upar se.
Is plan ne gehrai chuni. Ye khule hain, aur ye jaanna bhi exercise ka hissa hai:

- **CDC / change streams** — polling ko replace nahi karta, uske upar lagta hai.
  Fail hui mail ka koi naya change event nahi aata, isliye retry ke liye poller
  phir bhi chahiye. Resume tokens persist karna sabse zaruri detail hai.
- **Service layer adhoora** — Day 4 aur 13 do slice karte hain, pura codebase nahi.
  Jab jis area ko chhuo, tab karo — big-bang rewrite kabhi nahi.
- **Uploads API process se guzarte hain** — `express-fileupload` pura video `/tmp` pe
  buffer karta hai, phir Cloudinary pe re-upload. Ek worker dono transfer ke liye
  atka rehta hai. Signed direct-to-Cloudinary upload sahi fix hai.
- **JWT mein role baked hai, revocation nahi** — kisi ko demote ya ban karo, kuch
  nahi hota jab tak token expire na ho.
- **Go abhi bhi `users` collection seedha padhta hai** ek hand-copied struct se.
  Wo shared-schema dependency todna natural Day 15 hai.
- **Newsletter cron replicate nahi ho sakta**, aur container 09:00 IST ke aar-paar
  restart ho jaye toh pura din skip ho jaata hai. Dated unique index se run claim karo.
- **helmet, rate limiting, CORS allowlist**, aur courses/users pe cascade deletes.

---

## Progress kaise track karo

- **`NOTES.md`** — roz ek paragraph: aaj kya surprise kiya. Yahi Day 14 ke README
  ka raw material hai.
- **`BASELINE.md`** — Day 1 ke numbers, Day 13/14 pe compare.
- **Is doc mein** din tick karte jao.

Browser mein checkbox wala version bhi hai (Claude artifact), par **primary ye file
hai** — kyunki ye code ke saath rehti hai aur git history mein aati hai.
