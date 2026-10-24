# syntax=docker/dockerfile:1.7
#
# Reference image for conformance testing: Java PlantUML v1.2026.9 (MIT
# variant) with a pinned Java runtime, font set, Graphviz version and SVG
# rasterizer. Every download and base image is pinned by checksum or digest.
#
# Build for a single platform so references do not depend on the host:
#
#   docker build --platform linux/amd64 -t plantuml-go-reference tools/reference

ARG BASE_IMAGE=eclipse-temurin:21-jre-jammy@sha256:f04fb34e053148344e83317976114ec3f37e4b830ec8bdab5a2fe3cecd7d010b

# Graphviz 14.0.1 as built by upstream's own image (plantuml/plantuml:1.2026.8),
# which compiles it from source on the same base image. Compiling it here would
# need amd64 emulation on arm64 hosts, where the compiler crashes at random.
# Only the Graphviz files are taken from that image. The image of 1.2026.9
# builds the same Graphviz 14.0.1; keeping this one keeps the references free
# of changes from a rebuilt Graphviz.
ARG GRAPHVIZ_IMAGE=plantuml/plantuml:1.2026.8@sha256:d08610df482510844382caa4e016ba2bf7e3231f630f02ee12f250f3416c62b1

FROM ${GRAPHVIZ_IMAGE} AS graphviz

# Tools compiled once on the build platform: class files are portable.
FROM --platform=$BUILDPLATFORM eclipse-temurin:21-jdk-jammy@sha256:e0c60c487345d1dc9d0fc7b6f0496f3cc941e5132e09296cc17a6decc71b902b AS tools
ADD --chmod=644 --checksum=sha256:e84e45b37ad3030cf802e7be18a687d209eca55733a5679b1f1f1a515f7f70dd \
    https://repo1.maven.org/maven2/net/sourceforge/plantuml/plantuml-mit/1.2026.9/plantuml-mit-1.2026.9.jar \
    /opt/plantuml.jar
COPY FontMetricsDump.java ReferenceRender.java RunCli.java /src/
RUN javac --release 21 -cp /opt/plantuml.jar -d /opt/tools /src/FontMetricsDump.java /src/ReferenceRender.java /src/RunCli.java

FROM ${BASE_IMAGE}
ENV LANG=C.UTF-8
RUN apt-get update \
 && apt-get install -y --no-install-recommends fontconfig fonts-dejavu \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

# Keep upstream's install paths: the plugin directory is compiled into dot.
# As in upstream's image, the gd plugin is not usable (libgd is not installed);
# dot -c rebuilds the plugin configuration without it.
COPY --from=graphviz /usr/local/bin/ /usr/local/bin/
COPY --from=graphviz /usr/local/lib/ /usr/local/lib/
RUN echo /usr/local/lib > /etc/ld.so.conf.d/graphviz.conf \
 && echo /usr/local/lib/graphviz >> /etc/ld.so.conf.d/graphviz.conf \
 && ldconfig \
 && dot -c
ENV GRAPHVIZ_DOT=/usr/local/bin/dot

ADD --checksum=sha256:fa8c26495a187e592c501db15bf9e8a9fdc051d4b2b336b39703d5b59f912b9d \
    https://github.com/linebender/resvg/releases/download/v0.48.1/resvg-linux-x86_64.tar.gz \
    /tmp/resvg.tar.gz
RUN tar -xzf /tmp/resvg.tar.gz -C /usr/local/bin resvg && rm /tmp/resvg.tar.gz

ADD --chmod=644 --checksum=sha256:e84e45b37ad3030cf802e7be18a687d209eca55733a5679b1f1f1a515f7f70dd \
    https://repo1.maven.org/maven2/net/sourceforge/plantuml/plantuml-mit/1.2026.9/plantuml-mit-1.2026.9.jar \
    /opt/plantuml.jar
ENV PLANTUML_JAR=/opt/plantuml.jar

COPY --from=tools /opt/tools /opt/tools

WORKDIR /work
