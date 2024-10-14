FROM node:18 as build

WORKDIR /app

# Install dependencies
COPY package*.json ./
RUN npm install

# Copy application code
COPY . .

# Build React app
RUN chmod -R 755 /app/node_modules /app
RUN npm run build

# NGINX image to serve React app
FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80

# Start NGINX server
CMD ["nginx", "-g", "daemon off;"]
