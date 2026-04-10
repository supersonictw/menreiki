FROM keymetrics/pm2:24-alpine

RUN addgroup \
    -g 8383 \
    itsuka
RUN adduser -HD \
    -u 8383 \
    -G itsuka \
    -h /workplace \
    kotori

RUN mkdir -p /.npm /workplace
WORKDIR /workplace
ADD . /workplace

RUN chmod +x /workplace/docker-entrypoint.sh
RUN chown -R \
    8383:8383 \
    /.npm /workplace

USER 8383
RUN npm install

CMD ["/workplace/docker-entrypoint.sh"]
