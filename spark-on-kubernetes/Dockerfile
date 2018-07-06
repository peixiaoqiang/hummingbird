# VERSION 1
# Author: tim.xu
# Command format: Instruction [arguments / command] ..
# Base image to use, this must be set as the first line
FROM jdk:8u162

# Maintainer: docker_user <docker_user at email.com> (@docker_user)
MAINTAINER tim.xu xiaoxubeii@gmail.com

# Add Hadoop
ADD ./tools/hadoop-3.0.1 /usr/local/hadoop-3.0.1
# Add Spark2.3
ADD ./tools/spark-2.3.0 /usr/local/spark-2.3.0
# Add Tini
ADD ./tools/tini /sbin/tini

RUN  set -ex && \
ln -s /usr/local/hadoop-3.0.1 /usr/local/hadoop  && \
chgrp root /etc/passwd && chmod ug+rw /etc/passwd  && \
yum clean all && \
yum -y install  flex byacc libpcap ncurses openssl-devel openssl ncurses-devel libpcap-devel iftop traceroute snappy snappy-devel && \
ln -s /usr/local/spark-2.3.0 /usr/local/spark && \
touch /usr/local/spark/RELEASE && \
chmod +x /sbin/tini  && \
chown -R hadoop:hadoop /usr/local/spark

ENV  JAVA_HOME /usr/java/default
ENV  SPARK_HOME /usr/local/spark
ENV  HADOOP_HOME /usr/local/hadoop

WORKDIR /usr/local/spark/work-dir

ADD ./entrypoint.sh  /opt/
RUN chmod +x /opt/entrypoint.sh
ENTRYPOINT ["/opt/entrypoint.sh"]
