# VERSION 1
# Author: Tim.Xu
# Command format: Instruction [arguments / command] ..
# Base image to use, this must be set as the first line
FROM jdk:8u162
ARG SPARK_VERSION
ARG HADOOP_VERSION

# Maintainer: docker_user <docker_user at email.com> (@docker_user)
LABEL Maintainer="xiaoxubeii@gmail.com"

# Add Hadoop
ADD ./dist/hadoop-${SPARK_VERSION} /usr/local/hadoop
# Add Spark
ADD ./dist/spark-$P{SPARK_VERSION} /usr/local/spark
# Add Tini
ADD ./dist/tini /sbin/tini

ENV  JAVA_HOME /usr/java/default
ENV  SPARK_HOME /usr/local/spark
ENV  HADOOP_HOME /usr/local/hadoop