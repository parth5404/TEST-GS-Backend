const nodemailer = require('nodemailer');
const clgDev = require('./clgDev');
const axios=require('axios');

const emailSender = async (toEmail, subject, template, userFirstName, userLastName, extraData) => {
  try {
    const template = JSON.parse(template);
    const body = template.body;
    const extraData = JSON.parse(extraData);
    
    // Pass data in request body instead of headers since:
    // 1. Headers are meant for metadata, not large data payloads
    // 2. Headers have size limitations in some servers
    // 3. Complex objects in headers can cause encoding issues
    const requestData = {
      firstName: userFirstName,
      lastName: userLastName, 
      email: toEmail,
      subject: subject,
      body: body,
      template: template,
      extraData: extraData
    };
    
    const response = await axios.post('http://localhost:8080/send-email', requestData);

    // For real purpose 
    // const transporter = nodemailer.createTransport({
    //   host: process.env.MAIL_HOST,
    //   auth: {
    //     user: process.env.MAIL_USER,
    //     pass: process.env.MAIL_PASS,
    //   },
    // });



    // // For testing / development purpose
    // const transporter = nodemailer.createTransport({
    //   host: process.env.SMTP_HOST,
    //   port: process.env.SMTP_PORT,
    //   auth: {
    //     user: process.env.SMTP_EMAIL,
    //     pass: process.env.SMTP_PASSWORD,
    //   },
    // });

    // send mail
    // const info = await transporter.sendMail({
    //   from: "parth",
    //   to: toEmail,
    //   subject: subject,
    //   html: body,
    // });

    // return info;
  } catch (err) {
    clgDev(err.message);
    throw err;
  }
};

module.exports = emailSender;
