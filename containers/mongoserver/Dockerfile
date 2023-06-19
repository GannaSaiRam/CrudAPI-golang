FROM ubuntu

ARG port=$port 
ARG image=$image

COPY $image /
COPY start_mserver.sh /
RUN chmod +x /start_mserver.sh

CMD ["/start_mserver.sh"]

EXPOSE $port/tcp