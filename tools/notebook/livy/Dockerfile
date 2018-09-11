# VERSION 1
# Author: Tim.Xu
# Command format: Instruction [arguments / command] ..
# Base image to use, this must be set as the first line
FROM jdk:8u162

# Maintainer: docker_user <docker_user at email.com> (@docker_user)
LABEL maintainer="xiaoxubeii@gmail.com"

# Add Hadoop
ADD ./dist/hadoop /usr/local/hadoop
# Add Spark
ADD ./dist/spark /usr/local/spark
# Add Tini
ADD ./tini /sbin/tini
RUN chmod +x /sbin/tini

ADD ./dist/livy /livy

ENV JAVA_HOME /usr/java/default
ENV SPARK_HOME /usr/local/spark
ENV HADOOP_HOME /usr/local/hadoop
ENV LIVY_HOME /livy

ENTRYPOINT [ "/sbin/tini", "--" ]
WORKDIR ${LIVY_HOME}
CMD ["bin/livy-server", "start"]