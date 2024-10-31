ARG harbor_base_image_version
ARG harbor_base_namespace
FROM ${harbor_base_namespace}/harbor-log-base:${harbor_base_image_version}

COPY ./make/photon/log/rsyslog.conf /etc/rsyslog.conf
COPY ./make/photon/log/rsyslog_docker.conf /etc/rsyslog.d/
COPY ./make/photon/log/logrotate /etc/cron.hourly/
COPY ./make/photon/log/start.sh /usr/local/bin/start.sh

# Remove the original logrotate
RUN rm /etc/cron.daily/logrotate && \
    chmod +x /usr/local/bin/start.sh /etc/rsyslog.d/* && \
    chown -R 10000:10000 /etc/rsyslog.conf /etc/rsyslog.d/ /run /var/lib/logrotate/

HEALTHCHECK CMD netstat -ltun | grep 10514

VOLUME ["/var/log/docker/", "/run/", "/etc/logrotate.d/"]

CMD ["/usr/local/bin/start.sh"]
