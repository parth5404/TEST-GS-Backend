const accountCreationTemplate = (name) => {
	return `
	<!DOCTYPE html>
	<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Account Registration Confirmation</title>

    <style>
      body {
        background-color: #ffffff;
        font-family: Arial, sans-serif;
        font-size: 16px;
        line-height: 1.4;
        color: #333333;
        margin: 0;
        padding: 0;
      }

      .container {
        max-width: 600px;
        margin: 0 auto;
        padding: 20px;
        text-align: center;
      }

      .logo {
        max-width: 200px;
        margin-bottom: 20px;
      }

      .message {
        font-size: 18px;
        font-weight: bold;
        margin-bottom: 20px;
      }

      .body {
        font-size: 16px;
        margin-bottom: 20px;
      }

      .cta {
        display: inline-block;
        padding: 10px 20px;
        background-color: #ffd60a;
        color: #000000;
        text-decoration: none;
        border-radius: 5px;
        font-size: 16px;
        font-weight: bold;
        margin-top: 20px;
      }

      .support {
        font-size: 14px;
        color: #999999;
        margin-top: 20px;
      }

      .highlight {
        font-weight: bold;
      }
    </style>
  </head>

  <body>
    <div class="container">
      <!-- Header -->
      <div>
       
          <img class="logo" src="https://i.postimg.cc/Gt8KWn0f/GSLOGOEMAIL.gif" alt="GS-Academia" />
  
      </div>

      <!-- Body -->
      <div class="body">
        <p class="message">Account Registration Confirmation</p>

        <p>Dear ${name},</p>
        <p>Your account is successfully registered on GS Academia. We are happy to have you onboard !</p>

        <p>Please log in to your dashboard to start your journey.</p>
        

        <p>Happy journey !</p>
      </div>

      <!-- Footer -->
      <div class="support">
        <p>If you have any questions or need assistance, please feel free to reach out to us at <a href="mailto:gsacademia5404@gmail.com">Support</a>. We are here to help!</p>
      </div>
    </div>
  </body>
</html>
  
  `;
};

module.exports = accountCreationTemplate;