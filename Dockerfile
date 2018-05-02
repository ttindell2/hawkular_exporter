FROM scratch
LABEL MAINTAINER="Robert Tindell <tim.tindell@gmail.com>"

EXPOSE 9189

USER 1001
# Copy hawkular_exporter
ADD hawkular_exporter /

CMD ["/hawkular_exporter"]

