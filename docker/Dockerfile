FROM nginx:alpine

# Copy custom Nginx configuration to the container
COPY ./docker/nginx.conf /etc/nginx/conf.d/default.conf

VOLUME /usr/share/nginx/dist

EXPOSE 80

