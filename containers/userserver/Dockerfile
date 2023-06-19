FROM ubuntu

ARG port=$port
ARG image=$image

COPY $image /
COPY start_userver.sh /
RUN chmod +x /start_userver.sh

CMD ["/start_userver.sh"]

EXPOSE $port/tcp