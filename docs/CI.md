# GitHub Actions — is project ke through samajhna

---

## 1. CI hai kya, aur kyun

**Continuous Integration** = har push pe GitHub ek **saaf, khaali machine** leta hai,
tera code clone karta hai, aur tere checks chalata hai.

Sabse badi wajah: **"mere laptop pe toh chal raha tha" ka jawab.**

Tera laptop gandaa hai — usme purani `node_modules` padi hain, globally install kiye
packages hain, `.env` file hai. CI ke paas kuch nahi hota. Isliye wo wo bugs pakadta
hai jo tu kabhi na pakadta.

**Is repo ka ek zinda example:** `utils/emailSender.js` mein `require('axios')` hai, par
`axios` `package.json` mein hai hi nahi. Abhi ye sirf isliye chal raha hai kyunki
`razorpay` ki dependency ke roop mein npm ne usse hoist kar diya hai. Kabhi razorpay
update hua ya npm ka resolution badla — email ka har path toot jayega.

Ye Day 2 ka kaam hai. Lekin ye samajh le ki **CI aisi cheezein pakadne ke liye hi hota hai.**

---

## 2. File kahan rakhni hai

```
.github/workflows/ci.yml
```

GitHub **sirf** yahin dhundta hai. Naam kuch bhi ho (`ci.yml`, `test.yml`), par folder
exactly yahi hona chahiye. Folder galat = kuch nahi chalega, aur koi error bhi nahi aayega.

---

## 3. Structure — teen level

```
Workflow  (poori ci.yml file)
   └── Job        (alag machine, dusre jobs ke saath PARALLEL)
         └── Step (ek command, upar se neeche SEQUENTIAL)
```

Hamari file mein do jobs hain — `node` aur `go`. Wo **ek saath** chalte hain, alag-alag
machines pe. Isliye total time slow wale job jitna lagta hai, dono ka jod nahi.

Ek job ko dusre ka intezaar karwana ho:

```yaml
jobs:
  test:
    ...
  deploy:
    needs: test        # ← test pass hone ke baad hi chalega
```

---

## 4. `uses:` vs `run:`

Ye do hi tarah ke steps hote hain.

```yaml
- uses: actions/checkout@v4      # kisi aur ka banaya ready-made action
- run: npm ci                    # shell command
```

**`uses:`** — GitHub Marketplace se koi package. `actions/checkout` tera repo clone karta
hai (**hamesha pehla step**, iske bina runner khaali hota hai). `actions/setup-node` Node
install karta hai.

`@v4` version pin hai. Zaruri hai — warna kal ko wo action breaking change karega aur
tera CI bina kuch kiye red ho jayega.

**`run:`** — bash command, wahi jo tu terminal mein likhta hai.

---

## 5. Runner kya hai

```yaml
runs-on: ubuntu-latest
```

Ek fresh VM jo GitHub har job ke liye deta hai, aur job khatam hote hi phenk deta hai.
Kuch persist nahi hota. `windows-latest` aur `macos-latest` bhi hain, par mehnge aur dheeme
hain — Node/Go ke liye ubuntu hi sahi hai.

Public repo pe minutes free hain.

---

## 6. Caching

```yaml
- uses: actions/setup-node@v4
  with:
    cache: npm
    cache-dependency-path: backend_node/package-lock.json
```

Har run fresh machine pe hota hai, matlab har baar saari dependencies dobara download.
Cache isse bachata hai — npm ka download folder save karke agli baar restore kar deta hai.

`cache-dependency-path` isliye zaruri hai kyunki hamari lockfile root mein nahi hai.
Ye bataye bina action usse dhundh nahi paayega.

**Cache key lockfile ka hash hota hai.** Lockfile badla = naya cache. Isliye stale cache
ka dar nahi.

---

## 7. Local aur CI ek jaisa hona chahiye

Dhyan de ki CI wahi commands chala raha hai jo tu local pe chalata hai:

```yaml
- run: npm run lint      # tu bhi yahi chalata hai
- run: npm test          # aur yahi
```

Ye **parity** jaan-boojh ke hai. Agar CI kuch alag chalata, toh "local green, CI red"
wali paheli banti — aur wo debug karna sabse zyada frustrating cheez hai.

**Rule:** CI mein kabhi aisi command mat likhna jo tu local pe na chala sakta ho.

---

## 8. Is repo ka setup

**Job `node`:** checkout → Node 20 (cache ke saath) → `npm ci` → `npm run lint` → `npm test`

`npm ci` use ho raha hai, `npm install` nahi:
- `ci` lockfile ko **exactly** follow karta hai, `node_modules` pehle delete karta hai,
  lockfile kabhi modify nahi karta
- `install` lockfile update kar sakta hai

CI mein hamesha `ci` — tabhi build reproducible rehta hai.

**Job `go`:** checkout → Go (version `go.mod` se) → `go vet` → `go build` → `gofmt` check

`go-version-file: go_email/go.mod` — version do jagah maintain karne se bachne ke liye.
`go.mod` hi single source of truth hai.

---

## 9. Lint abhi green kaise hai — "ratchet"

