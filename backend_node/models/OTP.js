const mongoose = require('mongoose');
const clgDev = require('../utils/clgDev');
const emailSender = require('../utils/emailSender');
const emailOtpTemplate = require('../mail/templates/emailOtpTemplate');

const OTPSchema = new mongoose.Schema({
  otp: {
    type: String,
    required: true,
  },
  email: {
    type: String,
    required: true,
  },
  createdAt: {
    type: Date,
    default: Date.now,
    expires: '10m', // The document will automatically deleted after 10 minutes of its creation time
  },
});

const sendOtpEmail = async (toEmail, otp) => {
  try {
    const mailResponse = await emailSender(toEmail, 'Verification Email from GS-Academia', "emailOtpTemplate", "firstName", "lastName",JSON.stringify({
      "otp":otp
    }));
  } catch (err) {
    clgDev(`Error occurred while sending otp : ${err.message}`);
    throw err;
  }
};

// Send otp after OTP is created
OTPSchema.post('save', async function (doc) {
  try {
    await sendOtpEmail(this.email, this.otp);
  } catch (err) {
    clgDev(`Failed to send OTP email: ${err.message}`);
  }
});

module.exports = mongoose.model('OTP', OTPSchema);
