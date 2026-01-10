FROM clojure:temurin-25-lein-trixie-slim

WORKDIR /usr/src/app

CMD ["--help"]

COPY conf/main/logback.xml /usr/src/app/

COPY project.clj /usr/src/app/
RUN lein deps

RUN ln -s "/opt/java/openjdk/bin/java" "/bin/analyses"

ENV OTEL_TRACES_EXPORTER none

COPY . /usr/src/app

RUN lein uberjar && \
    cp target/analyses-standalone.jar .

ENTRYPOINT ["analyses", "-Dlogback.configurationFile=/usr/src/app/logback.xml", "-cp", ".:analyses-standalone.jar:/", "analyses.core"]