`backend_node/eslint.config.js` khol aur neeche wala `LEGACY RATCHET` block dekh.

Problem ye thi: rules ko `error` karte hi 9 purani galtiyan CI ko red kar deti thin.
**Din 1 se red CI bekaar hai** — banda usse ignore karna seekh jaata hai, aur phir asli
failure bhi miss ho jaata hai.

Do galat solutions:
- ❌ rule ko poore repo mein `warn` kar do → naya code bhi bach jayega
- ❌ CI red rehne do → signal marr gaya

Sahi solution — **ratchet**: rule sab jagah `error` hi hai, par sirf un 4 purani files
ko chhoot hai jinme pehle se problem thi. Naya code turant block hoga.

Har entry pe `TODO(Day N)` comment hai. Jaise-jaise tu wo file theek karega, us line ko
block se hata dena. Block khali = block delete.

**Wo block khud ek to-do list hai.** Aur asli teams legacy codebase pe linting isi
tarike se adopt karti hain.

Abhi status: **0 errors, 159 warnings.** Warnings CI ko fail nahi karti — wo tera backlog hai.

---

## 10. Jab CI red ho

1. GitHub pe repo → **Actions** tab
2. Red wala run kholo
3. Fail hua job kholo → fail hua step expand karo
4. Log padho — **poora padho, sirf aakhri line nahi.** Asli error aksar upar hota hai.

Phir wahi command local pe chala ke reproduce karo:

```bash
cd backend_node && npm run lint
cd go_email && go vet ./...
```

Local pe reproduce **nahi** ho raha? Toh wo environment ka farq hai — 90% cases mein
missing dependency (tere laptop pe hai, `package.json` mein nahi). Bilkul `axios` wala case.

---

## 11. Secrets

API keys workflow file mein **kabhi** mat likhna. Wo file public hai (public repo mein toh
sabko dikhti hai, aur private mein bhi har collaborator ko).

GitHub pe: **Settings → Secrets and variables → Actions**, phir:

```yaml
env:
  RAZORPAY_KEY: ${{ secrets.RAZORPAY_KEY }}
```

GitHub logs mein secrets ko apne aap mask karta hai. Par `echo $RAZORPAY_KEY` mat karna —
kabhi-kabhi masking bypass ho jaati hai (jaise base64 encode karke print karna).

**Aur ek zaruri baat:** `pull_request` events fork se aane pe secrets **nahi** milte —
ye jaan-boojh ke hai, warna koi bhi fork karke PR bhej ke tere secrets churaa leta.

---

## 12. Exercises

**1. Pehla run dekh**
Branch push kar aur GitHub pe **Actions** tab khol. Dono jobs ko chalte dekh. Kaunsa
pehle khatam hua? Har step ka time dekh — sabse zyada time kisme laga?

**2. Cache ka asar naapo**
Bina kuch badle dobara push kar (`git commit --allow-empty -m "ci: cache test"`).
"Install dependencies" step ki timing pehle wale run se compare kar.

**3. Jaan-boojh ke CI todo**
Kisi controller mein ek line add kar:
```js
const bilkulNayaVariable = undefinedWalaCheez;
```
Commit karne ki koshish kar — **husky pehle hi rok dega** (`no-undef` error hai).
Ab `git commit --no-verify` se hook bypass kar aur push kar de. Ab CI red hoga.
GitHub pe failure padh, phir revert kar de.

*(Seekh: hook aur CI do alag safety net hain. Hook tez hai par bypass ho sakta hai.
CI dheema hai par bypass nahi ho sakta. Isliye dono chahiye.)*

**4. Ratchet ko kaam karte dekho**
`eslint.config.js` ke legacy block se `'controllers/CourseProgressC.js'` line hata de,
phir `npm run lint` chala. Ab wo asli bug error banke aayega. Wapas add kar de —
ya behtar, us file mein `ErrorResponse` ka `require` add karke **abhi hi theek kar de**,
aur line permanently hata de. Ye Day 2 ka kaam pehle ho jayega.

**5. Ek naya step add kar**
`node` job mein Prettier check add kar:
```yaml
- name: Format check
  run: npm run format:check
```
Push kar. Ye **fail hoga** — poora codebase Prettier style mein nahi hai. Ab socho:
`npm run format` chala ke sab format kar dena chahiye?

*(Jawab: **nahi**, abhi nahi. Wo 40+ files ka diff banayega aur agle 14 din tere asli
changes usme dab jayenge. Behtar tarika: pre-commit hook ko sirf un files ko format
karne do jinhe tu waise bhi chhoo raha hai — jo abhi already set hai. Codebase
dheere-dheere apne aap saaf hoga. Ye step wapas hata de.)*

---

## Aage

- **Day 7** — jab asli tests aayenge, `npm test` step apne aap unhe chalane lagega.
  CI config badalne ki zarurat nahi. Achhe setup ki yahi pehchaan hai.
- **Day 14** — README mein badge:
  ```markdown
  ![CI](https://github.com/parth5404/TEST-GS-Backend/actions/workflows/ci.yml/badge.svg)
  ```
