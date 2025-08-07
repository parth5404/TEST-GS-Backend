FROM node:20

WORKDIR /app

# Copy only package.json first
COPY package*.json ./

# Install dependencies inside the container
RUN npm install

# Now copy the rest of the code
COPY . .

CMD ["npm", "run", "dev"]
