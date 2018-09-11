# VERSION 1
# Author: Tim.Xu
# Command format: Instruction [arguments / command] ..
# Base image to use, this must be set as the first line
FROM zeppelin

# Maintainer: docker_user <docker_user at email.com> (@docker_user)
LABEL maintainer="xiaoxubeii@gmail.com"

# Add Hadoop
ADD ./dist/hadoop /usr/local/hadoop
# Add Spark
ADD ./dist/spark /usr/local/spark
ENV SPARK_HOME /usr/local/spark
ENV HADOOP_HOME /usr/local/hadoop