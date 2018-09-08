FROM nginx

MAINTAINER maintainers@vikunja.io

COPY dist /usr/share/nginx/html
RUN rm /usr/share/nginx/html/js/*.map