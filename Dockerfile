FROM scratch
LABEL maintainer="roger@steneteg.org"

COPY hyperlink /
COPY static/ /static/
COPY templates/ /templates/

EXPOSE 8080
ENTRYPOINT ["/hyperlink"]