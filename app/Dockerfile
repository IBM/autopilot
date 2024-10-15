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

# Set permissions for non-root usage
RUN chmod -R g+rwx /var/cache/nginx /var/run /var/log/nginx

COPY --from=build /app/dist /usr/share/nginx/html

# Copy custom NGINX configuration
COPY nginx.conf /etc/nginx/nginx.conf
RUN rm /etc/nginx/conf.d/default.conf

EXPOSE 8080

# Start NGINX server
CMD ["nginx", "-g", "daemon off;"]