// ESLint 9 "flat config".
//
// Purana ESLint `.eslintrc.json` use karta tha. Naya version isi file ko dhundta
// hai — ek plain JS file jo array export karti hai. Array ke baad wale objects
// pehle walon ko override karte hain (CSS cascade jaisa).
//
// Ye config **is repo ke asli bugs** pakadne ke liye tuned hai, generic nahi.

const js = require('@eslint/js');
const globals = require('globals');

module.exports = [
  // Step 1: ESLint ka apna recommended set — safe, well-tested rules.
  js.configs.recommended,

  {
    files: ['**/*.js'],

    languageOptions: {
      ecmaVersion: 2022,
      // Ye codebase `require()` use karta hai, `import` nahi — isliye commonjs.
      sourceType: 'commonjs',
      globals: {
        // Bina iske ESLint ko `process`, `__dirname`, `console` ka pata nahi
        // hota aur wo har jagah no-undef error de deta.
        ...globals.node,
      },
    },

    rules: {
      // ── Ye rule tere repo mein ek ASLI BUG pakadta hai ────────────────────
      //
      // controllers/CourseProgressC.js `ErrorResponse` use karta hai par usse
      // require kabhi nahi karta. Har error path pe wahan ReferenceError aata
      // hai — catch block phir se wahi galti karta hai, toh unhandled rejection
      // ban jaata hai, clean 4xx nahi.
      //
      // Isliye ye 'error' hai, warn nahi. Aisi galti kabhi commit nahi honi chahiye.
      'no-undef': 'error',

      // Dead imports pakadta hai — jaise wo saat mail/templates/*.js jo import
      // toh hote hain par kabhi use nahi (Go ab templates render karta hai).
      //
      // 'warn' rakha hai kyunki abhi bahut saare hain. Ye teri to-do list hai,
      // commit blocker nahi. Day 3 mein saaf hoga.
      'no-unused-vars': [
        'warn',
        {
          argsIgnorePattern: '^_',
          // Express error middleware ka signature (err, req, res, next) hota hai
          // aur `next` use na hone pe bhi zaruri hai — warna Express usse error
          // handler manta hi nahi.
          caughtErrorsIgnorePattern: '^_',
        },
      ],

      // Repo mein abhi 18 console.log hain, ek `clgDev` helper ke hote hue bhi.
      // Day 7 pe ye sab pino se replace honge. Tab tak warn.
      'no-console': 'warn',

      // Chhupe hue bugs pakadne wale rules
      'no-unreachable': 'error',
      'no-dupe-keys': 'error',
      'no-duplicate-case': 'error',
      'no-fallthrough': 'error',
      'no-self-compare': 'error',
      'no-unmodified-loop-condition': 'error',

      // `await` bhoolna is codebase ki common galti hai —
      // jaise AuthC.js ka guestLogin, aur SubSectionC ka subSection.deleteOne().
      // 'require-await' pura fix nahi karta par kuch pakad leta hai.
      'require-await': 'warn',

      // Empty catch block matlab error chupa diya
      'no-empty': ['error', { allowEmptyCatch: false }],

      'eqeqeq': ['warn', 'smart'],
      'prefer-const': 'warn',
    },
  },

  // ───────────────────────────────────────────────────────────────────────────
  // LEGACY RATCHET
  //
  // Problem: rules ko 'error' karte hi 9 purani galtiyan CI ko red kar deti hain.
  // Din 1 se red CI bekaar hai — banda usse ignore karna seekh jaata hai.
  //
  // Galat solutions:
  //   ❌ rule ko poore repo mein 'warn' kar do → naya code bhi bach jayega
  //   ❌ CI red rehne do → signal marr gaya
  //
  // Sahi solution — "ratchet": rule sab jagah 'error' hi rahega, par SIRF in
  // purani files ko temporarily chhoot mili hai. Naya code turant block hoga.
  //
  // Jaise-jaise tu 14 din mein ye files theek karega, is block se line hata dena.
  // Block khali ho gaya = block delete. **Ye list khud ek to-do list hai.**
  //
  // Ye technique real hai — badi legacy codebases pe linting isi tarah adopt hoti hai.
  // ───────────────────────────────────────────────────────────────────────────
  {
    files: [
      'controllers/CourseProgressC.js',
      'controllers/CourseC.js',
      'models/Review.js',
      'models/User.js',
    ],
    rules: {
      // TODO(Day 2): controllers/CourseProgressC.js — ASLI BUG.
      //   `ErrorResponse` use hota hai par require nahi kiya gaya. Har error
      //   path pe ReferenceError, aur catch block wahi galti dohrata hai →
      //   unhandled rejection. Sirf ek `require` line add karni hai.
      //   Fix karte hi 'controllers/CourseProgressC.js' yahan se hata dena.
      'no-undef': 'warn',

      // TODO(Day 13): controllers/CourseC.js — `updates.hasOwnProperty(key)`.
      //   Object.create(null) ya jisme hasOwnProperty override ho, uspe crash
      //   karta hai. `Object.prototype.hasOwnProperty.call(updates, key)` sahi hai.
      'no-prototype-builtins': 'warn',

      // TODO(Day 13): models/Review.js — try { ... } catch (err) { throw err }
      //   Kuch karta hi nahi, sirf stack trace gandi karta hai. Hata do.
      'no-useless-catch': 'warn',

      // TODO(Day 3): models/User.js — email regex mein `\.` character class ke
      //   andar hai, jahan dot waise bhi literal hota hai. Harmless, par batata
      //   hai ki regex samajh ke nahi likha gaya. Waise bhi ye regex kaafi
      //   dhilaa hai — validation library behtar rahegi.
      'no-useless-escape': 'warn',
    },
  },

  {
    // Ye files lint nahi hongi
    ignores: ['node_modules/**', 'public/**', '_data/**', 'coverage/**'],
  },
];
