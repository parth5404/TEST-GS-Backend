const mongoose = require('mongoose');
const clgDev = require('../utils/clgDev');
const dotenv = require('dotenv').config();
const colors = require('colors');


const url=process.env.MONGO_URI;
const connectDB = async () => {
  try {
    const conn = await mongoose.connect(url);
    clgDev("MongoDB connected successfully".cyan.underline.bold);
  } catch (err) {
    clgDev(`${err.message}`.red.underline.bold);
    process.exit(1);
  }
}


// 2nd way to connect to mongo db
// const connectDB = () => {
//   mongoose.connect(process.env.MONGO_URI)
//     .then(() => clgDev("MongoDB connected successfully".cyan.underline.bold))
//     .catch((err) => {
//       clgDev(`${err.message}`.red.underline.bold);
//       process.exit(1);
//     });
// }


module.exports = connectDB;
