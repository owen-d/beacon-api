FROM nginx:1.13

RUN rm /etc/nginx/conf.d/default.conf

COPY proxy.conf /etc/nginx/proxy.template.conf
COPY run.sh /
RUN chmod +x /run.sh

ENTRYPOINT ["/run.sh"] 