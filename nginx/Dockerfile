FROM nginx:1.29-alpine

RUN rm -f /var/log/nginx/access.log /var/log/nginx/error.log && \
    mkdir -p /var/log/nginx /var/cache/nginx && \
    chown -R nginx:nginx \
        /var/log/nginx \
        /var/cache/nginx \
        /etc/nginx/conf.d

COPY nginx.conf /etc/nginx/nginx.conf
COPY templates/nginx.conf.template /etc/nginx/templates/default.conf.template

USER nginx
