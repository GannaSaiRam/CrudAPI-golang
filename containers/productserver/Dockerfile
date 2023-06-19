FROM ubuntu

ARG port=$port
ARG image=$image

COPY $image /
COPY start_pserver.sh /
RUN chmod +x /start_pserver.sh

CMD ["/start_pserver.sh"]

EXPOSE $port/tcp