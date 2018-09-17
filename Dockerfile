FROM nginx

MAINTAINER maintainers@vikunja.io

ADD nginx.conf /etc/nginx/nginx.conf
COPY dist /usr/share/nginx/html
RUN rm /usr/share/nginx/html/js/*.map